# Storage Account for temporary book files
resource "azurerm_storage_account" "main" {
  name                     = replace("${var.project_name}sa", "-", "")
  resource_group_name      = azurerm_resource_group.main.name
  location                 = azurerm_resource_group.main.location
  account_tier             = "Standard"
  account_replication_type = "LRS"
  
  blob_properties {
    delete_retention_policy {
      days = 7
    }
  }
  
  tags = var.tags
}

# Storage Container for books
resource "azurerm_storage_container" "books" {
  name                  = "books"
  storage_account_name  = azurerm_storage_account.main.name
  container_access_type = "private"
}

# Cosmos DB Account (NoSQL)
resource "azurerm_cosmosdb_account" "main" {
  name                = "${var.project_name}-cosmos"
  location            = azurerm_resource_group.main.location
  resource_group_name = azurerm_resource_group.main.name
  offer_type          = "Standard"
  kind                = "GlobalDocumentDB"
  
  consistency_policy {
    consistency_level = "Session"
  }
  
  geo_location {
    location          = azurerm_resource_group.main.location
    failover_priority = 0
  }
  
  capabilities {
    name = "EnableServerless"
  }
  
  tags = var.tags
}

# Cosmos DB SQL Database
resource "azurerm_cosmosdb_sql_database" "main" {
  name                = "flibusta_bot"
  resource_group_name = azurerm_resource_group.main.name
  account_name        = azurerm_cosmosdb_account.main.name
}

# Cosmos DB SQL Container for Users
resource "azurerm_cosmosdb_sql_container" "users" {
  name                = "users"
  resource_group_name = azurerm_resource_group.main.name
  account_name        = azurerm_cosmosdb_account.main.name
  database_name       = azurerm_cosmosdb_sql_database.main.name
  partition_key_paths = ["/telegram_user_id"]
  
  indexing_policy {
    indexing_mode = "consistent"
    
    included_path {
      path = "/*"
    }
  }
}

# Cosmos DB SQL Container for Sessions
resource "azurerm_cosmosdb_sql_container" "sessions" {
  name                = "sessions"
  resource_group_name = azurerm_resource_group.main.name
  account_name        = azurerm_cosmosdb_account.main.name
  database_name       = azurerm_cosmosdb_sql_database.main.name
  partition_key_paths = ["/user_id"]
  
  default_ttl = 3600 # Sessions expire after 1 hour
  
  indexing_policy {
    indexing_mode = "consistent"
    
    included_path {
      path = "/*"
    }
  }
}
