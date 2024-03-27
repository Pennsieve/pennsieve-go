package pennsieve

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/pennsieve/pennsieve-go/pkg/pennsieve/models/account"
)

type AccountService interface {
	GetPennsieveAccounts(ctx context.Context, accountType string) (*account.GetPennsieveAccountsResponse, error)
	CreateAccount(ctx context.Context, accountId string, accountType string, roleName string, externalId string) (*account.CreateAccountResponse, error)
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

func (a *accountService) CreateAccount(ctx context.Context, accountId string, accountType string, roleName string, externalId string) (*account.CreateAccountResponse, error) {
	postParams := fmt.Sprintf(`
		{
			"accountId": "%s",
			"accountType": "%s",
			"roleName": "%s",
			"externalId": "%s"
		}`, accountId, accountType, roleName, externalId)

	postParamsPayload := bytes.NewReader([]byte(postParams))

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/accounts", a.BaseUrl), postParamsPayload)
	if err != nil {
		return nil, err
	}

	if ctx == nil {
		ctx = req.Context()
	}

	res := account.CreateAccountResponse{}
	if err := a.Client.sendRequest(ctx, req, &res); err != nil {
		log.Println("sendRequest Error: ", err)
		return nil, err
	}

	return &res, nil
}
