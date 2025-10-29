variable "project_name" {
  description = "Name of the project, used as prefix for resources"
  type        = string
  default     = "flibusta-bot"
}

variable "resource_group_name" {
  description = "Name of the resource group"
  type        = string
  default     = "flibusta-bot-rg"
}

variable "location" {
  description = "Azure region for resources"
  type        = string
  default     = "eastus"
}

variable "environment" {
  description = "Environment name (dev, staging, prod)"
  type        = string
  default     = "prod"
}

variable "telegram_bot_token" {
  description = "Telegram Bot API token"
  type        = string
  sensitive   = true
}

variable "container_image" {
  description = "Docker container image for the bot"
  type        = string
  default     = "flibusta-bot:latest"
}

variable "container_cpu" {
  description = "CPU allocation for container (in cores)"
  type        = number
  default     = 0.5
}

variable "container_memory" {
  description = "Memory allocation for container (in GB)"
  type        = string
  default     = "1Gi"
}

variable "min_replicas" {
  description = "Minimum number of container replicas"
  type        = number
  default     = 1
}

variable "max_replicas" {
  description = "Maximum number of container replicas"
  type        = number
  default     = 3
}

variable "acs_data_location" {
  description = "Data location for Azure Communication Services"
  type        = string
  default     = "United States"
}

variable "tags" {
  description = "Tags to apply to all resources"
  type        = map(string)
  default = {
    Project     = "Flibusta Kindle Bot"
    ManagedBy   = "Terraform"
    Environment = "Production"
  }
}
