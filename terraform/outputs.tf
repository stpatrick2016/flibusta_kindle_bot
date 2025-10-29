output "resource_group_name" {
  description = "Name of the resource group"
  value       = azurerm_resource_group.main.name
}

output "container_registry_login_server" {
  description = "Login server for the container registry"
  value       = azurerm_container_registry.main.login_server
}

output "container_app_name" {
  description = "Name of the container app"
  value       = azurerm_container_app.bot.name
}

output "communication_service_connection_string" {
  description = "Connection string for Azure Communication Services"
  value       = azurerm_communication_service.main.primary_connection_string
  sensitive   = true
}

output "acs_sender_address" {
  description = "Sender email address for Azure Communication Services"
  value       = azurerm_email_communication_service_domain.main.from_sender_domain
}

output "cosmos_db_endpoint" {
  description = "Endpoint for Cosmos DB"
  value       = azurerm_cosmosdb_account.main.endpoint
}

output "storage_account_name" {
  description = "Name of the storage account"
  value       = azurerm_storage_account.main.name
}

output "application_insights_instrumentation_key" {
  description = "Instrumentation key for Application Insights"
  value       = azurerm_application_insights.main.instrumentation_key
  sensitive   = true
}

output "application_insights_connection_string" {
  description = "Connection string for Application Insights"
  value       = azurerm_application_insights.main.connection_string
  sensitive   = true
}

output "key_vault_uri" {
  description = "URI of the Key Vault"
  value       = azurerm_key_vault.main.vault_uri
}
