# Generate random suffix for unique resource names
resource "random_string" "suffix" {
  length  = 6
  special = false
  upper   = false
}

locals {
  name_suffix = random_string.suffix.result
  common_tags = merge(var.tags, {
    Environment = var.environment
    Project     = var.project_name
    ManagedBy   = "terraform"
  })
}

# Resource Group
resource "azurerm_resource_group" "main" {
  name     = var.resource_group_name
  location = var.location
  tags     = local.common_tags
}

# Virtual Network
resource "azurerm_virtual_network" "main" {
  name                = "fdip-vnet-${local.name_suffix}"
  resource_group_name = azurerm_resource_group.main.name
  location            = azurerm_resource_group.main.location
  address_space       = var.vnet_address_space
  tags                = local.common_tags
}

# Subnets
resource "azurerm_subnet" "app_service" {
  name                 = "fdip-app-service-subnet"
  resource_group_name  = azurerm_resource_group.main.name
  virtual_network_name = azurerm_virtual_network.main.name
  address_prefixes     = [var.subnet_address_prefixes[0]]
  
  delegation {
    name = "app-service-delegation"
    service_delegation {
      name    = "Microsoft.Web/serverFarms"
      actions = ["Microsoft.Network/virtualNetworks/subnets/action"]
    }
  }
}

resource "azurerm_subnet" "database" {
  name                 = "fdip-database-subnet"
  resource_group_name  = azurerm_resource_group.main.name
  virtual_network_name = azurerm_virtual_network.main.name
  address_prefixes     = [var.subnet_address_prefixes[1]]
  
  service_endpoints = ["Microsoft.Sql"]
}

# Network Security Groups
resource "azurerm_network_security_group" "app_service" {
  name                = "fdip-app-service-nsg"
  location            = azurerm_resource_group.main.location
  resource_group_name = azurerm_resource_group.main.name
  tags                = local.common_tags

  security_rule {
    name                       = "Allow-HTTP"
    priority                   = 100
    direction                  = "Inbound"
    access                     = "Allow"
    protocol                   = "Tcp"
    source_port_range          = "*"
    destination_port_range     = "80"
    source_address_prefix      = "*"
    destination_address_prefix = "*"
  }

  security_rule {
    name                       = "Allow-HTTPS"
    priority                   = 110
    direction                  = "Inbound"
    access                     = "Allow"
    protocol                   = "Tcp"
    source_port_range          = "*"
    destination_port_range     = "443"
    source_address_prefix      = "*"
    destination_address_prefix = "*"
  }
}

# Key Vault
resource "azurerm_key_vault" "main" {
  name                        = "fdip-kv-${local.name_suffix}"
  location                    = azurerm_resource_group.main.location
  resource_group_name         = azurerm_resource_group.main.name
  enabled_for_disk_encryption = true
  tenant_id                   = data.azurerm_client_config.current.tenant_id
  soft_delete_retention_days  = 7
  purge_protection_enabled    = false
  sku_name                   = var.key_vault_sku
  tags                       = local.common_tags
}

# Key Vault Access Policy
resource "azurerm_key_vault_access_policy" "terraform" {
  key_vault_id = azurerm_key_vault.main.id
  tenant_id    = data.azurerm_client_config.current.tenant_id
  object_id    = data.azurerm_client_config.current.object_id

  key_permissions = [
    "Get", "List", "Create", "Delete", "Update", "Import", "Backup", "Restore", "Recover", "Purge"
  ]

  secret_permissions = [
    "Get", "List", "Set", "Delete", "Backup", "Restore", "Recover", "Purge"
  ]

  certificate_permissions = [
    "Get", "List", "Create", "Delete", "Update", "Import", "Backup", "Restore", "Recover", "Purge"
  ]
}

# MariaDB Server
resource "azurerm_mariadb_server" "main" {
  name                = "fdip-db-${local.name_suffix}"
  location            = azurerm_resource_group.main.location
  resource_group_name = azurerm_resource_group.main.name
  tags                = local.common_tags

  administrator_login          = var.database_admin_username
  administrator_login_password = var.database_admin_password

  sku_name   = var.database_sku_name
  version    = "10.3"
  storage_mb = var.database_storage_mb

  backup_retention_days        = 7
  geo_redundant_backup_enabled = false
  auto_grow_enabled           = true

  public_network_access_enabled = false
  ssl_enforcement_enabled      = true
}

# MariaDB Database
resource "azurerm_mariadb_database" "main" {
  name                = var.database_name
  resource_group_name = azurerm_resource_group.main.name
  server_name         = azurerm_mariadb_server.main.name
  charset             = "utf8"
  collation           = "utf8_general_ci"
}

# Private Endpoint for Database
resource "azurerm_private_endpoint" "database" {
  name                = "fdip-db-private-endpoint"
  location            = azurerm_resource_group.main.location
  resource_group_name = azurerm_resource_group.main.name
  subnet_id           = azurerm_subnet.database.id
  tags                = local.common_tags

  private_service_connection {
    name                           = "fdip-db-private-connection"
    private_connection_resource_id = azurerm_mariadb_server.main.id
    subresource_names             = ["mariadbServer"]
    is_manual_connection          = false
  }
}

# App Service Plan
resource "azurerm_service_plan" "main" {
  name                = "fdip-app-plan-${local.name_suffix}"
  resource_group_name = azurerm_resource_group.main.name
  location            = azurerm_resource_group.main.location
  os_type             = "Linux"
  sku_name            = var.app_service_plan_sku
  tags                = local.common_tags
}

# App Service
resource "azurerm_linux_web_app" "backend" {
  name                = "fdip-backend-${local.name_suffix}"
  resource_group_name = azurerm_resource_group.main.name
  location            = azurerm_resource_group.main.location
  service_plan_id     = azurerm_service_plan.main.id
  tags                = local.common_tags

  site_config {
    application_stack {
      docker {
        registry_url = "https://index.docker.io"
        image_name   = "fdip-backend:latest"
      }
    }

    health_check_path = "/health"
    always_on         = true
  }

  app_settings = {
    "WEBSITES_ENABLE_APP_SERVICE_STORAGE" = "false"
    "DOCKER_ENABLE_CI"                   = "true"
    "PORT"                               = var.backend_port
    "ENV"                                = var.environment
  }

  lifecycle {
    ignore_changes = [
      app_settings["WEBSITE_RUN_FROM_PACKAGE"],
      app_settings["DOCKER_REGISTRY_SERVER_URL"],
      app_settings["DOCKER_REGISTRY_SERVER_USERNAME"],
      app_settings["DOCKER_REGISTRY_SERVER_PASSWORD"],
    ]
  }
}

# Static Web App
resource "azurerm_static_site" "frontend" {
  name                = "fdip-frontend-${local.name_suffix}"
  resource_group_name = azurerm_resource_group.main.name
  location            = azurerm_resource_group.main.location
  sku_tier            = var.static_web_app_sku
  tags                = local.common_tags
}

# Application Insights
resource "azurerm_application_insights" "main" {
  count               = var.enable_application_insights ? 1 : 0
  name                = "fdip-app-insights-${local.name_suffix}"
  location            = azurerm_resource_group.main.location
  resource_group_name = azurerm_resource_group.main.name
  application_type    = "web"
  tags                = local.common_tags
}

# Front Door (CDN)
resource "azurerm_frontdoor" "main" {
  count               = var.enable_front_door ? 1 : 0
  name                = "fdip-frontdoor-${local.name_suffix}"
  resource_group_name = azurerm_resource_group.main.name
  tags                = local.common_tags

  routing_rule {
    name               = "fdip-routing-rule"
    accepted_protocols = ["Http", "Https"]
    patterns_to_match   = ["/*"]
    frontend_endpoints  = ["fdip-frontend-endpoint"]
    forwarding_configuration {
      forwarding_protocol = "MatchRequest"
      backend_pool_name   = "fdip-backend-pool"
    }
  }

  backend_pool_load_balancing {
    name = "fdip-load-balancing"
  }

  backend_pool_health_probe {
    name = "fdip-health-probe"
  }

  backend_pool {
    name = "fdip-backend-pool"
    backend {
      host_header = azurerm_linux_web_app.backend.default_hostname
      address     = azurerm_linux_web_app.backend.default_hostname
      http_port   = 80
      https_port  = 443
    }

    load_balancing_name = "fdip-load-balancing"
    health_probe_name   = "fdip-health-probe"
  }

  frontend_endpoint {
    name                              = "fdip-frontend-endpoint"
    host_name                         = "fdip-frontdoor-${local.name_suffix}.azurefd.net"
    session_affinity_enabled          = true
    session_affinity_ttl_seconds      = 300
  }
}

# Data source for current Azure client configuration
data "azurerm_client_config" "current" {} 