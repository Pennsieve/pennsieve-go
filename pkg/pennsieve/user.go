package pennsieve

import (
	"context"
	"fmt"
	"github.com/pennsieve/pennsieve-go/pkg/pennsieve/models/user"
	"net/http"
)

type UserService interface {
	GetUser(ctx context.Context, options *user.UserOptions) (*user.User, error)
	SetBaseUrl(url string)
}

type userService struct {
	client  Client
	BaseUrl string
}

func NewUserService(client Client, baseUrl string) *userService {
	return &userService{
		client:  client,
		BaseUrl: baseUrl,
	}
}

func (c *userService) GetUser(ctx context.Context, options *user.UserOptions) (*user.User, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/user", c.BaseUrl), nil)
	if err != nil {
		return nil, err
	}

	if ctx == nil {
		ctx = req.Context()
	}

	res := user.User{}
	if err := c.client.SendRequest(ctx, req, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

func (s *userService) SetBaseUrl(url string) {
	s.BaseUrl = url
}
