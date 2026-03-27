// Package analyzer provides an HTTP client for the Python audio analyzer service.
package analyzer

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// DefaultURL is the default analyzer service endpoint.
const DefaultURL = "http://localhost:8000"

// AudioFeatures holds the DSP features extracted by the Python analyzer.
type AudioFeatures struct {
	BPM                  float64            `json:"bpm"`
	Key                  string             `json:"key"`
	SpectralCentroidMean float64            `json:"spectral_centroid_mean"`
	SpectralBandwidthMean float64           `json:"spectral_bandwidth_mean"`
	MFCCMeans            []float64          `json:"mfcc_means"`
	RMSEnergyMean        float64            `json:"rms_energy_mean"`
	ZeroCrossingRateMean float64            `json:"zero_crossing_rate_mean"`
	ChromaFeatures       map[string]float64 `json:"chroma_features"`
	OnsetStrengthMean    float64            `json:"onset_strength_mean"`
	DurationSeconds      float64            `json:"duration_seconds"`
	EnergyProfile        []float64          `json:"energy_profile"`
	MoodKeywords         []string           `json:"mood_keywords"`
}

// Client communicates with the Python analyzer service.
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// NewClient creates a new analyzer client.
func NewClient(baseURL string) *Client {
	if baseURL == "" {
		baseURL = DefaultURL
	}
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		httpClient: &http.Client{
			Timeout: 120 * time.Second, // audio analysis can be slow
		},
	}
}

// Analyze sends a local file path to the analyzer service and returns features.
func (c *Client) Analyze(filePath string) (*AudioFeatures, error) {
	data := url.Values{}
	data.Set("path", filePath)

	resp, err := c.httpClient.PostForm(c.baseURL+"/analyze", data)
	if err != nil {
		return nil, fmt.Errorf("failed to call analyzer service: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("analyzer returned status %d: %s", resp.StatusCode, string(body))
	}

	var features AudioFeatures
	if err := json.Unmarshal(body, &features); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &features, nil
}

// HealthCheck verifies the analyzer service is running.
func (c *Client) HealthCheck() error {
	resp, err := c.httpClient.Get(c.baseURL + "/health")
	if err != nil {
		return fmt.Errorf("analyzer service unreachable: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("analyzer service unhealthy: status %d", resp.StatusCode)
	}
	return nil
}

// Separate sends a local file path to the analyzer service to separate vocals
// and transcribe lyrics. It returns the transcribed lyrics text.
func (c *Client) Separate(filePath string) (string, error) {
	data := url.Values{}
	data.Set("path", filePath)

	resp, err := c.httpClient.PostForm(c.baseURL+"/separate", data)
	if err != nil {
		return "", fmt.Errorf("failed to call analyzer service: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("analyzer returned status %d: %s", resp.StatusCode, string(body))
	}

	var result map[string]string
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	return result["lyrics"], nil
}
