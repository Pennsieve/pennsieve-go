package models

import "time"

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

type APISession struct {
	Token        string
	IdToken      string
	Expiration   time.Time
	RefreshToken string
	IsRefreshed  bool
}

type APICredentials struct {
	ApiKey    string
	ApiSecret string
}
