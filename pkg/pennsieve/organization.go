package pennsieve

import (
	"context"
	"fmt"
	"github.com/pennsieve/pennsieve-go/pkg/pennsieve/models/organization"
	"net/http"
)

type OrganizationService interface {
	List(ctx context.Context) (*organization.GetOrganizationsResponse, error)
	Get(ctx context.Context, id string) (*organization.GetOrganizationResponse, error)
	SetBaseUrl(url string)
}

type organizationService struct {
	client  HTTPClient
	baseUrl string
}

func NewOrganizationService(client HTTPClient, baseUrl string) *organizationService {
	return &organizationService{
		client:  client,
		baseUrl: baseUrl,
	}
}

// List lists all the organizations that the user belongs to.
func (o *organizationService) List(ctx context.Context) (*organization.GetOrganizationsResponse, error) {

	req, err := http.NewRequest("GET", fmt.Sprintf("%s/organizatinons", o.baseUrl), nil)
	if err != nil {
		return nil, err
	}

	res := organization.GetOrganizationsResponse{}
	if err := o.client.sendRequest(ctx, req, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

// Get returns a single organization by id.
func (o *organizationService) Get(ctx context.Context, id string) (*organization.GetOrganizationResponse, error) {

	req, err := http.NewRequest("GET", fmt.Sprintf("%s/organizations/%s", o.baseUrl, id), nil)
	if err != nil {
		return nil, err
	}

	if ctx == nil {
		ctx = req.Context()
	}

	res := organization.GetOrganizationResponse{}
	if err := o.client.sendRequest(ctx, req, &res); err != nil {

		fmt.Println("ERROR: ", err)
		return nil, err
	}

	return &res, nil
}

func (s *organizationService) SetBaseUrl(url string) {
	s.baseUrl = url
}
