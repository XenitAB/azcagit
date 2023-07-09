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

data "azurerm_client_config" "current" {}
data "azuread_client_config" "current" {}
