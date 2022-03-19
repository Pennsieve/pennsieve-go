package pennsieve_go

import (
	"fmt"
	"github.com/spf13/viper"
	"net/http"
)

const (
	BaseURLV1 = "https://api.pennsieve.io"
)

type Pennsieve struct {
	BaseURL    string
	profile    string
	apiKey     string
	apiToken   string
	httpClient *http.Client

	Organization *OrganizationService
}

func NewClient(httpClient *http.Client) *Pennsieve {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	viper.SetConfigName("config")
	viper.SetConfigType("ini")
	viper.AddConfigPath("$HOME/.pennsieve")

	err := viper.ReadInConfig() // Find and read the config file
	if err != nil {             // Handle errors reading the config file
		panic(fmt.Errorf("Fatal error config file: %w \n", err))
	}

	c := &Pennsieve{httpClient: httpClient}
	c.Organization = &OrganizationService{client: c}

	fmt.Println(viper.AllKeys())

	return c
}
