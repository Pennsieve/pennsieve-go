package pennsieve

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentity"
	IdentityTypes "github.com/aws/aws-sdk-go-v2/service/cognitoidentity/types"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider/types"
	"github.com/golang-jwt/jwt"
	"github.com/pennsieve/pennsieve-go/pkg/pennsieve/models/authentication"
	"log"
	"math"
	"net/http"
	"strconv"
	"time"
)

type AuthenticationService interface {
	getCognitoConfig() (*authentication.CognitoConfig, error)
	ReAuthenticate() (*APISession, error)
	Authenticate(apiKey string, apiSecret string) (*APISession, error)
	GetAWSCredsForUser() *IdentityTypes.Credentials
	SetBaseUrl(url string)
	SetClient(client Client)
}

func NewAuthenticationService(client Client, baseUrl string) *authenticationService {
	return &authenticationService{
		client:  client,
		BaseUrl: baseUrl,
	}
}

type authenticationService struct {
	client  Client
	config  authentication.CognitoConfig
	BaseUrl string // BaseUrl is exposed in Auth service as we need to update to check new auth when switching profiles
}

// getCognitoConfig returns cognito urls from cloud.
func (s *authenticationService) getCognitoConfig() (*authentication.CognitoConfig, error) {

	req, _ := http.NewRequest("GET",
		fmt.Sprintf("%s/authentication/cognito-config", s.BaseUrl), nil)

	res := authentication.CognitoConfig{}
	if err := s.client.SendUnauthenticatedRequest(req.Context(), req, &res); err != nil {
		return nil, err
	}

	s.config = res
	return &res, nil
}

// ReAuthenticate updates authentication JWT and stores in local DB.
func (s *authenticationService) ReAuthenticate() (*APISession, error) {

	// Assert that credentials exist
	var emptyCredentials APICredentials
	if s.client.GetCredentials() == emptyCredentials {
		log.Panicln("Cannot call ReAuthenticate without prior Credentials")
	}

	// Stores the cognito config in the service object.
	_, err := s.getCognitoConfig()
	if err != nil {
		log.Println("Error getting config: ", err)
		return nil, err
	}

	params := &cognitoidentityprovider.InitiateAuthInput{
		AuthFlow: types.AuthFlowTypeUserPasswordAuth,
		AuthParameters: map[string]string{
			"USERNAME": s.client.GetCredentials().ApiKey,
			"PASSWORD": s.client.GetCredentials().ApiSecret,
		},
		ClientId: aws.String(s.config.TokenPool.AppClientID),
	}

	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion("us-east-1"),
	)
	if err != nil {
		log.Fatal(err)
	}

	svc := cognitoidentityprovider.NewFromConfig(cfg)
	authResponse, authError := svc.InitiateAuth(context.Background(), params)
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

	newSession := APISession{
		Token:        *accessToken,
		IdToken:      "",
		Expiration:   getTokenExpFromClaim(claims),
		RefreshToken: *refreshToken,
		IsRefreshed:  true,
	}

	s.client.SetSession(newSession)

	return &newSession, nil

}

func (s *authenticationService) Authenticate(apiKey string, apiSecret string) (*APISession, error) {

	// Get Cognito Configuration
	s.getCognitoConfig()

	clientID := aws.String(s.config.TokenPool.AppClientID)

	params := &cognitoidentityprovider.InitiateAuthInput{
		AuthFlow: types.AuthFlowTypeUserPasswordAuth,
		AuthParameters: map[string]string{
			"USERNAME": apiKey,
			"PASSWORD": apiSecret,
		},
		ClientId: clientID,
	}

	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion("us-east-1"),
	)
	if err != nil {
		log.Fatal(err)
	}

	svc := cognitoidentityprovider.NewFromConfig(cfg)

	authResponse, authError := svc.InitiateAuth(context.Background(), params)
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

	s.client.SetOrganization(orgIdInt, organizationNodeId)
	s.client.SetSession(creds)

	return &creds, nil
}

// GetAWSCredsForUser returns set of AWS credentials to allow user to upload data to upload bucket
func (s *authenticationService) GetAWSCredsForUser() *IdentityTypes.Credentials {

	// Authenticate with UserPool using API Key and Secret
	creds := s.client.GetAPICredentials()
	authReponse, _ := s.Authenticate(creds.ApiKey, creds.ApiSecret)

	poolId := s.config.IdentityPool.ID
	poolResource := fmt.Sprintf("cognito-idp.us-east-1.amazonaws.com/%s", s.config.TokenPool.ID)

	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion("us-east-1"),
	)
	if err != nil {
		log.Fatal(err)
	}

	svc := cognitoidentity.NewFromConfig(cfg)

	// Get an identity from Cognito's identity pool using authResponse from userpool
	idRes, err := svc.GetId(context.Background(), &cognitoidentity.GetIdInput{
		IdentityPoolId: aws.String(poolId),
		Logins: map[string]string{
			poolResource: authReponse.IdToken,
		},
	})

	// Exchange identity token for Credentials from Cognito Identity Pool
	credRes, err := svc.GetCredentialsForIdentity(context.Background(), &cognitoidentity.GetCredentialsForIdentityInput{
		IdentityId: idRes.IdentityId,
		Logins: map[string]string{
			poolResource: authReponse.IdToken,
		},
	})
	if err != nil {
		fmt.Println("Error getting cognito ID:", err)
	}

	return credRes.Credentials
}

func (s *authenticationService) SetBaseUrl(url string) {
	s.BaseUrl = url
}

func (s *authenticationService) SetClient(client Client) {
	s.client = client
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

	log.Println("Should be one hour:", sessionTokenExpiration.Unix()-time.Now().Unix())

	return sessionTokenExpiration
}

// AWSCredentialProviderWithExpiration provides AWS credentials
// This method is used by the upload service and is wrapped in Credentials Cache
type AWSCredentialProviderWithExpiration struct {
	AuthService AuthenticationService
}

func (p AWSCredentialProviderWithExpiration) Retrieve(ctx context.Context) (aws.Credentials, error) {

	log.Println("Retrieving new credentials from AWS Credentials Provider.")

	cognitoCredentials := p.AuthService.GetAWSCredsForUser()
	awsCredentials := aws.Credentials{
		AccessKeyID:     *cognitoCredentials.AccessKeyId,
		SecretAccessKey: *cognitoCredentials.SecretKey,
		SessionToken:    *cognitoCredentials.SessionToken,
		Source:          "AWS Cognito",
		CanExpire:       true,
		Expires:         *cognitoCredentials.Expiration,
	}

	return awsCredentials, nil
}
