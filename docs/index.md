---
layout: ""
page_title: "Provider: Multy"
description: |- The [Multy](https://multy.dev) Terraform provider allows you to deploy Multy resources that are
cloud-agnostic and can easily be moved across different clouds.
---

# Multy Provider

With the [Multy](https://multy.dev/) provider, you can easily leverage the Multy API to deploy resources. The provider
requires credentials to deploy resources into the respective cloud provider.

~> **NOTE:** This provider is currently in alpha. It is under development and will have breaking changes in future
versions. Feel free to contribute or leave feedback in
our [public repo](https://github.com/multycloud/terraform-provider-multy/).

## Example Usage

```terraform
terraform {
  required_providers {
    multy = {
      source = "multycloud/multy"
    }
  }
}

provider "multy" {
  api_key = "<multy_api_key>"
}
```