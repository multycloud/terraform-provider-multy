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
	"terraform-provider-multy/multy/validators"
)

type ResourceDatabaseType struct{}

var databaseAwsOutputs = map[string]attr.Type{
	"db_instance_id":                    types.StringType,
	"default_network_security_group_id": types.StringType,
	"db_subnet_group_id":                types.StringType,
}

var databaseAzureOutputs = map[string]attr.Type{
	"database_server_id": types.StringType,
}

var databaseGcpOutputs = map[string]attr.Type{
	"sql_database_instance_id": types.StringType,
}

func (r ResourceDatabaseType) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		MarkdownDescription: "Provides Multy Database resource",
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
				Description:   "Name of the database. If cloud is azure, name needs to be unique globally.",
				Required:      true,
				PlanModifiers: []tfsdk.AttributePlanModifier{common.RequiresReplaceIfCloudEq("azure")},
			},
			"engine": {
				Type:          mtypes.DbEngineType,
				Description:   fmt.Sprintf("Database engine. Available values are %v", mtypes.DbEngineType.GetAllValues()),
				Required:      true,
				Validators:    []tfsdk.AttributeValidator{validators.NewValidator(mtypes.DbEngineType)},
				PlanModifiers: []tfsdk.AttributePlanModifier{resource.RequiresReplace()},
			},
			"engine_version": {
				Type:          types.StringType,
				Description:   "Engine version",
				Required:      true,
				Validators:    []tfsdk.AttributeValidator{validators.StringInSliceValidator{Values: []string{"5.7", "8.0"}}},
				PlanModifiers: []tfsdk.AttributePlanModifier{resource.RequiresReplace()},
			},
			"storage_gb": {
				Type:        types.Int64Type,
				Description: "Size of database storage in gigabytes",
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
				Sensitive:   true,
				Required:    true,
			},
			"subnet_id": {
				Type:        types.StringType,
				Description: "Subnet associated with this database.",
				Required:    true,
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

			"cloud":    common.CloudsSchema,
			"location": common.LocationSchema,

			"hostname": {
				Type:        types.StringType,
				Description: "The hostname of the RDS instance.",
				Computed:    true,
			},
			"connection_username": {
				Type:        types.StringType,
				Description: "The username to connect to the database.",
				Computed:    true,
			},
			"aws": {
				Description: "AWS-specific ids of the underlying generated resources",
				Type:        types.ObjectType{AttrTypes: databaseAwsOutputs},
				Computed:    true,
			},
			"azure": {
				Description: "Azure-specific ids of the underlying generated resources",
				Type:        types.ObjectType{AttrTypes: databaseAzureOutputs},
				Computed:    true,
			},
			"gcp": {
				Description: "GCP-specific ids of the underlying generated resources",
				Type:        types.ObjectType{AttrTypes: databaseGcpOutputs},
				Computed:    true,
			},
			"resource_status": common.ResourceStatusSchema,
		},
	}, nil
}

func (r ResourceDatabaseType) NewResource(_ context.Context, p provider.Provider) (resource.Resource, diag.Diagnostics) {
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
		ResourceId: plan.Id.ValueString(),
		Resource:   convertFromDatabase(plan),
	})
	if err != nil {
		return Database{}, err
	}
	return convertToDatabase(vn), nil
}

func readDatabase(ctx context.Context, p Provider, state Database) (Database, error) {
	vn, err := p.Client.Client.ReadDatabase(ctx, &resourcespb.ReadDatabaseRequest{
		ResourceId: state.Id.ValueString(),
	})
	if err != nil {
		return Database{}, err
	}
	return convertToDatabase(vn), nil
}

func deleteDatabase(ctx context.Context, p Provider, state Database) error {
	_, err := p.Client.Client.DeleteDatabase(ctx, &resourcespb.DeleteDatabaseRequest{
		ResourceId: state.Id.ValueString(),
	})
	return err
}

type Database struct {
	Id                 types.String                                 `tfsdk:"id"`
	ResourceGroupId    types.String                                 `tfsdk:"resource_group_id"`
	Name               types.String                                 `tfsdk:"name"`
	Engine             mtypes.EnumValue[resourcespb.DatabaseEngine] `tfsdk:"engine"`
	EngineVersion      types.String                                 `tfsdk:"engine_version"`
	StorageGb          types.Int64                                  `tfsdk:"storage_gb"`
	Size               mtypes.EnumValue[commonpb.DatabaseSize_Enum] `tfsdk:"size"`
	Username           types.String                                 `tfsdk:"username"`
	Password           types.String                                 `tfsdk:"password"`
	SubnetId           types.String                                 `tfsdk:"subnet_id"`
	Cloud              mtypes.EnumValue[commonpb.CloudProvider]     `tfsdk:"cloud"`
	Location           mtypes.EnumValue[commonpb.Location]          `tfsdk:"location"`
	Hostname           types.String                                 `tfsdk:"hostname"`
	ConnectionUsername types.String                                 `tfsdk:"connection_username"`
	GcpOverridesObject types.Object                                 `tfsdk:"gcp_overrides"`
	AwsOutputs         types.Object                                 `tfsdk:"aws"`
	AzureOutputs       types.Object                                 `tfsdk:"azure"`
	GcpOutputs         types.Object                                 `tfsdk:"gcp"`
	ResourceStatus     types.Map                                    `tfsdk:"resource_status"`
}

func convertToDatabase(res *resourcespb.DatabaseResource) Database {
	return Database{
		Id:                 types.String{Value: res.CommonParameters.ResourceId},
		ResourceGroupId:    types.String{Value: res.CommonParameters.ResourceGroupId},
		Name:               types.String{Value: res.Name},
		Engine:             mtypes.DbEngineType.NewVal(res.Engine),
		EngineVersion:      types.String{Value: res.EngineVersion},
		StorageGb:          types.Int64{Value: res.StorageGb},
		Size:               mtypes.DbSizeType.NewVal(res.Size),
		Username:           types.String{Value: res.Username},
		Password:           types.String{Value: res.Password},
		SubnetId:           types.String{Value: res.SubnetId},
		Cloud:              mtypes.CloudType.NewVal(res.CommonParameters.CloudProvider),
		Location:           mtypes.LocationType.NewVal(res.CommonParameters.Location),
		Hostname:           types.String{Value: res.Host},
		ConnectionUsername: types.String{Value: res.ConnectionUsername},
		GcpOverridesObject: convertToDatabaseGcpOverrides(res.GcpOverride).GcpOverridesToObj(),
		AwsOutputs: common.OptionallyObj(res.AwsOutputs, types.Object{
			Attrs: map[string]attr.Value{
				"db_instance_id":                    common.DefaultToNull[types.String](res.GetAwsOutputs().GetDbInstanceId()),
				"default_network_security_group_id": common.DefaultToNull[types.String](res.GetAwsOutputs().GetDefaultNetworkSecurityGroupId()),
				"db_subnet_group_id":                common.DefaultToNull[types.String](res.GetAwsOutputs().GetDbSubnetGroupId()),
			},
			AttrTypes: databaseAwsOutputs,
		}),
		AzureOutputs: common.OptionallyObj(res.AzureOutputs, types.Object{
			Attrs: map[string]attr.Value{
				"database_server_id": common.DefaultToNull[types.String](res.GetAzureOutputs().GetDatabaseServerId()),
			},
			AttrTypes: databaseAzureOutputs,
		}),
		GcpOutputs: common.OptionallyObj(res.GcpOutputs, types.Object{
			Attrs: map[string]attr.Value{
				"sql_database_instance_id": common.DefaultToNull[types.String](res.GetGcpOutputs().GetSqlDatabaseInstanceId()),
			},
			AttrTypes: databaseGcpOutputs,
		}),
		ResourceStatus: common.GetResourceStatus(res.CommonParameters.GetResourceStatus()),
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
		SubnetId:      plan.SubnetId.Value,
		CommonParameters: &commonpb.ResourceCommonArgs{
			ResourceGroupId: plan.ResourceGroupId.Value,
			Location:        plan.Location.Value,
			CloudProvider:   plan.Cloud.Value,
		},
		GcpOverride: convertFromDatabaseGcpOverrides(plan.GetGcpOverrides()),
	}
}

func (v Database) UpdatePlan(_ context.Context, config Database, p Provider) (Database, []path.Path) {
	if config.Cloud.Value != commonpb.CloudProvider_GCP || p.Client.Gcp == nil {
		return v, nil
	}
	var requiresReplace []path.Path
	gcpOverrides := v.GetGcpOverrides()
	if o := config.GetGcpOverrides(); o == nil || o.Project.IsUnknown() {
		if gcpOverrides == nil {
			gcpOverrides = &DatabaseGcpOverrides{}
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

func (v Database) GetGcpOverrides() (o *DatabaseGcpOverrides) {
	if v.GcpOverridesObject.IsNull() || v.GcpOverridesObject.IsUnknown() {
		return
	}
	o = &DatabaseGcpOverrides{
		Project: v.GcpOverridesObject.Attrs["project"].(types.String),
	}
	return
}

func (o *DatabaseGcpOverrides) GcpOverridesToObj() types.Object {
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

type DatabaseGcpOverrides struct {
	Project types.String
}

func convertFromDatabaseGcpOverrides(ref *DatabaseGcpOverrides) *resourcespb.DatabaseGcpOverride {
	if ref == nil {
		return nil
	}

	return &resourcespb.DatabaseGcpOverride{Project: ref.Project.Value}
}

func convertToDatabaseGcpOverrides(ref *resourcespb.DatabaseGcpOverride) *DatabaseGcpOverrides {
	if ref == nil {
		return nil
	}

	return &DatabaseGcpOverrides{Project: common.DefaultToNull[types.String](ref.Project)}
}
