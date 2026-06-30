package provider

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

func newDeeplProvider() (*deeplProvider, error) {
	authKey := os.Getenv("DEEPL_AUTH_KEY")
	
	if authKey == "" {
		fmt.Println("Error: DEEPL_AUTH_KEY environment variable is not set")
		return nil, fmt.Errorf("Error: DEEPL_AUTH_KEY environment variable is not set")
	}

	config := &providerConfig{
		slug: "deepl",
		name: "DeepL",
	}
	
	return &deeplProvider{authKey: authKey, pConfig: config, freeVersion: true}, nil
}

type deeplProvider struct {
	authKey string
	freeVersion bool
	pConfig *providerConfig
}

func (p deeplProvider) config() *providerConfig {
	return p.pConfig
}

func (p deeplProvider) url() string {
	if p.freeVersion {
		return "https://api-free.deepl.com"
	}
	return "https://api.deepl.com"
}

type deeplLanguage struct {
	Key string `json:"lang"`
	Name string `json:"name"`
	Source bool `json:"usable_as_source"`
	Target bool `json:"usable_as_target"`
	Status string `json:"status"`
}

func (p deeplProvider) languages() ([]*Language, error) {
	url := p.url() + "/v3/languages"

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("Could not create request: %v", err)
	}

	req.Header.Set("Authorization", "DeepL-Auth-Key " + p.authKey)
	query := req.URL.Query()
	query.Add("resource", "translate_text")
	query.Add("include", "beta")
	req.URL.RawQuery = query.Encode()
	
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Could not make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Received non-200 response code: %d", resp.StatusCode)
	}

	var deeplLangs []deeplLanguage
	err = json.NewDecoder(resp.Body).Decode(&deeplLangs)
	if err != nil {
		return nil, fmt.Errorf("Could not decode response body: %v", err)
	}

	languages := make([]*Language, 0, len(deeplLangs))
	for _, lang := range deeplLangs {
		name := lang.Name
		if lang.Status != "stable" {
			name += " (" + lang.Status + ")"
		}

		languages = append(languages, &Language{
			Key:   lang.Key,
			Name:  name,
			Source: lang.Source,
			Target: lang.Target,
		})
	}
	return languages, nil
}

type deeplTranslateReq struct {
	Text       []string `json:"text"`
	TargetLang string   `json:"target_lang"`
	SourceLang string   `json:"source_lang,omitempty"`
}

type deeplTranslation struct {
	Translations []struct {
		DetectedSourceLang string `json:"detected_source_language"`
		Text               string `json:"text"`
	} `json:"translations"`
}

func (p deeplProvider) translateText(t *Translation) error {
	url := p.url() + "/v2/translate"

	reqBody := deeplTranslateReq{
		Text:       []string{t.Text},
		TargetLang: t.TargetLang.Key,
		SourceLang: t.SourceLang.Key,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("Could not marshal request body: %v", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return fmt.Errorf("Could not create request: %v", err)
	}

	req.Header.Set("Authorization", "DeepL-Auth-Key "+p.authKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("Could not make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Received non-200 response code: %d", resp.StatusCode)
	}

	var deeplResp deeplTranslation
	err = json.NewDecoder(resp.Body).Decode(&deeplResp)
	if err != nil {
		return fmt.Errorf("Could not decode response body: %v", err)
	}

	if len(deeplResp.Translations) == 0 {
		return fmt.Errorf("Could not find translations in response")
	}

	t.Translation = deeplResp.Translations[0].Text
	if t.SourceLang.Key != deeplResp.Translations[0].DetectedSourceLang {
		langs, err := p.languages()
		if err != nil {
			return err
		}
		for _, lang := range langs {
			if lang.Key == deeplResp.Translations[0].DetectedSourceLang {
				t.SourceLang = lang
				break
			}
		}	
	}

	return nil
}
