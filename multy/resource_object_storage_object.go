package multy

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/multycloud/multy/api/proto/resourcespb"
	"terraform-provider-multy/multy/common"
	"terraform-provider-multy/multy/mtypes"
	"terraform-provider-multy/multy/validators"
)

type ResourceObjectStorageObjectType struct{}

func (r ResourceObjectStorageObjectType) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		MarkdownDescription: "Provides Multy Object Storage Object resource",
		Attributes: map[string]tfsdk.Attribute{
			"id": {
				Type:          types.StringType,
				Computed:      true,
				PlanModifiers: []tfsdk.AttributePlanModifier{tfsdk.UseStateForUnknown()},
			},
			"name": {
				Type:        types.StringType,
				Description: "Name of object storage object",
				Required:    true,
			},
			"object_storage_id": {
				Type:        types.StringType,
				Description: "Id of object storage",
				Required:    true,
			},
			"content": {
				Type:        types.StringType,
				Description: "Content of the object",
				Required:    true,
			},
			"content_type": {
				Type:        types.StringType,
				Description: "Standard MIME type describing the format of the object data",
				Optional:    true,
				Validators:  []tfsdk.AttributeValidator{mtypes.NonEmptyStringValidator},
			},
			"acl": {
				Type:        mtypes.ObjectAclType,
				Description: "Access control for the given object. Can be public_read or private. Defaults to private.",
				Optional:    true,
				Computed:    true,
				Validators:  []tfsdk.AttributeValidator{validators.NewValidator(mtypes.ObjectAclType)},
			},
		},
	}, nil
}

func (r ResourceObjectStorageObjectType) NewResource(_ context.Context, p tfsdk.Provider) (tfsdk.Resource, diag.Diagnostics) {
	return resourceObjectStorageObject{
		p: *(p.(*Provider)),
	}, nil
}

type resourceObjectStorageObject struct {
	p Provider
}

func (r resourceObjectStorageObject) Create(ctx context.Context, req tfsdk.CreateResourceRequest, resp *tfsdk.CreateResourceResponse) {
	if !r.p.Configured {
		resp.Diagnostics.AddError(
			"Provider not configured",
			"The provider hasn't been configured before apply, likely because it depends on an unknown value from another resource. This leads to weird stuff happening, so we'd prefer if you didn't do that. Thanks!",
		)
		return
	}

	// Retrieve values from plan
	var plan ObjectStorageObject
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
	vn, err := c.Client.CreateObjectStorageObject(ctx, &resourcespb.CreateObjectStorageObjectRequest{
		Resource: r.convertResourcePlanToArgs(plan),
	})
	if err != nil {
		resp.Diagnostics.AddError("Error creating object_storage", common.ParseGrpcErrors(err))
		return
	}

	tflog.Trace(ctx, "created object_storage_object", map[string]interface{}{"object_storage_object_id": vn.CommonParameters.ResourceId})

	// Map response body to resource schema attribute
	state := r.convertResponseToResource(vn)

	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r resourceObjectStorageObject) Read(ctx context.Context, req tfsdk.ReadResourceRequest, resp *tfsdk.ReadResourceResponse) {
	// Get current state
	var state ObjectStorageObject
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
	vn, err := r.p.Client.Client.ReadObjectStorageObject(ctx, &resourcespb.ReadObjectStorageObjectRequest{ResourceId: state.Id.Value})
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

func (r resourceObjectStorageObject) Update(ctx context.Context, req tfsdk.UpdateResourceRequest, resp *tfsdk.UpdateResourceResponse) {
	var plan, state ObjectStorageObject
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
	vn, err := c.Client.UpdateObjectStorageObject(ctx, &resourcespb.UpdateObjectStorageObjectRequest{
		// fixme state vs plan
		ResourceId: state.Id.Value,
		Resource:   r.convertResourcePlanToArgs(plan),
	})
	if err != nil {
		resp.Diagnostics.AddError("Error updating object_storage", common.ParseGrpcErrors(err))
		return
	}

	tflog.Trace(ctx, "updated object_storage_object", map[string]interface{}{"object_storage_id": state.Id.Value})

	// Map response body to resource schema attribute & Set state
	state = r.convertResponseToResource(vn)

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r resourceObjectStorageObject) Delete(ctx context.Context, req tfsdk.DeleteResourceRequest, resp *tfsdk.DeleteResourceResponse) {
	var state ObjectStorageObject
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
	_, err = c.Client.DeleteObjectStorageObject(ctx, &resourcespb.DeleteObjectStorageObjectRequest{ResourceId: state.Id.Value})

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

func (r resourceObjectStorageObject) ImportState(ctx context.Context, req tfsdk.ImportResourceStateRequest, resp *tfsdk.ImportResourceStateResponse) {
	// Save the import identifier in the id attribute
	tfsdk.ResourceImportStatePassthroughID(ctx, tftypes.NewAttributePath().WithAttributeName("id"), req, resp)
}

type ObjectStorageObject struct {
	Id              types.String                                         `tfsdk:"id"`
	Name            string                                               `tfsdk:"name"`
	Acl             mtypes.EnumValue[resourcespb.ObjectStorageObjectAcl] `tfsdk:"acl"`
	ObjectStorageId string                                               `tfsdk:"object_storage_id"`
	Content         string                                               `tfsdk:"content"`
	ContentType     types.String                                         `tfsdk:"content_type"`
}

func (r resourceObjectStorageObject) convertResponseToResource(res *resourcespb.ObjectStorageObjectResource) ObjectStorageObject {
	return ObjectStorageObject{
		Id:              types.String{Value: res.CommonParameters.ResourceId},
		Name:            res.Name,
		Acl:             mtypes.ObjectAclType.NewVal(res.Acl),
		ObjectStorageId: res.ObjectStorageId,
		Content:         res.Content,
		ContentType:     common.DefaultToNull[types.String](res.ContentType),
	}
}

func (r resourceObjectStorageObject) convertResourcePlanToArgs(plan ObjectStorageObject) *resourcespb.ObjectStorageObjectArgs {
	return &resourcespb.ObjectStorageObjectArgs{
		Name:            plan.Name,
		Acl:             plan.Acl.Value,
		ObjectStorageId: plan.ObjectStorageId,
		Content:         plan.Content,
		ContentType:     common.NullToDefault[string](plan.ContentType),
	}
}
