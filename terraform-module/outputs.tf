output "azcagit_trigger_group_object_id" {
  description = "The Object ID for the azcagit trigger permission"
  value       = azuread_group.azcagit_trigger.id
}

output "container_app_environment" {
  description = "The Container App Environment"
  value       = azurerm_container_app_environment.this
}

output "platform_resource_group" {
  description = "The platform resource group"
  value       = azurerm_resource_group.platform
}

output "tenant_resource_group" {
  description = "The tenant resource group"
  value       = azurerm_resource_group.tenant
}
