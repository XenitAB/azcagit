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

data "azurerm_client_config" "current" {}
data "azuread_client_config" "current" {}
