variable "location" {
  description = "The location for the platform"
  type        = string
}

variable "location_short" {
  description = "The location shortname for the platform"
  type        = string
}

variable "environment" {
  description = "The environment name to use for the platform"
  type        = string
}

variable "name" {
  description = "The name to use for the platform"
  type        = string
}

variable "unique_suffix" {
  description = "Unique suffix that is used in globally unique resources names"
  type        = string
}

variable "git_config" {
  description = "The git configuration"
  type = object({
    url      = string
    branch   = string
    path     = string
    username = string
    secret   = string
  })
  sensitive = true
}

variable "azcagit_version" {
  description = "The version of azcagit to use"
  type        = string
}

variable "add_permissions_to_current_user" {
  description = "Enable if you want permissions be added to the current user"
  type        = bool
  default     = false
}

variable "network" {
  description = "The network configuration"
  type = object({
    virtual_network_address_space = string
    subnet_address_prefix         = string
  })
  default = {
    virtual_network_address_space = "10.0.0.0/16"
    subnet_address_prefix         = "10.0.0.0/20"
  }
}

variable "aad_resource_owner_object_ids" {
  description = "Add the list of object_ids as owners to the Azure AD applications, service principals and groups"
  type        = list(string)
  default     = []
}

variable "storage_configuration" {
  description = "The storage configuration"
  type = object({
    account_replication_type = string
    share_quota              = string
  })
  default = {
    account_replication_type = "ZRS"
    share_quota              = 128
  }
}
