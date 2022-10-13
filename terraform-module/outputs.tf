output "azcagit_trigger_group_object_id" {
  description = "The Object ID for the azcagit trigger permission"
  value       = azuread_group.azcagit_trigger.id
}
