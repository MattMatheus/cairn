terraform {
  required_providers {
    azurerm = {
      source  = "hashicorp/azurerm"
      version = "~> 4.0"
    }
  }
}

provider "azurerm" {
  features {}
}

locals {
  location    = "westus2"
  rg_name     = "athdev-wus2-centralkb-rg"
  name_prefix = "cairn-${var.environment_name}"

  tags = {
    Class          = "Technology"
    CostCenter     = "Ignite"
    Customer       = "Health Catalyst Internal"
    Product        = "Platform Services"
    AllocationType = "No Allocation Type"
    HCBusinessUnit = "Health Catalyst"
  }
}

resource "azurerm_resource_group" "rg" {
  name     = local.rg_name
  location = local.location
  tags     = local.tags
}

resource "azurerm_container_registry" "acr" {
  name                = var.acr_name
  resource_group_name = azurerm_resource_group.rg.name
  location            = azurerm_resource_group.rg.location
  sku                 = "Basic"
  admin_enabled       = false
  tags                = local.tags
}

resource "azurerm_storage_account" "workspace" {
  name                     = var.storage_account_name
  resource_group_name      = azurerm_resource_group.rg.name
  location                 = azurerm_resource_group.rg.location
  account_tier             = "Standard"
  account_replication_type = "LRS"
  tags                     = local.tags
}

resource "azurerm_storage_container" "workspace" {
  name                  = var.storage_container_name
  storage_account_id    = azurerm_storage_account.workspace.id
  container_access_type = "private"
}

resource "azurerm_log_analytics_workspace" "logs" {
  name                = "${local.name_prefix}-logs"
  location            = azurerm_resource_group.rg.location
  resource_group_name = azurerm_resource_group.rg.name
  sku                 = "PerGB2018"
  retention_in_days   = 30
  tags                = local.tags
}

resource "azurerm_user_assigned_identity" "indexer" {
  name                = "${local.name_prefix}-indexer-mi"
  location            = azurerm_resource_group.rg.location
  resource_group_name = azurerm_resource_group.rg.name
  tags                = local.tags
}

resource "azurerm_container_app_environment" "env" {
  name                       = "${local.name_prefix}-aca"
  location                   = azurerm_resource_group.rg.location
  resource_group_name        = azurerm_resource_group.rg.name
  log_analytics_workspace_id = azurerm_log_analytics_workspace.logs.id
  tags                       = local.tags
}

resource "azurerm_role_assignment" "acr_pull" {
  scope                = azurerm_container_registry.acr.id
  role_definition_name = "AcrPull"
  principal_id         = azurerm_user_assigned_identity.indexer.principal_id
}

resource "azurerm_role_assignment" "storage_reader" {
  scope                = azurerm_storage_account.workspace.id
  role_definition_name = "Storage Blob Data Reader"
  principal_id         = azurerm_user_assigned_identity.indexer.principal_id
}

resource "azurerm_role_assignment" "storage_contributor_user" {
  scope                = azurerm_storage_account.workspace.id
  role_definition_name = "Storage Blob Data Contributor"
  principal_id         = var.operator_principal_id
}

resource "azurerm_container_app" "indexer" {
  name                         = "${local.name_prefix}-indexer"
  container_app_environment_id = azurerm_container_app_environment.env.id
  resource_group_name          = azurerm_resource_group.rg.name
  revision_mode                = "Single"
  tags                         = local.tags

  identity {
    type         = "UserAssigned"
    identity_ids = [azurerm_user_assigned_identity.indexer.id]
  }

  registry {
    server   = azurerm_container_registry.acr.login_server
    identity = azurerm_user_assigned_identity.indexer.id
  }

  ingress {
    external_enabled = true
    target_port      = 8080
    transport        = "http"

    traffic_weight {
      percentage      = 100
      latest_revision = true
    }
  }

  template {
    container {
      name   = "indexer"
      image  = var.container_image
      cpu    = 0.5
      memory = "1Gi"

      env {
        name  = "CAIRN_INDEXER_WORKSPACE_ID"
        value = var.workspace_id
      }
      env {
        name  = "CAIRN_STORAGE_ACCOUNT"
        value = azurerm_storage_account.workspace.name
      }
      env {
        name  = "CAIRN_STORAGE_CONTAINER"
        value = azurerm_storage_container.workspace.name
      }
      env {
        name  = "CAIRN_STORAGE_PREFIX"
        value = var.storage_prefix
      }
      env {
        name  = "CAIRN_LOG_LEVEL"
        value = "info"
      }
    }

    min_replicas = var.min_replicas
    max_replicas = var.max_replicas
  }

  depends_on = [
    azurerm_role_assignment.acr_pull,
    azurerm_role_assignment.storage_reader,
  ]
}
