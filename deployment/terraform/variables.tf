variable "resource_group_name" {
  description = "Name of the resource group"
  type        = string
  default     = "fdip-rg"
}

variable "location" {
  description = "Azure region for resources"
  type        = string
  default     = "East US"
}

variable "environment" {
  description = "Environment name (dev, staging, prod)"
  type        = string
  default     = "prod"
}

variable "project_name" {
  description = "Project name for resource naming"
  type        = string
  default     = "fdip"
}

# Database Configuration
variable "database_name" {
  description = "Name of the MariaDB database"
  type        = string
  default     = "fdip"
}

variable "database_admin_username" {
  description = "Database administrator username"
  type        = string
  default     = "fdip_admin"
}

variable "database_admin_password" {
  description = "Database administrator password"
  type        = string
  sensitive   = true
}

variable "database_sku_name" {
  description = "Database SKU name"
  type        = string
  default     = "B_Gen5_1"
}

variable "database_storage_mb" {
  description = "Database storage in MB"
  type        = number
  default     = 5120
}

# App Service Configuration
variable "app_service_plan_sku" {
  description = "App Service Plan SKU"
  type        = string
  default     = "B1"
}

variable "app_service_plan_size" {
  description = "App Service Plan size"
  type        = string
  default     = "B1"
}

# Static Web Apps Configuration
variable "static_web_app_sku" {
  description = "Static Web App SKU"
  type        = string
  default     = "Free"
}

# Key Vault Configuration
variable "key_vault_sku" {
  description = "Key Vault SKU"
  type        = string
  default     = "standard"
}

# Tags
variable "tags" {
  description = "Tags to apply to all resources"
  type        = map(string)
  default = {
    Environment = "production"
    Project     = "fdip"
    ManagedBy   = "terraform"
  }
}

# Network Configuration
variable "vnet_address_space" {
  description = "VNet address space"
  type        = list(string)
  default     = ["10.0.0.0/16"]
}

variable "subnet_address_prefixes" {
  description = "Subnet address prefixes"
  type        = list(string)
  default     = ["10.0.1.0/24", "10.0.2.0/24"]
}

# Application Configuration
variable "backend_port" {
  description = "Backend application port"
  type        = number
  default     = 8080
}

variable "cors_origins" {
  description = "CORS allowed origins"
  type        = list(string)
  default     = ["https://*.azurestaticapps.net"]
}

# Monitoring Configuration
variable "enable_application_insights" {
  description = "Enable Application Insights"
  type        = bool
  default     = true
}

variable "enable_front_door" {
  description = "Enable Azure Front Door"
  type        = bool
  default     = true
} 