terraform {
  required_providers {
    azuread = {
      version = "2.39.0"
      source  = "hashicorp/azuread"
    }
    azurerm = {
      version = "3.64.0"
      source  = "hashicorp/azurerm"
    }
  }
}

provider "azurerm" {
  features {}
}

provider "azuread" {}

module "azcagit_we" {
  source = "../../terraform-module"

  environment                   = "dev"
  location_short                = "we"
  location                      = "West Europe"
  unique_suffix                 = ""
  name                          = "azcagit"
  git_config                    = var.git_config
  azcagit_version               = var.azcagit_version
  aad_resource_owner_object_ids = var.aad_resource_owner_object_ids
}

module "azcagit_ne" {
  source = "../../terraform-module"

  environment                   = "dev"
  location_short                = "ne"
  location                      = "North Europe"
  unique_suffix                 = var.unique_suffix
  name                          = "azcagit"
  git_config                    = var.git_config
  azcagit_version               = var.azcagit_version
  aad_resource_owner_object_ids = var.aad_resource_owner_object_ids
}
