package services

import (
	"encoding/base64"
	"encoding/json"
	"log"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/cyops-se/safe-import/si-inner/common"
	"github.com/cyops-se/safe-import/si-inner/types"
	outertypes "github.com/cyops-se/safe-import/si-outer/types"
	"github.com/cyops-se/safe-import/usvc"
	"github.com/nats-io/nats.go"
)

type RepositoryService struct {
	usvc.Usvc
}

func (svc *RepositoryService) Initialize(broker *usvc.UsvcBroker) {
	svc.InitializeService(broker, 1, "si-inner", "repos", "This service manages the repositories in the inner part of safe-import")
	svc.RegisterMethod("allitems", svc.allItems)
	svc.RegisterMethod("byid", svc.byId)
	svc.RegisterMethod("byurl", svc.byUrl)
	svc.RegisterMethod("create", svc.create)
	svc.RegisterMethod("update", svc.update)
	svc.RegisterMethod("approve", svc.approve)
	svc.RegisterMethod("deletebyid", svc.deletebyid)

	// We don't use settings right now
	if err := svc.LoadSettings(); err != nil {
		svc.SaveSettings() // Save default settings. Though we don't actually use the settings right now...
	}

	// Subscribe for job events that affect repo status
	broker.Subscribe("data.1.si-outer.jobs.>", svc.eventHandler)

	svc.Executor = svc.execute
	svc.SetTaskIdleTime(60 * 1) // every minute
	svc.execute()
}

func (svc *RepositoryService) execute() {
}

func (svc *RepositoryService) allItems(payload string) (interface{}, error) {
	var items []types.Repository
	if result := common.DB.Find(&items); result.Error != nil {
		svc.LogGeneric("error", "Error while accessing local database: %#v", result.Error)
		return nil, result.Error
	} else {
		// fmt.Println("result: ", result)
	}

	return items, nil
}

func (svc *RepositoryService) byId(payload string) (interface{}, error) {
	var args types.ByIdRequest
	if err := json.Unmarshal([]byte(payload), &args); err != nil {
		svc.LogGeneric("error", "Marshalling request to JSON failed: %#v", err)
		return nil, err
	}

	var item types.Repository
	if result := common.DB.Find(&item, "id = ?", args.ID); result.Error != nil {
		svc.LogGeneric("error", "Error while accessing local database: %#v", result.Error)
		return nil, result.Error
	}

	return item, nil
}

func (svc *RepositoryService) byUrl(payload string) (interface{}, error) {
	var args types.ByNameRequest
	if err := json.Unmarshal([]byte(payload), &args); err != nil {
		svc.LogGeneric("error", "Marshalling request to JSON failed: %#v", err)
		return nil, err
	}

	var item types.Repository
	if result := common.DB.Where(map[string]interface{}{args.Name: args.Value}).First(&item); result.Error != nil {
		svc.LogGeneric("error", "Error while accessing local database: %#v", result.Error)
		return nil, result.Error
	}

	return item, nil
}

func (svc *RepositoryService) create(payload string) (interface{}, error) {
	var item types.Repository
	if err := json.Unmarshal([]byte(payload), &item); err != nil {
		svc.LogGeneric("error", "Marshalling request to JSON failed: %#v", err)
		return nil, err
	}

	item.Hash = base64.StdEncoding.EncodeToString([]byte(item.URL))

	if result := common.DB.Create(&item); result.Error != nil {
		svc.LogGeneric("error", "Error while accessing local database: %#v", result.Error)
		return nil, result.Error
	}

	item.OuterPath = path.Join("/safe-import/outer", strconv.Itoa(int(item.ID)))
	item.InnerPath = path.Join("/safe-import/inner", strconv.Itoa(int(item.ID)))
	common.DB.Save(&item)

	return item, nil
}

func (svc *RepositoryService) approve(payload string) (interface{}, error) {
	var args types.ApproveRequest
	if err := json.Unmarshal([]byte(payload), &args); err != nil {
		svc.LogGeneric("error", "Marshalling request to JSON failed: %#v", err)
		return nil, err
	}

	item := &types.Repository{Type: args.Type, URL: args.URL}
	item.Hash = base64.StdEncoding.EncodeToString([]byte(item.URL))

	if result := common.DB.Create(&item); result.Error != nil {
		svc.LogGeneric("error", "Error while accessing local database: %#v", result.Error)
		return nil, result.Error
	}

	item.OuterPath = path.Join("/safe-import/outer", strconv.Itoa(int(item.ID)))
	item.InnerPath = path.Join("/safe-import/inner", strconv.Itoa(int(item.ID)))
	common.DB.Save(&item)

	msg := &types.ByIdRequest{ID: item.ID}
	svc.PublishData("approved", msg)

	return item, nil
}

func (svc *RepositoryService) update(payload string) (interface{}, error) {
	var item types.Repository
	if err := json.Unmarshal([]byte(payload), &item); err != nil {
		svc.LogGeneric("error", "Marshalling request to JSON failed: %#v", err)
		return nil, err
	}

	if result := common.DB.Save(&item); result.Error != nil {
		svc.LogGeneric("error", "Error while accessing local database: %#v", result.Error)
		return nil, result.Error
	}

	return item, nil
}

func (svc *RepositoryService) deleteall(payload string) (interface{}, error) {
	common.DB.Delete(&types.Repository{})
	return nil, nil
}

func (svc *RepositoryService) deletebyid(payload string) (interface{}, error) {
	var args types.ByIdRequest
	if err := json.Unmarshal([]byte(payload), &args); err != nil {
		svc.LogGeneric("error", "Marshalling request to JSON failed: %#v", err)
		return nil, err
	}

	var item types.Repository
	result := common.DB.Delete(&types.Repository{}, args.ID)
	if result.Error != nil {
		svc.LogGeneric("error", "Error while accessing local database: %#v", result.Error)
		return nil, result.Error
	}

	return item, nil
}

func (svc *RepositoryService) eventHandler(m *nats.Msg) {
	var event outertypes.Job
	if err := json.Unmarshal(m.Data, &event); err != nil {
		log.Println("ERROR: Failed to unmarshal log entry:", string(m.Data), err)
		return
	}

	log.Println("EVENT", event)

	// Find the repo
	var repo types.Repository
	if result := common.DB.First(&repo, event.Progress.ID); result.Error != nil {
		svc.LogGeneric("error", "Could not find repo from event, event: %#v, error: %#v", event, result.Error)
		return
	}

	// fmt.Println("REPO", repo)

	if strings.HasSuffix(m.Subject, "completed") {
		repo.LastSuccess = time.Now().UTC()
		repo.Available = true
	} else if strings.HasSuffix(m.Subject, "stopped") {
		repo.LastFailure = time.Now().UTC()
	} else if strings.HasSuffix(m.Subject, "infected") {
		repo.LastFailure = time.Now().UTC()
	} else if strings.HasSuffix(m.Subject, "failed") {
		repo.LastFailure = time.Now().UTC()
	}

	common.DB.Save(&repo)
}
