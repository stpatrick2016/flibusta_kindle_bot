# Azure Communication Services
resource "azurerm_communication_service" "main" {
  name                = "${var.project_name}-acs"
  resource_group_name = azurerm_resource_group.main.name
  data_location       = var.acs_data_location
  
  tags = var.tags
}

# Email Communication Service
resource "azurerm_email_communication_service" "main" {
  name                = "${var.project_name}-email"
  resource_group_name = azurerm_resource_group.main.name
  data_location       = var.acs_data_location
  
  tags = var.tags
}

# Email Communication Service Domain (AzureManagedDomain)
resource "azurerm_email_communication_service_domain" "main" {
  name             = "AzureManagedDomain"
  email_service_id = azurerm_email_communication_service.main.id
  
  domain_management = "AzureManaged"
  
  tags = var.tags
}

# Link Communication Service to Email
resource "azurerm_communication_service_email_domain_association" "main" {
  communication_service_id = azurerm_communication_service.main.id
  email_service_domain_id  = azurerm_email_communication_service_domain.main.id
}
