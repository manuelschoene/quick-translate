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
	ID          int64      `json:"id"`
	Text        string     `json:"text"`
	Translation string     `json:"translation"`
	CreatedAt   time.Time  `json:"createdAt"`
	ProviderID  ProviderID `json:"providerId"`
	SourceTag   *string    `json:"sourceTag"` // Detected language while detecting; nil if it was not resolved
	TargetTag   string     `json:"targetTag"`
	Detection   bool       `json:"detection"`
}
