# unfold
![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/aryannr97/unfold)
[![Go Reference](https://pkg.go.dev/badge/github.com/aryannr97/unfold.svg)](https://pkg.go.dev/github.com/aryannr97/unfold)
[![Codecov](https://img.shields.io/codecov/c/github/aryannr97/unfold)](https://app.codecov.io/gh/aryannr97/unfold)
[![Go Report Card](https://goreportcard.com/badge/github.com/aryannr97/unfold)](https://goreportcard.com/report/github.com/aryannr97/unfold)
[![Linter](https://img.shields.io/badge/Linter-golangci--lint-informational)](https://golangci-lint.run)
[![MIT license](https://img.shields.io/github/license/aryannr97/unfold)](https://github.com/aryannr97/unfold/blob/main/LICENSE)

A powerful command-line utility for cloud resource management and data decoding operations. `unfold` simplifies common tasks across Azure Marketplace, Google Workspace, and provides handy utilities for JWT.

## Features

### Azure Management
- **Private Audience Management**: Add/remove tenants and subscriptions from Azure Marketplace private audiences
- **Tenant Discovery**: Retrieve tenant information by subscription ID
- **Job Status Tracking**: Monitor Azure job execution status
- **Audience Search**: Check if tenants/subscriptions exist in private audiences

### Google Workspace Management
- **Group Membership**: Add/remove users from Google groups
- **Membership Search**: Check if email addresses are members of specific groups
- **Role Information**: View user roles within groups

### Decoding Utilities
- **JWT Decoding**: Parse and display JWT token claims in JSON format

## Installation

### Prerequisites
- Go 1.24 or later
- Appropriate cloud credentials (Azure and/or Google)

### Build from Source
```bash
git clone https://github.com/aryannr97/unfold.git
cd unfold
make build
```

### Install
```bash
# Build and install to your PATH
go install ./...
```

### Development Setup
```bash
# Install dependencies
go mod download

# Run tests
make test

# Run tests with coverage
make test-coverage

# Run linter
make lint

# Clean build artifacts
make clean

# See all available commands
make help
```

## Usage

### Azure Commands

#### Get Operations
```bash
# Get tenant by subscription ID
unfold azure get -t <subscription-id>

# Get job status
unfold azure get -s <job-id>
```

#### Configure Operations
```bash
# Add tenant to private audience
unfold azure configure -tid <tenant-id> -o <offer-name>

# Add subscription to private audience
unfold azure configure -sid <subscription-id> -o <offer-name>

# Remove tenant from private audience
unfold azure configure -r -tid <tenant-id> -o <offer-name>

# Remove subscription from private audience
unfold azure configure -r -sid <subscription-id> -o <offer-name>
```

#### Search Operations
```bash
# Search for tenant/subscription in private audience
unfold azure search -id <tenant-or-subscription-id> -o <offer-name>
```

### Google Commands

#### Search Operations
```bash
# Check if email exists in Google group
unfold google search -id <email-address> -g <group-id>
```

#### Configure Operations
```bash
# Add user to Google group
unfold google configure -id <email-address> -g <group-id>

# Remove user from Google group
unfold google configure -r -id <email-address> -g <group-id>
```

### Utility Commands

#### JWT Decoding
```bash
# Decode JWT token and display claims
unfold jwt decode <jwt-token>
```

## Configuration

### Environment Variables

#### Azure Configuration
Set the following environment variables for Azure functionality:

```bash
# Azure authentication credentials
export AZURE_CLIENT_ID="your-azure-client-id"
export AZURE_CLIENT_SECRET="your-azure-client-secret"
export AZURE_TENANT_ID="your-azure-tenant-id"

# Azure Marketplace configuration
export AZURE_OFFERS_PUBLISHER="your-publisher-name"
export AZURE_OFFERS_FILE="/path/to/azure-offers.yaml"

# Azure certificate file for identity verification
export AZURE_CERT_FILE="/path/to/azure-cert.pem"
```

#### Google Configuration
Set the following environment variables for Google functionality:

```bash
# Google service account credentials
export GOOGLE_KEYFILE="/path/to/google-service-account.json"

# Google Workspace domain configuration
export GOOGLE_GCP_DOMAIN="@yourdomain.com"

# Optional: JWK URL for token validation
export GOOGLE_JWK_URL="https://your-jwk-endpoint.com"
```

#### HTTP Timeout
Every request to Azure and Google (including OAuth token requests) is bounded by a per-request timeout:

```bash
# Optional: Go duration, defaults to 30s
export UNFOLD_HTTP_TIMEOUT="45s"
```

### Configuration Files

#### Azure Offers File
Create a YAML file (referenced by `AZURE_OFFERS_FILE`) with your marketplace offers:

```yaml
# azure-offers.yaml
offer-name-1:
  productDurableID: "product-durable-id-1"
offer-name-2:
  productDurableID: "product-durable-id-2"
my-marketplace-offer:
  productDurableID: "12345678-1234-1234-1234-123456789abc"
```

**Fields:**
- `offer-name`: The name you'll use with the `-o` flag in commands
- `productDurableID`: The Azure Marketplace product durable ID for your offer

#### Google Service Account Key File
Create a Google Cloud service account and download the JSON key file:

```json
{
  "type": "service_account",
  "project_id": "your-project-id",
  "private_key_id": "key-id",
  "private_key": "-----BEGIN PRIVATE KEY-----\n...\n-----END PRIVATE KEY-----\n",
  "client_email": "service-account@your-project.iam.gserviceaccount.com",
  "client_id": "client-id",
  "auth_uri": "https://accounts.google.com/o/oauth2/auth",
  "token_uri": "https://oauth2.googleapis.com/token",
  "auth_provider_x509_cert_url": "https://www.googleapis.com/oauth2/v1/certs",
  "client_x509_cert_url": "https://www.googleapis.com/robot/v1/metadata/x509/service-account%40your-project.iam.gserviceaccount.com"
}
```

### Setup Instructions

#### Azure Setup
If you don't already have Azure credentials configured:
1. Ensure you have an Azure App Registration in Azure Active Directory
2. Obtain client credentials (client ID and secret) if not already available
3. Verify your app has appropriate permissions for Microsoft Management and Graph APIs:
   - `https://management.azure.com`
   - `https://graph.microsoft.com`
4. Create your marketplace offers configuration in the YAML file
5. Configure certificate-based authentication for enhanced security

#### Google Setup
If you need to set up Google Workspace integration:
1. Ensure you have access to a Google Cloud Project
2. Verify the Cloud Identity API is enabled
3. Obtain a service account with appropriate permissions:
   - `https://www.googleapis.com/auth/cloud-identity.groups`
4. Download the service account JSON key file if not already available
5. Configure the `GOOGLE_GCP_DOMAIN` to match your organization's domain

## Development

### CI/CD Pipeline
This project uses GitHub Actions for continuous integration and deployment:
- **Automated testing** on all pull requests
- **Code linting** with golangci-lint
- **Test coverage** reporting via Codecov
- **Multi-version Go support** (currently Go 1.24.x)

### Local Development
```bash
# Setup development environment
git clone https://github.com/aryannr97/unfold.git
cd unfold
go mod download

# Run the full development cycle
make lint    # Check code quality
make test    # Run tests
make build   # Build the application
```

### Testing
```bash
# Run all tests
make test

# Run tests with coverage report
make test-coverage

# View coverage in browser
go tool cover -html=coverage.out
```

### Available Make Targets
```bash
make help           # Show all available targets
make lint           # Run golangci-lint
make test           # Run tests
make test-coverage  # Run tests with coverage
make build          # Build the application
make clean          # Clean build artifacts
```

## Dependencies

Key dependencies include:
- `github.com/golang-jwt/jwt/v5` - JWT token handling
- `golang.org/x/oauth2` - OAuth2 authentication
- `google.golang.org/api` - Google APIs client
- `gopkg.in/yaml.v2` - YAML configuration parsing

## Contributing

We welcome contributions! Please follow these steps:

1. **Fork the repository**
2. **Create a feature branch** (`git checkout -b feature/amazing-feature`)
3. **Follow development best practices**:
   - Write tests for new functionality
   - Run `make lint` to check code quality
   - Run `make test` to ensure all tests pass
   - Keep commits focused and atomic
4. **Commit your changes** (`git commit -m 'Add some amazing feature'`)
5. **Push to the branch** (`git push origin feature/amazing-feature`)
6. **Open a Pull Request**

### Code Quality
- All PRs must pass CI checks (linting, testing, building)
- Maintain test coverage for new features
- Follow Go best practices and conventions

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Support

For support and questions, please open an issue in the GitHub repository.