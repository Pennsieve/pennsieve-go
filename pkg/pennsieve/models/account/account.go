package account

type GetPennsieveAccountsResponse struct {
	AccountId   string `json:"accountId"`
	AccountType string `json:"accountType"`
}

type CreateAccountResponse struct {
	Uuid string `json:"uuid"`
}
