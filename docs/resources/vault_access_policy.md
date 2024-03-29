---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "multy_vault_access_policy Resource - terraform-provider-multy"
subcategory: ""
description: |-
  Provides Multy Object Storage resource
---

# multy_vault_access_policy (Resource)

Provides Multy Object Storage resource

## Example Usage

```terraform
resource "multy_vault" "web_app_vault" {
  name     = "web-app-vault-test"
  cloud    = "azure"
  location = "ireland"
}

resource "multy_vault_access_policy" "kv_ap" {
  vault_id = multy_vault.web_app_vault.id
  identity = multy_virtual_machine.vm.identity
  access   = "owner"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `access` (String) Access control, available values are [read write owner]
- `identity` (String) Identity of the resource that is being granted access to the `vault`
- `vault_id` (String) Id of the associated vault

### Read-Only

- `aws` (Object) AWS-specific ids of the underlying generated resources (see [below for nested schema](#nestedatt--aws))
- `azure` (Object) Azure-specific ids of the underlying generated resources (see [below for nested schema](#nestedatt--azure))
- `gcp` (Object) GCP-specific ids of the underlying generated resources (see [below for nested schema](#nestedatt--gcp))
- `id` (String) The ID of this resource.
- `resource_status` (Map of String) Statuses of underlying created resources

<a id="nestedatt--aws"></a>
### Nested Schema for `aws`

Read-Only:

- `iam_policy_arn` (String)


<a id="nestedatt--azure"></a>
### Nested Schema for `azure`

Read-Only:

- `key_vault_access_policy_id` (String)


<a id="nestedatt--gcp"></a>
### Nested Schema for `gcp`

Read-Only:

- `secret_manager_secret_iam_membership_ids` (List of String)


