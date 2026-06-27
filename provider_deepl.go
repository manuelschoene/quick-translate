package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

type DeepLProvider struct {
	authKey string
}

func InitDeepLProvider() (Provider, error) {
	authKey := os.Getenv("DEEPL_AUTH_KEY")

	if authKey == "" {
		return nil, fmt.Errorf("Error: DEEPL_AUTH_KEY environment variable is not set")
	}

	return DeepLProvider{authKey}, nil
}

type DeepLRequest struct {
	Text       []string `json:"text"`
	TargetLang string   `json:"target_lang"`
	SourceLang string   `json:"source_lang,omitempty"`
}

type DeepLResponse struct {
	Translations []struct {
		DetectedSourceLang string `json:"detected_source_language"`
		Text               string `json:"text"`
	} `json:"translations"`
}

func (p DeepLProvider) TranslateText(pt PendingTranslation) (string, string, error) {
	url := "https://api-free.deepl.com/v2/translate"

	reqBody := DeepLRequest{
		Text:       []string{pt.Content},
		TargetLang: pt.TargetLang,
		SourceLang: pt.SourceLang,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", "", fmt.Errorf("Error marshaling request body: %v", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", "", fmt.Errorf("Error creating request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "DeepL-Auth-Key "+p.authKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", "", fmt.Errorf("Error making request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("Error: received non-200 response code: %d", resp.StatusCode)
	}

	var deeplResp DeepLResponse
	err = json.NewDecoder(resp.Body).Decode(&deeplResp)
	if err != nil {
		return "", "", fmt.Errorf("Error decoding response body: %v", err)
	}

	if len(deeplResp.Translations) == 0 {
		return "", "", fmt.Errorf("Error: no translations found in response")
	}

	return deeplResp.Translations[0].Text, deeplResp.Translations[0].DetectedSourceLang, nil
}
