package pennsieve

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cognitoidentity"
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
	refreshToken := aws.String(s.client.APISession.RefreshToken)

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

	sessionConfig := APISession{

		Token:        *accessToken,
		RefreshToken: *refreshToken,
		Expiration:   getTokenExpFromClaim(claims),
		IsRefreshed:  true,
	}

	apiConfig := APICredentials{
		ApiKey:    s.client.APICredentials.ApiKey,
		ApiSecret: s.client.APICredentials.ApiSecret}

	s.client.APISession = sessionConfig
	s.client.APICredentials = apiConfig

	return nil, nil
}

func (s *AuthenticationService) ReAuthenticate() (*APISession, error) {

	// Assert that credentials exist
	var emptyCredentials APICredentials
	if s.client.APICredentials == emptyCredentials {
		log.Panicln("Cannot call ReAuthenticate without prior Credentials")
	}

	s.getCognitoConfig()

	// Use API-key and Secret from credentials and get new token
	params := &cognito.InitiateAuthInput{
		AuthFlow: aws.String("USER_PASSWORD_AUTH"),
		AuthParameters: map[string]*string{
			"USERNAME": aws.String(s.client.APICredentials.ApiKey),
			"PASSWORD": aws.String(s.client.APICredentials.ApiSecret),
		},
		ClientId: aws.String(s.config.TokenPool.AppClientID),
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

	s.client.APISession.Token = *accessToken
	s.client.APISession.RefreshToken = *refreshToken
	s.client.APISession.Expiration = getTokenExpFromClaim(claims)
	s.client.APISession.IsRefreshed = true

	fmt.Println("Expiration in reauth", s.client.APISession.Expiration, s.client.APICredentials.ApiKey)
	fmt.Printf("\n%s\n", s.client.APISession.Token)

	return &s.client.APISession, nil

}

func (s *AuthenticationService) Authenticate(apiKey string, apiSecret string) (*APISession, error) {

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
	//var iss string

	if x, found := claims["custom:organization_node_id"]; found {
		if organizationNodeId, ok = x.(string); !ok {
			log.Panicln("Organization Node ID is not a string in the claim")
		}
	} else {
		log.Panicln("Claim does not contain an organization identifier")
	}

	if x, found := claims["custom:organization_id"]; found {
		if orgId, ok = x.(string); !ok {
			log.Panicln("OrgID is not an int in the claim")
		}
	} else {
		log.Panicln("Claim does not contain an OrgID")
	}

	//if x, found := claims["iss"]; found {
	//	if iss, ok = x.(string); !ok {
	//		log.Panicln("iss is not an int in the claim")
	//	}
	//} else {
	//	log.Panicln("Claim does not contain an iss")
	//}

	orgIdInt, _ := strconv.Atoi(orgId)
	creds := APISession{
		Token:        *accessToken,
		RefreshToken: *refreshToken,
		IdToken:      *idTokenJwt,
		Expiration:   getTokenExpFromClaim(claims),
		IsRefreshed:  false,
	}

	s.client.OrganizationNodeId = organizationNodeId
	s.client.APISession = creds
	s.client.OrganizationId = orgIdInt

	return &creds, nil
}

// GetAWSCredsForUser returns set of AWS credentials to allow user to upload data to upload bucket
func (s *AuthenticationService) GetAWSCredsForUser() *cognitoidentity.Credentials {

	//TODO: Return IdentityPoolId in getCognitoConfig

	// Authenticate with UserPool using API Key and Secret
	authReponse, _ := s.client.Authentication.Authenticate(s.client.APICredentials.ApiKey, s.client.APICredentials.ApiSecret)

	// Create new AWS session
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("us-east-1"),
	})
	if err != nil {
		fmt.Println("Error creating session", err)
	}

	// Get an identity from Cognito's identity pool using authResponse from userpool
	svc := cognitoidentity.New(sess)
	idRes, err := svc.GetId(&cognitoidentity.GetIdInput{
		IdentityPoolId: aws.String("us-east-1:a980feea-d99b-437d-98b8-dc23b3bb972c"),
		Logins: map[string]*string{
			"cognito-idp.us-east-1.amazonaws.com/us-east-1_uCQXlh5nG": aws.String(authReponse.IdToken),
		},
	})
	if err != nil {
		fmt.Println("Error getting cognito ID:", err)
	}

	// Exchange identity token for Credentials from Cognito Identity Pool
	credRes, err := svc.GetCredentialsForIdentity(&cognitoidentity.GetCredentialsForIdentityInput{
		IdentityId: idRes.IdentityId,
		Logins: map[string]*string{
			"cognito-idp.us-east-1.amazonaws.com/us-east-1_uCQXlh5nG": aws.String(authReponse.IdToken),
		},
	})
	if err != nil {
		fmt.Println(err)
	}

	// Update Credentials in the Pennsieve Client
	s.client.AWSCredentials = credRes.Credentials

	return s.client.AWSCredentials
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
