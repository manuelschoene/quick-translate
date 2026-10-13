package models

import (
	"time"
)

type Language struct {
	Tag    string `json:"tag"`
	Name   string `json:"name"`
	Source bool   `json:"source"`
	Target bool   `json:"target"`
	Stable bool   `json:"stable"`
}

type Translation struct {
	ID          int64
	Text        string
	Translation string
	CreatedAt   time.Time
	ProviderID  ProviderID
	SourceTag   *string
	TargetTag   string
	Detection   bool
}
