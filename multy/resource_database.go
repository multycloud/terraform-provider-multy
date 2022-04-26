package multy

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/multycloud/multy/api/proto/commonpb"
	"github.com/multycloud/multy/api/proto/resourcespb"
	"terraform-provider-multy/multy/common"
	"terraform-provider-multy/multy/mtypes"
	"terraform-provider-multy/multy/validators"
)

type ResourceDatabaseType struct{}

func (r ResourceDatabaseType) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		MarkdownDescription: "Provides Multy Database resource",
		Attributes: map[string]tfsdk.Attribute{
			"id": {
				Type:          types.StringType,
				Computed:      true,
				PlanModifiers: []tfsdk.AttributePlanModifier{tfsdk.UseStateForUnknown()},
			},
			"name": {
				Type:          types.StringType,
				Description:   "Name of the database. If cloud is azure, name needs to be unique globally.",
				Required:      true,
				PlanModifiers: []tfsdk.AttributePlanModifier{common.RequiresReplaceIfCloudEq("azure")},
			},
			"engine": {
				Type:          mtypes.DbEngineType,
				Description:   fmt.Sprintf("Database engine. Available values are %v", mtypes.DbEngineType.GetAllValues()),
				Required:      true,
				Validators:    []tfsdk.AttributeValidator{validators.NewValidator(mtypes.DbEngineType)},
				PlanModifiers: []tfsdk.AttributePlanModifier{tfsdk.RequiresReplace()},
			},
			"engine_version": {
				Type:          types.StringType,
				Description:   "Engine version",
				Required:      true,
				Validators:    []tfsdk.AttributeValidator{validators.StringInSliceValidator{Values: []string{"5.7", "8.0"}}},
				PlanModifiers: []tfsdk.AttributePlanModifier{tfsdk.RequiresReplace()},
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
				// TODO: validate length
			},

			"cloud":    common.CloudsSchema,
			"location": common.LocationSchema,

			"hostname": {
				Type:        types.StringType,
				Description: "The hostname of the RDS instance.",
				Computed:    true,
			},
		},
	}, nil
}

func (r ResourceDatabaseType) NewResource(_ context.Context, p tfsdk.Provider) (tfsdk.Resource, diag.Diagnostics) {
	return MultyResource[Database]{
		p:          *(p.(*Provider)),
		createFunc: createDatabase,
		updateFunc: updateDatabase,
		readFunc:   readDatabase,
		deleteFunc: deleteDatabase,
	}, nil
}

func createDatabase(ctx context.Context, p Provider, plan Database) (Database, error) {
	vn, err := p.Client.Client.CreateDatabase(ctx, &resourcespb.CreateDatabaseRequest{
		Resource: convertFromDatabase(plan),
	})
	if err != nil {
		return Database{}, err
	}
	return convertToDatabase(vn), nil
}

func updateDatabase(ctx context.Context, p Provider, plan Database) (Database, error) {
	vn, err := p.Client.Client.UpdateDatabase(ctx, &resourcespb.UpdateDatabaseRequest{
		ResourceId: plan.Id.Value,
		Resource:   convertFromDatabase(plan),
	})
	if err != nil {
		return Database{}, err
	}
	return convertToDatabase(vn), nil
}

func readDatabase(ctx context.Context, p Provider, state Database) (Database, error) {
	vn, err := p.Client.Client.ReadDatabase(ctx, &resourcespb.ReadDatabaseRequest{
		ResourceId: state.Id.Value,
	})
	if err != nil {
		return Database{}, err
	}
	return convertToDatabase(vn), nil
}

func deleteDatabase(ctx context.Context, p Provider, state Database) error {
	_, err := p.Client.Client.DeleteDatabase(ctx, &resourcespb.DeleteDatabaseRequest{
		ResourceId: state.Id.Value,
	})
	return err
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
	Hostname      types.String                                 `tfsdk:"hostname"`
}

func convertToDatabase(res *resourcespb.DatabaseResource) Database {
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
		Hostname:      types.String{Value: res.Host},
	}
}

func convertFromDatabase(plan Database) *resourcespb.DatabaseArgs {
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
