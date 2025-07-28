#!/bin/bash

# FDIP Platform - Master Deployment Script
# This script orchestrates the complete deployment process

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
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

print_header() {
    echo -e "${BLUE}================================${NC}"
    echo -e "${BLUE}$1${NC}"
    echo -e "${BLUE}================================${NC}"
}

# Function to check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Check prerequisites
check_prerequisites() {
    print_header "Checking Prerequisites"
    
    local missing_tools=()
    
    if ! command_exists az; then
        missing_tools+=("Azure CLI")
    fi
    
    if ! command_exists terraform; then
        missing_tools+=("Terraform")
    fi
    
    if ! command_exists docker; then
        missing_tools+=("Docker")
    fi
    
    if ! command_exists node; then
        missing_tools+=("Node.js")
    fi
    
    if ! command_exists npm; then
        missing_tools+=("npm")
    fi
    
    if ! command_exists go; then
        missing_tools+=("Go")
    fi
    
    if [ ${#missing_tools[@]} -ne 0 ]; then
        print_error "Missing required tools:"
        for tool in "${missing_tools[@]}"; do
            echo "  - $tool"
        done
        echo
        print_status "Please install the missing tools and try again."
        exit 1
    fi
    
    print_status "All prerequisites are installed."
}

# Check Azure login
check_azure_login() {
    print_header "Checking Azure Login"
    
    if ! az account show &> /dev/null; then
        print_error "Not logged into Azure. Please run 'az login' first."
        exit 1
    fi
    
    local account_name=$(az account show --query user.name -o tsv)
    print_status "Logged into Azure as: $account_name"
}

# Deploy infrastructure
deploy_infrastructure() {
    print_header "Deploying Infrastructure"
    
    cd terraform
    
    # Initialize Terraform
    print_status "Initializing Terraform..."
    terraform init
    
    # Plan the deployment
    print_status "Planning deployment..."
    terraform plan -out=tfplan
    
    # Ask for confirmation
    echo
    read -p "Do you want to proceed with the deployment? (y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        print_warning "Deployment cancelled."
        exit 0
    fi
    
    # Apply the deployment
    print_status "Applying Terraform configuration..."
    terraform apply tfplan
    
    # Clean up plan file
    rm tfplan
    
    cd ..
    
    print_status "Infrastructure deployment completed."
}

# Deploy backend
deploy_backend() {
    print_header "Deploying Backend"
    
    cd scripts
    
    # Make script executable
    chmod +x deploy-backend.sh
    
    # Run backend deployment
    ./deploy-backend.sh
    
    cd ..
    
    print_status "Backend deployment completed."
}

# Deploy frontend
deploy_frontend() {
    print_header "Deploying Frontend"
    
    cd scripts
    
    # Make script executable
    chmod +x deploy-frontend.sh
    
    # Run frontend deployment
    ./deploy-frontend.sh
    
    cd ..
    
    print_status "Frontend deployment completed."
}

# Display deployment summary
show_deployment_summary() {
    print_header "Deployment Summary"
    
    cd terraform
    
    # Get URLs from Terraform outputs
    local backend_url=$(terraform output -raw app_service_url 2>/dev/null || echo "Not available")
    local frontend_url=$(terraform output -raw static_web_app_url 2>/dev/null || echo "Not available")
    local resource_group=$(terraform output -raw resource_group_name 2>/dev/null || echo "Not available")
    
    cd ..
    
    echo
    print_status "✅ Deployment completed successfully!"
    echo
    echo "📋 Deployment Details:"
    echo "  Resource Group: $resource_group"
    echo "  Backend URL: $backend_url"
    echo "  Frontend URL: $frontend_url"
    echo
    echo "🔧 Next Steps:"
    echo "  1. Test the application by visiting the frontend URL"
    echo "  2. Configure custom domain (optional)"
    echo "  3. Set up monitoring and alerts"
    echo "  4. Configure SSL certificates"
    echo "  5. Set up backup procedures"
    echo
    echo "📚 Documentation:"
    echo "  - Azure Portal: https://portal.azure.com"
    echo "  - Application Insights: https://portal.azure.com/#view/Microsoft_Azure_Monitoring/AzureMonitoringBrowseBlade/~/overview"
    echo "  - Static Web Apps: https://portal.azure.com/#view/HubsExtension/BrowseResource/resourceType/Microsoft.Web%2FstaticSites"
    echo
    print_status "For troubleshooting, see: deployment/README.md"
}

# Show usage
show_usage() {
    echo "Usage: $0 [OPTIONS]"
    echo
    echo "Options:"
    echo "  --infrastructure-only    Deploy only infrastructure"
    echo "  --backend-only          Deploy only backend"
    echo "  --frontend-only         Deploy only frontend"
    echo "  --skip-checks           Skip prerequisite checks"
    echo "  --help                  Show this help message"
    echo
    echo "Examples:"
    echo "  $0                      # Full deployment"
    echo "  $0 --infrastructure-only # Deploy only infrastructure"
    echo "  $0 --backend-only       # Deploy only backend (requires infrastructure)"
}

# Parse command line arguments
INFRASTRUCTURE_ONLY=false
BACKEND_ONLY=false
FRONTEND_ONLY=false
SKIP_CHECKS=false

while [[ $# -gt 0 ]]; do
    case $1 in
        --infrastructure-only)
            INFRASTRUCTURE_ONLY=true
            shift
            ;;
        --backend-only)
            BACKEND_ONLY=true
            shift
            ;;
        --frontend-only)
            FRONTEND_ONLY=true
            shift
            ;;
        --skip-checks)
            SKIP_CHECKS=true
            shift
            ;;
        --help)
            show_usage
            exit 0
            ;;
        *)
            print_error "Unknown option: $1"
            show_usage
            exit 1
            ;;
    esac
done

# Main deployment function
main() {
    print_header "FDIP Platform Deployment"
    
    # Check prerequisites unless skipped
    if [ "$SKIP_CHECKS" = false ]; then
        check_prerequisites
        check_azure_login
    fi
    
    # Determine what to deploy
    if [ "$INFRASTRUCTURE_ONLY" = true ]; then
        deploy_infrastructure
    elif [ "$BACKEND_ONLY" = true ]; then
        deploy_backend
    elif [ "$FRONTEND_ONLY" = true ]; then
        deploy_frontend
    else
        # Full deployment
        deploy_infrastructure
        deploy_backend
        deploy_frontend
        show_deployment_summary
    fi
    
    print_status "Deployment process completed."
}

# Run the main function
main "$@" 