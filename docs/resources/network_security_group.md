---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "multy_network_security_group Resource - terraform-provider-multy"
subcategory: ""
description: |-
  Provides Multy Network Security Group resource
---

# multy_network_security_group (Resource)

Provides Multy Network Security Group resource

## Example Usage

```terraform
resource "multy_virtual_network" "vn" {
  name       = "test_nsg"
  cidr_block = "10.0.0.0/16"
  cloud      = "azure"
  location   = "eu_west_1"
}

resource "multy_network_security_group" "nsg" {
  name               = "test_nsg"
  virtual_network_id = multy_virtual_network.vn.id
  cloud              = "azure"
  location           = "eu_west_1"
  rule {
    protocol   = "tcp"
    priority   = 120
    from_port  = 22
    to_port    = 22
    cidr_block = "0.0.0.0/0"
    direction  = "both"
  }
  rule {
    protocol   = "tcp"
    priority   = 130
    from_port  = 443
    to_port    = 444
    cidr_block = "0.0.0.0/0"
    direction  = "both"
  }
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `cloud` (String) Cloud provider to deploy resource into. Accepted values are `aws`, `azure` or `gcp`
- `location` (String) Location to deploy resource into. Read more about regions in [documentation](https://docs.multy.dev/regions)
- `name` (String) Name of Network Security Group
- `virtual_network_id` (String) ID of `virtual_network` resource

### Optional

- `gcp_overrides` (Attributes) GCP-specific attributes that will be set if this resource is deployed in GCP (see [below for nested schema](#nestedatt--gcp_overrides))
- `rule` (Block List) Network rule block definition (see [below for nested schema](#nestedblock--rule))

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


<a id="nestedblock--rule"></a>
### Nested Schema for `rule`

Required:

- `cidr_block` (String) CIDR block of network rule
- `direction` (String) Direction of network rule. Accepted values are `ingress`, `egress` or `both`
- `from_port` (Number) From port of network rule port range. Value must be in between 0 and 65535
- `priority` (Number) Priority of network rule. Value must be in between 0 and 0
- `protocol` (String) Protocol of network rule. Accepted values are `tcp`, `udp` or `icmp`
- `to_port` (Number) To port of network rule port range. Value must be in between 0 and 65535


<a id="nestedatt--aws"></a>
### Nested Schema for `aws`

Read-Only:

- `security_group_id` (String)


<a id="nestedatt--azure"></a>
### Nested Schema for `azure`

Read-Only:

- `network_security_group_id` (String)


<a id="nestedatt--gcp"></a>
### Nested Schema for `gcp`

Read-Only:

- `compute_firewall_ids` (List of String)


