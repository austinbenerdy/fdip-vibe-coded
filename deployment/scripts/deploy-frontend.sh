#!/bin/bash

# FDIP Frontend Deployment Script
# This script deploys the React frontend to Azure Static Web Apps

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
    
    if ! command -v node &> /dev/null; then
        print_error "Node.js is not installed. Please install it first."
        exit 1
    fi
    
    if ! command -v npm &> /dev/null; then
        print_error "npm is not installed. Please install it first."
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
    
    # Get the Static Web App name from Terraform output
    STATIC_WEB_APP_NAME=$(cd ../terraform && terraform output -raw static_web_app_name)
    RESOURCE_GROUP=$(cd ../terraform && terraform output -raw resource_group_name)
    APP_SERVICE_URL=$(cd ../terraform && terraform output -raw app_service_url)
    
    print_status "Static Web App Name: $STATIC_WEB_APP_NAME"
    print_status "Resource Group: $RESOURCE_GROUP"
    print_status "Backend API URL: $APP_SERVICE_URL"
}

# Build the React application
build_frontend() {
    print_status "Building React application..."
    
    cd ../../frontend
    
    # Install dependencies
    print_status "Installing npm dependencies..."
    npm install
    
    # Create .env file with backend URL
    print_status "Creating environment configuration..."
    cat > .env << EOF
REACT_APP_API_URL=$APP_SERVICE_URL
REACT_APP_STRIPE_PUBLISHABLE_KEY=$STRIPE_PUBLISHABLE_KEY
EOF
    
    # Build the application
    print_status "Building production build..."
    npm run build
    
    print_status "Frontend build completed successfully."
}

# Deploy to Azure Static Web Apps
deploy_to_static_web_apps() {
    print_status "Deploying to Azure Static Web Apps..."
    
    STATIC_WEB_APP_NAME=$(cd ../terraform && terraform output -raw static_web_app_name)
    RESOURCE_GROUP=$(cd ../terraform && terraform output -raw resource_group_name)
    
    # Get the deployment token
    DEPLOYMENT_TOKEN=$(az staticwebapp secrets list --name "$STATIC_WEB_APP_NAME" --resource-group "$RESOURCE_GROUP" --query properties.apiKey -o tsv)
    
    if [ -z "$DEPLOYMENT_TOKEN" ]; then
        print_error "Could not get deployment token. Please check if the Static Web App exists."
        exit 1
    fi
    
    # Deploy using Azure CLI
    print_status "Uploading build files..."
    az staticwebapp create \
        --name "$STATIC_WEB_APP_NAME" \
        --resource-group "$RESOURCE_GROUP" \
        --source ../../frontend/build \
        --api-key "$DEPLOYMENT_TOKEN" \
        --deployment-token "$DEPLOYMENT_TOKEN"
    
    print_status "Deployment to Static Web Apps completed."
}

# Configure Static Web Apps settings
configure_static_web_apps() {
    print_status "Configuring Static Web Apps settings..."
    
    STATIC_WEB_APP_NAME=$(cd ../terraform && terraform output -raw static_web_app_name)
    RESOURCE_GROUP=$(cd ../terraform && terraform output -raw resource_group_name)
    
    # Configure app settings
    az staticwebapp appsettings set \
        --name "$STATIC_WEB_APP_NAME" \
        --resource-group "$RESOURCE_GROUP" \
        --setting-names \
        REACT_APP_API_URL="$APP_SERVICE_URL" \
        REACT_APP_STRIPE_PUBLISHABLE_KEY="$STRIPE_PUBLISHABLE_KEY"
    
    print_status "Static Web Apps settings configured."
}

# Configure routing for SPA
configure_routing() {
    print_status "Configuring routing for Single Page Application..."
    
    STATIC_WEB_APP_NAME=$(cd ../terraform && terraform output -raw static_web_app_name)
    RESOURCE_GROUP=$(cd ../terraform && terraform output -raw resource_group_name)
    
    # Create routes configuration
    cat > routes.json << EOF
{
  "routes": [
    {
      "route": "/*",
      "serve": "/index.html",
      "statusCode": 200
    }
  ],
  "navigationFallback": {
    "rewrite": "/index.html"
  }
}
EOF
    
    # Apply routes configuration
    az staticwebapp routes set \
        --name "$STATIC_WEB_APP_NAME" \
        --resource-group "$RESOURCE_GROUP" \
        --routes routes.json
    
    # Clean up
    rm routes.json
    
    print_status "Routing configuration applied."
}

# Configure CORS for the backend
configure_cors() {
    print_status "Configuring CORS for backend..."
    
    APP_SERVICE_NAME=$(cd ../terraform && terraform output -raw app_service_name)
    RESOURCE_GROUP=$(cd ../terraform && terraform output -raw resource_group_name)
    STATIC_WEB_APP_URL=$(cd ../terraform && terraform output -raw static_web_app_url)
    
    # Update CORS settings in App Service
    az webapp cors add \
        --name "$APP_SERVICE_NAME" \
        --resource-group "$RESOURCE_GROUP" \
        --allowed-origins "$STATIC_WEB_APP_URL"
    
    print_status "CORS configuration updated."
}

# Check deployment status
check_deployment_status() {
    print_status "Checking deployment status..."
    
    STATIC_WEB_APP_NAME=$(cd ../terraform && terraform output -raw static_web_app_name)
    STATIC_WEB_APP_URL=$(cd ../terraform && terraform output -raw static_web_app_url)
    
    # Wait for the deployment to complete
    print_status "Waiting for deployment to complete..."
    sleep 30
    
    # Check if the app is responding
    if curl -f -s "$STATIC_WEB_APP_URL" > /dev/null; then
        print_status "✅ Frontend deployment successful!"
        print_status "Static Web App URL: $STATIC_WEB_APP_URL"
    else
        print_warning "⚠️  Static Web App might still be deploying. Please check the Azure portal."
        print_status "You can check the deployment status in the Azure portal."
    fi
}

# Prompt for Stripe publishable key
get_stripe_key() {
    print_status "Configuring Stripe integration..."
    
    read -p "Enter your Stripe Publishable Key (pk_test_... or pk_live_...): " STRIPE_PUBLISHABLE_KEY
    
    if [ -z "$STRIPE_PUBLISHABLE_KEY" ]; then
        print_warning "No Stripe key provided. Payment features will not work."
        STRIPE_PUBLISHABLE_KEY=""
    else
        print_status "Stripe key configured."
    fi
}

# Main deployment function
main() {
    print_status "Starting FDIP frontend deployment..."
    
    check_prerequisites
    check_azure_login
    get_terraform_outputs
    get_stripe_key
    build_frontend
    deploy_to_static_web_apps
    configure_static_web_apps
    configure_routing
    configure_cors
    check_deployment_status
    
    print_status "Frontend deployment completed!"
    
    # Display final URLs
    echo
    print_status "Deployment Summary:"
    print_status "Frontend URL: $(cd ../terraform && terraform output -raw static_web_app_url)"
    print_status "Backend API URL: $(cd ../terraform && terraform output -raw app_service_url)"
    echo
    print_status "Next steps:"
    print_status "1. Test the application by visiting the frontend URL"
    print_status "2. Configure custom domain if needed"
    print_status "3. Set up monitoring and alerts"
    print_status "4. Configure SSL certificates"
}

# Run the main function
main "$@" 