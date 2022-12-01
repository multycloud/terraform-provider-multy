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
)

type ResourceKubernetesClusterType struct{}

var kubernetesClusterAwsOutputs = map[string]attr.Type{
	"eks_cluster_id": types.StringType,
	"iam_role_arn":   types.StringType,
}

var kubernetesClusterAzureOutputs = map[string]attr.Type{
	"aks_cluster_id": types.StringType,
}

var kubernetesClusterGcpOutputs = map[string]attr.Type{
	"gke_cluster_id":        types.StringType,
	"service_account_email": types.StringType,
}

var kubernetesClusterSchema = tfsdk.Schema{
	MarkdownDescription: "Provides Multy Kubernetes Cluster resource",
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
			Description:   "Name of the cluster",
			Required:      true,
			PlanModifiers: []tfsdk.AttributePlanModifier{resource.RequiresReplace()},
		},
		"virtual_network_id": {
			Type:          types.StringType,
			Description:   "Virtual network where cluster and associated node pools should be in.",
			Required:      true,
			PlanModifiers: []tfsdk.AttributePlanModifier{resource.RequiresReplace()},
		},
		"service_cidr": {
			Type:          types.StringType,
			Description:   "CIDR block for service nodes.",
			Computed:      true,
			Optional:      true,
			PlanModifiers: []tfsdk.AttributePlanModifier{resource.RequiresReplace()},
		},
		"default_node_pool": {
			Attributes:  tfsdk.SingleNestedAttributes(getKubernetesNodePoolAttrs()),
			Description: "Default node pool to associate with this cluster.",
			Required:    true,
		},
		"endpoint": {
			Type:        types.StringType,
			Description: "Endpoint of the kubernetes cluster.",
			Computed:    true,
		},
		"ca_certificate": {
			Type:        types.StringType,
			Description: "Base64 encoded certificate data required to communicate with your cluster.",
			Computed:    true,
			Sensitive:   true,
		},
		"kube_config_raw": {
			Type:        types.StringType,
			Description: "Raw Kubernetes config to be used by kubectl and other compatible tools.",
			Computed:    true,
			Sensitive:   true,
		},
		"cloud":    common.CloudsSchema,
		"location": common.LocationSchema,
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
		"aws": {
			Description: "AWS-specific ids of the underlying generated resources",
			Type:        types.ObjectType{AttrTypes: kubernetesClusterAwsOutputs},
			Computed:    true,
		},
		"azure": {
			Description: "Azure-specific ids of the underlying generated resources",
			Type:        types.ObjectType{AttrTypes: kubernetesClusterAzureOutputs},
			Computed:    true,
		},
		"gcp": {
			Description: "GCP-specific ids of the underlying generated resources",
			Type:        types.ObjectType{AttrTypes: kubernetesClusterGcpOutputs},
			Computed:    true,
		},
		"resource_status": common.ResourceStatusSchema,
	},
}

func (r ResourceKubernetesClusterType) NewResource(_ context.Context, p provider.Provider) resource.Resource {
	return MultyResource[KubernetesCluster]{
		p:          *(p.(*Provider)),
		createFunc: createKubernetesCluster,
		updateFunc: updateKubernetesCluster,
		readFunc:   readKubernetesCluster,
		deleteFunc: deleteKubernetesCluster,
		name:       "multy_kubernetes_cluster",
		schema:     kubernetesClusterSchema,
	}
}

func createKubernetesCluster(ctx context.Context, p Provider, plan KubernetesCluster) (KubernetesCluster, error) {
	vn, err := p.Client.Client.CreateKubernetesCluster(ctx, &resourcespb.CreateKubernetesClusterRequest{
		Resource: convertFromKubernetesCluster(plan),
	})
	if err != nil {
		return KubernetesCluster{}, err
	}
	return convertToKubernetesCluster(vn), nil
}

func updateKubernetesCluster(ctx context.Context, p Provider, plan KubernetesCluster) (KubernetesCluster, error) {
	vn, err := p.Client.Client.UpdateKubernetesCluster(ctx, &resourcespb.UpdateKubernetesClusterRequest{
		ResourceId: plan.Id.ValueString(),
		Resource:   convertFromKubernetesCluster(plan),
	})
	if err != nil {
		return KubernetesCluster{}, err
	}
	return convertToKubernetesCluster(vn), nil
}

func readKubernetesCluster(ctx context.Context, p Provider, state KubernetesCluster) (KubernetesCluster, error) {
	vn, err := p.Client.Client.ReadKubernetesCluster(ctx, &resourcespb.ReadKubernetesClusterRequest{
		ResourceId: state.Id.ValueString(),
	})
	if err != nil {
		return KubernetesCluster{}, err
	}
	return convertToKubernetesCluster(vn), nil
}

func deleteKubernetesCluster(ctx context.Context, p Provider, state KubernetesCluster) error {
	_, err := p.Client.Client.DeleteKubernetesCluster(ctx, &resourcespb.DeleteKubernetesClusterRequest{
		ResourceId: state.Id.ValueString(),
	})
	return err
}

type KubernetesCluster struct {
	Id               types.String                             `tfsdk:"id"`
	Name             types.String                             `tfsdk:"name"`
	VirtualNetworkId types.String                             `tfsdk:"virtual_network_id"`
	ServiceCidr      types.String                             `tfsdk:"service_cidr"`
	Cloud            mtypes.EnumValue[commonpb.CloudProvider] `tfsdk:"cloud"`
	Location         mtypes.EnumValue[commonpb.Location]      `tfsdk:"location"`
	ResourceGroupId  types.String                             `tfsdk:"resource_group_id"`

	DefaultNodePool KubernetesNodePool `tfsdk:"default_node_pool"`

	GcpOverridesObject types.Object `tfsdk:"gcp_overrides"`

	Endpoint      types.String `tfsdk:"endpoint"`
	CaCertificate types.String `tfsdk:"ca_certificate"`
	KubeConfigRaw types.String `tfsdk:"kube_config_raw"`

	AwsOutputs     types.Object `tfsdk:"aws"`
	AzureOutputs   types.Object `tfsdk:"azure"`
	GcpOutputs     types.Object `tfsdk:"gcp"`
	ResourceStatus types.Map    `tfsdk:"resource_status"`
}

func convertToKubernetesCluster(res *resourcespb.KubernetesClusterResource) KubernetesCluster {
	return KubernetesCluster{
		Id:                 types.StringValue(res.CommonParameters.ResourceId),
		Name:               types.StringValue(res.Name),
		VirtualNetworkId:   types.StringValue(res.VirtualNetworkId),
		ServiceCidr:        types.StringValue(res.ServiceCidr),
		Cloud:              mtypes.CloudType.NewVal(res.CommonParameters.CloudProvider),
		Location:           mtypes.LocationType.NewVal(res.CommonParameters.Location),
		ResourceGroupId:    types.StringValue(res.CommonParameters.ResourceGroupId),
		DefaultNodePool:    convertToKubernetesNodePool(res.GetDefaultNodePool()),
		GcpOverridesObject: convertToKubernetesClusterGcpOverrides(res.GcpOverride).GcpOverridesToObj(),
		Endpoint:           types.StringValue(res.Endpoint),
		CaCertificate:      types.StringValue(res.CaCertificate),
		KubeConfigRaw:      types.StringValue(res.KubeConfigRaw),
		AwsOutputs: common.OptionallyObj(res.AwsOutputs, kubernetesClusterAwsOutputs, map[string]attr.Value{
			"eks_cluster_id": common.DefaultToNull[types.String](res.GetAwsOutputs().GetEksClusterId()),
			"iam_role_arn":   common.DefaultToNull[types.String](res.GetAwsOutputs().GetIamRoleArn()),
		}),
		AzureOutputs: common.OptionallyObj(res.AzureOutputs, kubernetesClusterAzureOutputs, map[string]attr.Value{
			"aks_cluster_id": common.DefaultToNull[types.String](res.GetAzureOutputs().GetAksClusterId()),
		}),
		GcpOutputs: common.OptionallyObj(res.GcpOutputs, kubernetesClusterGcpOutputs, map[string]attr.Value{
			"gke_cluster_id":        common.DefaultToNull[types.String](res.GetGcpOutputs().GetGkeClusterId()),
			"service_account_email": common.DefaultToNull[types.String](res.GetGcpOutputs().GetServiceAccountEmail()),
		}),
		ResourceStatus: common.GetResourceStatus(res.CommonParameters.GetResourceStatus()),
	}
}

func convertFromKubernetesCluster(plan KubernetesCluster) *resourcespb.KubernetesClusterArgs {
	return &resourcespb.KubernetesClusterArgs{
		CommonParameters: &commonpb.ResourceCommonArgs{
			Location:        plan.Location.Value,
			CloudProvider:   plan.Cloud.Value,
			ResourceGroupId: plan.ResourceGroupId.ValueString(),
		},
		Name:             plan.Name.ValueString(),
		VirtualNetworkId: plan.VirtualNetworkId.ValueString(),
		ServiceCidr:      plan.ServiceCidr.ValueString(),
		DefaultNodePool:  convertFromKubernetesNodePool(plan.DefaultNodePool),
		GcpOverride:      convertFromKubernetesClusterGcpOverrides(plan.GetGcpOverrides()),
	}
}

func (v KubernetesCluster) UpdatePlan(_ context.Context, config KubernetesCluster, p Provider) (KubernetesCluster, []path.Path) {
	if config.Cloud.Value != commonpb.CloudProvider_GCP {
		return v, nil
	}
	var requiresReplace []path.Path
	gcpOverrides := v.GetGcpOverrides()
	if o := config.GetGcpOverrides(); o == nil || o.Project.IsUnknown() {
		if gcpOverrides == nil {
			gcpOverrides = &KubernetesClusterGcpOverrides{}
		}

		gcpOverrides.Project = types.StringValue(p.Client.Gcp.Project)

		v.GcpOverridesObject = gcpOverrides.GcpOverridesToObj()
		requiresReplace = append(requiresReplace, path.Root("gcp_overrides").AtName("project"))
	}
	return v, requiresReplace
}

func (v KubernetesCluster) GetGcpOverrides() (o *KubernetesClusterGcpOverrides) {
	if v.GcpOverridesObject.IsNull() || v.GcpOverridesObject.IsUnknown() {
		return
	}
	o = &KubernetesClusterGcpOverrides{
		Project: v.GcpOverridesObject.Attributes()["project"].(types.String),
	}
	return
}

func (o *KubernetesClusterGcpOverrides) GcpOverridesToObj() types.Object {
	attrTypes := map[string]attr.Type{
		"project": types.StringType,
	}
	if o == nil {
		return types.ObjectNull(attrTypes)
	}
	result, _ := types.ObjectValue(attrTypes, map[string]attr.Value{"project": o.Project})
	return result
}

type KubernetesClusterGcpOverrides struct {
	Project types.String
}

func convertFromKubernetesClusterGcpOverrides(ref *KubernetesClusterGcpOverrides) *resourcespb.KubernetesClusterOverrides {
	if ref == nil {
		return nil
	}

	return &resourcespb.KubernetesClusterOverrides{Project: ref.Project.ValueString()}
}

func convertToKubernetesClusterGcpOverrides(ref *resourcespb.KubernetesClusterOverrides) *KubernetesClusterGcpOverrides {
	if ref == nil {
		return nil
	}

	return &KubernetesClusterGcpOverrides{Project: common.DefaultToNull[types.String](ref.Project)}
}
