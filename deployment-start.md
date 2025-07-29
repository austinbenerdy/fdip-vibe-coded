# FDIP Platform - Deployment Start Guide

This guide will walk you through deploying your FDIP platform to Azure step by step.

## 🎯 Prerequisites Checklist

Before you begin, make sure you have:

- [ ] **Azure Account** with billing enabled
- [ ] **GitHub Account** (for CI/CD)
- [ ] **Stripe Account** (for payment processing)
- [ ] **Domain Name** (optional but recommended)

## 📋 Step-by-Step Deployment Instructions

### **Step 1: Install Required Tools**

#### **1.1 Install Azure CLI**
```bash
# For Ubuntu/Debian
curl -sL https://aka.ms/InstallAzureCLIDeb | sudo bash

# For Fedora/RHEL/CentOS
sudo dnf install azure-cli

# For macOS
brew install azure-cli

# For Windows
# Download from: https://docs.microsoft.com/cli/azure/install-azure-cli-windows
```

#### **1.2 Install Terraform**
```bash
# For Ubuntu/Debian
curl -fsSL https://apt.releases.hashicorp.com/gpg | sudo apt-key add -
sudo apt-add-repository "deb [arch=amd64] https://apt.releases.hashicorp.com $(lsb_release -cs) main"
sudo apt-get update && sudo apt-get install terraform

# For Fedora/RHEL/CentOS
# Method 1: Using dnf (recommended)
sudo dnf install -y dnf-plugins-core
sudo dnf config-manager --add-repo https://rpm.releases.hashicorp.com/fedora/hashicorp.repo
sudo dnf install terraform

# Method 2: Direct download (if Method 1 fails)
# curl -O https://releases.hashicorp.com/terraform/1.5.0/terraform_1.5.0_linux_amd64.zip
# sudo unzip terraform_1.5.0_linux_amd64.zip -d /usr/local/bin/

# For macOS
brew install terraform

# For Windows
# Download from: https://www.terraform.io/downloads.html
```

#### **1.3 Install Docker**
```bash
# For Ubuntu/Debian
sudo apt-get update
sudo apt-get install docker.io
sudo usermod -aG docker $USER

# For Fedora/RHEL/CentOS
sudo dnf install -y docker
sudo systemctl enable docker
sudo systemctl start docker
sudo usermod -aG docker $USER

# For macOS
brew install --cask docker

# For Windows
# Download Docker Desktop from: https://www.docker.com/products/docker-desktop
```

#### **1.4 Install Node.js and npm**
```bash
# For Ubuntu/Debian
curl -fsSL https://deb.nodesource.com/setup_18.x | sudo -E bash -
sudo apt-get install -y nodejs

# For Fedora/RHEL/CentOS
curl -fsSL https://rpm.nodesource.com/setup_18.x | sudo bash -
sudo dnf install -y nodejs

# For macOS
brew install node

# For Windows
# Download from: https://nodejs.org/
```

#### **1.5 Install Go**
```bash
# For Ubuntu/Debian
sudo apt-get install golang-go

# For Fedora/RHEL/CentOS
sudo dnf install golang

# For macOS
brew install go

# For Windows
# Download from: https://golang.org/dl/
```

### **Step 2: Azure Account Setup**

#### **2.1 Login to Azure**
```bash
az login
```
This will open a browser window. Sign in with your Azure account.

#### **2.2 Set Subscription (if you have multiple)**
```bash
# List your subscriptions
az account list --output table

# Set the subscription you want to use
az account set --subscription "Your-Subscription-Name"
```

#### **2.3 Create Service Principal (for CI/CD)**
```bash
# Create service principal
az ad sp create-for-rbac --name "fdip-deployment" --role contributor \
    --scopes /subscriptions/$(az account show --query id -o tsv) \
    --sdk-auth
```

**Save the output** - you'll need it for GitHub Actions secrets.

### **Step 3: Prepare Your Application**

#### **3.1 Clone Your Repository**
```bash
git clone <your-repository-url>
cd fdip
```

#### **3.2 Test Your Application Locally**
```bash
# Test backend
cd backend
go mod download
go test ./...
go run main.go

# Test frontend (in another terminal)
cd frontend
npm install
npm start
```

### **Step 4: Configure Secrets**

#### **4.1 Generate JWT Secret**
```bash
# Generate a secure JWT secret
openssl rand -base64 32
```
**Save this value** - you'll need it later.

#### **4.2 Get Stripe Keys**
1. Go to [Stripe Dashboard](https://dashboard.stripe.com/)
2. Navigate to Developers → API Keys
3. Copy your **Publishable Key** and **Secret Key**
4. **Save both values** - you'll need them later.

#### **4.3 Generate Database Password**
```bash
# Generate a secure database password
openssl rand -base64 16
```
**Save this value** - you'll need it later.

### **Step 5: Deploy Infrastructure**

#### **5.1 Navigate to Deployment Directory**
```bash
cd deployment
```

#### **5.2 Initialize Terraform**
```bash
cd terraform
terraform init
```

#### **5.3 Create Terraform Variables File**
```bash
# Create terraform.tfvars file
cat > terraform.tfvars << EOF
database_admin_password = "YOUR_DATABASE_PASSWORD_HERE"
location = "East US"
environment = "prod"
EOF
```

**Replace `YOUR_DATABASE_PASSWORD_HERE`** with the password you generated in Step 4.3.

#### **5.4 Deploy Infrastructure**
```bash
# Plan the deployment
terraform plan

# Apply the deployment
terraform apply
```

**Wait for this to complete** - it may take 10-15 minutes.

#### **5.5 Save Output Values**
```bash
# Save important values
terraform output -json > outputs.json
```

### **Step 6: Deploy Backend**

#### **6.1 Navigate to Scripts Directory**
```bash
cd ../scripts
```

#### **6.2 Run Backend Deployment**
```bash
./deploy-backend.sh
```

**When prompted, enter:**
- JWT Secret (from Step 4.1)
- Database Password (from Step 4.3)
- Stripe Secret Key (from Step 4.2)
- Stripe Webhook Secret (from Stripe Dashboard)

### **Step 7: Deploy Frontend**

#### **7.1 Run Frontend Deployment**
```bash
./deploy-frontend.sh
```

**When prompted, enter:**
- Stripe Publishable Key (from Step 4.2)

### **Step 8: Verify Deployment**

#### **8.1 Check Backend Health**
```bash
# Get the backend URL from Terraform outputs
cd ../terraform
BACKEND_URL=$(terraform output -raw app_service_url)

# Test the health endpoint
curl -f "$BACKEND_URL/health"
```

#### **8.2 Check Frontend**
```bash
# Get the frontend URL from Terraform outputs
FRONTEND_URL=$(terraform output -raw static_web_app_url)

# Open in browser
echo "Frontend URL: $FRONTEND_URL"
```

### **Step 9: Set Up CI/CD (Optional)**

#### **9.1 Configure GitHub Secrets**
1. Go to your GitHub repository
2. Navigate to Settings → Secrets and variables → Actions
3. Add the following secrets:

```
AZURE_CREDENTIALS = [Service Principal JSON from Step 2.3]
AZURE_WEBAPP_NAME = [App Service name from terraform output]
AZURE_STATIC_WEB_APP_NAME = [Static Web App name from terraform output]
RESOURCE_GROUP = [Resource Group name from terraform output]
DATABASE_ADMIN_PASSWORD = [Your database password]
REACT_APP_API_URL = [Backend URL from terraform output]
REACT_APP_STRIPE_PUBLISHABLE_KEY = [Your Stripe publishable key]
AZURE_STATIC_WEB_APPS_API_TOKEN = [Get from Azure Portal]
```

#### **9.2 Push GitHub Actions Workflow**
```bash
# Copy the workflow file
cp deployment/github-actions/deploy.yml .github/workflows/
git add .github/workflows/deploy.yml
git commit -m "Add deployment workflow"
git push
```

### **Step 10: Post-Deployment Configuration**

#### **10.1 Configure Custom Domain (Optional)**
1. Go to Azure Portal
2. Navigate to your Static Web App
3. Go to Custom domains
4. Add your domain and configure DNS

#### **10.2 Set Up Monitoring**
1. Go to Azure Portal
2. Navigate to Application Insights
3. Set up alerts for:
   - High response times
   - Error rates
   - Availability

#### **10.3 Configure Backup**
1. Go to Azure Portal
2. Navigate to your MariaDB database
3. Configure backup retention and schedule

## 🔧 Troubleshooting

### **Common Issues**

#### **Issue: Terraform fails with permission errors**
**Solution:**
```bash
# Make sure you're logged in with the right account
az login
az account show
```

#### **Issue: Backend deployment fails**
**Solution:**
```bash
# Check App Service logs
az webapp log tail --name <app-service-name> --resource-group <resource-group>
```

#### **Issue: Frontend not loading**
**Solution:**
```bash
# Check Static Web Apps configuration
az staticwebapp show --name <static-web-app-name> --resource-group <resource-group>
```

#### **Issue: Database connection fails**
**Solution:**
1. Check if the database is running
2. Verify firewall rules
3. Check connection string

### **Useful Commands**

```bash
# Check all resources
az resource list --resource-group <resource-group-name>

# View App Service logs
az webapp log tail --name <app-service-name> --resource-group <resource-group>

# Check database status
az mariadb server show --name <db-server-name> --resource-group <resource-group>

# Get deployment status
az webapp deployment list --name <app-service-name> --resource-group <resource-group>
```

## 📞 Support

If you encounter issues:

1. **Check the logs** using the commands above
2. **Review the README** in the deployment directory
3. **Check Azure Portal** for resource status
4. **Verify your secrets** are correctly configured

## 🎉 Success!

Once you've completed all steps, your FDIP platform should be running on Azure with:

- ✅ **Backend API** running on Azure App Service
- ✅ **Frontend** hosted on Azure Static Web Apps
- ✅ **Database** running on Azure Database for MariaDB
- ✅ **Secrets** managed in Azure Key Vault
- ✅ **Monitoring** enabled with Application Insights
- ✅ **CDN** configured with Azure Front Door

**Total monthly cost: ~$33.55**

## 📚 Next Steps

1. **Test all features** of your application
2. **Set up monitoring alerts**
3. **Configure custom domain**
4. **Set up backup procedures**
5. **Plan for scaling** as your user base grows

---

**Need help?** Check the main `deployment/README.md` for detailed documentation. 