package account

type GetPennsieveAccountsResponse struct {
	AccountId   string `json:"accountId"`
	AccountType string `json:"accountType"`
}

type CreateAccountResponse struct {
	Uuid string `json:"uuid"`
}

type AccountResponse struct {
	Uuid        string `json:"uuid"`
	AccountId   string `json:"accountId"`
	AccountType string `json:"accountType"`
	RoleName    string `json:"roleName"`
	ExternalId  string `json:"externalId"`
	UserId      string `json:"userId"`
	Status      string `json:"status"`
}

type DeleteAccountResponse struct {
	Uuid        string `json:"uuid"`
	AccountId   string `json:"accountId"`
	AccountType string `json:"accountType"`
	RoleName    string `json:"roleName"`
}
