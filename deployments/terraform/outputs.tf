output "container_app_fqdn" {
  description = "FQDN for the Cairn indexer Container App — use as remote_index.url in workspace config."
  value       = "https://${azurerm_container_app.indexer.latest_revision_fqdn}"
}

output "acr_login_server" {
  description = "ACR login server for building and pushing the indexer image."
  value       = azurerm_container_registry.acr.login_server
}

output "managed_identity_client_id" {
  description = "Client ID of the indexer user-assigned managed identity."
  value       = azurerm_user_assigned_identity.indexer.client_id
}

output "storage_account_name" {
  description = "Storage account name for workspace sync config."
  value       = azurerm_storage_account.workspace.name
}
