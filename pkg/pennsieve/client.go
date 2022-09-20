package pennsieve

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"
)

const (
	BaseURLV1           = "https://api.pennsieve.io"
	BaseURLV2           = "https://api2.pennsieve.io"
	DefaultUploadBucket = "pennsieve-prod-uploads-v2-use1"
)

type Client struct {
	APISession     APISession
	APICredentials APICredentials
	HTTPClient     *http.Client

	OrganizationNodeId string
	OrganizationId     int
	UploadBucket       string

	Organization   *OrganizationService
	Authentication *AuthenticationService
	User           *UserService
	Dataset        *DatasetService
	Manifest       *ManifestService
}

type APISession struct {
	Token        string
	IdToken      string
	Expiration   time.Time
	RefreshToken string
	IsRefreshed  bool
}

type APICredentials struct {
	ApiKey    string
	ApiSecret string
}

type errorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// NewClient creates a new Pennsieve HTTP client.
func NewClient(baseUrlV1 string, baseUrlV2 string) *Client {

	c := &Client{
		HTTPClient: &http.Client{
			Timeout: time.Minute,
		},
	}
	c.Organization = &OrganizationService{client: c, baseUrl: baseUrlV1}
	c.Authentication = &AuthenticationService{client: c, BaseUrl: baseUrlV1}
	c.User = &UserService{client: c, BaseUrl: baseUrlV1}
	c.Dataset = &DatasetService{Client: c, BaseUrl: baseUrlV1}
	c.Manifest = &ManifestService{client: c, baseUrl: baseUrlV2}

	c.UploadBucket = DefaultUploadBucket

	return c
}

func (c *Client) SetBasePathForServices(baseUrlV1 string, baseUrlV2 string) {

	c.Organization.baseUrl = baseUrlV1
	c.Authentication.BaseUrl = baseUrlV1
	c.User.BaseUrl = baseUrlV1
	c.Dataset.BaseUrl = baseUrlV1
	c.Manifest.baseUrl = baseUrlV2
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

// SendRequest sends a http request with the appropriate Pennsieve headers and auth.
// The method checks if the token is valid and refreshes the token if not.
func (c *Client) SendRequest(ctx context.Context, req *http.Request, v interface{}) error {

	// Check Expiration Time for current session and refresh if necessary
	if time.Now().After(c.APISession.Expiration.Add(-5 * time.Minute)) {
		log.Println("Refreshing token")

		// We are using reAuthenticate instead of refresh pathway as eventually, the refresh-token
		// also expires and there is no real reason why we don't just re-authenticate.`
		_, err := c.Authentication.ReAuthenticate()

		if err != nil {
			log.Println("Error authenticating:", err)
			return err
		}
	}

	req = req.WithContext(ctx)

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json; charset=utf-8")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.APISession.Token))
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
