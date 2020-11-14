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
		svc.LogGeneric("error", "Marshalling request to JSON failed: %#v", err)
		return nil, err
	}

	jobrequest := otypes.WaitRequest{URL: request.URL}
	if response, err := jobsSvc.RequestMessage("requesturlwait", jobrequest); err == nil {
		var r otypes.WaitResponse
		if err := json.Unmarshal([]byte(response.Payload), &r); err != nil {
			svc.LogGeneric("error", "Marshalling request to JSON failed: %#v", err)
			return nil, err
		}

		fmt.Println("RESPONSE:", r)

		if r.Success {
			if err, exitcode, infos := common.Scan(r.Filename); err != nil {
				fmt.Println("SCAN failed", err, exitcode, infos)
				svc.LogError("SCAN failed", err)
				return nil, err
			}

			wd, _ := os.Getwd()
			oldname := filepath.VolumeName(wd) + r.Filename
			innerpath := strings.Replace(r.Filename, "outer", "inner", 1)

			folder := path.Dir(innerpath)
			os.MkdirAll(folder, os.ModeDir)
			os.Remove(innerpath)

			if err := os.Symlink(oldname, innerpath); err != nil {
				return nil, err
			}

			msg := types.HttpDownloadResponse{URL: request.URL, Filename: innerpath}
			return msg, nil
		}
	} else {
		svc.LogError("Failed to request job from si-outer", err)
	}

	return nil, nil
}
