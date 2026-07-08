package provider

import (
	"fmt"

	"quick-translate/internal/models"
)

type provider interface {
	languages() ([]*models.Language, error)
	translate(*models.Translation) error
}

type Meta struct {
	Slug, Name        string
	LanguageDetection bool
}

type Service struct {
	current  *Meta
	all      []*Meta
	provider provider
}

func NewService() (*Service, error) {
	if err := initFile(); err != nil {
		return nil, err
	}

	all, m, p, err := loadConfig()
	if err != nil {
		return nil, err
	}

	return &Service{
		current:  m,
		all:      all,
		provider: p,
	}, nil
}

func (s *Service) Active() *Meta {
	return s.current
}

func (s *Service) List() []*Meta {
	return s.all
}

func (s *Service) Use(slug string) error {
	for _, m := range s.all {
		if m.Slug == slug {
			p, err := loadProvider(m)
			if err != nil {
				return err
			}

			s.current = m
			s.provider = p

			return nil
		}
	}

	return fmt.Errorf("Invalid provider: '%s'", slug)
}

func (s *Service) Languages() ([]*models.Language, error) {
	return s.provider.languages()
}

func (s *Service) Translate(t *models.Translation) error {
	return s.provider.translate(t)
}

func (m *Meta) provider() provider {
	switch m.Slug {
	case "deepl":
		return &deepl{}
	}

	return nil
}
