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
    azapi = {
      source  = "Azure/azapi"
      version = "1.7.0"
    }
    random = {
      source  = "hashicorp/random"
      version = "3.5.1"
    }
  }
}

provider "azurerm" {
  features {}
}

provider "azapi" {}

data "azurerm_client_config" "current" {}
data "azuread_client_config" "current" {}
