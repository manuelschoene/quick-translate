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

type LanguagePreferences struct {
	Source, Target string
}

type Translation struct {
	ID                                                          int
	CreatedAt                                                   time.Time
	Source, Target, Text, Translation, DetectedSource, Provider string
}
