package pennsieve

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/pennsieve/pennsieve-go/pkg/pennsieve/models/account"
)

type AccountService interface {
	GetPennsieveAccounts(ctx context.Context, accountType string) (*account.GetPennsieveAccountsResponse, error)
}

type accountService struct {
	Client  PennsieveHTTPClient
	BaseUrl string
}

func NewAccountService(client PennsieveHTTPClient, baseUrl string) *accountService {
	return &accountService{
		Client:  client,
		BaseUrl: baseUrl,
	}
}

func (a *accountService) GetPennsieveAccounts(ctx context.Context, accountType string) (*account.GetPennsieveAccountsResponse, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/pennsieve-accounts/%s", a.BaseUrl, accountType), nil)
	if err != nil {
		return nil, err
	}

	if ctx == nil {
		ctx = req.Context()
	}

	res := account.GetPennsieveAccountsResponse{}
	if err := a.Client.sendRequest(ctx, req, &res); err != nil {

		log.Println("AccountService: SendRequest Error in Get: ", err)
		return nil, err
	}

	return &res, nil
}
