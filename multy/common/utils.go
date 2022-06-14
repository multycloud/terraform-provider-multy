package common

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/multycloud/multy/api/proto/errorspb"
	"github.com/multycloud/multy/api/proto/resourcespb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"strings"
)

const DebugMode = true

func StringToRuleDirection(dir string) resourcespb.Direction {
	if strings.EqualFold(dir, "both") {
		dir = "both_directions"
	}
	return resourcespb.Direction(resourcespb.Direction_value[strings.ToUpper(dir)])
}

func RuleDirectionToString(dir resourcespb.Direction) string {
	if strings.EqualFold(dir.String(), "BOTH_DIRECTIONS") {
		return "both"
	}
	return strings.ToLower(dir.String())
}

func ParseGrpcErrors(err error) string {
	s, ok := status.FromError(err)
	if !ok {
		return err.Error()
	}

	str := ""

	// user screwed up
	if s.Code() == codes.InvalidArgument {
		str += s.Message()
		for _, detail := range s.Details() {
			if e, ok := detail.(*errorspb.ResourceValidationError); ok {
				str += "\n" + e.ErrorMessage
			}
		}
	} else if s.Code() == codes.Internal {
		str += "something went wrong: " + s.Message()
		if DebugMode {
			for _, detail := range s.Details() {
				if e, ok := detail.(*errorspb.InternalErrorDetails); ok {
					str += "\n" + e.ErrorMessage
				}
			}
		}
	} else if s.Code() == codes.FailedPrecondition {
		str += s.Message()
		for _, detail := range s.Details() {
			if e, ok := detail.(*errorspb.DeploymentErrorDetails); ok {
				str += "\n" + e.ErrorMessage
			}
		}
	} else if s.Code() == codes.Unavailable {
		str = "Server is unavailable. Please try again in a few minutes.\n"
		if DebugMode {
			str += s.Message()
		}
	} else if s.Code() == codes.NotFound {
		str = s.Message()
	} else if s.Code() == codes.PermissionDenied {
		str = "Permission denied: " + s.Message()
	} else {
		str += s.String()
	}
	return str

}

func DefaultToNull[OutT attr.Value](t any) OutT {
	var s attr.Value
	switch t.(type) {
	case string:
		s = types.String{
			Null:  t.(string) == "",
			Value: t.(string),
		}
	case int, int64, int32:
		s = types.Int64{
			Null:  t.(int64) == 0,
			Value: t.(int64),
		}
	}
	return s.(OutT)
}

func NullToDefault[OutT any, T attr.Value](t T) OutT {
	returnVal := new(OutT)
	value, err := t.ToTerraformValue(nil)
	if err != nil {
		panic(err)
	}
	if value.IsNull() {
		return *returnVal
	}

	err = value.As(returnVal)
	if err != nil {
		panic(err)
	}

	return *returnVal
}

func DefaultSliceToNull[T attr.Value](t []T) []T {
	if len(t) == 0 {
		return nil
	}
	return t
}
