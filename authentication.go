package pennsieve

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	cognito "github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	"github.com/golang-jwt/jwt"
	"log"
	"math"
	"net/http"
	"strconv"
	"time"
)

type CognitoConfig struct {
	Region    string    `json:"region"`
	UserPool  UserPool  `json:"userPool"`
	TokenPool TokenPool `json:"tokenPool"`
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

type credentials struct {
	apiKey       string
	apiSecret    string
	Token        string
	Expiration   time.Time
	RefreshToken string
	IsRefreshed  bool
}

type AuthenticationService struct {
	client        *Client
	config        CognitoConfig
	cognitoClient *cognito.CognitoIdentityProvider
}

func (s *AuthenticationService) getCognitoConfig() (*CognitoConfig, error) {

	req, _ := http.NewRequest("GET",
		fmt.Sprintf("%s/authentication/cognito-config", s.client.BaseURL), nil)

	res := CognitoConfig{}
	if err := s.client.sendUnauthenticatedRequest(req.Context(), req, &res); err != nil {
		return nil, err
	}

	s.config = res
	return &res, nil
}

func (s *AuthenticationService) refreshToken() (*string, error) {

	s.getCognitoConfig()

	clientID := aws.String(s.config.TokenPool.AppClientID)
	refreshToken := aws.String(s.client.Credentials.RefreshToken)

	params := &cognito.InitiateAuthInput{
		AuthFlow: aws.String("REFRESH_TOKEN_AUTH"),
		AuthParameters: map[string]*string{
			"REFRESH_TOKEN": refreshToken,
		},
		ClientId: clientID,
	}

	sess, _ := session.NewSession(&aws.Config{
		Region: aws.String("us-east-1"),
	})
	svc := cognito.New(sess)

	authResponse, authError := svc.InitiateAuth(params)
	if authError != nil {
		fmt.Println("Error = ", authError)
		return nil, authError
	}

	// Parse JWT and extract token expiration
	accessToken := authResponse.AuthenticationResult.AccessToken
	idTokenJwt := authResponse.AuthenticationResult.IdToken
	token, err := jwt.Parse(*idTokenJwt, nil)
	if token == nil {
		return nil, err
	}
	claims, _ := token.Claims.(jwt.MapClaims)

	sessionConfig := credentials{
		apiKey:       s.client.Credentials.apiKey,
		apiSecret:    s.client.Credentials.apiSecret,
		Token:        *accessToken,
		RefreshToken: *refreshToken,
		Expiration:   getTokenExpFromClaim(claims),
		IsRefreshed:  true,
	}

	s.client.Credentials = sessionConfig

	return nil, nil
}

func (s *AuthenticationService) reAuthenticate() (*credentials, error) {

	// Assert that credentials exist
	var emptyCredentials credentials
	if s.client.Credentials == emptyCredentials {
		log.Panicln("Cannot call reAuthenticate without prior Credentials")
	}

	// Use API-key and Secret from credentials and get new token
	clientID := aws.String(s.config.TokenPool.AppClientID)
	params := &cognito.InitiateAuthInput{
		AuthFlow: aws.String("USER_PASSWORD_AUTH"),
		AuthParameters: map[string]*string{
			"USERNAME": &s.client.Credentials.apiKey,
			"PASSWORD": &s.client.Credentials.apiSecret,
		},
		ClientId: clientID,
	}

	sess, _ := session.NewSession(&aws.Config{
		Region: aws.String("us-east-1"),
	})
	svc := cognito.New(sess)

	authResponse, authError := svc.InitiateAuth(params)
	if authError != nil {

		fmt.Println("Error = ", authError)
		return nil, authError
	}

	accessToken := authResponse.AuthenticationResult.AccessToken
	idTokenJwt := authResponse.AuthenticationResult.IdToken
	refreshToken := authResponse.AuthenticationResult.RefreshToken

	// Parse JWT and extract Organization for current user
	token, err := jwt.Parse(*idTokenJwt, nil)
	if token == nil {
		return nil, err
	}
	claims, _ := token.Claims.(jwt.MapClaims)

	s.client.Credentials.Token = *accessToken
	s.client.Credentials.RefreshToken = *refreshToken
	s.client.Credentials.Expiration = getTokenExpFromClaim(claims)
	s.client.Credentials.IsRefreshed = true

	return &s.client.Credentials, nil

}

func (s *AuthenticationService) Authenticate(apiKey string, apiSecret string) (*credentials, error) {

	username := aws.String(apiKey)
	password := aws.String(apiSecret)

	// Get Cognito Configuration
	s.getCognitoConfig()

	clientID := aws.String(s.config.TokenPool.AppClientID)

	params := &cognito.InitiateAuthInput{
		AuthFlow: aws.String("USER_PASSWORD_AUTH"),
		AuthParameters: map[string]*string{
			"USERNAME": username,
			"PASSWORD": password,
		},
		ClientId: clientID,
	}

	sess, _ := session.NewSession(&aws.Config{
		Region: aws.String("us-east-1"),
	})
	svc := cognito.New(sess)

	authResponse, authError := svc.InitiateAuth(params)
	if authError != nil {

		fmt.Println("Error = ", authError)
		return nil, authError
	}

	accessToken := authResponse.AuthenticationResult.AccessToken
	idTokenJwt := authResponse.AuthenticationResult.IdToken
	refreshToken := authResponse.AuthenticationResult.RefreshToken

	// Parse JWT and extract Organization for current user
	token, err := jwt.Parse(*idTokenJwt, nil)
	if token == nil {
		return nil, err
	}
	claims, _ := token.Claims.(jwt.MapClaims)

	var organizationNodeId string
	var orgId string
	var ok bool

	if x, found := claims["custom:organization_node_id"]; found {
		if organizationNodeId, ok = x.(string); !ok {
			log.Panicln("Organization Node ID is not a string in the claim")
		}
	} else {
		fmt.Println(claims)
		log.Panicln("Claim does not contain an organization identifier")
	}

	if x, found := claims["custom:organization_id"]; found {
		if orgId, ok = x.(string); !ok {
			log.Panicln("OrgID is not an int in the claim")
		}
	} else {
		log.Panicln("Claim does not contain an OrgID")
	}

	orgIdInt, _ := strconv.Atoi(orgId)
	creds := credentials{
		apiKey:       apiKey,
		apiSecret:    apiSecret,
		Token:        *accessToken,
		RefreshToken: *refreshToken,
		Expiration:   getTokenExpFromClaim(claims),
		IsRefreshed:  false,
	}

	s.client.OrganizationNodeId = organizationNodeId
	s.client.Credentials = creds
	s.client.OrganizationId = orgIdInt

	return &creds, nil
}

// getTokenExpFromClaim grabs the token expiration timestamp from claims
func getTokenExpFromClaim(claims jwt.MapClaims) time.Time {
	var tokenExp float64
	var ok bool
	if x, found := claims["exp"]; found {
		if tokenExp, ok = x.(float64); !ok {
			log.Panicln("Token Expire time is not a long in the claim")
		}
	} else {
		log.Panicln("Claim does not contain a Token Expire time")
	}

	integ, decim := math.Modf(tokenExp)
	sessionTokenExpiration := time.Unix(int64(integ), int64(decim*(1e9)))

	return sessionTokenExpiration
}
