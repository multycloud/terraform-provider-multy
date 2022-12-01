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
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/multycloud/multy/api/proto/commonpb"
	"github.com/multycloud/multy/api/proto/resourcespb"
	"terraform-provider-multy/multy/common"
	"terraform-provider-multy/multy/mtypes"
	"terraform-provider-multy/multy/validators"
)

type ResourceVirtualMachineType struct{}

var virtualMachineSchema = tfsdk.Schema{
	MarkdownDescription: "Provides Multy Virtual Machine resource",
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
			Description:   "Name of Virtual Machine",
			Required:      true,
			PlanModifiers: []tfsdk.AttributePlanModifier{common.RequiresReplaceIfCloudEq("azure")},
		},
		"size": {
			Type:          mtypes.VmSizeType,
			Description:   fmt.Sprintf("Size of Virtual Machine. Accepted values are %s", common.StringSliceToDocsMarkdown(mtypes.VmSizeType.GetAllValues())),
			Required:      true,
			Validators:    []tfsdk.AttributeValidator{validators.NewValidator(mtypes.VmSizeType)},
			PlanModifiers: []tfsdk.AttributePlanModifier{resource.RequiresReplace()},
		},
		"subnet_id": {
			Type:          types.StringType,
			Description:   "ID of `subnet` resource",
			Required:      true,
			PlanModifiers: []tfsdk.AttributePlanModifier{resource.RequiresReplace()},
		},
		"network_interface_ids": {
			Type:        types.ListType{ElemType: types.StringType},
			Description: "IDs of `network_interface` resource",
			Optional:    true,
		},
		"network_security_group_ids": {
			Type:        types.ListType{ElemType: types.StringType},
			Description: "IDs of `network_security_group` resource",
			Optional:    true,
		},
		"user_data_base64": {
			Type:          types.StringType,
			Description:   "User Data script of Virtual Machine that will run on instance launch",
			Optional:      true,
			PlanModifiers: []tfsdk.AttributePlanModifier{resource.RequiresReplace()},
		},
		"availability_zone": {
			Type:          types.Int64Type,
			Description:   "Availability zone where this machine should be placed",
			Optional:      true,
			Computed:      true,
			PlanModifiers: []tfsdk.AttributePlanModifier{resource.RequiresReplace()},
			Validators:    []tfsdk.AttributeValidator{mtypes.NonEmptyIntValidator},
		},
		"public_ssh_key": {
			Type:          types.StringType,
			Description:   "Public SSH Key of Virtual Machine",
			Optional:      true,
			PlanModifiers: []tfsdk.AttributePlanModifier{resource.RequiresReplace(), validators.IgnoringWhitespace},
		},
		"public_ip_id": {
			Type:        types.StringType,
			Description: "ID of `public_ip` resource. Cannot be used with `generate_public_ip`",
			Optional:    true,
			Validators:  []tfsdk.AttributeValidator{mtypes.NonEmptyStringValidator},
		},
		"generate_public_ip": {
			Type:          types.BoolType,
			Description:   "If true, a public IP will be automatically generated. Cannot be used with `public_ip_id`",
			Optional:      true,
			Computed:      true,
			PlanModifiers: []tfsdk.AttributePlanModifier{common.RequiresReplaceIfCloudEq("aws")},
		},
		"image_reference": {
			Description: "Virtual Machine image definition",
			Attributes: tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
				"os": {
					Type:          mtypes.ImageOsDistroType,
					Description:   fmt.Sprintf("Operating System of Virtual Machine. Accepted values are %s", common.StringSliceToDocsMarkdown(mtypes.ImageOsDistroType.GetAllValues())),
					Required:      true,
					Validators:    []tfsdk.AttributeValidator{validators.NewValidator(mtypes.ImageOsDistroType)},
					PlanModifiers: []tfsdk.AttributePlanModifier{resource.RequiresReplace(), resource.UseStateForUnknown()},
				},
				"version": {
					Type:          types.StringType,
					Description:   "OS Version",
					Required:      true,
					PlanModifiers: []tfsdk.AttributePlanModifier{resource.RequiresReplace(), resource.UseStateForUnknown()},
				},
			}),
			// make this optional + computed and handle unknown values
			Required:      true,
			PlanModifiers: []tfsdk.AttributePlanModifier{resource.UseStateForUnknown()},
		},
		"aws_overrides": {
			Description: "AWS-specific attributes that will be set if this resource is deployed in AWS",
			Attributes: tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
				"instance_type": {
					Type:          types.StringType,
					Description:   fmt.Sprintf("The instance type to use for the instance."),
					Optional:      true,
					PlanModifiers: []tfsdk.AttributePlanModifier{common.RequiresReplaceIfCloudEq("aws"), resource.UseStateForUnknown()},
					Validators:    []tfsdk.AttributeValidator{mtypes.NonEmptyStringValidator},
				},
			}),
			Optional:      true,
			PlanModifiers: []tfsdk.AttributePlanModifier{common.RequiresReplaceIfCloudEq("aws")},
		},
		"azure_overrides": {
			Description: "Azure-specific attributes that will be set if this resource is deployed in Azure",
			Attributes: tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
				"size": {
					Type:          types.StringType,
					Description:   fmt.Sprintf("The size to use for the instance."),
					Optional:      true,
					PlanModifiers: []tfsdk.AttributePlanModifier{common.RequiresReplaceIfCloudEq("azure"), resource.UseStateForUnknown()},
					Validators:    []tfsdk.AttributeValidator{mtypes.NonEmptyStringValidator},
				},
			}),
			Optional:      true,
			PlanModifiers: []tfsdk.AttributePlanModifier{common.RequiresReplaceIfCloudEq("azure")},
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

		"public_ip": {
			Type:        types.StringType,
			Description: "Public IP of Virtual Machine",
			Computed:    true,
		},
		"identity": {
			Type:        types.StringType,
			Description: "Identity of Virtual Machine",
			Computed:    true,
		},
		"resource_status": common.ResourceStatusSchema,
	},
}

func (r ResourceVirtualMachineType) NewResource(_ context.Context, p provider.Provider) resource.Resource {
	return MultyResource[VirtualMachine]{
		p:          *(p.(*Provider)),
		createFunc: createVirtualMachine,
		updateFunc: updateVirtualMachine,
		readFunc:   readVirtualMachine,
		deleteFunc: deleteVirtualMachine,
		name:       "multy_virtual_machine",
		schema:     virtualMachineSchema,
	}
}

func createVirtualMachine(ctx context.Context, p Provider, plan VirtualMachine) (VirtualMachine, error) {
	pIpId := !plan.PublicIpId.IsNull() && !plan.PublicIpId.IsUnknown() && plan.PublicIpId.ValueString() != ""
	pIp := !plan.GeneratePublicIp.IsNull() && !plan.GeneratePublicIp.IsUnknown() && plan.GeneratePublicIp.ValueBool()
	// fixme check isnt working
	if pIpId && pIp {
		return VirtualMachine{}, fmt.Errorf("cannot set both public_ip and public_ip_id")
	}

	// Create new order from plan values
	vm, err := p.Client.Client.CreateVirtualMachine(ctx, &resourcespb.CreateVirtualMachineRequest{
		Resource: convertFromVirtualMachine(plan),
	})
	if err != nil {
		return VirtualMachine{}, err
	}

	tflog.Trace(ctx, "created virtual machine", map[string]interface{}{"virtual_machine_id": vm.CommonParameters.ResourceId})

	return convertToVirtualMachine(vm), nil
}

func updateVirtualMachine(ctx context.Context, p Provider, plan VirtualMachine) (VirtualMachine, error) {
	pIpId := !plan.PublicIpId.IsNull() && !plan.PublicIpId.IsUnknown() && plan.PublicIpId.ValueString() != ""
	pIp := !plan.GeneratePublicIp.IsNull() && !plan.GeneratePublicIp.IsUnknown() && plan.GeneratePublicIp.ValueBool()
	// fixme check isnt working
	if pIpId && pIp {
		return VirtualMachine{}, fmt.Errorf("cannot set both public_ip and public_ip_id")
	}

	// Create new order from plan values
	vm, err := p.Client.Client.UpdateVirtualMachine(ctx, &resourcespb.UpdateVirtualMachineRequest{
		ResourceId: plan.Id.ValueString(),
		Resource:   convertFromVirtualMachine(plan),
	})
	if err != nil {
		return VirtualMachine{}, err
	}

	tflog.Trace(ctx, "updated virtual machine", map[string]interface{}{"virtual_machine_id": vm.CommonParameters.ResourceId})

	return convertToVirtualMachine(vm), nil
}

func readVirtualMachine(ctx context.Context, p Provider, state VirtualMachine) (VirtualMachine, error) {
	vm, err := p.Client.Client.ReadVirtualMachine(ctx, &resourcespb.ReadVirtualMachineRequest{ResourceId: state.Id.ValueString()})
	if err != nil {
		return VirtualMachine{}, err
	}

	state = convertToVirtualMachine(vm)
	return convertToVirtualMachine(vm), nil
}

func deleteVirtualMachine(ctx context.Context, p Provider, state VirtualMachine) error {
	_, err := p.Client.Client.DeleteVirtualMachine(ctx, &resourcespb.DeleteVirtualMachineRequest{ResourceId: state.Id.ValueString()})
	return err
}

func convertToVirtualMachine(res *resourcespb.VirtualMachineResource) VirtualMachine {
	return VirtualMachine{
		Id:                      types.StringValue(res.CommonParameters.ResourceId),
		ResourceGroupId:         types.StringValue(res.CommonParameters.ResourceGroupId),
		Name:                    types.StringValue(res.Name),
		Size:                    mtypes.VmSizeType.NewVal(res.VmSize),
		SubnetId:                types.StringValue(res.SubnetId),
		NetworkInterfaceIds:     common.DefaultSliceToNull(common.TypesStringToStringSlice(res.NetworkInterfaceIds)),
		NetworkSecurityGroupIds: common.DefaultSliceToNull(common.TypesStringToStringSlice(res.NetworkSecurityGroupIds)),
		UserDataBase64:          common.DefaultToNull[types.String](res.UserDataBase64),
		PublicSshKey:            common.DefaultToNull[types.String](res.PublicSshKey),
		PublicIpId:              common.DefaultToNull[types.String](res.PublicIpId),
		GeneratePublicIp:        types.BoolValue(res.GeneratePublicIp),
		PublicIp:                types.StringValue(res.PublicIp),
		Identity:                types.StringValue(res.IdentityId),
		ImageReference: &ImageReference{
			OS:      mtypes.ImageOsDistroType.NewVal(res.ImageReference.Os),
			Version: types.StringValue(res.ImageReference.Version),
		},
		AvailabilityZone:   types.Int64Value(int64(res.AvailabilityZone)),
		AwsOverrides:       convertToVirtualMachineAwsOverrides(res.AwsOverride),
		AzureOverrides:     convertToVirtualMachineAzureOverrides(res.AzureOverride),
		GcpOverridesObject: convertToVirtualMachineGcpOverrides(res.GcpOverride).GcpOverridesToObj(),
		Cloud:              mtypes.CloudType.NewVal(res.CommonParameters.CloudProvider),
		Location:           mtypes.LocationType.NewVal(res.CommonParameters.Location),
		ResourceStatus:     common.GetResourceStatus(res.CommonParameters.GetResourceStatus()),
	}
}

func convertFromVirtualMachine(plan VirtualMachine) *resourcespb.VirtualMachineArgs {
	return &resourcespb.VirtualMachineArgs{
		CommonParameters: &commonpb.ResourceCommonArgs{
			Location:        plan.Location.Value,
			CloudProvider:   plan.Cloud.Value,
			ResourceGroupId: plan.ResourceGroupId.ValueString(),
		},
		Name:                    plan.Name.ValueString(),
		NetworkInterfaceIds:     common.StringSliceToTypesString(plan.NetworkInterfaceIds),
		NetworkSecurityGroupIds: common.StringSliceToTypesString(plan.NetworkSecurityGroupIds),
		VmSize:                  plan.Size.Value,
		UserDataBase64:          plan.UserDataBase64.ValueString(),
		SubnetId:                plan.SubnetId.ValueString(),
		PublicSshKey:            plan.PublicSshKey.ValueString(),
		PublicIpId:              plan.PublicIpId.ValueString(),
		GeneratePublicIp:        plan.GeneratePublicIp.ValueBool(),
		AvailabilityZone:        int32(plan.AvailabilityZone.ValueInt64()),
		ImageReference:          convertFromImageRef(plan.ImageReference),
		AwsOverride:             convertFromVirtualMachineAwsOverrides(plan.AwsOverrides),
		AzureOverride:           convertFromVirtualMachineAzureOverrides(plan.AzureOverrides),
		GcpOverride:             convertFromVirtualMachineGcpOverrides(plan.GetGcpOverrides()),
	}
}

func convertFromImageRef(ref *ImageReference) *resourcespb.ImageReference {
	if ref == nil {
		return nil
	}

	return &resourcespb.ImageReference{
		Os:      ref.OS.Value,
		Version: ref.Version.ValueString(),
	}
}

func convertFromVirtualMachineAwsOverrides(ref *VirtualMachineAwsOverrides) *resourcespb.VirtualMachineAwsOverride {
	if ref == nil {
		return nil
	}

	return &resourcespb.VirtualMachineAwsOverride{
		InstanceType: ref.InstanceType.ValueString(),
	}
}

func convertToVirtualMachineAwsOverrides(ref *resourcespb.VirtualMachineAwsOverride) *VirtualMachineAwsOverrides {
	if ref == nil {
		return nil
	}

	return &VirtualMachineAwsOverrides{InstanceType: common.DefaultToNull[types.String](ref.InstanceType)}
}
func convertFromVirtualMachineAzureOverrides(ref *VirtualMachineAzureOverrides) *resourcespb.VirtualMachineAzureOverride {
	if ref == nil {
		return nil
	}

	return &resourcespb.VirtualMachineAzureOverride{
		Size: ref.Size.ValueString(),
	}
}

func convertToVirtualMachineAzureOverrides(ref *resourcespb.VirtualMachineAzureOverride) *VirtualMachineAzureOverrides {
	if ref == nil {
		return nil
	}

	return &VirtualMachineAzureOverrides{Size: common.DefaultToNull[types.String](ref.Size)}
}

type VirtualMachine struct {
	Id                      types.String                           `tfsdk:"id"`
	ResourceGroupId         types.String                           `tfsdk:"resource_group_id"`
	Name                    types.String                           `tfsdk:"name"`
	Size                    mtypes.EnumValue[commonpb.VmSize_Enum] `tfsdk:"size"`
	SubnetId                types.String                           `tfsdk:"subnet_id"`
	NetworkInterfaceIds     []types.String                         `tfsdk:"network_interface_ids"`
	NetworkSecurityGroupIds []types.String                         `tfsdk:"network_security_group_ids"`
	UserDataBase64          types.String                           `tfsdk:"user_data_base64"`
	PublicSshKey            types.String                           `tfsdk:"public_ssh_key"`
	PublicIpId              types.String                           `tfsdk:"public_ip_id"`
	GeneratePublicIp        types.Bool                             `tfsdk:"generate_public_ip"`
	PublicIp                types.String                           `tfsdk:"public_ip"`
	Identity                types.String                           `tfsdk:"identity"`
	ImageReference          *ImageReference                        `tfsdk:"image_reference"`
	AvailabilityZone        types.Int64                            `tfsdk:"availability_zone"`
	AwsOverrides            *VirtualMachineAwsOverrides            `tfsdk:"aws_overrides"`
	AzureOverrides          *VirtualMachineAzureOverrides          `tfsdk:"azure_overrides"`
	GcpOverridesObject      types.Object                           `tfsdk:"gcp_overrides"`

	Cloud          mtypes.EnumValue[commonpb.CloudProvider] `tfsdk:"cloud"`
	Location       mtypes.EnumValue[commonpb.Location]      `tfsdk:"location"`
	ResourceStatus types.Map                                `tfsdk:"resource_status"`
}

type ImageReference struct {
	OS      mtypes.EnumValue[resourcespb.ImageReference_OperatingSystemDistribution] `tfsdk:"os"`
	Version types.String                                                             `tfsdk:"version"`
}

type VirtualMachineAwsOverrides struct {
	InstanceType types.String `tfsdk:"instance_type"`
}
type VirtualMachineAzureOverrides struct {
	Size types.String `tfsdk:"size"`
}

func convertFromVirtualMachineGcpOverrides(ref *VirtualMachineGcpOverrides) *resourcespb.VirtualMachineGcpOverride {
	if ref == nil {
		return nil
	}

	return &resourcespb.VirtualMachineGcpOverride{Project: ref.Project.ValueString()}
}

func convertToVirtualMachineGcpOverrides(ref *resourcespb.VirtualMachineGcpOverride) *VirtualMachineGcpOverrides {
	if ref == nil {
		return nil
	}

	return &VirtualMachineGcpOverrides{Project: common.DefaultToNull[types.String](ref.Project)}
}

func (v VirtualMachine) GetGcpOverrides() (o *VirtualMachineGcpOverrides) {
	if v.GcpOverridesObject.IsNull() || v.GcpOverridesObject.IsUnknown() {
		return
	}
	o = &VirtualMachineGcpOverrides{
		Project: v.GcpOverridesObject.Attributes()["project"].(types.String),
	}
	return
}

func (o *VirtualMachineGcpOverrides) GcpOverridesToObj() types.Object {
	attrTypes := map[string]attr.Type{
		"project": types.StringType,
	}
	if o == nil {
		return types.ObjectNull(attrTypes)
	}
	result, _ := types.ObjectValue(attrTypes, map[string]attr.Value{"project": o.Project})
	return result
}

type VirtualMachineGcpOverrides struct {
	Project types.String
}

func (v VirtualMachine) UpdatePlan(_ context.Context, config VirtualMachine, p Provider) (VirtualMachine, []path.Path) {
	if config.Cloud.Value != commonpb.CloudProvider_GCP || p.Client.Gcp == nil {
		return v, nil
	}
	var requiresReplace []path.Path
	gcpOverrides := v.GetGcpOverrides()
	if o := config.GetGcpOverrides(); o == nil || o.Project.IsUnknown() {
		if gcpOverrides == nil {
			gcpOverrides = &VirtualMachineGcpOverrides{}
		}

		gcpOverrides.Project = types.StringValue(p.Client.Gcp.Project)

		v.GcpOverridesObject = gcpOverrides.GcpOverridesToObj()
		requiresReplace = append(requiresReplace, path.Root("gcp_overrides").AtName("project"))
	}
	return v, requiresReplace
}
