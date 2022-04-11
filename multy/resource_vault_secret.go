package multy

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/multycloud/multy/api/proto/resourcespb"
)

type ResourceVaultSecretType struct{}

func (r ResourceVaultSecretType) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		MarkdownDescription: "Provides Multy Object Storage resource",
		Attributes: map[string]tfsdk.Attribute{
			"id": {
				Type:          types.StringType,
				Computed:      true,
				PlanModifiers: []tfsdk.AttributePlanModifier{tfsdk.UseStateForUnknown()},
			},
			"name": {
				Type:        types.StringType,
				Description: "Name of the secret",
				Required:    true,
			},
			"value": {
				Type:        types.StringType,
				Description: "Secret value",
				Required:    true,
			},
			"vault_id": {
				Type:          types.StringType,
				Description:   "Secret value",
				Required:      true,
				PlanModifiers: []tfsdk.AttributePlanModifier{tfsdk.RequiresReplace()},
			},
		},
	}, nil
}

func (r ResourceVaultSecretType) NewResource(_ context.Context, p tfsdk.Provider) (tfsdk.Resource, diag.Diagnostics) {
	return MultyResource[VaultSecret]{
		p:          *(p.(*Provider)),
		createFunc: createVaultSecret,
		updateFunc: updateVaultSecret,
		readFunc:   readVaultSecret,
		deleteFunc: deleteVaultSecret,
	}, nil
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
		ResourceId: plan.Id.Value,
		Resource:   convertFromVaultSecret(plan),
	})
	if err != nil {
		return VaultSecret{}, err
	}
	return convertToVaultSecret(vn), nil
}

func readVaultSecret(ctx context.Context, p Provider, state VaultSecret) (VaultSecret, error) {
	vn, err := p.Client.Client.ReadVaultSecret(ctx, &resourcespb.ReadVaultSecretRequest{
		ResourceId: state.Id.Value,
	})
	if err != nil {
		return VaultSecret{}, err
	}
	return convertToVaultSecret(vn), nil
}

func deleteVaultSecret(ctx context.Context, p Provider, state VaultSecret) error {
	_, err := p.Client.Client.DeleteVaultSecret(ctx, &resourcespb.DeleteVaultSecretRequest{
		ResourceId: state.Id.Value,
	})
	return err
}

type VaultSecret struct {
	Id      types.String `tfsdk:"id"`
	VaultId types.String `tfsdk:"vault_id"`
	Name    types.String `tfsdk:"name"`
	Value   types.String `tfsdk:"value"`
}

func convertToVaultSecret(res *resourcespb.VaultSecretResource) VaultSecret {
	return VaultSecret{
		Id:      types.String{Value: res.CommonParameters.ResourceId},
		VaultId: types.String{Value: res.VaultId},
		Name:    types.String{Value: res.Name},
		Value:   types.String{Value: res.Value},
	}
}

func convertFromVaultSecret(plan VaultSecret) *resourcespb.VaultSecretArgs {
	return &resourcespb.VaultSecretArgs{
		Name:    plan.Name.Value,
		Value:   plan.Value.Value,
		VaultId: plan.VaultId.Value,
	}
}
