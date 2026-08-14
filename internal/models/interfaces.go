package models

type Provider interface {
	Translate(source string, target string, text string) (string, string, error)
	Languages() ([]*Language, error)
}
