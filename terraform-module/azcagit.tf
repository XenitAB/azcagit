locals {
  git_user_string = "${var.git_config.username}:${var.git_config.secret}@" == ":@" ? "" : "${var.git_config.username}:${var.git_config.secret}@"
  git_full_url    = "https://${local.git_user_string}${var.git_config.url}"
}

resource "azuread_application" "azcagit" {
  display_name = "sp-${local.eln}-azcagit"
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

resource "azurerm_role_assignment" "azcagit_platform_metrics_publisher" {
  scope                = azurerm_resource_group.platform.id
  role_definition_name = "Monitoring Metrics Publisher"
  principal_id         = azuread_service_principal.azcagit.object_id
}

resource "azurerm_role_assignment" "azcagit_tenant" {
  scope                = azurerm_resource_group.tenant.id
  role_definition_name = "Contributor"
  principal_id         = azuread_service_principal.azcagit.object_id
}

resource "azurerm_container_app" "azcagit" {
  name                         = "azcagit"
  container_app_environment_id = azurerm_container_app_environment.this.id
  resource_group_name          = azurerm_resource_group.platform.name
  revision_mode                = "Single"

  template {
    container {
      name  = "azcagit"
      image = "ghcr.io/xenitab/azcagit:${var.azcagit_version}"
      args = [
        "--resource-group-name", azurerm_resource_group.tenant.name,
        "--environment", var.environment,
        "--subscription-id", data.azurerm_client_config.current.subscription_id,
        "--managed-environment-id", azurerm_container_app_environment.this.id,
        "--key-vault-name", azurerm_key_vault.tenant_kv.name,
        "--own-resource-group-name", azurerm_resource_group.platform.name,
        "--container-registry-server", azurerm_container_registry.tenant.login_server,
        "--container-registry-username", azurerm_container_registry.tenant.admin_username,
        "--location", azurerm_resource_group.tenant.location,
        "--dapr-topic-name", azurerm_servicebus_topic.azcagit_trigger.name,
        "--reconcile-interval", "5m",
        "--git-branch", var.git_config.branch,
        "--git-yaml-path", var.git_config.path,
        "--notifications-enabled"
      ]

      env {
        name        = "GIT_URL"
        secret_name = "git-url"
      }
      env {
        name        = "CONTAINER_REGISTRY_PASSWORD"
        secret_name = "container-registry-password"
      }
      env {
        name        = "AZURE_TENANT_ID"
        secret_name = "azure-tenant-id"
      }
      env {
        name        = "AZURE_CLIENT_ID"
        secret_name = "azure-client-id"
      }
      env {
        name        = "AZURE_CLIENT_SECRET"
        secret_name = "azure-client-secret"
      }

      memory = "0.5Gi"
      cpu    = "0.25"
    }

    min_replicas = 1
    max_replicas = 1
  }

  secret {
    name  = "git-url"
    value = local.git_full_url
  }
  secret {
    name  = "container-registry-password"
    value = azurerm_container_registry.tenant.admin_password
  }
  secret {
    name  = "azure-tenant-id"
    value = data.azurerm_client_config.current.tenant_id
  }
  secret {
    name  = "azure-client-id"
    value = azuread_application.azcagit.application_id
  }
  secret {
    name  = "azure-client-secret"
    value = azuread_application_password.azcagit.value
  }

  dapr {
    app_id       = "azcagit"
    app_port     = 8080
    app_protocol = "http"
  }
}
