package provider

import (
	"fmt"
)

func NewService() (*ProviderService, error) {
	p, err := initProvider("deepl")
	if err != nil {
		return nil, err
	}

	return &ProviderService{
		provider: p,
	}, nil
}

type ProviderService struct {
	provider provider
}

type provider interface {
	config() *providerConfig
	translateText(*Translation) error
	languages() ([]*Language, error)
}

type Language struct {
	Key, Name string
	Source, Target bool
}

type Translation struct {
	Text, Translation string
	SourceLang, TargetLang *Language
}

type providerConfig struct {
	slug, name string
}

func (ps *ProviderService) ActiveProvider() string {
	return ps.provider.config().slug
}

func (ps *ProviderService) ListProviders() map[string]string {
	return map[string]string{
		ps.provider.config().slug: ps.provider.config().name,
	}
}

func (ps *ProviderService) UseProvider(pSlug string) error {
	p, err := initProvider(pSlug)
	if err != nil {
		return err
	}

	ps.provider = p
	return nil
}

func (ps *ProviderService) TranslateText(t *Translation) error {
	return ps.provider.translateText(t)
}

func (ps *ProviderService) Languages() ([]*Language, error) {
	return ps.provider.languages()
}

func initProvider(pSlug string) (provider, error) {
	switch pSlug {
	case "deepl":
		p, err := newDeeplProvider()
		if err != nil {
			return nil, err
		}
		return p, nil
	}

	return nil, fmt.Errorf("Unknown provider slug: %s", pSlug)
}
