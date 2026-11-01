package models

// Referenced in the sqlc.yml for overrides. Used inside the database as primary key for the providers table. Do not use Iota and do not change the values, as they are stored in the database and referenced by other tables.
type ProviderID int64

const (
	DeepL ProviderID = 1
)

// The slugs the config file names the providers by.
var providerSlugs = map[ProviderID]string{
	DeepL: "deepl",
}

// Returns the slug of the provider, empty for an unknown ID.
func (id ProviderID) String() string {
	return providerSlugs[id]
}

// Returns the provider ID for a slug; false if no provider has that slug.
func ProviderIDFromSlug(slug string) (ProviderID, bool) {
	for id, s := range providerSlugs {
		if s == slug {
			return id, true
		}
	}

	return 0, false
}
