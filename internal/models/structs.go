package models

import (
	"time"
)

type Language struct {
	Tag, Name              string
	Source, Target, Stable bool
}

type LanguagePreferences struct {
	Source, Target string
}

type Translation struct {
	ID                                                          int
	CreatedAt                                                   time.Time
	Source, Target, Text, Translation, DetectedSource, Provider string
}
