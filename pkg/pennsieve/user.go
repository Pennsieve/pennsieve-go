package pennsieve

import (
	"context"
	"fmt"
	"github.com/pennsieve/pennsieve-go/pkg/pennsieve/models/user"
	"net/http"
)

type UserService interface {
	GetUser(ctx context.Context) (*user.User, error)
	SwitchOrganization(ctx context.Context, organizationId string) error
	SetBaseUrl(url string)
}

type userService struct {
	client  PennsieveHTTPClient
	BaseUrl string
}

func NewUserService(client PennsieveHTTPClient, baseUrl string) *userService {
	return &userService{
		client:  client,
		BaseUrl: baseUrl,
	}
}

func (s *userService) GetUser(ctx context.Context) (*user.User, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/user", s.BaseUrl), nil)
	if err != nil {
		return nil, err
	}

	if ctx == nil {
		ctx = req.Context()
	}

	res := user.User{}
	if err := s.client.sendRequest(ctx, req, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

// SwitchOrganization switches the user's active organization server-side.
func (s *userService) SwitchOrganization(ctx context.Context, organizationId string) error {
	req, err := http.NewRequest("PUT",
		fmt.Sprintf("%s/session/switch-organization?organization_id=%s", s.BaseUrl, organizationId), nil)
	if err != nil {
		return err
	}

	if ctx == nil {
		ctx = req.Context()
	}

	if err := s.client.sendRequest(ctx, req, nil); err != nil {
		return err
	}

	return nil
}

func (s *userService) SetBaseUrl(url string) {
	s.BaseUrl = url
}
