package mtypes

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
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

func (s EnumValue[T]) Type(_ context.Context) attr.Type {
	return s.Typ
}

func (s EnumValue[T]) ToTerraformValue(_ context.Context) (tftypes.Value, error) {
	if s.Null {
		return tftypes.NewValue(tftypes.String, nil), nil
	}
	if s.Unknown {
		return tftypes.NewValue(tftypes.String, tftypes.UnknownValue), nil
	}
	if s.strValue != nil {
		return tftypes.NewValue(tftypes.String, s.strValue), nil
	}

	return tftypes.NewValue(tftypes.String, strings.ToLower(s.Value.String())), nil
}

func (s EnumValue[T]) Equal(other attr.Value) bool {
	o, ok := other.(EnumValue[T])
	if !ok {
		return false
	}
	if s.Null != o.Null {
		return false
	}
	if s.Unknown != o.Unknown {
		return false
	}
	return s.Typ.Equal(o.Typ) && s.Value.String() == o.Value.String()
}

func (s EnumValue[T]) Validate() error {
	if s.Unknown || s.Null {
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
	elemTotal := len(n.ValueMap)
	if !n.allowDefaultValue {
		elemTotal -= 1
	}
	if elemTotal <= 0 {
		return nil
	}
	allowedVals := make([]string, elemTotal)
	for k, val := range n.ValueMap {
		if val == 0 && !n.allowDefaultValue {
			continue
		}
		i := val
		if !n.allowDefaultValue {
			i -= 1
		}
		allowedVals[i] = strings.ToLower(k)
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
	o, ok := t.(EnumType[T])
	if !ok {
		return false
	}
	return maps.Equal(n.ValueMap, o.ValueMap)
}

func (n EnumType[T]) String() string {
	return "types.EnumType"
}

func (n EnumType[T]) ApplyTerraform5AttributePathStep(_ tftypes.AttributePathStep) (interface{}, error) {
	return nil, fmt.Errorf("unimplemented")
}

type nonEmptyValueValidator[T comparable] struct {
}

var NonEmptyStringValidator = nonEmptyValueValidator[string]{}
var NonEmptyIntValidator = nonEmptyValueValidator[int]{}

func (n nonEmptyValueValidator[T]) Description(_ context.Context) string {
	//TODO implement me
	panic("implement me")
}

func (n nonEmptyValueValidator[T]) MarkdownDescription(_ context.Context) string {
	//TODO implement me
	panic("implement me")
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
