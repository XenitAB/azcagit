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

variable "location" {
  description = "The location for the platform"
  type        = string
  default     = "West Europe"
}

variable "location_short" {
  description = "The location shortname for the platform"
  type        = string
  default     = "we"
}

variable "environment" {
  description = "The environment name to use for the platform"
  type        = string
  default     = "dev"
}

variable "unique_suffix" {
  description = "Unique suffix that is used in globally unique resources names"
  type        = string
  default     = ""
}

variable "add_permissions_to_current_user" {
  description = "Enable if you want permissions be added to the current user"
  type        = bool
  default     = false
}

variable "aad_resource_owner_object_ids" {
  description = "Add the list of object_ids as owners to the Azure AD applications, service principals and groups"
  type        = list(string)
  default     = []
}
