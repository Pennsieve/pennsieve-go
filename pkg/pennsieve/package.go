package pennsieve

import (
	"context"
	"fmt"
	"github.com/pennsieve/pennsieve-go/pkg/pennsieve/models/ps_package"
	"log"
	"net/http"
)

type PackageService interface {
	GetPresignedUrl(ctx context.Context, packageId string, short bool) (*ps_package.GetPresignedUrlResponse, error)
	GetPackageSources(ctx context.Context, packageId string) (*ps_package.GetPackageSourcesResponse, error)
	SetBaseUrl(url string, url2 string)
}

type packageService struct {
	client   PennsieveHTTPClient
	baseUrl  string
	baseUrl2 string
}

func NewPackageService(client PennsieveHTTPClient, baseUrl string, baseUrl2 string) *packageService {
	return &packageService{
		client:   client,
		baseUrl:  baseUrl,
		baseUrl2: baseUrl2,
	}

}

func (p *packageService) SetBaseUrl(url string, url2 string) {
	p.baseUrl = url
	p.baseUrl2 = url2
}

// GetPackageSources returns an array of package resource files.
func (p *packageService) GetPackageSources(ctx context.Context, packageId string) (*ps_package.GetPackageSourcesResponse, error) {

	req, err := http.NewRequest("GET", fmt.Sprintf("%s/packages/%s/sources-paged", p.baseUrl, packageId), nil)
	if err != nil {
		return nil, err
	}

	if ctx == nil {
		ctx = req.Context()
	}

	fmt.Println(req.URL.String())
	res := ps_package.GetPackageSourcesResponse{}
	if err := p.client.sendRequest(ctx, req, &res); err != nil {
		log.Println("Error getting sources for package:", err)
		return nil, err
	}

	return &res, nil
}

// GetPresignedUrl returns a pre-signed URL for a file in a package.
func (p *packageService) GetPresignedUrl(ctx context.Context, packageId string, short bool) (*ps_package.GetPresignedUrlResponse, error) {

	// --- Get Sources for Package ---
	sources, err := p.GetPackageSources(ctx, packageId)
	if err != nil {
		log.Println("Error getting presigned url for package:", err)
		return nil, err
	}

	// --- Get Presigned URL for sources ---
	var resultFileInfo []ps_package.PresignedFileInfo
	for i, _ := range sources.Results {
		fileId := sources.Results[i].Content.ID
		req, err := http.NewRequest("GET",
			fmt.Sprintf("%s/packages/%s/files/%d?short=%t", p.baseUrl, packageId, fileId, short), nil)
		if err != nil {
			return nil, err
		}

		res := ps_package.GetPresignedUrlDTO{}
		if err := p.client.sendRequest(ctx, req, &res); err != nil {
			return nil, err
		}

		fileInfo := ps_package.PresignedFileInfo{
			URL:  res.URL,
			Name: sources.Results[i].Content.Filename,
		}

		resultFileInfo = append(resultFileInfo, fileInfo)
	}

	return &ps_package.GetPresignedUrlResponse{
		Files: resultFileInfo,
	}, nil

}
