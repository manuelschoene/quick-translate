package provider

import (
	"net/http"
	"time"
)

// How long a provider gets to answer before its request is given up on. Go applies no timeout of its own,
// and a translation is made while the core holds its lock, so a request that never returns would freeze
// every other call into the application rather than just itself.
const requestTimeout = 30 * time.Second

// The client every provider makes its requests with. Shared, so connections are reused across translations
// instead of a new one being opened for each of them.
var client = &http.Client{Timeout: requestTimeout}

// The slug the DeepL provider is identified by in the configuration file and in the history.
const SlugDeepl = "deepl"

// Returns a map of all supported provider slugs and if the provider supports language detection.
func All() map[string]bool {
	return map[string]bool{
		SlugDeepl: true,
	}
}

// Creates a new instance of the DeepL provider. The instance is not configured and must be filled with the required configuration values before use.
func NewDeepl() *Deepl {
	return &Deepl{}
}
