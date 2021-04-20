package services

import (
	"encoding/json"

	"github.com/cyops-se/safe-import/si-outer/common"
	"github.com/cyops-se/safe-import/si-outer/types"
	"github.com/cyops-se/safe-import/usvc"
)

type RepoService struct {
	usvc.Usvc
	repos map[int]*types.Job
}

func (svc *RepoService) Initialize(broker *usvc.UsvcBroker) {
	svc.InitializeService(broker, 1, "si-outer", "repos", "Takes care of repositories (only directory based repos supported at the moment)")
	svc.RegisterMethod("allitems", svc.findAll)
	svc.RegisterMethod("findbyid", svc.findById)
	svc.RegisterMethod("create", svc.create)
	svc.RegisterMethod("update", svc.update)
	svc.RegisterMethod("deletebyid", svc.deleteById)

	svc.repos = make(map[int]*types.Job, 1)
}

func (svc *RepoService) findAll(payload string) (interface{}, error) {
	var items []types.Repository
	common.DB.Find(&items)
	return items, nil
}

func (svc *RepoService) findById(payload string) (interface{}, error) {
	var item types.ByIdRequest
	if err := json.Unmarshal([]byte(payload), &item); err != nil {
		svc.LogGeneric("error", "Marshalling request to JSON failed: %#v", err)
		return nil, err
	}

	var repo types.Repository
	common.DB.Find(&repo, item.ID)
	return item, nil
}

func (svc *RepoService) create(payload string) (interface{}, error) {
	var item types.Repository
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

func (svc *RepoService) update(payload string) (interface{}, error) {
	var item types.Repository
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

func (svc *RepoService) deleteById(payload string) (interface{}, error) {
	var item types.ByIdRequest
	if err := json.Unmarshal([]byte(payload), &item); err != nil {
		svc.LogGeneric("error", "Marshalling request to JSON failed: %#v", err)
		return nil, err
	}

	if result := common.DB.Delete(&types.Repository{}, item.ID); result.Error != nil {
		svc.LogGeneric("error", "Failed to query database, error: %#v", result.Error)
		return nil, result.Error
	}

	return item, nil
}
