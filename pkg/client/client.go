package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/shimshimney/pkg/api"
)

type Client struct {
	BaseURL string
	HTTP    *http.Client
}

func New(baseURL string) *Client {
	return &Client{
		BaseURL: strings.TrimRight(baseURL, "/"),
		HTTP: &http.Client{Timeout: 10 * time.Second},
	}
}

func (c *Client) doJSON(method, path string, payload any, out any) error {
	var body []byte
	if payload != nil {
		var err error
		body, err = json.Marshal(payload)
		if err != nil {
			return err
		}
	}
	url := c.BaseURL + path
	req, err := http.NewRequest(method, url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("request to %s failed with status %s", url, resp.Status)
	}
	if out == nil {
		return nil
	}
	return json.NewDecoder(resp.Body).Decode(out)
}

func (c *Client) Health() error {
	resp, err := c.HTTP.Get(c.BaseURL + "/health")
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("health endpoint returned %s", resp.Status)
	}
	return nil
}

func (c *Client) Register(req api.RegisterRequest) error {
	return c.doJSON(http.MethodPost, "/register", req, nil)
}

func (c *Client) Heartbeat(req api.HeartbeatRequest) error {
	return c.doJSON(http.MethodPost, "/heartbeat", req, nil)
}

func (c *Client) Rebuild() error {
	return c.doJSON(http.MethodPost, "/rebuild", nil, nil)
}

func (c *Client) ListPods() ([]api.PodStatus, error) {
	var result []api.PodStatus
	if err := c.doJSON(http.MethodGet, "/pods", nil, &result); err != nil {
		return nil, err
	}
	return result, nil
}
