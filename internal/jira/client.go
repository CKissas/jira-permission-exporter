package jira

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"jira-permission-exporter/internal/model"
)

type Client struct {
	BaseURL    string
	Email      string
	APIToken   string
	HTTPClient *http.Client
}

func NewClient(baseURL, email, apiToken string) *Client {
	return &Client{
		BaseURL:  baseURL,
		Email:    email,
		APIToken: apiToken,
		HTTPClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

func (c *Client) doRequest(path string) ([]byte, error) {
	fullURL := c.BaseURL + path

	req, err := http.NewRequest(http.MethodGet, fullURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.SetBasicAuth(c.Email, c.APIToken)
	req.Header.Set("Accept", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request jira: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("jira returned %s: %s", resp.Status, string(body))
	}

	return body, nil
}

func (c *Client) GetPermissionSchemes() ([]model.PermissionScheme, error) {
	body, err := c.doRequest("/rest/api/3/permissionscheme")
	if err != nil {
		return nil, err
	}

	var result model.PermissionSchemesResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parse permission schemes response: %w", err)
	}

	return result.PermissionSchemes, nil
}

func (c *Client) GetPermissionSchemesExpanded() ([]model.PermissionScheme, error) {
	body, err := c.doRequest("/rest/api/3/permissionscheme?expand=all")
	if err != nil {
		return nil, err
	}

	var result model.PermissionSchemesResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parse expanded permission schemes response: %w", err)
	}

	return result.PermissionSchemes, nil
}

func (c *Client) GetAllProjects() ([]model.ProjectSummary, error) {
	var allProjects []model.ProjectSummary
	startAt := 0
	maxResults := 50

	for {
		path := "/rest/api/3/project/search?startAt=" + strconv.Itoa(startAt) + "&maxResults=" + strconv.Itoa(maxResults)

		body, err := c.doRequest(path)
		if err != nil {
			return nil, err
		}

		var result model.ProjectSearchResponse
		if err := json.Unmarshal(body, &result); err != nil {
			return nil, fmt.Errorf("parse project search response: %w", err)
		}

		allProjects = append(allProjects, result.Values...)

		if result.IsLast || len(result.Values) == 0 {
			break
		}

		startAt += result.MaxResults
	}

	return allProjects, nil
}

func (c *Client) GetProjectPermissionScheme(projectKeyOrID string) (*model.ProjectPermissionScheme, error) {
	path := "/rest/api/3/project/" + url.PathEscape(projectKeyOrID) + "/permissionscheme"

	body, err := c.doRequest(path)
	if err != nil {
		return nil, err
	}

	var result model.ProjectPermissionScheme
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parse project permission scheme response for %s: %w", projectKeyOrID, err)
	}

	return &result, nil
}

func (c *Client) GetProjectRoles(projectKeyOrID string) (model.ProjectRolesMap, error) {
	path := "/rest/api/3/project/" + url.PathEscape(projectKeyOrID) + "/role"

	body, err := c.doRequest(path)
	if err != nil {
		return nil, err
	}

	var result model.ProjectRolesMap
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parse project roles response for %s: %w", projectKeyOrID, err)
	}

	return result, nil
}

func (c *Client) GetProjectRoleDetail(projectKeyOrID string, roleURL string) (*model.ProjectRoleDetail, error) {
	path := roleURL

	if strings.HasPrefix(roleURL, c.BaseURL) {
		path = strings.TrimPrefix(roleURL, c.BaseURL)
	}

	body, err := c.doRequest(path)
	if err != nil {
		return nil, err
	}

	var result model.ProjectRoleDetail
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parse project role detail response for %s: %w", projectKeyOrID, err)
	}

	return &result, nil
}
