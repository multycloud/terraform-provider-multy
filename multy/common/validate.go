package common

var (
	Clouds    = []string{"aws", "azure"}
	Locations = []string{"ireland", "uk", "us-east"}
)

//func ValidateLocation(i interface{}, k string) (warnings []string, errors []error) {
//	v, ok := i.(string)
//	if !ok {
//		errors = append(errors, fmt.Errorf("expected type of %s to be string", k))
//		return warnings, errors
//	}
//
//	location := common_proto.Location(common_proto.Location_value[strings.ToUpper(v)])
//	if location == 0 {
//		errors = append(errors, fmt.Errorf("expected %s to be one of %v, got %s", k, GetLocationNames(), v))
//	}
//
//	return warnings, errors
//}
