#!/bin/bash

# FDIP Backend Deployment Script
# This script deploys the Go backend to Azure App Service

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if required tools are installed
check_prerequisites() {
    print_status "Checking prerequisites..."
    
    if ! command -v az &> /dev/null; then
        print_error "Azure CLI is not installed. Please install it first."
        exit 1
    fi
    
    if ! command -v docker &> /dev/null; then
        print_error "Docker is not installed. Please install it first."
        exit 1
    fi
    
    if ! command -v go &> /dev/null; then
        print_error "Go is not installed. Please install it first."
        exit 1
    fi
    
    print_status "All prerequisites are installed."
}

# Check if user is logged into Azure
check_azure_login() {
    print_status "Checking Azure login status..."
    
    if ! az account show &> /dev/null; then
        print_error "Not logged into Azure. Please run 'az login' first."
        exit 1
    fi
    
    print_status "Logged into Azure as: $(az account show --query user.name -o tsv)"
}

# Get Terraform outputs
get_terraform_outputs() {
    print_status "Getting Terraform outputs..."
    
    if [ ! -f "../terraform/terraform.tfstate" ]; then
        print_error "Terraform state file not found. Please run 'terraform apply' first."
        exit 1
    fi
    
    # Source the outputs
    eval $(cd ../terraform && terraform output -json | jq -r 'to_entries | .[] | "export " + .key + "=\"" + .value + "\""')
    
    print_status "App Service Name: $app_service_name"
    print_status "Resource Group: $resource_group_name"
}

# Build the Docker image
build_docker_image() {
    print_status "Building Docker image..."
    
    cd ../../backend
    
    # Build the image
    docker build -t fdip-backend:latest .
    
    print_status "Docker image built successfully."
}

# Deploy to Azure App Service
deploy_to_app_service() {
    print_status "Deploying to Azure App Service..."
    
    # Get the App Service name from Terraform output
    APP_SERVICE_NAME=$(cd ../terraform && terraform output -raw app_service_name)
    RESOURCE_GROUP=$(cd ../terraform && terraform output -raw resource_group_name)
    
    # Deploy using Azure CLI
    az webapp config container set \
        --name "$APP_SERVICE_NAME" \
        --resource-group "$RESOURCE_GROUP" \
        --docker-custom-image-name fdip-backend:latest \
        --docker-registry-server-url https://index.docker.io
    
    print_status "Deployment configuration updated."
}

# Set environment variables
set_environment_variables() {
    print_status "Setting environment variables..."
    
    APP_SERVICE_NAME=$(cd ../terraform && terraform output -raw app_service_name)
    RESOURCE_GROUP=$(cd ../terraform && terraform output -raw resource_group_name)
    
    # Get database connection info
    DB_HOST=$(cd ../terraform && terraform output -raw database_server_fqdn)
    DB_NAME=$(cd ../terraform && terraform output -raw database_name)
    DB_USER=$(cd ../terraform && terraform output -raw backend_environment_variables | jq -r '.DB_USER')
    
    # Set environment variables
    az webapp config appsettings set \
        --name "$APP_SERVICE_NAME" \
        --resource-group "$RESOURCE_GROUP" \
        --settings \
        DB_HOST="$DB_HOST" \
        DB_PORT="3306" \
        DB_USER="$DB_USER" \
        DB_NAME="$DB_NAME" \
        PORT="8080" \
        ENV="production" \
        CORS_ORIGIN="https://$(cd ../terraform && terraform output -raw static_web_app_default_hostname)"
    
    print_status "Environment variables set."
}

# Configure secrets in Key Vault
configure_secrets() {
    print_status "Configuring secrets in Key Vault..."
    
    KEY_VAULT_NAME=$(cd ../terraform && terraform output -raw key_vault_name)
    RESOURCE_GROUP=$(cd ../terraform && terraform output -raw resource_group_name)
    
    # Prompt for secrets
    read -s -p "Enter JWT Secret: " JWT_SECRET
    echo
    read -s -p "Enter Database Password: " DB_PASSWORD
    echo
    read -s -p "Enter Stripe Secret Key: " STRIPE_SECRET_KEY
    echo
    read -s -p "Enter Stripe Webhook Secret: " STRIPE_WEBHOOK_SECRET
    echo
    
    # Store secrets in Key Vault
    az keyvault secret set --vault-name "$KEY_VAULT_NAME" --name "JWT-SECRET" --value "$JWT_SECRET"
    az keyvault secret set --vault-name "$KEY_VAULT_NAME" --name "DB-PASSWORD" --value "$DB_PASSWORD"
    az keyvault secret set --vault-name "$KEY_VAULT_NAME" --name "STRIPE-SECRET-KEY" --value "$STRIPE_SECRET_KEY"
    az keyvault secret set --vault-name "$KEY_VAULT_NAME" --name "STRIPE-WEBHOOK-SECRET" --value "$STRIPE_WEBHOOK_SECRET"
    
    print_status "Secrets stored in Key Vault."
}

# Configure App Service to use Key Vault
configure_key_vault_integration() {
    print_status "Configuring Key Vault integration..."
    
    APP_SERVICE_NAME=$(cd ../terraform && terraform output -raw app_service_name)
    RESOURCE_GROUP=$(cd ../terraform && terraform output -raw resource_group_name)
    KEY_VAULT_NAME=$(cd ../terraform && terraform output -raw key_vault_name)
    
    # Enable managed identity
    az webapp identity assign \
        --name "$APP_SERVICE_NAME" \
        --resource-group "$RESOURCE_GROUP"
    
    # Get the principal ID
    PRINCIPAL_ID=$(az webapp identity show --name "$APP_SERVICE_NAME" --resource-group "$RESOURCE_GROUP" --query principalId -o tsv)
    
    # Grant access to Key Vault
    az keyvault set-policy \
        --name "$KEY_VAULT_NAME" \
        --object-id "$PRINCIPAL_ID" \
        --secret-permissions get list
    
    # Set app settings to use Key Vault references
    az webapp config appsettings set \
        --name "$APP_SERVICE_NAME" \
        --resource-group "$RESOURCE_GROUP" \
        --settings \
        JWT_SECRET="@Microsoft.KeyVault(SecretUri=https://$KEY_VAULT_NAME.vault.azure.net/secrets/JWT-SECRET/)" \
        DB_PASSWORD="@Microsoft.KeyVault(SecretUri=https://$KEY_VAULT_NAME.vault.azure.net/secrets/DB-PASSWORD/)" \
        STRIPE_SECRET_KEY="@Microsoft.KeyVault(SecretUri=https://$KEY_VAULT_NAME.vault.azure.net/secrets/STRIPE-SECRET-KEY/)" \
        STRIPE_WEBHOOK_SECRET="@Microsoft.KeyVault(SecretUri=https://$KEY_VAULT_NAME.vault.azure.net/secrets/STRIPE-WEBHOOK-SECRET/)"
    
    print_status "Key Vault integration configured."
}

# Restart the App Service
restart_app_service() {
    print_status "Restarting App Service..."
    
    APP_SERVICE_NAME=$(cd ../terraform && terraform output -raw app_service_name)
    RESOURCE_GROUP=$(cd ../terraform && terraform output -raw resource_group_name)
    
    az webapp restart --name "$APP_SERVICE_NAME" --resource-group "$RESOURCE_GROUP"
    
    print_status "App Service restarted."
}

# Check deployment status
check_deployment_status() {
    print_status "Checking deployment status..."
    
    APP_SERVICE_NAME=$(cd ../terraform && terraform output -raw app_service_name)
    RESOURCE_GROUP=$(cd ../terraform && terraform output -raw resource_group_name)
    
    # Wait for the app to be ready
    print_status "Waiting for App Service to be ready..."
    sleep 30
    
    # Check if the app is responding
    APP_URL="https://$APP_SERVICE_NAME.azurewebsites.net"
    
    if curl -f -s "$APP_URL/health" > /dev/null; then
        print_status "✅ Backend deployment successful!"
        print_status "App Service URL: $APP_URL"
    else
        print_warning "⚠️  App Service might still be starting up. Please check the logs."
        print_status "You can check the logs with: az webapp log tail --name $APP_SERVICE_NAME --resource-group $RESOURCE_GROUP"
    fi
}

# Main deployment function
main() {
    print_status "Starting FDIP backend deployment..."
    
    check_prerequisites
    check_azure_login
    get_terraform_outputs
    build_docker_image
    deploy_to_app_service
    set_environment_variables
    configure_secrets
    configure_key_vault_integration
    restart_app_service
    check_deployment_status
    
    print_status "Backend deployment completed!"
}

# Run the main function
main "$@" 