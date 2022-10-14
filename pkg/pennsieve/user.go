package pennsieve

import (
	"context"
	"fmt"
	"github.com/pennsieve/pennsieve-go/pkg/pennsieve/models/user"
	"net/http"
)

type UserService interface {
	GetUser(ctx context.Context) (*user.User, error)
	SetBaseUrl(url string)
}

type userService struct {
	client  HTTPClient
	BaseUrl string
}

func NewUserService(client HTTPClient, baseUrl string) *userService {
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

func (s *userService) SetBaseUrl(url string) {
	s.BaseUrl = url
}
