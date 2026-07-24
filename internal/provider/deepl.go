package provider

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"quick-translate/internal/models"
)

type Deepl struct {
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

// Fetches all available languages from the DeepL API. If the request fails, an error is returned.
func (p Deepl) Languages() ([]*models.Language, error) {
	url := p.url() + "/v3/languages"

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("Could not create HTTP request: %w", err)
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

// Translates the given text from the source language to the target language using the DeepL API. If the request fails, an error is returned. The function returns the translated text and the detected source language.
func (p Deepl) Translate(source string, target string, text string) (string, string, error) {
	url := p.url() + "/v2/translate"

	body, err := p.buildTranslateBody(source, target, text)
	if err != nil {
		return "", "", err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return "", "", fmt.Errorf("Could not create HTTP request: %w", err)
	}

	req.Header.Set("Authorization", "DeepL-Auth-Key "+p.AuthKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := makeRequest(req)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	return decodeTranslateResponse(resp)
}

// Returns the base URL for the DeepL API, depending on whether the free version is used or not.
func (p Deepl) url() string {
	if p.FreeVersion {
		return "https://api-free.deepl.com"
	}
	return "https://api.deepl.com"
}

// Builds the request body for a DeepL translation request. If the text is longer than 127KiB, an error is returned.
func (p Deepl) buildTranslateBody(source string, target string, text string) ([]byte, error) {
	buf := new(bytes.Buffer)
	buf.WriteString(text)
	if buf.Len() > 127*1024 {
		return nil, fmt.Errorf("Cannot translate text longer than 127KiB. Try splitting the text into smaller chunks.")
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
		Text:      []string{text},
		Target:    target,
		Source:    source,
		Formality: formality,
		ModelType: modelType,
	}

	json, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("Could not encode request body: %v", err)
	}

	return json, nil
}

// Fires an HTTP request and returns the response. If the response code is not 200, an error is returned with a message based on the response code. The request body is not closed in this function, it is the caller's responsibility to close it after use.
func makeRequest(req *http.Request) (*http.Response, error) {
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		switch resp.StatusCode {
		case http.StatusUnauthorized:
			err = fmt.Errorf("Unauthorized: Please check your DeepL API key.")
		case http.StatusTooManyRequests:
			err = fmt.Errorf("Too Many Requests: You have exceeded the rate limit for the DeepL API. Try again later.")
		default:
			body, readErr := io.ReadAll(resp.Body)
			if readErr == nil {
				err = fmt.Errorf("Received non-200 HTTP response code: %d. Response body: %s", resp.StatusCode, string(body))
			}
		}

		resp.Body.Close()
		return nil, err
	}

	return resp, nil
}

// Decodes the response body of a DeepL language list request into a slice of Language models. If the response body cannot be decoded, an error is returned.
func decodeLanguageResponse(resp *http.Response) ([]*models.Language, error) {
	var deeplLangs []deeplLanguage

	err := json.NewDecoder(resp.Body).Decode(&deeplLangs)
	if err != nil {
		return nil, fmt.Errorf("Could not decode response body: %v", err)
	}

	languages := make([]*models.Language, 0, len(deeplLangs))

	for _, lang := range deeplLangs {
		languages = append(languages, &models.Language{
			Tag:    lang.Key,
			Name:   lang.Name,
			Source: lang.IsSource,
			Target: lang.IsTarget,
			Stable: lang.Status == "stable",
		})
	}

	return languages, nil
}

// Decodes the response body of a DeepL translation request into the translated text and detected source language. If the response body cannot be decoded, or if no translations are found, an error is returned.
func decodeTranslateResponse(resp *http.Response) (string, string, error) {
	var deeplResp deeplTranslation

	err := json.NewDecoder(resp.Body).Decode(&deeplResp)
	if err != nil {
		return "", "", fmt.Errorf("Could not decode response body: %v", err)
	}

	if len(deeplResp.Translations) == 0 {
		return "", "", fmt.Errorf("Could not find any translations in HTTP response.")
	}

	return deeplResp.Translations[0].Text, strings.ToLower(deeplResp.Translations[0].DetectedSource), nil
}
