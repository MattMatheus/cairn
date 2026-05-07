variable "environment_name" {
  description = "Short name used in generated resource names (e.g. pod-a-dev)."
  type        = string
}

variable "acr_name" {
  description = "Globally unique Azure Container Registry name (alphanumeric, 5-50 chars)."
  type        = string
}

variable "storage_account_name" {
  description = "Globally unique storage account name (lowercase alphanumeric, 3-24 chars)."
  type        = string
}

variable "storage_container_name" {
  description = "Blob container name for workspace files."
  type        = string
  default     = "pod-a"
}

variable "storage_prefix" {
  description = "Optional prefix within the blob container."
  type        = string
  default     = ""
}

variable "workspace_id" {
  description = "Cairn workspace ID served by this indexer (e.g. cairn:workspace:pod-a)."
  type        = string
}

variable "container_image" {
  description = "Full image reference for the Cairn indexer (e.g. <acr>.azurecr.io/cairn-indexer:latest)."
  type        = string
}

variable "operator_principal_id" {
  description = "Entra object ID of the operator user or group that needs Storage Blob Data Contributor access for cairn sync."
  type        = string
}

variable "min_replicas" {
  description = "Minimum Container App replica count. Set to 0 to allow scale-to-zero."
  type        = number
  default     = 0
}

variable "max_replicas" {
  description = "Maximum Container App replica count."
  type        = number
  default     = 3
}
