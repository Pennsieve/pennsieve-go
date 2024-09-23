module github.com/pennsieve/pennsieve-go

go 1.18

//replace github.com/pennsieve/pennsieve-go-api => ../pennsieve-go-api

require (
	github.com/aws/aws-sdk-go-v2 v1.17.8
	github.com/aws/aws-sdk-go-v2/config v1.18.14
	github.com/aws/aws-sdk-go-v2/service/cognitoidentity v1.14.0
	github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider v1.20.0
	github.com/golang-jwt/jwt v3.2.2+incompatible
	github.com/pennsieve/pennsieve-go-api v1.1.0
	github.com/pennsieve/pennsieve-go-core v1.11.0
	github.com/stretchr/testify v1.8.1
)

require (
	github.com/aws/aws-sdk-go-v2/credentials v1.13.14 // indirect
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.12.23 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.1.32 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.4.26 // indirect
	github.com/aws/aws-sdk-go-v2/internal/ini v1.3.30 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.9.23 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.12.3 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssooidc v1.14.3 // indirect
	github.com/aws/aws-sdk-go-v2/service/sts v1.18.4 // indirect
	github.com/aws/smithy-go v1.13.5 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/google/uuid v1.3.0 // indirect
	github.com/lib/pq v1.10.7 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
