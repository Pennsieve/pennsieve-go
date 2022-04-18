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

type Credentials struct {
	apiKey       string
	apiSecret    string
	Token        string
	IdToken      string
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

	fmt.Println(s.client.BaseURL)

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

	sessionConfig := Credentials{
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

func (s *AuthenticationService) reAuthenticate() (*Credentials, error) {

	// Assert that credentials exist
	var emptyCredentials Credentials
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

func (s *AuthenticationService) Authenticate(apiKey string, apiSecret string) (*Credentials, error) {

	username := aws.String(apiKey)
	password := aws.String(apiSecret)

	// Get Cognito Configuration
	fmt.Println("AUTHSERVICE:", s.client.BaseURL)
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
	var iss string

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

	if x, found := claims["iss"]; found {
		if iss, ok = x.(string); !ok {
			log.Panicln("iss is not an int in the claim")
		}
	} else {
		log.Panicln("Claim does not contain an iss")
	}

	orgIdInt, _ := strconv.Atoi(orgId)
	creds := Credentials{
		apiKey:       apiKey,
		apiSecret:    apiSecret,
		Token:        *accessToken,
		RefreshToken: *refreshToken,
		IdToken:      *idTokenJwt,
		Expiration:   getTokenExpFromClaim(claims),
		IsRefreshed:  false,
	}

	s.client.OrganizationNodeId = organizationNodeId
	s.client.Credentials = creds
	s.client.OrganizationId = orgIdInt

	// COGNITO IDENTITY CODE
	fmt.Println(iss)

	fmt.Println("AWS CREDS: ", s.client.AWSCredentials)

	return &creds, nil
}

// GetAWSCredsForUser returns set of AWS credentials to allow user to upload data to upload bucket
func (s *AuthenticationService) GetAWSCredsForUser() *cognitoidentity.Credentials {

	//TODO: Return IdentityPoolId in getCognitoConfig

	// Authenticate with UserPool using API Key and Secret
	authReponse, _ := s.client.Authentication.Authenticate(s.client.Credentials.apiKey, s.client.Credentials.apiSecret)

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
