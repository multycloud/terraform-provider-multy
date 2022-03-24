package common

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/multycloud/multy/api/proto/common"
	"strings"
)

func SetLocation(providerLocation common.Location, actualLocation common.Location,
	effectiveLocationField *types.String, userProvidedLocationField *types.String) {
	if providerLocation != actualLocation {
		// this was an override, so set the location explictly
		userProvidedLocationField.Value = strings.ToLower(actualLocation.String())
	} else {
		userProvidedLocationField.Null = true
	}

	effectiveLocationField.Value = strings.ToLower(actualLocation.String())
}
