output "azcagit_trigger_group_object_id" {
  description = "The Object ID for the azcagit trigger permission"
  value       = azuread_group_member.azcagit_trigger.group_object_id
}
