terraform {
  required_providers {
    azuread = {
      source  = "hashicorp/azuread"
      version = ">=2.45.0"
    }
    azurerm = {
      source  = "hashicorp/azurerm"
      version = ">=3.80.0"
    }
    azapi = {
      source  = "Azure/azapi"
      version = ">=1.10.0"
    }
    random = {
      source  = "hashicorp/random"
      version = ">=3.5.1"
    }
  }
}

provider "azurerm" {
  features {}
}

data "azurerm_client_config" "current" {}
data "azuread_client_config" "current" {}
