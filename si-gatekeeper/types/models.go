package types

import (
	"gorm.io/gorm"
)

type CachedItem struct {
	gorm.Model
	Name      string `json:"name"`
	Filename  string `json:"filename"`
	Path      string `json:"path"`
	Approved  bool   `json:"approved"`
	Available bool   `json:"available"`
}

type InfectionInfo struct {
	gorm.Model
	Filename       string `json:"filename"`
	OriginalPath   string `json:"originalpath"`
	QuarantinePath string `json:"quarantinepath"`
	VirusName      string `json:"virusname"`
}
