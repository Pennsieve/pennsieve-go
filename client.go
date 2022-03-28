package pennsieve

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"
)

const (
	BaseURLV1        = "https://api.pennsieve.io"
	DevelopmentURLV1 = "https://api.pennsieve.net"
)

type Client struct {
	BaseURL     string
	Credentials credentials
	HTTPClient  *http.Client

	OrganizationNodeId string
	OrganizationId     int

	Organization   *OrganizationService
	Authentication *AuthenticationService
	User           *UserService
}

// NewClient creates a new Pennsieve HTTP client.
func NewClient() *Client {

	c := &Client{
		BaseURL: BaseURLV1,
		HTTPClient: &http.Client{
			Timeout: time.Minute,
		},
	}
	c.Organization = &OrganizationService{client: c}
	c.Authentication = &AuthenticationService{client: c}
	c.User = &UserService{client: c}

	return c
}

type errorResponse struct {
	*http.Request
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// sendUnauthenticatedRequest sends a http request without authentication
func (c *Client) sendUnauthenticatedRequest(ctx context.Context, req *http.Request, v interface{}) error {
	req = req.WithContext(ctx)

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json; charset=utf-8")

	res, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}

	defer res.Body.Close()

	// Try to unmarshall into errorResponse
	if res.StatusCode != http.StatusOK {
		var errRes errorResponse
		if err = json.NewDecoder(res.Body).Decode(&errRes); err == nil {
			return errors.New(errRes.Message)
		}

		return fmt.Errorf("unknown error, status code: %d", res.StatusCode)
	}

	if err = json.NewDecoder(res.Body).Decode(&v); err != nil {
		return err
	}

	return nil
}

// sendRequest sends a http request with the appropriate Pennsieve headers and auth.
// The method checks if the token is valid and refreshes the token if not.
func (c *Client) sendRequest(ctx context.Context, req *http.Request, v interface{}) error {

	// Check Expiration Time for current session and refresh if necessary
	if time.Now().After(c.Credentials.Expiration.Add(-5 * time.Minute)) {
		fmt.Println("Refreshing token")

		// We are using reAuthenticate instead of refresh pathway as eventually, the refresh-token
		// also expires and there is no real reason why we don't just re-authenticate.
		c.Authentication.reAuthenticate()
	}

	req = req.WithContext(ctx)

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json; charset=utf-8")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.Credentials.Token))
	req.Header.Set("X-ORGANIZATION-ID", c.OrganizationNodeId)

	res, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}

	defer res.Body.Close()

	// Try to unmarshall into errorResponse
	if res.StatusCode != http.StatusOK {
		var errRes errorResponse
		if err = json.NewDecoder(res.Body).Decode(&errRes); err == nil {
			return errors.New(errRes.Message)
		}

		return fmt.Errorf("unknown error, status code: %d", res.StatusCode)
	}

	if err = json.NewDecoder(res.Body).Decode(&v); err != nil {
		return err
	}

	return nil
}
