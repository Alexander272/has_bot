package homeassistant

import (
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/goccy/go-json"
)

type Config struct {
	Url   string
	Token string
}

type Client struct {
	url   string
	token string
	http  *http.Client
}

type EntityState struct {
	EntityID          string `json:"entity_id"`
	State            string `json:"state"`
	UnitOfMeasurement string `json:"unit_of_measurement"`
}

func NewClient(conf Config) *Client {
	return &Client{
		url:   conf.Url,
		token: conf.Token,
		http: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c *Client) GetState(entityID string) (*EntityState, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/states/%s", c.url, entityID), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
	}

	var state EntityState
	if err := json.Unmarshal(body, &state); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	state.EntityID = entityID

	return &state, nil
}