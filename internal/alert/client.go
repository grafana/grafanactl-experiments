package alert

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/grafana/grafanactl/internal/config"
	"k8s.io/client-go/rest"
)

const basePath = "/api/prometheus/grafana/api/v1/rules"

// Client fetches alert rules and groups from the Prometheus-compatible API.
type Client struct {
	httpClient *http.Client
	host       string
}

// NewClient creates a new alert client.
func NewClient(cfg config.NamespacedRESTConfig) (*Client, error) {
	httpClient, err := rest.HTTPClientFor(&cfg.Config)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP client: %w", err)
	}
	return &Client{httpClient: httpClient, host: cfg.Host}, nil
}

// ListOptions configures filtering for List operations.
type ListOptions struct {
	RuleUID    string
	GroupName  string
	FolderUID  string
	GroupLimit int
}

// List returns rules matching the given options.
func (c *Client) List(ctx context.Context, opts ListOptions) (*RulesResponse, error) {
	params := url.Values{}
	if opts.RuleUID != "" {
		params.Set("rule_uid", opts.RuleUID)
	}
	if opts.GroupName != "" {
		params.Set("rule_group", opts.GroupName)
	}
	if opts.FolderUID != "" {
		params.Set("folder_uid", opts.FolderUID)
	}
	if opts.GroupLimit > 0 {
		params.Set("group_limit", strconv.Itoa(opts.GroupLimit))
	}

	path := basePath
	if len(params) > 0 {
		path += "?" + params.Encode()
	}

	return c.doRequest(ctx, path)
}

// GetRule returns a single rule by UID.
func (c *Client) GetRule(ctx context.Context, uid string) (*RuleStatus, error) {
	resp, err := c.List(ctx, ListOptions{RuleUID: uid})
	if err != nil {
		return nil, err
	}

	for _, group := range resp.Data.Groups {
		for i := range group.Rules {
			if group.Rules[i].UID == uid {
				return &group.Rules[i], nil
			}
		}
	}
	return nil, fmt.Errorf("rule %s not found", uid)
}

// ListGroups returns all groups.
func (c *Client) ListGroups(ctx context.Context) ([]RuleGroup, error) {
	resp, err := c.List(ctx, ListOptions{})
	if err != nil {
		return nil, err
	}
	return resp.Data.Groups, nil
}

// GetGroup returns a single group by name with all its rules.
func (c *Client) GetGroup(ctx context.Context, name string) (*RuleGroup, error) {
	resp, err := c.List(ctx, ListOptions{GroupName: name})
	if err != nil {
		return nil, err
	}

	for i := range resp.Data.Groups {
		if resp.Data.Groups[i].Name == name {
			return &resp.Data.Groups[i], nil
		}
	}
	return nil, fmt.Errorf("group %s not found", name)
}

func (c *Client) doRequest(ctx context.Context, path string) (*RulesResponse, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.host+path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request failed with status %d", resp.StatusCode)
	}

	var result RulesResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}
