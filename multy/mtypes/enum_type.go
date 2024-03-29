package mtypes

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/multycloud/multy/api/proto/commonpb"
	"github.com/multycloud/multy/api/proto/resourcespb"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
	"google.golang.org/protobuf/reflect/protoreflect"
	"strings"
)

var (
	CloudType = EnumType[commonpb.CloudProvider]{
		ValueMap: commonpb.CloudProvider_value,
	}
	LocationType = EnumType[commonpb.Location]{
		ValueMap: commonpb.Location_value,
	}
	VmSizeType = EnumType[commonpb.VmSize_Enum]{
		ValueMap: commonpb.VmSize_Enum_value,
	}
	OperatingSystemType = EnumType[commonpb.OperatingSystem_Enum]{
		ValueMap: commonpb.OperatingSystem_Enum_value,
	}
	ObjectAclType = EnumType[resourcespb.ObjectStorageObjectAcl]{
		ValueMap:          resourcespb.ObjectStorageObjectAcl_value,
		allowDefaultValue: true,
	}
	DbEngineType = EnumType[resourcespb.DatabaseEngine]{
		ValueMap: resourcespb.DatabaseEngine_value,
	}
	DbSizeType = EnumType[commonpb.DatabaseSize_Enum]{
		ValueMap: commonpb.DatabaseSize_Enum_value,
	}
	RouteDestinationType = EnumType[resourcespb.RouteDestination]{
		ValueMap: resourcespb.RouteDestination_value,
	}
	VaultAclType = EnumType[resourcespb.VaultAccess_Enum]{
		ValueMap: resourcespb.VaultAccess_Enum_value,
	}
	ImageOsDistroType = EnumType[resourcespb.ImageReference_OperatingSystemDistribution]{
		ValueMap: resourcespb.ImageReference_OperatingSystemDistribution_value,
	}
)

type ProtoEnum interface {
	~int32
	String() string
	Number() protoreflect.EnumNumber
}

type EnumType[T ProtoEnum] struct {
	ValueMap          map[string]int32
	allowDefaultValue bool
}

type EnumValue[T ProtoEnum] struct {
	Typ     EnumType[T]
	Value   T
	Null    bool
	Unknown bool

	strValue *string
}

func (n EnumType[T]) NewVal(val T) EnumValue[T] {
	return EnumValue[T]{Value: val, Typ: n}
}

func (n EnumType[T]) ZeroVal() EnumValue[T] {
	return EnumValue[T]{Typ: n}
}

// ValueType should return the attr.Value type returned by
// ValueFromTerraform. The returned attr.Value can be any null, unknown,
// or known value for the type, as this is intended for type detection
// and improving error diagnostics.
func (n EnumType[T]) ValueType(context.Context) attr.Value {
	return types.String{}
}

func (s EnumValue[T]) Type(_ context.Context) attr.Type {
	return s.Typ
}

func (s EnumValue[T]) ToTerraformValue(_ context.Context) (tftypes.Value, error) {
	if s.IsNull() {
		return tftypes.NewValue(tftypes.String, nil), nil
	}
	if s.IsUnknown() {
		return tftypes.NewValue(tftypes.String, tftypes.UnknownValue), nil
	}
	if s.strValue != nil {
		return tftypes.NewValue(tftypes.String, s.strValue), nil
	}

	return tftypes.NewValue(tftypes.String, strings.ToLower(s.Value.String())), nil
}

func (s EnumValue[T]) IsNull() bool {
	return s.Null
}

func (s EnumValue[T]) IsUnknown() bool {
	return s.Unknown
}

func (s EnumValue[T]) String() string {
	if s.strValue == nil {
		return ""
	}
	return *s.strValue
}

func (s EnumValue[T]) Equal(other attr.Value) bool {
	o, ok := other.(EnumValue[T])
	if !ok {
		return false
	}
	if s.IsNull() != o.IsNull() {
		return false
	}
	if s.IsUnknown() != o.IsUnknown() {
		return false
	}
	return s.Typ.Equal(o.Typ) && s.Value.String() == o.Value.String()
}

func (s EnumValue[T]) Validate() error {
	if s.IsUnknown() || s.IsNull() {
		return nil
	}

	if s.strValue == nil && (s.Typ.allowDefaultValue || s.Value != T(0)) {
		return nil
	}

	allowedVals := s.Typ.GetAllValues()

	if s.strValue == nil || !slices.Contains(allowedVals, *s.strValue) {
		return fmt.Errorf("value %s it not defined, must be one of %v", *s.strValue, allowedVals)
	}

	return nil
}

func (n EnumType[T]) GetAllValues() []string {
	var elemTotal int32
	for _, v := range n.ValueMap {
		if v >= elemTotal {
			elemTotal = v + 1
		}
	}
	if elemTotal <= 0 {
		return nil
	}

	allVals := make([]string, elemTotal)
	for k, val := range n.ValueMap {
		if val == 0 && !n.allowDefaultValue {
			continue
		}
		allVals[val] = strings.ToLower(k)
	}
	var allowedVals []string
	for _, v := range allVals {
		if len(v) > 0 {
			allowedVals = append(allowedVals, v)
		}
	}

	return allowedVals
}

func (n EnumType[T]) TerraformType(_ context.Context) tftypes.Type {
	return tftypes.String
}

func (n EnumType[T]) ValueFromTerraform(_ context.Context, in tftypes.Value) (attr.Value, error) {
	if !in.IsKnown() {
		return EnumValue[T]{Unknown: true}, nil
	}
	if in.IsNull() {
		return EnumValue[T]{Null: true}, nil
	}
	var s string
	err := in.As(&s)
	if err != nil {
		return nil, err
	}

	if val, ok := n.ValueMap[strings.ToUpper(s)]; ok {
		return EnumValue[T]{Value: T(val), strValue: &s}, nil
	} else {
		return EnumValue[T]{Value: T(0), strValue: &s}, nil
	}
}

func (n EnumType[T]) Equal(t attr.Type) bool {
	// TODO: fix this
	_, ok := t.(EnumType[T])
	if !ok {
		return false
	}
	//return maps.Equal(n.ValueMap, o.ValueMap)

	return true
}

func (n EnumType[T]) String() string {

	return fmt.Sprintf("types.EnumType[%v]", maps.Keys(n.ValueMap))
}

func (n EnumType[T]) ApplyTerraform5AttributePathStep(_ tftypes.AttributePathStep) (interface{}, error) {
	return nil, fmt.Errorf("unimplemented")
}

type nonEmptyValueValidator[T comparable] struct {
}

var NonEmptyStringValidator = nonEmptyValueValidator[string]{}
var NonEmptyIntValidator = nonEmptyValueValidator[int]{}

func (n nonEmptyValueValidator[T]) Description(_ context.Context) string {
	return ""
}

func (n nonEmptyValueValidator[T]) MarkdownDescription(_ context.Context) string {
	return ""
}

func (n nonEmptyValueValidator[T]) Validate(ctx context.Context, req tfsdk.ValidateAttributeRequest, resp *tfsdk.ValidateAttributeResponse) {
	value, err := req.AttributeConfig.ToTerraformValue(ctx)
	if err != nil {
		panic(err)
	}
	if !value.IsFullyKnown() || value.IsNull() {
		return
	}

	val := new(T)
	err = value.As(val)
	if err != nil {
		panic(err)
	}

	if *val == *new(T) {
		resp.Diagnostics.AddAttributeError(
			req.AttributePath,
			"Invalid value",
			fmt.Sprintf("value cannot be empty, but was \"%v\"", *val),
		)
	}
}
