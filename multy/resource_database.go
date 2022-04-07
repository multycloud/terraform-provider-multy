package multy

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/multycloud/multy/api/proto/commonpb"
	"github.com/multycloud/multy/api/proto/resourcespb"
	"terraform-provider-multy/multy/common"
	"terraform-provider-multy/multy/mtypes"
	"terraform-provider-multy/multy/validators"
)

type ResourceDatabaseType struct{}

func (r ResourceDatabaseType) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Attributes: map[string]tfsdk.Attribute{
			"id": {
				Type:          types.StringType,
				Computed:      true,
				PlanModifiers: []tfsdk.AttributePlanModifier{tfsdk.UseStateForUnknown()},
			},
			"name": {
				Type:        types.StringType,
				Description: "Name of the database",
				Required:    true,
			},
			"engine": {
				Type:        mtypes.DbEngineType,
				Description: fmt.Sprintf("Database engine. Available values are %v", mtypes.DbEngineType.GetAllValues()),
				Required:    true,
				Validators:  []tfsdk.AttributeValidator{validators.NewValidator(mtypes.DbEngineType)},
			},
			"engine_version": {
				Type:        types.StringType,
				Description: "Engine version",
				Required:    true,
			},
			"storage_gb": {
				Type:        types.Int64Type,
				Description: "Database engine. Available values are",
				Required:    true,
			},
			"size": {
				Type:        mtypes.DbSizeType,
				Description: fmt.Sprintf("Database size. Available values are %v", mtypes.DbSizeType.GetAllValues()),
				Required:    true,
				Validators:  []tfsdk.AttributeValidator{validators.NewValidator(mtypes.DbSizeType)},
			},
			"username": {
				Type:        types.StringType,
				Description: "Username for the database user",
				Required:    true,
			},
			"password": {
				Type:        types.StringType,
				Description: "Password for the database user",
				Required:    true,
			},
			"subnet_ids": {
				Type:        types.ListType{ElemType: types.StringType},
				Description: "Subnets associated with this database. At least 2 in different availability zones are required.",
				Required:    true,
			},

			"cloud":    common.CloudsSchema,
			"location": common.LocationSchema,

			"host": {
				Type:        types.StringType,
				Description: "Database endpoint to connect to",
				Computed:    true,
			},
		},
	}, nil
}

func (r ResourceDatabaseType) NewResource(_ context.Context, p tfsdk.Provider) (tfsdk.Resource, diag.Diagnostics) {
	return resourceDatabase{
		p: *(p.(*Provider)),
	}, nil
}

type resourceDatabase struct {
	p Provider
}

func (r resourceDatabase) Create(ctx context.Context, req tfsdk.CreateResourceRequest, resp *tfsdk.CreateResourceResponse) {
	if !r.p.Configured {
		resp.Diagnostics.AddError(
			"Provider not configured",
			"The provider hasn't been configured before apply, likely because it depends on an unknown value from another resource. This leads to weird stuff happening, so we'd prefer if you didn't do that. Thanks!",
		)
		return
	}

	// Retrieve values from plan
	var plan Database
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
	vn, err := c.Client.CreateDatabase(ctx, &resourcespb.CreateDatabaseRequest{
		Resource: r.convertResourcePlanToArgs(plan),
	})
	if err != nil {
		resp.Diagnostics.AddError("Error creating network_interface", common.ParseGrpcErrors(err))
		return
	}

	tflog.Trace(ctx, "created network_interface", map[string]interface{}{"network_interface_id": vn.CommonParameters.ResourceId})

	// Map response body to resource schema attribute
	state := r.convertResponseToResource(vn)

	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r resourceDatabase) Read(ctx context.Context, req tfsdk.ReadResourceRequest, resp *tfsdk.ReadResourceResponse) {
	// Get current state
	var state Database
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

	// Get network_interface from API and then update what is in state from what the API returns
	vn, err := r.p.Client.Client.ReadDatabase(ctx, &resourcespb.ReadDatabaseRequest{ResourceId: state.Id.Value})
	if err != nil {
		resp.Diagnostics.AddError("Error getting network_interface", common.ParseGrpcErrors(err))
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

func (r resourceDatabase) Update(ctx context.Context, req tfsdk.UpdateResourceRequest, resp *tfsdk.UpdateResourceResponse) {
	var plan, state Database
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

	// Update network_interface
	vn, err := c.Client.UpdateDatabase(ctx, &resourcespb.UpdateDatabaseRequest{
		// fixme state vs plan
		ResourceId: state.Id.Value,
		Resource:   r.convertResourcePlanToArgs(plan),
	})
	if err != nil {
		resp.Diagnostics.AddError("Error updating network_interface", common.ParseGrpcErrors(err))
		return
	}

	tflog.Trace(ctx, "updated network_interface", map[string]interface{}{"network_interface_id": state.Id.Value})

	// Map response body to resource schema attribute & Set state
	state = r.convertResponseToResource(vn)

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r resourceDatabase) Delete(ctx context.Context, req tfsdk.DeleteResourceRequest, resp *tfsdk.DeleteResourceResponse) {
	var state Database
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

	// Delete network_interface
	_, err = c.Client.DeleteDatabase(ctx, &resourcespb.DeleteDatabaseRequest{ResourceId: state.Id.Value})

	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting network_interface",
			common.ParseGrpcErrors(err),
		)
		return
	}

	// Remove resource from state
	resp.State.RemoveResource(ctx)
}

func (r resourceDatabase) ImportState(ctx context.Context, req tfsdk.ImportResourceStateRequest, resp *tfsdk.ImportResourceStateResponse) {
	// Save the import identifier in the id attribute
	tfsdk.ResourceImportStatePassthroughID(ctx, tftypes.NewAttributePath().WithAttributeName("id"), req, resp)
}

type Database struct {
	Id            types.String                                 `tfsdk:"id"`
	Name          types.String                                 `tfsdk:"name"`
	Engine        mtypes.EnumValue[resourcespb.DatabaseEngine] `tfsdk:"engine"`
	EngineVersion types.String                                 `tfsdk:"engine_version"`
	StorageGb     types.Int64                                  `tfsdk:"storage_gb"`
	Size          mtypes.EnumValue[commonpb.DatabaseSize_Enum] `tfsdk:"size"`
	Username      types.String                                 `tfsdk:"username"`
	Password      types.String                                 `tfsdk:"password"`
	SubnetIds     []types.String                               `tfsdk:"subnet_ids"`
	Cloud         mtypes.EnumValue[commonpb.CloudProvider]     `tfsdk:"cloud"`
	Location      mtypes.EnumValue[commonpb.Location]          `tfsdk:"location"`
	Host          types.String                                 `tfsdk:"host"`
}

func (r resourceDatabase) convertResponseToResource(res *resourcespb.DatabaseResource) Database {
	return Database{
		Id:            types.String{Value: res.CommonParameters.ResourceId},
		Name:          types.String{Value: res.Name},
		Engine:        mtypes.DbEngineType.NewVal(res.Engine),
		EngineVersion: types.String{Value: res.EngineVersion},
		StorageGb:     types.Int64{Value: res.StorageGb},
		Size:          mtypes.DbSizeType.NewVal(res.Size),
		Username:      types.String{Value: res.Username},
		Password:      types.String{Value: res.Password},
		SubnetIds:     common.TypesStringToStringSlice(res.SubnetIds),
		Cloud:         mtypes.CloudType.NewVal(res.CommonParameters.CloudProvider),
		Location:      mtypes.LocationType.NewVal(res.CommonParameters.Location),
		Host:          types.String{Value: res.Host},
	}
}

func (r resourceDatabase) convertResourcePlanToArgs(plan Database) *resourcespb.DatabaseArgs {
	return &resourcespb.DatabaseArgs{
		Name:          plan.Name.Value,
		Engine:        plan.Engine.Value,
		EngineVersion: plan.EngineVersion.Value,
		StorageGb:     plan.StorageGb.Value,
		Size:          plan.Size.Value,
		Username:      plan.Username.Value,
		Password:      plan.Password.Value,
		SubnetIds:     common.StringSliceToTypesString(plan.SubnetIds),
		CommonParameters: &commonpb.ResourceCommonArgs{
			Location:      plan.Location.Value,
			CloudProvider: plan.Cloud.Value,
		},
	}
}
