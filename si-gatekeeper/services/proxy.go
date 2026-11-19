package services

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/cyops-se/safe-import/si-gatekeeper/common"
	"github.com/cyops-se/safe-import/si-gatekeeper/system"
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

	os.MkdirAll(common.OuterDir(), 0755)
	os.MkdirAll(common.InnerDir(), 0755)
	os.MkdirAll(common.StagingDir(), 0750)
	os.MkdirAll(system.QuarantineDir(), 0750)

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

	log.Println("Processing request:", request.Method, request.URL)

	jobrequest := otypes.WaitRequest{URL: request.URL, Method: request.Method, Body: request.Body}
	for _, v := range request.Headers {
		jobrequest.Headers = append(jobrequest.Headers, otypes.NameValue{Name: v.Name, Value: v.Value})
	}

	svc.PublishString("file.download.start", request.URL)
	response, err := jobsSvc.RequestMessage("requesturlwait", jobrequest)
	if err != nil {
		svc.LogError("Failed to request job from si-outer", err)
		svc.PublishString("file.download.fail", request.URL)
		return nil, err
	}

	var r otypes.WaitResponse
	if err := json.Unmarshal([]byte(response.Payload), &r); err != nil {
		svc.LogGeneric("error", "Marshalling response JSON payload to response failed: %#v, %s", err, response.Payload)
		svc.PublishString("file.download.fail", request.URL)
		return nil, err
	}

	if !r.Success {
		svc.PublishString("file.download.fail", request.URL)
		return nil, fmt.Errorf("si-outer failed to download %s", request.URL)
	}
	svc.PublishString("file.download.success", request.URL)

	// si-outer is the least trusted component: verify that the reported file is inside the outer area
	rel, err := common.RelativeToRoot(common.OuterDir(), r.Filename)
	if err != nil {
		svc.LogGeneric("alert", "si-outer reported a file outside the outer area: %s", err.Error())
		svc.PublishString("file.download.fail", request.URL)
		return nil, err
	}

	// Copy the file to the gatekeeper's private staging area BEFORE scanning it, so that
	// si-outer can not change the content after it has been scanned
	wd, _ := os.Getwd()
	src := filepath.VolumeName(wd) + r.Filename
	staged, sha, err := common.CopyToStaging(src)
	if err != nil {
		svc.LogError("Failed to copy downloaded file to staging area", err)
		svc.PublishString("file.download.fail", request.URL)
		return nil, err
	}
	defer os.Remove(staged) // No-op if it has been moved to inner or quarantine

	if request.NoScan {
		svc.LogGeneric("warning", "Virus scan SKIPPED (noscan set in white list) for %s, sha256 %s", request.URL, sha)
	} else {
		svc.PublishString("file.download.scan.start", request.URL)
		if err, exitcode, infos := common.Scan(staged); err != nil {
			if exitcode == 1 {
				for _, info := range infos {
					svc.LogInfection(fmt.Sprintf("VIRUS: %s", info.VirusName), "%s (url %s, sha256 %s)", info.Filename, request.URL, sha)
					svc.PublishEventMessage("proxy.infection", info)
				}
			} else {
				svc.LogError(fmt.Sprintf("SCAN failed (exitcode: %d)", exitcode), err)
			}
			svc.PublishString("file.download.scan.fail", request.URL)
			return nil, fmt.Errorf("file %s was rejected by the virus scan (exit code %d)", request.URL, exitcode)
		}
		svc.PublishString("file.download.scan.success", request.URL)
	}

	innerpath := filepath.Join(common.InnerDir(), rel)
	svc.PublishString("file.download.link.start", request.URL)
	if err := common.MoveToInner(staged, filepath.VolumeName(wd)+innerpath); err != nil {
		svc.LogError("Failed to move scanned file to inner area", err)
		svc.PublishString("file.download.link.fail", request.URL)
		return nil, err
	}
	svc.PublishString("file.download.link.success", request.URL)
	svc.LogGeneric("info", "Imported %s (sha256 %s)", request.URL, sha)

	msg := types.HttpDownloadResponse{URL: request.URL, Filename: innerpath}
	for _, v := range r.Headers {
		msg.Headers = append(msg.Headers, types.NameValue{Name: v.Name, Value: v.Value})
	}
	return msg, nil
}
