package provider

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"quick-translate/internal/models"
)

type deepl struct {
	AuthKey     string `yaml:"auth_key"`
	FreeVersion bool   `yaml:"free_version"`
	FastMode    bool   `yaml:"fast_mode"`
	Formality   string `yaml:"formality"`
}

type deeplLanguage struct {
	Key      string `json:"lang"`
	Name     string `json:"name"`
	IsSource bool   `json:"usable_as_source"`
	IsTarget bool   `json:"usable_as_target"`
	Status   string `json:"status"`
}

type deeplTranslateReq struct {
	Text      []string `json:"text"`
	Target    string   `json:"target_lang"`
	Source    string   `json:"source_lang,omitempty"`
	Formality string   `json:"formality"`
	ModelType string   `json:"model_type"`
}

type deeplTranslation struct {
	Translations []struct {
		DetectedSource string `json:"detected_source_language"`
		Text           string `json:"text"`
	} `json:"translations"`
}

func (p deepl) languages() ([]*models.Language, error) {
	fmt.Println(p)

	url := p.url() + "/v3/languages"

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("Could not create request: %v", err)
	}

	req.Header.Set("Authorization", "DeepL-Auth-Key "+p.AuthKey)

	query := req.URL.Query()
	query.Add("resource", "translate_text")
	query.Add("include", "beta")
	req.URL.RawQuery = query.Encode()

	resp, err := makeRequest(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return decodeLanguageResponse(resp)
}

func (p deepl) translate(t *models.Translation) error {
	url := p.url() + "/v2/translate"
	body, err := p.buildBody(t)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("Could not create request: %v", err)
	}

	req.Header.Set("Authorization", "DeepL-Auth-Key "+p.AuthKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := makeRequest(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return decodeTranslateResponse(resp, t)
}

func (p deepl) url() string {
	if p.FreeVersion {
		return "https://api-free.deepl.com"
	}
	return "https://api.deepl.com"
}

func (p deepl) buildBody(t *models.Translation) ([]byte, error) {
	buf := new(bytes.Buffer)
	buf.WriteString(t.Text)
	if buf.Len() > 127*1024 {
		return nil, fmt.Errorf("Text exceeds maximum length of 127KiB")
	}

	var formality string
	switch p.Formality {
	case "formal":
		formality = "prefer_more"
	case "informal":
		formality = "prefer_less"
	default:
		formality = "default"
	}

	var modelType string
	if p.FastMode {
		modelType = "latency_optimized"
	} else {
		modelType = "prefer_quality_optimized"
	}

	body := &deeplTranslateReq{
		Text:      []string{t.Text},
		Target:    t.TargetLanguage.Key,
		Source:    t.SourceLanguage.Key,
		Formality: formality,
		ModelType: modelType,
	}

	json, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("Could not marshal request body: %v", err)
	}

	return json, nil
}

func makeRequest(req *http.Request) (*http.Response, error) {
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Could not make request: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("Received non-200 response code: %d", resp.StatusCode)
	}

	return resp, nil
}

func decodeLanguageResponse(resp *http.Response) ([]*models.Language, error) {
	var deeplLangs []deeplLanguage
	err := json.NewDecoder(resp.Body).Decode(&deeplLangs)
	if err != nil {
		return nil, fmt.Errorf("Could not decode response body: %v", err)
	}

	languages := make([]*models.Language, 0, len(deeplLangs))

	for _, lang := range deeplLangs {
		name := lang.Name
		if lang.Status != "stable" {
			name += " (" + lang.Status + ")"
		}

		languages = append(languages, &models.Language{
			Key:      lang.Key,
			Name:     name,
			IsSource: lang.IsSource,
			IsTarget: lang.IsTarget,
		})
	}

	return languages, nil
}

func decodeTranslateResponse(resp *http.Response, t *models.Translation) error {
	var deeplResp deeplTranslation
	err := json.NewDecoder(resp.Body).Decode(&deeplResp)
	if err != nil {
		return fmt.Errorf("Could not decode response body: %v", err)
	}

	if len(deeplResp.Translations) == 0 {
		return fmt.Errorf("Could not find translations in response")
	}

	t.Translation = deeplResp.Translations[0].Text
	t.DetectedSourceLanguage = deeplResp.Translations[0].DetectedSource

	return nil
}
