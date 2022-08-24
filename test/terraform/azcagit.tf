locals {
  git_user_string = "${var.git_config.username}:${var.git_config.secret}@" == ":@" ? "" : "${var.git_config.username}:${var.git_config.secret}@"
  git_full_url    = "https://${local.git_user_string}${var.git_config.url}"
}

resource "azuread_application" "azcagit" {
  display_name = "sp-azcagit"
}

resource "azuread_service_principal" "azcagit" {
  application_id = azuread_application.azcagit.application_id
}

resource "azuread_application_password" "azcagit" {
  application_object_id = azuread_application.azcagit.object_id
  end_date              = timeadd(timestamp(), "87600h") # 10 years

  lifecycle {
    ignore_changes = [
      end_date
    ]
  }
}

resource "azurerm_role_assignment" "azcagit_platform" {
  scope                = azurerm_resource_group.platform.id
  role_definition_name = "Contributor"
  principal_id         = azuread_service_principal.azcagit.object_id
}

resource "azurerm_role_assignment" "azcagit_tenant" {
  scope                = azurerm_resource_group.tenant.id
  role_definition_name = "Contributor"
  principal_id         = azuread_service_principal.azcagit.object_id
}

resource "azapi_resource" "container_app_azcagit" {
  type                      = "Microsoft.App/containerapps@2022-03-01"
  name                      = "azcagit"
  parent_id                 = azurerm_resource_group.platform.id
  location                  = azurerm_resource_group.platform.location
  schema_validation_enabled = false

  body = jsonencode({
    properties = {
      managedEnvironmentId = azapi_resource.managed_environment.id
      configuration = {
        activeRevisionsMode = "Single"
        secrets = [
          {
            name  = "git-url"
            value = local.git_full_url
          },
          {
            name  = "azure-tenant-id"
            value = data.azurerm_client_config.current.tenant_id
          },
          {
            name  = "azure-client-id"
            value = azuread_application.azcagit.application_id
          },
          {
            name  = "azure-client-secret"
            value = azuread_application_password.azcagit.value
          }
        ]
      }
      template = {
        containers = [
          {
            name  = "azcagit"
            image = "ghcr.io/xenitab/azcagit:v0.0.1"
            args = [
              "--resource-group-name", azurerm_resource_group.tenant.name,
              "--subscription-id", data.azurerm_client_config.current.subscription_id,
              "--managed-environment-id", azapi_resource.managed_environment.id,
              "--location", azurerm_resource_group.tenant.location,
              "--reconcile-interval", "30s",
              "--checkout-path", "/tmp/gitops",
              "--git-url", local.git_full_url,
              "--git-branch", var.git_config.branch,
              "--git-yaml-path", var.git_config.path
            ]
            env = [
              {
                name      = "GIT_URL"
                secretRef = "git-url"
              },
              {
                name      = "AZURE_TENANT_ID"
                secretRef = "azure-tenant-id"
              },
              {
                name      = "AZURE_CLIENT_ID"
                secretRef = "azure-client-id"
              },
              {
                name      = "AZURE_CLIENT_SECRET"
                secretRef = "azure-client-secret"
              },
            ]
            resources = {
              cpu    = ".25"
              memory = ".5Gi"
            }
          }
        ]
        scale = {
          minReplicas = 1
          maxReplicas = 1
          rules       = []
        }
      }
    }
  })

  response_export_values = ["properties"]
}
