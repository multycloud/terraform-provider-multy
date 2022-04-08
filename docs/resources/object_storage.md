---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "multy_object_storage Resource - terraform-provider-multy"
subcategory: ""
description: |-
  
---

# multy_object_storage (Resource)





<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `cloud` (String) CloudType provider to deploy resource into. Accepted values are `aws`, `azure`,
- `location` (String) LocationType to deploy resource into. Accepted values are `us_east`, `ireland`, `uk`,
- `name` (String) Name of Virtual Network

### Optional

- `versioning` (Boolean) If true, versioning will be enabled to `object_storage_object`

### Read-Only

- `id` (String) The ID of this resource.

