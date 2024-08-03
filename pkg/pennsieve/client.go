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

type APISession struct {
	Token        string
	IdToken      string
	Expiration   time.Time
	RefreshToken string
	IsRefreshed  bool
}

type APIParams struct {
	ApiKey        string
	ApiSecret     string
	Port          string
	ApiHost       string
	ApiHost2      string
	UploadBucket  string
	UseConfigFile bool
	Profile       string
}

type errorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type Client struct {
	APISession APISession
	aPIParams  APIParams
	HTTPClient *http.Client

	OrganizationNodeId string
	OrganizationId     int

	Organization   OrganizationService
	Authentication AuthenticationService
	User           UserService
	Dataset        DatasetService
	Manifest       ManifestService
	Discover       DiscoverService
	Account        AccountService
	Package        PackageService
}

// NewClient creates a new Pennsieve HTTP client.
func NewClient(params APIParams) *Client {

	c := &Client{
		APISession:         APISession{},
		aPIParams:          params,
		HTTPClient:         &http.Client{Timeout: time.Minute},
		OrganizationNodeId: "",
		OrganizationId:     0,
	}

	c.Authentication = NewAuthenticationService(c, params.ApiHost)
	c.Organization = NewOrganizationService(c, params.ApiHost)
	c.User = NewUserService(c, params.ApiHost)
	c.Dataset = NewDatasetService(c, params.ApiHost)
	c.Discover = NewDiscoverService(c, params.ApiHost)
	c.Manifest = NewManifestService(c, params.ApiHost2)
	c.Account = NewAccountService(c, params.ApiHost2)
	c.Package = NewPackageService(c, params.ApiHost, params.ApiHost2)

	c.Authentication.getCognitoConfig()

	return c
}

type PennsieveHTTPClient interface {
	sendUnauthenticatedRequest(ctx context.Context, req *http.Request, v interface{}) error
	sendRequest(ctx context.Context, req *http.Request, v interface{}) error
	GetAPIParams() *APIParams
	SetSession(s APISession)
	SetOrganization(orgId int, orgNodeId string)
	Updateparams(params APIParams)
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
func (c *Client) sendRequest(ctx context.Context, req *http.Request, v interface{}) error {

	// Check Expiration Time for current session and refresh if necessary
	if time.Now().After(c.APISession.Expiration.Add(-5 * time.Minute)) {
		log.Println("Refreshing token")

		// We are using reAuthenticate instead of refresh pathway as eventually, the refresh-token
		// also expires and there is no real reason why we don't just re-authenticate.`
		_, err := c.Authentication.Authenticate(c.aPIParams.ApiKey, c.aPIParams.ApiSecret)

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
	if res.StatusCode != http.StatusOK && res.StatusCode != http.StatusCreated {
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

func (c *Client) SetSession(s APISession) {
	c.APISession = s
}

func (c *Client) GetAPIParams() *APIParams {
	return &c.aPIParams
}

func (c *Client) SetOrganization(orgId int, orgNodeId string) {
	c.OrganizationId = orgId
	c.OrganizationNodeId = orgNodeId
}

func (c *Client) Updateparams(params APIParams) {
	c.aPIParams = params

	c.Organization.SetBaseUrl(params.ApiHost)
	c.Authentication.SetBaseUrl(params.ApiHost)
	c.User.SetBaseUrl(params.ApiHost)
	c.Dataset.SetBaseUrl(params.ApiHost)
	c.Manifest.SetBaseUrl(params.ApiHost2)
	c.Account.SetBaseUrl(params.ApiHost2)
	c.Package.SetBaseUrl(params.ApiHost, params.ApiHost2)

}
