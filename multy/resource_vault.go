package multy

import (
	"context"
	"terraform-provider-multy/multy/common"
	"terraform-provider-multy/multy/mtypes"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/multycloud/multy/api/proto/commonpb"
	"github.com/multycloud/multy/api/proto/resourcespb"
)

type ResourceVaultType struct{}

func (r ResourceVaultType) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		MarkdownDescription: "Provides Multy Vault resource",
		Attributes: map[string]tfsdk.Attribute{
			"id": {
				Type:          types.StringType,
				Computed:      true,
				PlanModifiers: []tfsdk.AttributePlanModifier{tfsdk.UseStateForUnknown()},
			},
			"name": {
				Type:          types.StringType,
				Description:   "Name of vault resource",
				Required:      true,
				PlanModifiers: []tfsdk.AttributePlanModifier{tfsdk.RequiresReplace()},
			},
			"cloud":    common.CloudsSchema,
			"location": common.LocationSchema,
		},
	}, nil
}

func (r ResourceVaultType) NewResource(_ context.Context, p tfsdk.Provider) (tfsdk.Resource, diag.Diagnostics) {
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
	Id       types.String                             `tfsdk:"id"`
	Name     types.String                             `tfsdk:"name"`
	Cloud    mtypes.EnumValue[commonpb.CloudProvider] `tfsdk:"cloud"`
	Location mtypes.EnumValue[commonpb.Location]      `tfsdk:"location"`
}

func convertToVault(res *resourcespb.VaultResource) Vault {
	return Vault{
		Id:       types.String{Value: res.CommonParameters.ResourceId},
		Name:     types.String{Value: res.Name},
		Cloud:    mtypes.CloudType.NewVal(res.CommonParameters.CloudProvider),
		Location: mtypes.LocationType.NewVal(res.CommonParameters.Location),
	}
}

func convertFromVault(plan Vault) *resourcespb.VaultArgs {
	return &resourcespb.VaultArgs{
		CommonParameters: &commonpb.ResourceCommonArgs{
			Location:      plan.Location.Value,
			CloudProvider: plan.Cloud.Value,
		},
		Name: plan.Name.Value,
	}
}
