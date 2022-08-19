package multy

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/multycloud/multy/api/proto/commonpb"
	"github.com/multycloud/multy/api/proto/resourcespb"
	"terraform-provider-multy/multy/common"
	"terraform-provider-multy/multy/mtypes"
)

type ResourceVaultType struct{}

var vaultAzureOutputs = map[string]attr.Type{
	"key_vault_id": types.StringType,
}

func (r ResourceVaultType) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		MarkdownDescription: "Provides Multy Vault resource",
		Attributes: map[string]tfsdk.Attribute{
			"id": {
				Type:          types.StringType,
				Computed:      true,
				PlanModifiers: []tfsdk.AttributePlanModifier{resource.UseStateForUnknown()},
			},
			"resource_group_id": {
				Type:          types.StringType,
				Computed:      true,
				PlanModifiers: []tfsdk.AttributePlanModifier{resource.UseStateForUnknown()},
			},
			"name": {
				Type:          types.StringType,
				Description:   "Name of vault resource",
				Required:      true,
				PlanModifiers: []tfsdk.AttributePlanModifier{resource.RequiresReplace()},
			},
			"gcp_overrides": {
				Description: "GCP-specific attributes that will be set if this resource is deployed in GCP",
				Attributes: tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
					"project": {
						Type:          types.StringType,
						Description:   fmt.Sprintf("The project to use for this resource."),
						Optional:      true,
						Computed:      true,
						PlanModifiers: []tfsdk.AttributePlanModifier{common.RequiresReplaceIfCloudEq("gcp"), resource.UseStateForUnknown()},
						Validators:    []tfsdk.AttributeValidator{mtypes.NonEmptyStringValidator},
					},
				}),
				Optional: true,
				Computed: true,
			},
			"azure": {
				Description: "Azure-specific ids of the underlying generated resources",
				Type:        types.ObjectType{AttrTypes: vaultAzureOutputs},
				Computed:    true,
			},
			"cloud":           common.CloudsSchema,
			"location":        common.LocationSchema,
			"resource_status": common.ResourceStatusSchema,
		},
	}, nil
}

func (r ResourceVaultType) NewResource(_ context.Context, p provider.Provider) (resource.Resource, diag.Diagnostics) {
	return MultyResource[Vault]{
		p:          *(p.(*Provider)),
		createFunc: createVault,
		updateFunc: updateVault,
		readFunc:   readVault,
		deleteFunc: deleteVault,
	}, nil
}

func createVault(ctx context.Context, p Provider, plan Vault) (Vault, error) {
	vn, err := p.Client.Client.CreateVault(ctx, &resourcespb.CreateVaultRequest{
		Resource: convertFromVault(plan),
	})
	if err != nil {
		return Vault{}, err
	}
	return convertToVault(vn), nil
}

func updateVault(ctx context.Context, p Provider, plan Vault) (Vault, error) {
	vn, err := p.Client.Client.UpdateVault(ctx, &resourcespb.UpdateVaultRequest{
		ResourceId: plan.Id.Value,
		Resource:   convertFromVault(plan),
	})
	if err != nil {
		return Vault{}, err
	}
	return convertToVault(vn), nil
}

func readVault(ctx context.Context, p Provider, state Vault) (Vault, error) {
	vn, err := p.Client.Client.ReadVault(ctx, &resourcespb.ReadVaultRequest{
		ResourceId: state.Id.Value,
	})
	if err != nil {
		return Vault{}, err
	}
	return convertToVault(vn), nil
}

func deleteVault(ctx context.Context, p Provider, state Vault) error {
	_, err := p.Client.Client.DeleteVault(ctx, &resourcespb.DeleteVaultRequest{
		ResourceId: state.Id.Value,
	})
	return err
}

type Vault struct {
	Id              types.String                             `tfsdk:"id"`
	Name            types.String                             `tfsdk:"name"`
	Cloud           mtypes.EnumValue[commonpb.CloudProvider] `tfsdk:"cloud"`
	Location        mtypes.EnumValue[commonpb.Location]      `tfsdk:"location"`
	ResourceGroupId types.String                             `tfsdk:"resource_group_id"`

	GcpOverridesObject types.Object `tfsdk:"gcp_overrides"`
	AzureOutputs       types.Object `tfsdk:"azure"`
	ResourceStatus     types.Map    `tfsdk:"resource_status"`
}

func convertToVault(res *resourcespb.VaultResource) Vault {
	return Vault{
		Id:                 types.String{Value: res.CommonParameters.ResourceId},
		Name:               types.String{Value: res.Name},
		Cloud:              mtypes.CloudType.NewVal(res.CommonParameters.CloudProvider),
		Location:           mtypes.LocationType.NewVal(res.CommonParameters.Location),
		ResourceGroupId:    types.String{Value: res.CommonParameters.ResourceGroupId},
		GcpOverridesObject: convertToVaultGcpOverrides(res.GcpOverride).GcpOverridesToObj(),
		AzureOutputs: common.OptionallyObj(res.AzureOutputs, types.Object{
			Attrs: map[string]attr.Value{
				"key_vault_id": common.DefaultToNull[types.String](res.GetAzureOutputs().GetKeyVaultId()),
			},
			AttrTypes: vaultAzureOutputs,
		}),
		ResourceStatus: common.GetResourceStatus(res.CommonParameters.GetResourceStatus()),
	}
}

func convertFromVault(plan Vault) *resourcespb.VaultArgs {
	return &resourcespb.VaultArgs{
		CommonParameters: &commonpb.ResourceCommonArgs{
			ResourceGroupId: plan.ResourceGroupId.Value,
			Location:        plan.Location.Value,
			CloudProvider:   plan.Cloud.Value,
		},
		Name:        plan.Name.Value,
		GcpOverride: convertFromVaultGcpOverrides(plan.GetGcpOverrides()),
	}
}

func (v Vault) UpdatePlan(_ context.Context, config Vault, p Provider) (Vault, []path.Path) {
	if config.Cloud.Value != commonpb.CloudProvider_GCP || p.Client.Gcp == nil {
		return v, nil
	}
	var requiresReplace []path.Path
	gcpOverrides := v.GetGcpOverrides()
	if o := config.GetGcpOverrides(); o == nil || o.Project.Unknown {
		if gcpOverrides == nil {
			gcpOverrides = &VaultGcpOverrides{}
		}

		gcpOverrides.Project = types.String{
			Unknown: false,
			Null:    false,
			Value:   p.Client.Gcp.Project,
		}

		v.GcpOverridesObject = gcpOverrides.GcpOverridesToObj()
		requiresReplace = append(requiresReplace, path.Root("gcp_overrides").AtName("project"))
	}
	return v, requiresReplace
}

func (v Vault) GetGcpOverrides() (o *VaultGcpOverrides) {
	if v.GcpOverridesObject.Null || v.GcpOverridesObject.Unknown {
		return
	}
	o = &VaultGcpOverrides{
		Project: v.GcpOverridesObject.Attrs["project"].(types.String),
	}
	return
}

func (o *VaultGcpOverrides) GcpOverridesToObj() types.Object {
	result := types.Object{
		Unknown: false,
		Null:    false,
		AttrTypes: map[string]attr.Type{
			"project": types.StringType,
		},
		Attrs: map[string]attr.Value{
			"project": types.String{Null: true},
		},
	}
	if o != nil {
		result.Attrs = map[string]attr.Value{
			"project": o.Project,
		}
	}

	return result
}

type VaultGcpOverrides struct {
	Project types.String
}

func convertFromVaultGcpOverrides(ref *VaultGcpOverrides) *resourcespb.VaultGcpOverride {
	if ref == nil {
		return nil
	}

	return &resourcespb.VaultGcpOverride{Project: ref.Project.Value}
}

func convertToVaultGcpOverrides(ref *resourcespb.VaultGcpOverride) *VaultGcpOverrides {
	if ref == nil {
		return nil
	}

	return &VaultGcpOverrides{Project: common.DefaultToNull[types.String](ref.Project)}
}
