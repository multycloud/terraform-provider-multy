package multy

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"terraform-provider-multy/multy/common"
)

type MultyResource[T any] struct {
	p          Provider
	createFunc func(ctx context.Context, p Provider, plan T) (T, error)
	updateFunc func(ctx context.Context, p Provider, plan T) (T, error)
	readFunc   func(ctx context.Context, p Provider, state T) (T, error)
	deleteFunc func(ctx context.Context, p Provider, state T) error
}

func (r MultyResource[T]) Create(ctx context.Context, req tfsdk.CreateResourceRequest, resp *tfsdk.CreateResourceResponse) {
	if !r.p.Configured {
		resp.Diagnostics.AddError(
			"Provider not configured",
			"The provider hasn't been configured before apply, likely because it depends on an unknown value from another resource. This leads to weird stuff happening, so we'd prefer if you didn't do that. Thanks!",
		)
		return
	}

	// Retrieve values from plan
	plan := new(T)
	diags := req.Plan.Get(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	c := r.p.Client
	ctx, err := c.AddHeaders(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error encoding credentials", err.Error())
		return
	}

	state, err := r.createFunc(ctx, r.p, *plan)

	if err != nil {
		resp.Diagnostics.AddError("Error creating resource", common.ParseGrpcErrors(err))
		return
	}

	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r MultyResource[T]) Read(ctx context.Context, req tfsdk.ReadResourceRequest, resp *tfsdk.ReadResourceResponse) {
	c := r.p.Client
	ctx, err := c.AddHeaders(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error encoding credentials", err.Error())
		return
	}

	// Get current state
	state := new(T)
	diags := req.State.Get(ctx, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	newState, err := r.readFunc(ctx, r.p, *state)
	if s, ok := status.FromError(err); ok && s.Code() == codes.NotFound {
		tflog.Info(ctx, "Resource doesn't exist, deleting")
		diags = resp.State.Set(ctx, newState)
		resp.Diagnostics.Append(diags...)
		// Remove MultyResource from state
		resp.State.RemoveResource(ctx)
	} else if err != nil {
		resp.Diagnostics.AddError("Error reading resource", common.ParseGrpcErrors(err))
		return
	} else {
		diags = resp.State.Set(ctx, newState)
		resp.Diagnostics.Append(diags...)
	}
}

func (r MultyResource[T]) Update(ctx context.Context, req tfsdk.UpdateResourceRequest, resp *tfsdk.UpdateResourceResponse) {
	plan := new(T)
	// Get plan values
	resp.Diagnostics.Append(req.Plan.Get(ctx, plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	c := r.p.Client
	ctx, err := c.AddHeaders(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error encoding credentials", err.Error())
		return
	}

	state, err := r.updateFunc(ctx, r.p, *plan)

	if err != nil {
		resp.Diagnostics.AddError("Error updating resource", common.ParseGrpcErrors(err))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r MultyResource[T]) Delete(ctx context.Context, req tfsdk.DeleteResourceRequest, resp *tfsdk.DeleteResourceResponse) {
	state := new(T)
	resp.Diagnostics.Append(req.State.Get(ctx, state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	c := r.p.Client
	ctx, err := c.AddHeaders(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error encoding credentials", err.Error())
		return
	}

	err = r.deleteFunc(ctx, r.p, *state)

	if s, ok := status.FromError(err); ok && (s.Code() == codes.NotFound || s.Code() == codes.OK) {
		tflog.Info(ctx, "Resource was already deleted")
		// Remove MultyResource from state
		resp.State.RemoveResource(ctx)
	} else {
		resp.Diagnostics.AddError(
			"Error deleting resource",
			common.ParseGrpcErrors(err),
		)
	}
}

type planUpdater[T any] interface {
	UpdatePlan(ctx context.Context, config T, p Provider) (T, []*tftypes.AttributePath)
}

func (r MultyResource[T]) ModifyPlan(ctx context.Context, req tfsdk.ModifyResourcePlanRequest, resp *tfsdk.ModifyResourcePlanResponse) {
	plan := new(T)
	diags := req.Plan.Get(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if diags.HasError() {
		tflog.Warn(ctx, "Unable to parse plan when modifying it")
		return
	}

	config := new(T)
	diags = req.Config.Get(ctx, config)
	resp.Diagnostics.Append(diags...)
	if diags.HasError() {
		tflog.Warn(ctx, "Unable to parse config when modifying it")
		return
	}

	if c, ok := (any(*plan)).(planUpdater[T]); ok {
		tflog.Info(ctx, "Updating plan")
		newPlan, requiresReplace := c.UpdatePlan(ctx, *config, r.p)
		resp.Plan.Set(ctx, newPlan)
		resp.RequiresReplace = requiresReplace
	} else {
		tflog.Info(ctx, "Not updating plan because it doesn't implement planUpdater")
	}

}

func (r MultyResource[T]) ImportState(ctx context.Context, req tfsdk.ImportResourceStateRequest, resp *tfsdk.ImportResourceStateResponse) {
	// Save the import identifier in the id attribute
	tfsdk.ResourceImportStatePassthroughID(ctx, tftypes.NewAttributePath().WithAttributeName("id"), req, resp)
}
