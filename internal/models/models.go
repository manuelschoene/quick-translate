package models

import (
	"time"
)

type Language struct {
	Key, Name          string
	IsSource, IsTarget bool
}

type Translation struct {
	DetectedSourceLanguage, Text, Translation, Provider string
	SourceLanguage                                      *Language
	TargetLanguage                                      *Language
	Timestamp                                           time.Time
}
