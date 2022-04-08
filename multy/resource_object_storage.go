package multy

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/multycloud/multy/api/proto/commonpb"
	"github.com/multycloud/multy/api/proto/resourcespb"
	"terraform-provider-multy/multy/common"
	"terraform-provider-multy/multy/mtypes"
)

type ResourceObjectStorageType struct{}

func (r ResourceObjectStorageType) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
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
				Description: "Name of Virtual Network",
				Required:    true,
			},
			"versioning": {
				Type:        types.BoolType,
				Description: "If true, versioning will be enabled to `object_storage_object`",
				Optional:    true,
				Computed:    true,
			},
			"cloud":    common.CloudsSchema,
			"location": common.LocationSchema,
		},
	}, nil
}

func (r ResourceObjectStorageType) NewResource(_ context.Context, p tfsdk.Provider) (tfsdk.Resource, diag.Diagnostics) {
	return resourceObjectStorage{
		p: *(p.(*Provider)),
	}, nil
}

type resourceObjectStorage struct {
	p Provider
}

func (r resourceObjectStorage) Create(ctx context.Context, req tfsdk.CreateResourceRequest, resp *tfsdk.CreateResourceResponse) {
	if !r.p.Configured {
		resp.Diagnostics.AddError(
			"Provider not configured",
			"The provider hasn't been configured before apply, likely because it depends on an unknown value from another resource. This leads to weird stuff happening, so we'd prefer if you didn't do that. Thanks!",
		)
		return
	}

	// Retrieve values from plan
	var plan ObjectStorage
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	c := r.p.Client
	ctx, err := c.AddHeaders(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error communicating with server", err.Error())
		return
	}

	// Create new order from plan values
	vn, err := c.Client.CreateObjectStorage(ctx, &resourcespb.CreateObjectStorageRequest{
		Resource: r.convertResourcePlanToArgs(plan),
	})
	if err != nil {
		resp.Diagnostics.AddError("Error creating object_storage", common.ParseGrpcErrors(err))
		return
	}

	tflog.Trace(ctx, "created object_storage", map[string]interface{}{"object_storage_id": vn.CommonParameters.ResourceId})

	// Map response body to resource schema attribute
	state := r.convertResponseToResource(vn)

	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r resourceObjectStorage) Read(ctx context.Context, req tfsdk.ReadResourceRequest, resp *tfsdk.ReadResourceResponse) {
	// Get current state
	var state ObjectStorage
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	c := r.p.Client
	ctx, err := c.AddHeaders(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error communicating with server", err.Error())
		return
	}

	// Get object_storage from API and then update what is in state from what the API returns
	vn, err := r.p.Client.Client.ReadObjectStorage(ctx, &resourcespb.ReadObjectStorageRequest{ResourceId: state.Id.Value})
	if err != nil {
		resp.Diagnostics.AddError("Error getting object_storage", common.ParseGrpcErrors(err))
		return
	}

	// Map response body to resource schema attribute & Set state
	state = r.convertResponseToResource(vn)
	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r resourceObjectStorage) Update(ctx context.Context, req tfsdk.UpdateResourceRequest, resp *tfsdk.UpdateResourceResponse) {
	var plan, state ObjectStorage
	// Get plan values
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	// Get current state
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	c := r.p.Client
	ctx, err := c.AddHeaders(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error communicating with server", err.Error())
		return
	}

	// Update object_storage
	vn, err := c.Client.UpdateObjectStorage(ctx, &resourcespb.UpdateObjectStorageRequest{
		// fixme state vs plan
		ResourceId: state.Id.Value,
		Resource:   r.convertResourcePlanToArgs(plan),
	})
	if err != nil {
		resp.Diagnostics.AddError("Error updating object_storage", common.ParseGrpcErrors(err))
		return
	}

	tflog.Trace(ctx, "updated object_storage", map[string]interface{}{"object_storage_id": state.Id.Value})

	// Map response body to resource schema attribute & Set state
	state = r.convertResponseToResource(vn)

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r resourceObjectStorage) Delete(ctx context.Context, req tfsdk.DeleteResourceRequest, resp *tfsdk.DeleteResourceResponse) {
	var state ObjectStorage
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	c := r.p.Client
	ctx, err := c.AddHeaders(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error communicating with server", err.Error())
		return
	}

	// Delete object_storage
	_, err = c.Client.DeleteObjectStorage(ctx, &resourcespb.DeleteObjectStorageRequest{ResourceId: state.Id.Value})

	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting object_storage",
			common.ParseGrpcErrors(err),
		)
		return
	}

	// Remove resource from state
	resp.State.RemoveResource(ctx)
}

func (r resourceObjectStorage) ImportState(ctx context.Context, req tfsdk.ImportResourceStateRequest, resp *tfsdk.ImportResourceStateResponse) {
	// Save the import identifier in the id attribute
	tfsdk.ResourceImportStatePassthroughID(ctx, tftypes.NewAttributePath().WithAttributeName("id"), req, resp)
}

type ObjectStorage struct {
	Id         types.String                             `tfsdk:"id"`
	Name       types.String                             `tfsdk:"name"`
	Versioning types.Bool                               `tfsdk:"versioning"`
	Cloud      mtypes.EnumValue[commonpb.CloudProvider] `tfsdk:"cloud"`
	Location   mtypes.EnumValue[commonpb.Location]      `tfsdk:"location"`
}

func (r resourceObjectStorage) convertResponseToResource(res *resourcespb.ObjectStorageResource) ObjectStorage {
	return ObjectStorage{
		Id:         types.String{Value: res.CommonParameters.ResourceId},
		Name:       types.String{Value: res.Name},
		Versioning: types.Bool{Value: res.Versioning},
		Cloud:      mtypes.CloudType.NewVal(res.CommonParameters.CloudProvider),
		Location:   mtypes.LocationType.NewVal(res.CommonParameters.Location),
	}
}

func (r resourceObjectStorage) convertResourcePlanToArgs(plan ObjectStorage) *resourcespb.ObjectStorageArgs {
	return &resourcespb.ObjectStorageArgs{
		CommonParameters: &commonpb.ResourceCommonArgs{
			Location:      plan.Location.Value,
			CloudProvider: plan.Cloud.Value,
		},
		Name:       plan.Name.Value,
		Versioning: plan.Versioning.Value,
	}
}
