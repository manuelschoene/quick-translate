package config

// The number of entries kept in the history if the option is not set in the configuration file.
const defaultMaxEntries = 100

// Returns the maximum number of history entries. Zero disables the history.
//
//wails:ignore
func (s *Service) HistoryLimit() int {
	return s.config.historyLimit
}
