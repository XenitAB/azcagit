terraform {
  required_providers {
    azuread = {
      version = "2.45.0"
      source  = "hashicorp/azuread"
    }
    azurerm = {
      version = "3.80.0"
      source  = "hashicorp/azurerm"
    }
  }
}

provider "azurerm" {
  features {}
}

provider "azuread" {}

module "azcagit" {
  source = "../../terraform-module"

  environment                     = var.environment
  location_short                  = var.location_short
  location                        = var.location
  unique_suffix                   = var.unique_suffix
  name                            = "azcagit"
  git_config                      = var.git_config
  azcagit_version                 = var.azcagit_version
  add_permissions_to_current_user = var.add_permissions_to_current_user
  aad_resource_owner_object_ids   = var.aad_resource_owner_object_ids
}
