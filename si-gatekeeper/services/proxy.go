package services

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/cyops-se/safe-import/si-gatekeeper/common"
	"github.com/cyops-se/safe-import/si-gatekeeper/types"
	otypes "github.com/cyops-se/safe-import/si-outer/types"
	"github.com/cyops-se/safe-import/usvc"
)

var jobsSvc *usvc.UsvcStub

type ProxyService struct {
	usvc.Usvc
}

func (svc *ProxyService) Initialize(broker *usvc.UsvcBroker) {
	svc.InitializeService(broker, 1, "si-gatekeeper", "proxy", "Proxies request from si-inner to si-outer")
	svc.RegisterMethod("allitems", svc.allItems)
	svc.RegisterMethod("byfieldname", svc.byFieldName)
	svc.RegisterMethod("update", svc.update)
	svc.RegisterMethod("prune", svc.prune)
	svc.RegisterMethod("httpget", svc.httpGet)

	jobsSvc = usvc.CreateStub(broker, "jobs", "si-outer", 1)

	// We don't use settings right now
	if err := svc.LoadSettings(); err != nil {
		svc.SaveSettings() // Save default settings. Though we don't actually use the settings right now...
	}

	svc.Executor = svc.execute
	svc.SetTaskIdleTime(60 * 1) // every minute
	svc.execute()
}

func (svc *ProxyService) execute() {
}

func (svc *ProxyService) allItems(payload string) (interface{}, error) {
	var items []types.CachedItem
	common.DB.Find(&items)
	return items, nil
}

func (svc *ProxyService) byFieldName(payload string) (interface{}, error) {
	var args types.ByNameRequest
	if err := json.Unmarshal([]byte(payload), &args); err != nil {
		svc.LogGeneric("error", "Marshalling request to JSON failed: %#v", err)
		return nil, err
	}

	var items []types.CachedItem
	if result := common.DB.Where(map[string]interface{}{args.Name: args.Value}).Find(&items); result.Error != nil {
		svc.LogGeneric("error", "Failed to query database, error: %#v", result.Error)
		return nil, result.Error
	}

	return items, nil
}

func (svc *ProxyService) update(payload string) (interface{}, error) {
	var item types.CachedItem
	if err := json.Unmarshal([]byte(payload), &item); err != nil {
		svc.LogGeneric("error", "Marshalling request to JSON failed: %#v", err)
		return nil, err
	}

	if result := common.DB.Save(&item); result.Error != nil {
		svc.LogGeneric("error", "Failed to query database, error: %#v", result.Error)
		return nil, result.Error
	}

	return item, nil
}

func (svc *ProxyService) prune(payload string) (interface{}, error) {
	return nil, fmt.Errorf("Method not yet implemented")
}

func (svc *ProxyService) httpGet(payload string) (interface{}, error) {
	var request types.HttpDownloadRequest
	if err := json.Unmarshal([]byte(payload), &request); err != nil {
		svc.LogGeneric("error", "Marshalling request payload JSON to request failed: %#v, %s", err, payload)
		return nil, err
	}

	// log.Println("Processing request:", request)

	jobrequest := otypes.WaitRequest{URL: request.URL, Method: request.Method, Body: request.Body}
	for _, v := range request.Headers {
		jobrequest.Headers = append(jobrequest.Headers, otypes.NameValue{Name: v.Name, Value: v.Value})
	}

	svc.PublishString("file.download.start", request.URL)
	if response, err := jobsSvc.RequestMessage("requesturlwait", jobrequest); err == nil {
		var r otypes.WaitResponse
		if err := json.Unmarshal([]byte(response.Payload), &r); err != nil {
			svc.LogGeneric("error", "Marshalling response JSON payload to response failed: %#v, %s", err, response.Payload)
			svc.PublishString("file.download.fail", request.URL)
			return nil, err
		}

		svc.PublishString("file.download.success", request.URL)
		if r.Success {
			if request.NoScan == false {
				svc.PublishString("file.download.scan.start", request.URL)
				if err, exitcode, infos := common.Scan(r.Filename); err != nil {
					if exitcode == 1 {
						for _, info := range infos {
							err = fmt.Errorf("VIRUS: %s", info.VirusName)
							svc.LogInfection(info.Filename, "VIRUS: %s", info.VirusName)
							svc.PublishEventMessage("proxy.infection", info)
							svc.PublishString("file.download.scan.fail", request.URL)
						}
					} else {
						svc.LogError(fmt.Sprintf("SCAN failed (exitcode: %d)", exitcode), err)
						svc.PublishString("file.download.scan.fail", request.URL)
					}

					return nil, err
				}

				svc.PublishString("file.download.scan.success", request.URL)
			}

			wd, _ := os.Getwd()
			oldname := filepath.VolumeName(wd) + r.Filename
			innerpath := strings.Replace(r.Filename, "outer", "inner", 1)

			folder := path.Dir(innerpath)
			os.MkdirAll(folder, os.ModeDir)
			os.Remove(innerpath)

			svc.PublishString("file.download.link.start", request.URL)
			if err := os.Symlink(oldname, innerpath); err != nil {
				svc.PublishString("file.download.link.fail", request.URL)
				return nil, err
			}

			svc.PublishString("file.download.link.success", request.URL)

			msg := types.HttpDownloadResponse{URL: request.URL, Filename: innerpath}
			for _, v := range r.Headers {
				msg.Headers = append(msg.Headers, types.NameValue{Name: v.Name, Value: v.Value})
			}
			return msg, nil
		}
	} else {
		svc.LogError("Failed to request job from si-outer", err)
	}

	return nil, nil
}
