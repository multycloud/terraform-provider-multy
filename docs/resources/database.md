---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "multy_database Resource - terraform-provider-multy"
subcategory: ""
description: |-
  Provides Multy Database resource
---

# multy_database (Resource)

Provides Multy Database resource

## Example Usage

```terraform
resource "random_password" "password" {
  length = 16
}

resource "multy_database" "example_db" {
  cloud          = "aws"
  location       = "us_east_1"
  storage_gb     = 10
  name           = "multydb"
  engine         = "mysql"
  engine_version = "5.7"
  username       = "multyadmin"
  password       = random_password.password.result
  size           = "micro"
  subnet_ids     = multy_subnet.subnet.id
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `cloud` (String) Cloud provider to deploy resource into. Accepted values are `aws`, `azure` or `gcp`
- `engine` (String) Database engine. Available values are [mysql postgres mariadb]
- `engine_version` (String) Engine version
- `location` (String) Location to deploy resource into. Read more about regions in [documentation](https://docs.multy.dev/regions)
- `name` (String) Name of the database. If cloud is azure, name needs to be unique globally.
- `password` (String, Sensitive) Password for the database user
- `size` (String) Database size. Available values are [micro medium small]
- `storage_gb` (Number) Size of database storage in gigabytes
- `subnet_id` (String) Subnet associated with this database.
- `username` (String) Username for the database user

### Optional

- `gcp_overrides` (Attributes) GCP-specific attributes that will be set if this resource is deployed in GCP (see [below for nested schema](#nestedatt--gcp_overrides))

### Read-Only

- `aws` (Object) AWS-specific ids of the underlying generated resources (see [below for nested schema](#nestedatt--aws))
- `azure` (Object) Azure-specific ids of the underlying generated resources (see [below for nested schema](#nestedatt--azure))
- `connection_username` (String) The username to connect to the database.
- `gcp` (Object) GCP-specific ids of the underlying generated resources (see [below for nested schema](#nestedatt--gcp))
- `hostname` (String) The hostname of the RDS instance.
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

- `db_instance_id` (String)
- `db_subnet_group_id` (String)
- `default_network_security_group_id` (String)


<a id="nestedatt--azure"></a>
### Nested Schema for `azure`

Read-Only:

- `database_server_id` (String)


<a id="nestedatt--gcp"></a>
### Nested Schema for `gcp`

Read-Only:

- `sql_database_instance_id` (String)


