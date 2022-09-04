terraform {
  required_providers {
    azuread = {
      version = "2.19.1"
      source  = "hashicorp/azuread"
    }
    azurerm = {
      version = "3.8.0"
      source  = "hashicorp/azurerm"
    }
    azapi = {
      version = "0.3.0"
      source  = "Azure/azapi"
    }
  }
}

provider "azurerm" {
  features {}
}

provider "azuread" {}

module "azcagit_we" {
  source = "../../terraform-module"

  environment     = "dev"
  location_short  = "we"
  location        = "West Europe"
  unique_suffix   = ""
  name            = "azcagit"
  git_config      = var.git_config
  azcagit_version = var.azcagit_version
}

module "azcagit_ne" {
  source = "../../terraform-module"

  environment     = "dev"
  location_short  = "ne"
  location        = "North Europe"
  unique_suffix   = ""
  name            = "azcagit"
  git_config      = var.git_config
  azcagit_version = var.azcagit_version
}
