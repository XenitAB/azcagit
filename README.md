# azcagit

Welcome to `azcagit` (_/ɑːsk/ /eɪ/ /ɡɪt/ - ask a git_)!

It is a MVP of using [GitOps](https://opengitops.dev/#principles) with [Azure Container Apps](https://azure.microsoft.com/en-us/services/container-apps/#overview). You could call it an Azure Container App GitOps Engine.

> Please note: This is an early version and will have lots of breaking changes in the future.

## Overview

![overview](docs/overview.png "Overview of azcagit")

In the test scenario, two resource groups are created:

- platform
- tenant

Platform is used for what we call "platform services", in this case the virtual network, Container Apps Managed Environment, Azure Container Registry and things to connect to using Dapr (like Storage Account or a Service Bus). In here, `azcagit` will also be setup to take care of the reconciliation of the tenant resource group.

Tenant is used only to synchronize the Container Apps manifests. The Container Apps that are created by `azcagit` will reside here.

The manifests are in the same format as Kubernetes manifests ([Kubernetes Resource Model aka KRM](https://cloud.google.com/blog/topics/developers-practitioners/build-platform-krm-part-2-how-kubernetes-resource-model-works)), but with a hard coupling to the [Azure Container Apps specification](https://docs.microsoft.com/en-us/azure/templates/microsoft.app/containerapps?pivots=deployment-language-arm-template).

An example manifest:

```yaml
kind: AzureContainerApp
apiVersion: aca.xenit.io/v1alpha1
metadata:
  name: foobar
spec:
  app:
    location: west europe
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
        scale:
          minReplicas: 1
          maxReplicas: 1
```

YAML-files can contain one or more documents (with `---` as a document separator). As of right now, all files in the git repository path (configured with `--git-path` when launching `azcagit`) needs to pass validation for any deletion to occur (deletion will be disabled if any manifests contains validation errors).

## Things TODO in the future

- [ ] Append secrets to Container Apps from KeyVault
- [x] ~~Better error handling of validation failures (should deletion be stopped?)~~ _stop reconciliation on any parsing error_
- [ ] Push git commit status (like [Flux notification-controller](https://fluxcd.io/docs/components/notification/provider/#git-commit-status))
- [ ] Health checks
- [ ] Metrics

## Usage

`azcagit` will connect to a git repository (over https) and syncronize it on an interval. If changes are identified, it will push them to Azure. It can create, update and delete Azure Container Apps.

The easiest way to test it is using the terraform code which you can find in `test/terraform`. You may have to update a few names to get it working.

## Local development

### libgit2

`libgit2` (`v1.3.0`) is required to be [installed locally](https://github.com/libgit2/libgit2#installation).

If you are using Arch based distribution, please follow the guidelines below:

```shell
cd /tmp
git clone https://aur.archlinux.org/libgit2-git.git
cd libgit2-git
sed -i 's|\(git+https://github.com/libgit2/libgit2.git\)|\1#tag=v1.3.0|g' PKGBUILD
makepkg -si
```

### Configuration parameters

#### Environment variables

Create an environment file named `.tmp/env`.

The following parameters can be used

```env
RG_NAME=resource_group_name
SUB_ID=azure_subscription_id
ME_ID=azure_container_apps_managed_environment_id
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
```

Make sure url doesn't contain `https://`, it will be appended by terraform. If you want to use a private repository, add username and secret (PAT). Path is where you want `azcagit` to start traversing the directory tree.

Run the following to setup an environment (you may have to change storage account names etc): `make terraform-up`

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
