package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"quick-translate/internal/models"
)

const (
	deeplURL          = "https://api.deepl.com"
	deeplFreeURL      = "https://api-free.deepl.com"
	deeplMaxText      = 127 * 1024 // DeepL refuses request bodies over 128 KiB; the rest is left for the JSON around the text.
	deeplMaxErrorBody = 4 * 1024   // How much of an unexpected error response is read into the error.
)

// The settings of the DeepL provider, as they are configured.
type deeplSettings struct {
	AuthKey     string `yaml:"auth_key"`
	FreeVersion bool   `yaml:"free_version"`
	FastMode    bool   `yaml:"fast_mode"`
	Formality   string `yaml:"formality"`
}

// Fails if the auth key is missing or the formality is unknown.
func (s deeplSettings) validate() error {
	if len(s.AuthKey) == 0 {
		return errors.New("auth_key not set")
	}

	switch s.Formality {
	case "", "default", "formal", "informal":
		return nil
	default:
		return fmt.Errorf("formality %q unknown, use default, formal or informal", s.Formality)
	}
}

type deepl struct {
	settings deeplSettings
	client   *http.Client
}

type deeplLanguage struct {
	Key      string `json:"lang"`
	Name     string `json:"name"`
	IsSource bool   `json:"usable_as_source"`
	IsTarget bool   `json:"usable_as_target"`
	Status   string `json:"status"`
}

type deeplTranslateRequest struct {
	Text      []string `json:"text"`
	Target    string   `json:"target_lang"`
	Source    string   `json:"source_lang,omitempty"`
	Formality string   `json:"formality"`
	ModelType string   `json:"model_type"`
}

type deeplTranslateResponse struct {
	Translations []struct {
		DetectedSource string `json:"detected_source_language"`
		Text           string `json:"text"`
	} `json:"translations"`
}

// Builds the DeepL provider from its settings.
func newDeepl(decode func(into any) error, env environment) (implementation, error) {
	var settings deeplSettings
	if err := decode(&settings); err != nil {
		return nil, err
	}

	if err := settings.validate(); err != nil {
		return nil, err
	}

	return &deepl{settings: settings, client: env.client}, nil
}

// Fetches the languages DeepL can translate text from and to, including the beta ones.
func (p *deepl) languages(ctx context.Context) ([]models.Language, error) {
	query := url.Values{}
	query.Add("resource", "translate_text")
	query.Add("include", "beta")

	var response []deeplLanguage
	if err := p.do(ctx, http.MethodGet, "/v3/languages", query, nil, &response); err != nil {
		return nil, err
	}

	languages := make([]models.Language, 0, len(response))
	for _, lang := range response {
		languages = append(languages, models.Language{
			Tag:    lang.Key,
			Name:   lang.Name,
			Source: lang.IsSource,
			Target: lang.IsTarget,
			Stable: lang.Status == "stable",
		})
	}

	return languages, nil
}

// Translates the text. The detected language is lowercased.
func (p *deepl) translate(ctx context.Context, request Request) (Result, error) {
	if len(request.Text) > deeplMaxText {
		return Result{}, errors.New("text longer than 127 KiB, split it into smaller parts")
	}

	body, err := json.Marshal(deeplTranslateRequest{
		Text:      []string{request.Text},
		Target:    request.Target,
		Source:    request.Source,
		Formality: p.formality(),
		ModelType: p.modelType(),
	})
	if err != nil {
		return Result{}, fmt.Errorf("encode request: %w", err)
	}

	var response deeplTranslateResponse
	if err := p.do(ctx, http.MethodPost, "/v2/translate", nil, body, &response); err != nil {
		return Result{}, err
	}

	if len(response.Translations) == 0 {
		return Result{}, errors.New("response holds no translation")
	}

	return Result{
		Text:     response.Translations[0].Text,
		Detected: strings.ToLower(response.Translations[0].DetectedSource),
	}, nil
}

// Maps the configured formality onto DeepL's preference values.
func (p *deepl) formality() string {
	switch p.settings.Formality {
	case "formal":
		return "prefer_more"
	case "informal":
		return "prefer_less"
	default:
		return "default"
	}
}

// Picks the faster model in fast mode, the better one otherwise.
func (p *deepl) modelType() string {
	if p.settings.FastMode {
		return "latency_optimized"
	}

	return "prefer_quality_optimized"
}

// Returns the API's base URL for the configured plan.
func (p *deepl) url() string {
	if p.settings.FreeVersion {
		return deeplFreeURL
	}

	return deeplURL
}

// Sends an authorized request and decodes a 200 response into out. A body is sent as JSON.
func (p *deepl) do(ctx context.Context, method string, path string, query url.Values, body []byte, out any) error {
	var reader io.Reader
	if body != nil {
		reader = bytes.NewReader(body)
	}

	req, err := http.NewRequestWithContext(ctx, method, p.url()+path, reader)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Authorization", "DeepL-Auth-Key "+p.settings.AuthKey)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	if len(query) != 0 {
		req.URL.RawQuery = query.Encode()
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return fmt.Errorf("request failed, check your internet connection: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return deeplStatusError(resp)
	}

	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}

	return nil
}

// What a status means, for the statuses DeepL documents. Translating answers 400, 403, 404, 413, 414, 429, 456,
// 500, 504 and 529, listing the languages 400, 401, 403, 429, 500 and 503.
var deeplStatuses = map[int]string{
	http.StatusBadRequest:            "invalid request",
	http.StatusUnauthorized:          "auth key rejected, check the auth key",
	http.StatusForbidden:             "auth key rejected, check the auth key",
	http.StatusNotFound:              "not found",
	http.StatusRequestEntityTooLarge: "text too large, split it into smaller parts",
	http.StatusRequestURITooLong:     "request URL too long",
	http.StatusTooManyRequests:       "too many requests, try again later",
	456:                              "quota exceeded for this billing period",
	http.StatusInternalServerError:   "internal server error, try again later",
	http.StatusServiceUnavailable:    "service unavailable, try again later",
	http.StatusGatewayTimeout:        "gateway timeout, try again later",
	529:                              "too many requests, try again later",
}

// The body DeepL answers every non-200 status with.
type deeplErrorResponse struct {
	Message string `json:"message"`
	Code    string `json:"code"`
}

// Turns a non-200 response into an error carrying the API's message. A body that is not the documented JSON is
// used as it is.
func deeplStatusError(resp *http.Response) error {
	meaning, ok := deeplStatuses[resp.StatusCode]
	if !ok {
		meaning = "unexpected response"
	}

	message := deeplErrorMessage(resp.Body)
	if len(message) == 0 {
		return fmt.Errorf("%s (status %d)", meaning, resp.StatusCode)
	}

	return fmt.Errorf("%s (status %d): %s", meaning, resp.StatusCode, message)
}

// Reads the message of an error response, or the start of the raw body if it is not the documented JSON. Empty
// if the body is empty or unreadable.
func deeplErrorMessage(body io.Reader) string {
	raw, err := io.ReadAll(io.LimitReader(body, deeplMaxErrorBody))
	if err != nil {
		return ""
	}

	var response deeplErrorResponse
	if err := json.Unmarshal(raw, &response); err == nil {
		if len(response.Message) != 0 {
			return response.Message
		}

		return response.Code
	}

	return strings.TrimSpace(string(raw))
}
