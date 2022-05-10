## Overview

This repository is the brain of Multy which translates Multy resources into cloud specific infrastructure.

The Multy Engine is a GRPC server that translates a multy infrastructure resource request into cloud specific resources.

## File Structure

```bash
.
├── docs # automatically generated documentation
│    └── resources
├── examples # example snippets for docs 
│    ├── data
│    ├── provider
│    └── resources
├── multy
│    ├── common
│    ├── data
│    ├── mtypes
│    ├── validators # parameter validation functions
│    └── *.go # resource definition
├── tests
│    ├── data
│    ├── provider
│    ├── resources # resource specific tests
│    └── use_case
│        └── web_app # full web_app example 
└── tools
```

## Technologies

- Golang (>1.18)
- Terraform Plugin Framework ([link](https://www.terraform.io/plugin/framework))

## Running locally

0. Setup local server

If you want to locally run a server
Follow the overview guide from

2. Clone repository

```bash
git clone https://github.com/multycloud/terraform-provider-multy.git
cd terraform-provider-multy
```

2. Setup project

```bash
# adjust $OS_ARCH
make build
```

- Set environment variables

```bash
export MULY_API_KEY=#API_KEY_IN_LOCAL_DB#
```

Configure your AWS or Azure accounts as per the [Getting Started docs](https://docs.multy.dev/getting-started)

3. Create a configuration

You can use the configurations from `./tests/resources/xx.go` to deploy resources or create your own. To use the local
server/provider, you need to configure the provider as it follows:

```hcl
terraform {
  required_providers {
    multy = {
      version = "0.0.1"                    # instead of using the hashicorp registry
      source  = "hashicorp.com/dev/multy"  # this is use the local configuration from `make`
    }
  }
}

provider "multy" {
  server_endpoint = "localhost:8000" # run local server
  aws             = {}
  azure           = {}
}
```

4. Deploy configuration

Run the following commands to apply your configuration

**Note: if not on a free tier, these command will deploy resources into your account which might incur a cost.**

To avoid deploying resources, run the server locally with the `--dry-run` flag.

```bash
terraform init    # download the terraform providers 
terraform plan    # outputs what would be deployed if configuration is applied
terraform apply   # deploy infrastructure
```

To delete resources created, run:

```bash
terraform destroy
```

5. Run tests

To run all tests:

```bash
TF_ACC=1 TF_VAR_cloud=aws USER_SECRET_PREFIX=#MULTY_API_KEY_FROM_DB# go test ./multy/... -v
```

To run a single test:

```bash
TF_ACC=1 TF_VAR_cloud=aws USER_SECRET_PREFIX=#MULTY_API_KEY_FROM_DB# go test ./multy/... -v -run Acc/object_storage_object
```

6. Generate docs

When you make a change to a resource or its documentation, you need to regenerate the Terraform docs by running:

```bash
make docs
```

You need to commit these changes into your PR.