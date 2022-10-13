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