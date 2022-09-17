package authentication

type CognitoConfig struct {
	Region       string       `json:"region"`
	UserPool     UserPool     `json:"userPool"`
	TokenPool    TokenPool    `json:"tokenPool"`
	IdentityPool IdentityPool `json:"identityPool"`
}
type UserPool struct {
	Region      string `json:"region"`
	ID          string `json:"id"`
	AppClientID string `json:"appClientId"`
}
type TokenPool struct {
	Region      string `json:"region"`
	ID          string `json:"id"`
	AppClientID string `json:"appClientId"`
}

type IdentityPool struct {
	Region string `json:"region"`
	ID     string `json:"id"`
}
