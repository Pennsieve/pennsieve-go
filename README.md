# pennsieve-go

Go Client Library for the Pennsieve Platform

## Overview

`pennsieve-go` is the official Go SDK for interacting with the [Pennsieve Platform](https://www.pennsieve.io/), a cloud-based scientific data management platform. This library provides a comprehensive set of APIs to manage datasets, packages, files, organizations, users, and more.

## Features

- Authentication with Pennsieve API using API keys or Cognito
- Dataset management (create, read, update, delete)
- Package and file operations
- Organization and user management
- Time series data support
- Manifest operations for workspace management
- Discover API integration for public datasets
- Account management functionality

## Installation

```bash
go get github.com/pennsieve/pennsieve-go
```

## Requirements

- Go 1.23 or later
- Valid Pennsieve account and API credentials

## Quick Start

```go
package main

import (
    "fmt"
    "github.com/pennsieve/pennsieve-go/pkg/pennsieve"
)

func main() {
    // Initialize client with API credentials
    params := pennsieve.APIParams{
        ApiKey:    "your-api-key",
        ApiSecret: "your-api-secret",
    }
    
    client := pennsieve.NewClient(params)
    
    // Authenticate
    err := client.Authenticate()
    if err != nil {
        panic(err)
    }
    
    // Work with datasets, packages, etc.
    datasets, err := client.GetDatasets()
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("Found %d datasets\n", len(datasets))
}
```

## Configuration

The client can be configured using:

1. **Direct API Parameters**: Pass `APIParams` struct with your credentials
2. **Configuration File**: Use a local config file with profiles (set `UseConfigFile: true`)
3. **Environment Variables**: Set relevant environment variables for API endpoints

### API Endpoints

- **API v1**: `https://api.pennsieve.io` (default)
- **API v2**: `https://api2.pennsieve.io` (default)
- **Upload Bucket**: `pennsieve-prod-uploads-v2-use1` (default)

## Core Components

### Authentication
- Cognito-based authentication
- API key/secret authentication
- Token refresh support

### Data Management
- **Datasets**: Create, list, update, and delete datasets
- **Packages**: Manage data packages within datasets
- **Files**: Upload, download, and manage files
- **Manifests**: Handle workspace manifest operations

### Organization & Users
- Manage organizations
- User account operations
- Team collaboration features

### Specialized Features
- **Time Series**: Support for time series data
- **Discover API**: Access public datasets
- **Account Management**: User account operations

## Testing

Run the test suite:

```bash
go test ./...
```

Run tests with coverage:

```bash
go test -cover ./...
```

## Project Structure

```
pennsieve-go/
├── pkg/
│   └── pennsieve/
│       ├── models/           # Data models
│       │   ├── account/
│       │   ├── authentication/
│       │   ├── dataset/
│       │   ├── discover/
│       │   ├── organization/
│       │   ├── ps_package/
│       │   ├── timeseries/
│       │   └── user/
│       ├── account.go        # Account management
│       ├── authentication.go # Authentication logic
│       ├── client.go         # Main client implementation
│       ├── dataset.go        # Dataset operations
│       ├── discover.go       # Discover API
│       ├── manifest.go       # Manifest operations
│       ├── organization.go   # Organization management
│       ├── package.go        # Package operations
│       ├── timeseries.go     # Time series support
│       ├── user.go          # User management
│       └── *_test.go        # Test files
├── go.mod                   # Go module definition
├── go.sum                   # Dependency checksums
└── README.md               # This file
```

## Creating Releases with Semantic Versioning

This project follows [Semantic Versioning](https://semver.org/) (SemVer) for release management.

### Semantic Versioning Format

Releases use the format: `vMAJOR.MINOR.PATCH`

- **MAJOR** version: Incompatible API changes
- **MINOR** version: Backwards-compatible new functionality
- **PATCH** version: Backwards-compatible bug fixes

Examples: `v1.0.0`, `v1.2.3`, `v2.0.0`

### Creating a GitHub Release

1. **Ensure all changes are merged to main**:
   ```bash
   git checkout main
   git pull origin main
   ```

2. **Create and push a version tag**:
   ```bash
   # For a new patch release (bug fixes)
   git tag v1.0.1
   
   # For a new minor release (new features)
   git tag v1.1.0
   
   # For a new major release (breaking changes)
   git tag v2.0.0
   
   # Push the tag to GitHub
   git push origin v1.0.1
   ```

3. **Create the GitHub Release**:
   - Go to the repository's [Releases page](https://github.com/pennsieve/pennsieve-go/releases)
   - Click "Draft a new release"
   - Select the tag you just created
   - Set the release title (e.g., "v1.0.1 - Bug Fixes")
   - Add release notes describing:
     - New features (for minor releases)
     - Bug fixes
     - Breaking changes (for major releases)
     - Migration guides if applicable
   - Check "Set as the latest release" if appropriate
   - Click "Publish release"

### Release Guidelines

- **Pre-release versions**: Use suffixes like `-alpha`, `-beta`, `-rc.1` for pre-releases (e.g., `v1.0.0-beta.1`)
- **Breaking changes**: Always increment MAJOR version
- **New features**: Increment MINOR version, reset PATCH to 0
- **Bug fixes only**: Increment PATCH version
- **Documentation**: Update CHANGELOG.md (if present) with release notes
- **Testing**: Ensure all tests pass before creating a release

## Dependencies

Key dependencies include:
- AWS SDK for Go v2 (Cognito authentication)
- JWT for token handling
- Pennsieve Go Core libraries
- Testify for testing

See `go.mod` for the complete dependency list.

## License

This project is licensed under the terms specified by Pennsieve, Inc. See the LICENSE file for details.

## Support

- [Pennsieve Documentation](https://docs.pennsieve.io/)
- [API Reference](https://docs.pennsieve.io/reference)
- [GitHub Issues](https://github.com/pennsieve/pennsieve-go/issues)

## Contact

For questions and support, please contact the Pennsieve team or open an issue on GitHub.
