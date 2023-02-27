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
