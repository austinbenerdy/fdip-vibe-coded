# Resource Group
output "resource_group_name" {
  description = "Name of the resource group"
  value       = azurerm_resource_group.main.name
}

output "resource_group_location" {
  description = "Location of the resource group"
  value       = azurerm_resource_group.main.location
}

# Virtual Network
output "virtual_network_name" {
  description = "Name of the virtual network"
  value       = azurerm_virtual_network.main.name
}

output "virtual_network_id" {
  description = "ID of the virtual network"
  value       = azurerm_virtual_network.main.id
}

# Database
output "database_server_name" {
  description = "Name of the MariaDB server"
  value       = azurerm_mariadb_server.main.name
}

output "database_server_fqdn" {
  description = "FQDN of the MariaDB server"
  value       = azurerm_mariadb_server.main.fqdn
}

output "database_name" {
  description = "Name of the database"
  value       = azurerm_mariadb_database.main.name
}

output "database_connection_string" {
  description = "Database connection string (without password)"
  value       = "Server=${azurerm_mariadb_server.main.fqdn};Database=${azurerm_mariadb_database.main.name};User Id=${var.database_admin_username};"
  sensitive   = true
}

# App Service
output "app_service_name" {
  description = "Name of the App Service"
  value       = azurerm_linux_web_app.backend.name
}

output "app_service_url" {
  description = "URL of the App Service"
  value       = "https://${azurerm_linux_web_app.backend.default_hostname}"
}

output "app_service_default_hostname" {
  description = "Default hostname of the App Service"
  value       = azurerm_linux_web_app.backend.default_hostname
}

# Static Web App
output "static_web_app_name" {
  description = "Name of the Static Web App"
  value       = azurerm_static_site.frontend.name
}

output "static_web_app_url" {
  description = "URL of the Static Web App"
  value       = "https://${azurerm_static_site.frontend.default_host_name}"
}

output "static_web_app_default_hostname" {
  description = "Default hostname of the Static Web App"
  value       = azurerm_static_site.frontend.default_host_name
}

# Key Vault
output "key_vault_name" {
  description = "Name of the Key Vault"
  value       = azurerm_key_vault.main.name
}

output "key_vault_uri" {
  description = "URI of the Key Vault"
  value       = azurerm_key_vault.main.vault_uri
}

# Application Insights
output "application_insights_name" {
  description = "Name of the Application Insights resource"
  value       = var.enable_application_insights ? azurerm_application_insights.main[0].name : null
}

output "application_insights_instrumentation_key" {
  description = "Instrumentation key for Application Insights"
  value       = var.enable_application_insights ? azurerm_application_insights.main[0].instrumentation_key : null
  sensitive   = true
}

output "application_insights_connection_string" {
  description = "Connection string for Application Insights"
  value       = var.enable_application_insights ? azurerm_application_insights.main[0].connection_string : null
  sensitive   = true
}

# Front Door
output "front_door_name" {
  description = "Name of the Front Door"
  value       = var.enable_front_door ? azurerm_frontdoor.main[0].name : null
}

output "front_door_url" {
  description = "URL of the Front Door"
  value       = var.enable_front_door ? "https://${azurerm_frontdoor.main[0].frontend_endpoint[0].host_name}" : null
}

# Environment Variables Template
output "backend_environment_variables" {
  description = "Template for backend environment variables"
  value = {
    DB_HOST     = azurerm_mariadb_server.main.fqdn
    DB_PORT     = "3306"
    DB_USER     = "${var.database_admin_username}@${azurerm_mariadb_server.main.name}"
    DB_NAME     = azurerm_mariadb_database.main.name
    PORT        = var.backend_port
    ENV         = var.environment
    CORS_ORIGIN = "https://${azurerm_static_site.frontend.default_host_name}"
  }
}

output "frontend_environment_variables" {
  description = "Template for frontend environment variables"
  value = {
    REACT_APP_API_URL = "https://${azurerm_linux_web_app.backend.default_hostname}"
  }
}

# Deployment Instructions
output "deployment_instructions" {
  description = "Instructions for completing the deployment"
  value = <<-EOT
    ========================================
    FDIP Platform Deployment Complete!
    ========================================
    
    Backend API URL: https://${azurerm_linux_web_app.backend.default_hostname}
    Frontend URL: https://${azurerm_static_site.frontend.default_host_name}
    ${var.enable_front_door ? "CDN URL: https://${azurerm_frontdoor.main[0].frontend_endpoint[0].host_name}" : ""}
    
    Next Steps:
    1. Set up environment variables in App Service
    2. Deploy your backend code to the App Service
    3. Deploy your frontend code to the Static Web App
    4. Configure custom domain (optional)
    5. Set up monitoring dashboards
    
    Database Connection:
    - Server: ${azurerm_mariadb_server.main.fqdn}
    - Database: ${azurerm_mariadb_database.main.name}
    - Username: ${var.database_admin_username}
    
    Key Vault:
    - Name: ${azurerm_key_vault.main.name}
    - URI: ${azurerm_key_vault.main.vault_uri}
    
    ${var.enable_application_insights ? "Application Insights: ${azurerm_application_insights.main[0].name}" : ""}
    
    For detailed instructions, see: deployment/README.md
    EOT
} 