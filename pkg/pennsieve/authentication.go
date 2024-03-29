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

// AWSEndpoints can be used to set custom endpoints for Cognito.
// The default value will provide the default endpoints.
// Modifications only affect subsequent calls to NewAuthenticationService and
// not any already existing instances of authenticationService.
var AWSEndpoints = AWSCognitoEndpoints{}

type AWSCognitoEndpoints struct {
	IdentityProviderEndpoint string
	IdentityEndpoint         string
}

// Reset resets the endpoints to the default values.
func (e *AWSCognitoEndpoints) Reset() {
	e.IdentityEndpoint = ""
	e.IdentityProviderEndpoint = ""
}

// IsEmpty returns true if no custom endpoints have been set.
func (e *AWSCognitoEndpoints) IsEmpty() bool {
	return e.IdentityProviderEndpoint == "" && e.IdentityEndpoint == ""
}

type AuthenticationService interface {
	getCognitoConfig() (*authentication.CognitoConfig, error)
	ReAuthenticate() (*APISession, error)
	Authenticate(apiKey string, apiSecret string) (*APISession, error)
	GetAWSCredsForUser() *IdentityTypes.Credentials
	SetBaseUrl(url string)
	SetClient(client *Client)
}

func NewAuthenticationService(client PennsieveHTTPClient, baseUrl string) *authenticationService {
	cfg := newAwsConfig()
	return &authenticationService{
		client:    client,
		BaseUrl:   baseUrl,
		awsConfig: cfg,
	}
}

func newAwsConfig() aws.Config {
	loadOptions := []func(*config.LoadOptions) error{config.WithRegion("us-east-1")}

	if !AWSEndpoints.IsEmpty() {
		endpointMap := map[string]string{
			cognitoidentityprovider.ServiceID: AWSEndpoints.IdentityProviderEndpoint,
			cognitoidentity.ServiceID:         AWSEndpoints.IdentityEndpoint,
		}
		endpointResolver := aws.EndpointResolverWithOptionsFunc(func(service string, region string, options ...interface{}) (aws.Endpoint, error) {
			if endpoint := endpointMap[service]; endpoint != "" {
				return aws.Endpoint{
					URL: endpoint,
				}, nil
			}
			// Returning EndpointNotFoundError will cause service to fallback to default
			return aws.Endpoint{}, &aws.EndpointNotFoundError{}
		})
		loadOptions = append(loadOptions, config.WithEndpointResolverWithOptions(endpointResolver))

	}

	cfg, err := config.LoadDefaultConfig(context.TODO(),
		loadOptions...,
	)
	if err != nil {
		log.Fatal(err)
	}
	return cfg
}

type authenticationService struct {
	client    PennsieveHTTPClient
	config    authentication.CognitoConfig
	BaseUrl   string // BaseUrl is exposed in Auth service as we need to update to check new auth when switching profiles
	awsConfig aws.Config
}

// getCognitoConfig returns cognito urls from cloud.
func (s *authenticationService) getCognitoConfig() (*authentication.CognitoConfig, error) {

	req, _ := http.NewRequest("GET",
		fmt.Sprintf("%s/authentication/cognito-config", s.BaseUrl), nil)

	res := authentication.CognitoConfig{}
	if err := s.client.sendUnauthenticatedRequest(req.Context(), req, &res); err != nil {
		return nil, err
	}

	s.config = res

	return &res, nil
}

// ReAuthenticate updates authentication JWT and stores in local DB.
func (s *authenticationService) ReAuthenticate() (*APISession, error) {

	// Assert that credentials exist
	var emptyParams APIParams
	if *s.client.GetAPIParams() == emptyParams {
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
			"USERNAME": s.client.GetAPIParams().ApiKey,
			"PASSWORD": s.client.GetAPIParams().ApiSecret,
		},
		ClientId: aws.String(s.config.TokenPool.AppClientID),
	}

	svc := cognitoidentityprovider.NewFromConfig(s.awsConfig)
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

	svc := cognitoidentityprovider.NewFromConfig(s.awsConfig)

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
			log.Panicln("OrgID is not an string in the claim")
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

	orgIdInt, err := strconv.Atoi(orgId)
	if err != nil {
		log.Panicf("OrgID: %q is not an int\n", orgIdInt)

	}
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
	creds := s.client.GetAPIParams()

	authResponse, err := s.Authenticate(creds.ApiKey, creds.ApiSecret)
	if err != nil {
		log.Println("Error authenticating: ", err)
	}

	poolId := s.config.IdentityPool.ID
	poolResource := fmt.Sprintf("cognito-idp.us-east-1.amazonaws.com/%s", s.config.TokenPool.ID)

	svc := cognitoidentity.NewFromConfig(s.awsConfig)

	// Get an identity from Cognito's identity pool using authResponse from userpool
	idRes, err := svc.GetId(context.Background(), &cognitoidentity.GetIdInput{
		IdentityPoolId: aws.String(poolId),
		Logins: map[string]string{
			poolResource: authResponse.IdToken,
		},
	})

	// Exchange identity token for Credentials from Cognito Identity Pool
	credRes, err := svc.GetCredentialsForIdentity(context.Background(), &cognitoidentity.GetCredentialsForIdentityInput{
		IdentityId: idRes.IdentityId,
		Logins: map[string]string{
			poolResource: authResponse.IdToken,
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

func (s *authenticationService) SetClient(client *Client) {
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
