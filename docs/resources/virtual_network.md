---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "multy_virtual_network Resource - terraform-provider-multy"
subcategory: ""
description: |-
  Provides Multy Virtual Network resource
---

# multy_virtual_network (Resource)

Provides Multy Virtual Network resource

## Example Usage

```terraform
resource multy_virtual_network vn {
  for_each = ["aws", "azure"]

  name       = "dev-vn"
  cidr_block = "10.0.0.0/16"
  cloud      = each.key
  location   = "us_east_1"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `cidr_block` (String) CIDR Block of Virtual Network
- `cloud` (String) Cloud provider to deploy resource into. Accepted values are `aws`, `azure` or `gcp`
- `location` (String) Location to deploy resource into. Read more about regions in [documentation](https://docs.multy.dev/regions)
- `name` (String) Name of Virtual Network

### Optional

- `gcp_overrides` (Attributes) GCP-specific attributes that will be set if this resource is deployed in GCP (see [below for nested schema](#nestedatt--gcp_overrides))

### Read-Only

- `aws` (Object) AWS-specific ids of the underlying generated resources (see [below for nested schema](#nestedatt--aws))
- `azure` (Object) Azure-specific ids of the underlying generated resources (see [below for nested schema](#nestedatt--azure))
- `gcp` (Object) GCP-specific ids of the underlying generated resources (see [below for nested schema](#nestedatt--gcp))
- `id` (String) The ID of this resource.
- `resource_group_id` (String)
- `resource_status` (Map of String) Statuses of underlying created resources

<a id="nestedatt--gcp_overrides"></a>
### Nested Schema for `gcp_overrides`

Optional:

- `project` (String) The project to use for this resource.


<a id="nestedatt--aws"></a>
### Nested Schema for `aws`

Read-Only:

- `default_security_group_id` (String)
- `internet_gateway_id` (String)
- `vpc_id` (String)


<a id="nestedatt--azure"></a>
### Nested Schema for `azure`

Read-Only:

- `local_route_table_id` (String)
- `virtual_network_id` (String)


<a id="nestedatt--gcp"></a>
### Nested Schema for `gcp`

Read-Only:

- `compute_network_id` (String)
- `default_compute_firewall_id` (String)


