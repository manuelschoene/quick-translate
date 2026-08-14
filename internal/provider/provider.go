package provider

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
