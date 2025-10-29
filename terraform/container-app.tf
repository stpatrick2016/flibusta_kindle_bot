# Container Registry
resource "azurerm_container_registry" "main" {
  name                = replace("${var.project_name}acr", "-", "")
  resource_group_name = azurerm_resource_group.main.name
  location            = azurerm_resource_group.main.location
  sku                 = "Basic"
  admin_enabled       = true
  
  tags = var.tags
}

# Container Apps Environment
resource "azurerm_container_app_environment" "main" {
  name                       = "${var.project_name}-env"
  location                   = azurerm_resource_group.main.location
  resource_group_name        = azurerm_resource_group.main.name
  log_analytics_workspace_id = azurerm_log_analytics_workspace.main.id
  
  tags = var.tags
}

# Container App
resource "azurerm_container_app" "bot" {
  name                         = var.project_name
  container_app_environment_id = azurerm_container_app_environment.main.id
  resource_group_name          = azurerm_resource_group.main.name
  revision_mode                = "Single"
  
  template {
    min_replicas = var.min_replicas
    max_replicas = var.max_replicas
    
    container {
      name   = "bot"
      image  = "${azurerm_container_registry.main.login_server}/${var.container_image}"
      cpu    = var.container_cpu
      memory = var.container_memory
      
      env {
        name  = "TELEGRAM_BOT_TOKEN"
        secret_name = "telegram-bot-token"
      }
      
      env {
        name  = "ACS_CONNECTION_STRING"
        secret_name = "acs-connection-string"
      }
      
      env {
        name  = "ACS_SENDER_ADDRESS"
        value = azurerm_email_communication_service_domain.main.from_sender_domain
      }
      
      env {
        name  = "DATABASE_CONNECTION_STRING"
        secret_name = "database-connection-string"
      }
      
      env {
        name  = "AZURE_STORAGE_CONNECTION_STRING"
        secret_name = "storage-connection-string"
      }
      
      env {
        name  = "APPLICATIONINSIGHTS_CONNECTION_STRING"
        value = azurerm_application_insights.main.connection_string
      }
      
      env {
        name  = "LOG_LEVEL"
        value = "info"
      }
    }
  }
  
  secret {
    name  = "telegram-bot-token"
    value = var.telegram_bot_token
  }
  
  secret {
    name  = "acs-connection-string"
    value = azurerm_communication_service.main.primary_connection_string
  }
  
  secret {
    name  = "database-connection-string"
    value = azurerm_cosmosdb_account.main.connection_strings[0]
  }
  
  secret {
    name  = "storage-connection-string"
    value = azurerm_storage_account.main.primary_connection_string
  }
  
  registry {
    server               = azurerm_container_registry.main.login_server
    username             = azurerm_container_registry.main.admin_username
    password_secret_name = "registry-password"
  }
  
  secret {
    name  = "registry-password"
    value = azurerm_container_registry.main.admin_password
  }
  
  ingress {
    external_enabled = false
    target_port      = 8080
    
    traffic_weight {
      latest_revision = true
      percentage      = 100
    }
  }
  
  tags = var.tags
}
