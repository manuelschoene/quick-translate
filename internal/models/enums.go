package models

// Referenced in the sqlc.yml for overrides. Used inside the database as primary key for the providers table. Do not use Iota and do not change the values, as they are stored in the database and referenced by other tables.
type ProviderID int64

const (
	DeepL ProviderID = 1
)
