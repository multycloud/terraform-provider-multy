package multy

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/multycloud/multy/api/proto/resourcespb"
	"terraform-provider-multy/multy/common"
)

type ResourceVaultSecretType struct{}

var vaultSecretAwsOutputs = map[string]attr.Type{
	"ssm_parameter_arn": types.StringType,
}

var vaultSecretAzureOutputs = map[string]attr.Type{
	"key_vault_secret_id": types.StringType,
}

var vaultSecretGcpOutputs = map[string]attr.Type{
	"secret_manager_secret_id":         types.StringType,
	"secret_manager_secret_version_id": types.StringType,
}

var vaultSecretSchema = tfsdk.Schema{
	MarkdownDescription: "Provides Multy Object Storage resource",
	Attributes: map[string]tfsdk.Attribute{
		"id": {
			Type:          types.StringType,
			Computed:      true,
			PlanModifiers: []tfsdk.AttributePlanModifier{resource.UseStateForUnknown()},
		},
		"name": {
			Type:          types.StringType,
			Description:   "Name of the secret",
			Required:      true,
			PlanModifiers: []tfsdk.AttributePlanModifier{resource.RequiresReplace()},
		},
		"value": {
			Type:        types.StringType,
			Description: "Secret value",
			Required:    true,
		},
		"vault_id": {
			Type:          types.StringType,
			Description:   "Id of `vault` to store the secret in",
			Required:      true,
			PlanModifiers: []tfsdk.AttributePlanModifier{resource.RequiresReplace()},
		},
		"aws": {
			Description: "AWS-specific ids of the underlying generated resources",
			Type:        types.ObjectType{AttrTypes: vaultSecretAwsOutputs},
			Computed:    true,
		},
		"azure": {
			Description: "Azure-specific ids of the underlying generated resources",
			Type:        types.ObjectType{AttrTypes: vaultSecretAzureOutputs},
			Computed:    true,
		},
		"gcp": {
			Description: "GCP-specific ids of the underlying generated resources",
			Type:        types.ObjectType{AttrTypes: vaultSecretGcpOutputs},
			Computed:    true,
		},
		"resource_status": common.ResourceStatusSchema,
	},
}

func (r ResourceVaultSecretType) NewResource(_ context.Context, p provider.Provider) resource.Resource {
	return MultyResource[VaultSecret]{
		p:          *(p.(*Provider)),
		createFunc: createVaultSecret,
		updateFunc: updateVaultSecret,
		readFunc:   readVaultSecret,
		deleteFunc: deleteVaultSecret,
		name:       "multy_vault_secret",
		schema:     vaultSecretSchema,
	}
}

func createVaultSecret(ctx context.Context, p Provider, plan VaultSecret) (VaultSecret, error) {
	vn, err := p.Client.Client.CreateVaultSecret(ctx, &resourcespb.CreateVaultSecretRequest{
		Resource: convertFromVaultSecret(plan),
	})
	if err != nil {
		return VaultSecret{}, err
	}
	return convertToVaultSecret(vn), nil
}

func updateVaultSecret(ctx context.Context, p Provider, plan VaultSecret) (VaultSecret, error) {
	vn, err := p.Client.Client.UpdateVaultSecret(ctx, &resourcespb.UpdateVaultSecretRequest{
		ResourceId: plan.Id.ValueString(),
		Resource:   convertFromVaultSecret(plan),
	})
	if err != nil {
		return VaultSecret{}, err
	}
	return convertToVaultSecret(vn), nil
}

func readVaultSecret(ctx context.Context, p Provider, state VaultSecret) (VaultSecret, error) {
	vn, err := p.Client.Client.ReadVaultSecret(ctx, &resourcespb.ReadVaultSecretRequest{
		ResourceId: state.Id.ValueString(),
	})
	if err != nil {
		return VaultSecret{}, err
	}
	return convertToVaultSecret(vn), nil
}

func deleteVaultSecret(ctx context.Context, p Provider, state VaultSecret) error {
	_, err := p.Client.Client.DeleteVaultSecret(ctx, &resourcespb.DeleteVaultSecretRequest{
		ResourceId: state.Id.ValueString(),
	})
	return err
}

type VaultSecret struct {
	Id             types.String `tfsdk:"id"`
	VaultId        types.String `tfsdk:"vault_id"`
	Name           types.String `tfsdk:"name"`
	Value          types.String `tfsdk:"value"`
	AwsOutputs     types.Object `tfsdk:"aws"`
	AzureOutputs   types.Object `tfsdk:"azure"`
	GcpOutputs     types.Object `tfsdk:"gcp"`
	ResourceStatus types.Map    `tfsdk:"resource_status"`
}

func convertToVaultSecret(res *resourcespb.VaultSecretResource) VaultSecret {
	return VaultSecret{
		Id:      types.StringValue(res.CommonParameters.ResourceId),
		VaultId: types.StringValue(res.VaultId),
		Name:    types.StringValue(res.Name),
		Value:   types.StringValue(res.Value),
		AwsOutputs: common.OptionallyObj(res.AwsOutputs, vaultSecretAwsOutputs, map[string]attr.Value{
			"ssm_parameter_arn": common.DefaultToNull[types.String](res.GetAwsOutputs().GetSsmParameterArn()),
		}),
		AzureOutputs: common.OptionallyObj(res.AzureOutputs, vaultSecretAzureOutputs, map[string]attr.Value{
			"key_vault_secret_id": common.DefaultToNull[types.String](res.GetAzureOutputs().GetKeyVaultSecretId()),
		}),
		GcpOutputs: common.OptionallyObj(res.GcpOutputs, vaultSecretGcpOutputs, map[string]attr.Value{
			"secret_manager_secret_id":         common.DefaultToNull[types.String](res.GetGcpOutputs().GetSecretManagerSecretId()),
			"secret_manager_secret_version_id": common.DefaultToNull[types.String](res.GetGcpOutputs().GetSecretManagerSecretVersionId()),
		}),
		ResourceStatus: common.GetResourceStatus(res.CommonParameters.GetResourceStatus()),
	}
}

func convertFromVaultSecret(plan VaultSecret) *resourcespb.VaultSecretArgs {
	return &resourcespb.VaultSecretArgs{
		Name:    plan.Name.ValueString(),
		Value:   plan.Value.ValueString(),
		VaultId: plan.VaultId.ValueString(),
	}
}
