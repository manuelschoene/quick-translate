package config

// The preferred languages as BCP 47 tags. Not validated besides being strings; empty means no preference.
type LanguagePreferences struct {
	Source string `yaml:"source"`
	Target string `yaml:"target"`
}

// Returns the preferred source and target languages.
//
//wails:ignore
func (s *Service) Preferences() LanguagePreferences {
	return s.config.preferences
}
