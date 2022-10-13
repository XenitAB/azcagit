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
