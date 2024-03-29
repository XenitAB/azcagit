# azcagit

Welcome to `azcagit` (_/ɑːsk/ /eɪ/ /ɡɪt/ - ask a git_)!

It's [GitOps](https://opengitops.dev/#principles) for [Azure Container Apps](https://azure.microsoft.com/en-us/services/container-apps/#overview). You can call it a GitOps Engine for Azure Container Apps.

> Please note: This is an early version and will have lots of breaking changes in the future.

## Changelog

Below, large (and eventually breaking) will be documented:

### v0.0.19

Refactored `azcagit` to run as an [Azure Container App Job](https://github.com/XenitAB/azcagit/pull/57) on a schedule. Lots of breaking changes.

**BREAKING CHANGES**

- trigger-client cli parameter: `--namespace` instead of `--fully-qualified-namespace` (note: don't use the full name anymore)
- trigger-client cli parameter: `--queue` instead of `--topic`
- CosmosDB is used for cache
- Service Bus is now basic

### v0.0.18

Support for `AzureContainerJob` was added.

**BREAKING CHANGES**

- `spec.remoteSecrets[].appSecretName` has been renamed `spec.remoteSecrets[].secretName`
- `apiVersion` has been updated from `aca.xenit.io/v1alpha1` to `aca.xenit.io/v1alpha2` for kind `AzureContainerApp`

## Overview

![overview](docs/overview.png "Overview of azcagit")

In the test scenario, two resource groups are created:

- platform
- tenant

Platform is used for what we call "platform services", in this case the virtual network, Container Apps Managed Environment, Azure Container Registry and things to connect to using Dapr (like Storage Account or a Service Bus). In here, `azcagit` will also be setup to take care of the reconciliation of the tenant resource group.

Tenant is used only to synchronize the Container Apps manifests. The Container Apps that are created by `azcagit` will reside here.

The manifests are in the same format as Kubernetes manifests ([Kubernetes Resource Model aka KRM](https://cloud.google.com/blog/topics/developers-practitioners/build-platform-krm-part-2-how-kubernetes-resource-model-works)), but with a hard coupling to the [Azure Container Apps specification](https://docs.microsoft.com/en-us/azure/templates/microsoft.app/containerapps?pivots=deployment-language-arm-template) for `spec.app` when using `kind: AzureContainerApp` and [Azure Container Jobs specification](https://learn.microsoft.com/en-us/azure/templates/microsoft.app/jobs?pivots=deployment-language-arm-template) for `spec.job` when using `kind: AzureContainerJob`. Auto generated schemas can be found in the [schemas](schemas/) directory.

An example manifest of an app:

```yaml
kind: AzureContainerApp
apiVersion: aca.xenit.io/v1alpha2
metadata:
  name: foobar
spec:
  locationFilter:
    - West Europe
  remoteSecrets:
    - secretName: connection-string
      remoteSecretName: mssql-connection-string
  replacements:
    images:
      - imageName: "mcr.microsoft.com/azuredocs/containerapps-helloworld"
        newImageTag: "v0.1"
  app:
    properties:
      configuration:
        activeRevisionsMode: Single
      template:
        containers:
          - name: simple-hello-world-container
            image: mcr.microsoft.com/azuredocs/containerapps-helloworld:latest
            resources:
              cpu: 0.25
              memory: .5Gi
            env:
              - name: CONNECTION_STRING
                secretRef: connection-string
              - name: MEANING_WITH_LIFE
                value: "forty two"
        scale:
          minReplicas: 1
          maxReplicas: 1
```

example manifest of a job:

```yaml
kind: AzureContainerJob
apiVersion: aca.xenit.io/v1alpha2
metadata:
  name: foobar
spec:
  locationFilter:
    - West Europe
  remoteSecrets:
    - secretName: connection-string
      remoteSecretName: mssql-connection-string
  replacements:
    images:
      - imageName: "mcr.microsoft.com/k8se/quickstart-jobs"
        newImageTag: "latest"
  job:
    properties:
      configuration:
        scheduleTriggerConfig:
          cronExpression: "*/5 * * * *"
          parallelism: 1
          replicaCompletionCount: 1
        replicaRetryLimit: 1
        replicaTimeout: 1800
        triggerType: Schedule
      template:
        containers:
          - name: main
            image: mcr.microsoft.com/k8se/quickstart-jobs:foobar
            resources:
              cpu: 0.25
              memory: .5Gi
```

YAML-files can contain one or more documents (with `---` as a document separator). As of right now, all files in the git repository path (configured with `--git-path` when launching `azcagit`) needs to pass validation for any deletion to occur (deletion will be disabled if any manifests contains validation errors).

## Features

- Synchronize git repository (using https only, public and private) to a specific resource group
- Choose what folder in the git repository to synchronize
- Trigger manual synchronization using CLI
- Populate Container Apps secrets from Azure KeyVault
- Populate Container Apps registries with default registry credential
- Send notifications to the git commits
- Filter locations, making it possible to specify in the manifest what regions can run the app
- Push custom metrics to Azure monitor
- Functionality to replace the image tag using `spec.replacements.images`

## Frequently Asked Questions

> What happens if a manifest can't be parsed?

Reconciliation will stop an no changes (add/delete/update) will be made until the parse error is fixed.

> What happens if a secret in the KeyVault is defined in a manifest but doesn't exist?

Reconciliation will stop an no changes (add/delete/update) will be made until the secret is added to the KeyVault or it's removed from the manifest.

> What happens if a secret is changed in the KeyVault?

The Container App will be updated at the next reconcile.

> What happens if I add the tag `aca.xenit.io=true` to a Container App in the tenant resource group, without the app being defined in a manifest?

It will be removed at the next reconcile.

> What happens if I remove the tag `aca.xenit.io=true` from a Container App in the tenant resource group, while still having a manifest for it?

It won't be reconciled anymore. Depending on the order, a few apps before will still be reconciled but none after.

> What happens if I add the tag `aca.xenit.io=true` to a Container App in the tenant resource group, while it's also defined in a manifest?

It will be updated based on the manifest.

> What properties, as of now, can't be used even though they are defined in the Azure Container Apps specification?

- `spec.app.properties.managedEnvironmentID`: it's defined by azcagit
- `spec.app.location`: it's defined by azcagit

> What git providers are supported?

Most likely those supported by go-git, but with that said has only Azure DevOps (Azure Repositories) and GitHub been tested. If you need it with on-prem/enterprise variants of them or another git provider doesn't work as expected, create an issue or PR.

> Are private git repositories supported?

Yes, as long as you provide credentials in `--git-url` like `https://username:token@provider.io/repo`.

> I'm using a public repository without credentials but `azcagit` throws an error that it needs credentials, isn't it supported to use public repositories without credentials?

It is supported to use public repositories without credentials, but if you have enabled notifications (`--notifications-enabled`) then credentials are required to be able to push the git status to the commit.

> How does a notification look?

In GitHub, a successful notification looks like this:

![example-notification](docs/example-notification.png "Example of a notification in GitHub")

> Is multi region supported?

It sure is! You can find an example for the setup using terraform [here](test/terraform-multi-region/main.tf). We've also recorded a short video showing it in action:

[![Watch the video](docs/multi-region-thumbnail.jpg)](https://youtu.be/9SwfSIfa6I0)

> What is the location filter feature?

It makes it possible to specify `spec.locationFilter` with an array of what Azure regions are allowed to run this specific app.

![multi-region-location-filter](docs/multi-region-location-filter.png "Example of a notification in GitHub")

> How does the location filter work?

- No change if `spec.locationFilter` isn't defined
- No change if `spec.locationFilter` is an empty list
- No change if `spec.locationFilter` contains the location of azcagit (defined with `--location`)
- If `spec.locationFilter` has a value, of values, where it or none of them match the location of azcagit - we'll skip it (only logged with `--debug` enabled)

> Where can I find the custom metrics?

If you open the `azcagit` container app (in the platform resource group) and go to Monitoring and then Metrics, you can choose the namespace azcagit and then the specific metrics you want to look at.

![custom-metrics](docs/custom-metrics.png "Example custom metrics in Azure")

> How does the image tag replacement work?

If an image replacement is configured, it will match for the image name and if found it will apply the newImageTag.

## Things TODO in the future

- [x] Append secrets to Container Apps from KeyVault
- [x] ~~Better error handling of validation failures (should deletion be stopped?)~~ _stop reconciliation on any parsing error_
- [x] Push git commit status (like [Flux notification-controller](https://fluxcd.io/docs/components/notification/provider/#git-commit-status))
- [ ] Health checks
- [x] Metrics
- [x] Manually trigger reconcile
- [x] Enforce Location for app
- [x] Add Container Registry credentials by default
- [x] Add location filter

## Usage

`azcagit` will connect to a git repository (over https) and syncronize it on an interval. If changes are identified, it will push them to Azure. It can create, update and delete Azure Container Apps.

The easiest way to test it is using the terraform code which you can find in `test/terraform`. You may have to update a few names to get it working.

### Manually trigger reconcile

If you have used the example terraform, there will be a service bus created with a queue. `azcagit-trigger` will start and then trigger `azcagit-reconcile` when a message is received on the queue.

You can use `azcagit-trigger-client` to trigger it:

```go
go run ./trigger-client -n namespace -q queue
```

Please note that this requires you to be authenticated with either the Azure CLI and have access to publish to this topic with your current user, or use environment varaibles with a service principal that has access.

## Local development

### Configuration parameters

#### Environment variables

Create an environment file named `.tmp/env`.

The following parameters can be used

```env
RG_NAME=resource_group_name
SUB_ID=azure_subscription_id
ME_ID=azure_container_apps_managed_environment_id
KV_NAME=kvcontainerapps
GIT_URL_AND_CREDS=git_url_with_optional_credentials
```

The `GIT_URL_AND_CREDS` can be in either the format `https://github.com/simongottschlag/aca-test-yaml.git` or `https://username:secret@github.com/simongottschlag/aca-test-yaml-priv.git`.

To use a service principal with it's client_id and client_secret, add the following environment variables:

```env
TENANT_ID=azure_tenant_id
CLIENT_ID=service_principal_application_id
CLIENT_SECRET=service_principal_secret
```

### Terraform

Create a file named `.tmp/lab.tfvars` and configure the following:

```terraform
git_config = {
  url      = "github.com/simongottschlag/aca-test-yaml.git"
  branch   = "main"
  path     = "yaml"
  username = ""
  secret   = ""
}

azcagit_version = "vX.Y.Z"
```

Make sure url doesn't contain `https://`, it will be appended by terraform. If you want to use a private repository, add username and secret (PAT). Path is where you want `azcagit` to start traversing the directory tree.

Run the following to setup an environment (you may have to change storage account names etc): `make terraform-up`

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
