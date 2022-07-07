module github.com/pennsieve/pennsieve-go

go 1.18

//replace github.com/pennsieve/pennsieve-go-api => ../pennsieve-go-api

require (
	github.com/aws/aws-sdk-go v1.43.25
	github.com/golang-jwt/jwt v3.2.2+incompatible
	github.com/pennsieve/pennsieve-go-api v0.2.1
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
)
