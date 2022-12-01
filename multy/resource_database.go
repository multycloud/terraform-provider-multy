package multy

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/attr"
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

var databaseSchema = tfsdk.Schema{
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
}

func (r ResourceDatabaseType) NewResource(_ context.Context, p provider.Provider) resource.Resource {
	return MultyResource[Database]{
		p:          *(p.(*Provider)),
		createFunc: createDatabase,
		updateFunc: updateDatabase,
		readFunc:   readDatabase,
		deleteFunc: deleteDatabase,
		name:       "multy_database",
		schema:     databaseSchema,
	}
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
		Id:                 types.StringValue(res.CommonParameters.ResourceId),
		ResourceGroupId:    types.StringValue(res.CommonParameters.ResourceGroupId),
		Name:               types.StringValue(res.Name),
		Engine:             mtypes.DbEngineType.NewVal(res.Engine),
		EngineVersion:      types.StringValue(res.EngineVersion),
		StorageGb:          types.Int64Value(res.StorageGb),
		Size:               mtypes.DbSizeType.NewVal(res.Size),
		Username:           types.StringValue(res.Username),
		Password:           types.StringValue(res.Password),
		SubnetId:           types.StringValue(res.SubnetId),
		Cloud:              mtypes.CloudType.NewVal(res.CommonParameters.CloudProvider),
		Location:           mtypes.LocationType.NewVal(res.CommonParameters.Location),
		Hostname:           types.StringValue(res.Host),
		ConnectionUsername: types.StringValue(res.ConnectionUsername),
		GcpOverridesObject: convertToDatabaseGcpOverrides(res.GcpOverride).GcpOverridesToObj(),
		AwsOutputs: common.OptionallyObj(res.AwsOutputs, databaseAwsOutputs, map[string]attr.Value{
			"db_instance_id":                    common.DefaultToNull[types.String](res.GetAwsOutputs().GetDbInstanceId()),
			"default_network_security_group_id": common.DefaultToNull[types.String](res.GetAwsOutputs().GetDefaultNetworkSecurityGroupId()),
			"db_subnet_group_id":                common.DefaultToNull[types.String](res.GetAwsOutputs().GetDbSubnetGroupId()),
		}),
		AzureOutputs: common.OptionallyObj(res.AzureOutputs, databaseAzureOutputs, map[string]attr.Value{
			"database_server_id": common.DefaultToNull[types.String](res.GetAzureOutputs().GetDatabaseServerId()),
		}),
		GcpOutputs: common.OptionallyObj(res.GcpOutputs, databaseGcpOutputs, map[string]attr.Value{
			"sql_database_instance_id": common.DefaultToNull[types.String](res.GetGcpOutputs().GetSqlDatabaseInstanceId()),
		}),
		ResourceStatus: common.GetResourceStatus(res.CommonParameters.GetResourceStatus()),
	}
}

func convertFromDatabase(plan Database) *resourcespb.DatabaseArgs {
	return &resourcespb.DatabaseArgs{
		Name:          plan.Name.ValueString(),
		Engine:        plan.Engine.Value,
		EngineVersion: plan.EngineVersion.ValueString(),
		StorageGb:     plan.StorageGb.ValueInt64(),
		Size:          plan.Size.Value,
		Username:      plan.Username.ValueString(),
		Password:      plan.Password.ValueString(),
		SubnetId:      plan.SubnetId.ValueString(),
		CommonParameters: &commonpb.ResourceCommonArgs{
			ResourceGroupId: plan.ResourceGroupId.ValueString(),
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

		gcpOverrides.Project = types.StringValue(p.Client.Gcp.Project)

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
		Project: v.GcpOverridesObject.Attributes()["project"].(types.String),
	}
	return
}

func (o *DatabaseGcpOverrides) GcpOverridesToObj() types.Object {
	attrTypes := map[string]attr.Type{
		"project": types.StringType,
	}
	if o == nil {
		return types.ObjectNull(attrTypes)
	}
	result, _ := types.ObjectValue(attrTypes, map[string]attr.Value{"project": o.Project})
	return result
}

type DatabaseGcpOverrides struct {
	Project types.String
}

func convertFromDatabaseGcpOverrides(ref *DatabaseGcpOverrides) *resourcespb.DatabaseGcpOverride {
	if ref == nil {
		return nil
	}

	return &resourcespb.DatabaseGcpOverride{Project: ref.Project.ValueString()}
}

func convertToDatabaseGcpOverrides(ref *resourcespb.DatabaseGcpOverride) *DatabaseGcpOverrides {
	if ref == nil {
		return nil
	}

	return &DatabaseGcpOverrides{Project: common.DefaultToNull[types.String](ref.Project)}
}
