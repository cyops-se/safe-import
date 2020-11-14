package services

import (
	"encoding/json"
	"fmt"

	"github.com/cyops-se/safe-import/si-gatekeeper/common"
	"github.com/cyops-se/safe-import/si-gatekeeper/types"
	"github.com/cyops-se/safe-import/usvc"
)

type CacheService struct {
	usvc.Usvc
}

func (svc *CacheService) Initialize(broker *usvc.UsvcBroker) {
	svc.InitializeService(broker, 1, "si-gatekeeper", "proxy", "Proxies request from si-inner to si-outer")
	svc.RegisterMethod("allitems", svc.allItems)
	svc.RegisterMethod("byfieldname", svc.byFieldName)
	svc.RegisterMethod("update", svc.update)
	svc.RegisterMethod("prune", svc.prune)

	// We don't use settings right now
	if err := svc.LoadSettings(); err != nil {
		svc.SaveSettings() // Save default settings. Though we don't actually use the settings right now...
	}

	svc.Executor = svc.execute
	svc.SetTaskIdleTime(60 * 1) // every minute
	svc.execute()
}

func (svc *CacheService) execute() {
}

func (svc *CacheService) allItems(payload string) (interface{}, error) {
	var items []types.CachedItem
	common.DB.Find(&items)
	return items, nil
}

func (svc *CacheService) byFieldName(payload string) (interface{}, error) {
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

func (svc *CacheService) update(payload string) (interface{}, error) {
	return nil, fmt.Errorf("Method not yet implemented")
}

func (svc *CacheService) prune(payload string) (interface{}, error) {
	return nil, fmt.Errorf("Method not yet implemented")
}
