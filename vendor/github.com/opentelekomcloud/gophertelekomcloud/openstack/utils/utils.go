package utils

import (
	"os"
	"reflect"
	"strings"

	golangsdk "github.com/opentelekomcloud/gophertelekomcloud"
)

func DeleteNotPassParams(params *map[string]interface{}, notPassParams []string) {
	for _, i := range notPassParams {
		delete(*params, i)
	}
}

// merges two interfaces. In cases where a value is defined for both 'overridingInterface' and
// 'inferiorInterface' the value in 'overridingInterface' will take precedence.
func MergeInterfaces(overridingInterface, inferiorInterface interface{}) interface{} {
	switch overriding := overridingInterface.(type) {
	case map[string]interface{}:
		interfaceMap, ok := inferiorInterface.(map[string]interface{})
		if !ok {
			return overriding
		}
		for k, v := range interfaceMap {
			if overridingValue, ok := overriding[k]; ok {
				overriding[k] = MergeInterfaces(overridingValue, v)
			} else {
				overriding[k] = v
			}
		}
	case []interface{}:
		list, ok := inferiorInterface.([]interface{})
		if !ok {
			return overriding
		}
		overriding = append(overriding, list...)
		return overriding
	case nil:
		// mergeClouds(nil, map[string]interface{...}) -> map[string]interface{...}
		v, ok := inferiorInterface.(map[string]interface{})
		if ok {
			return v
		}
	}
	// We don't want to override with empty values
	if reflect.DeepEqual(overridingInterface, nil) || reflect.DeepEqual(reflect.Zero(reflect.TypeOf(overridingInterface)).Interface(), overridingInterface) {
		return inferiorInterface
	} else {
		return overridingInterface
	}
}

func PrependString(item string, slice []string) []string {
	result := make([]string, len(slice)+1)
	result[0] = item
	for i, v := range slice {
		result[i+1] = v
	}
	return result
}

func In(item interface{}, slice interface{}) bool {
	for _, it := range slice.([]interface{}) {
		if reflect.DeepEqual(item, it) {
			return true
		}
	}
	return false
}

// GetRegion returns the region that was specified in the auth options. If a
// region was not set it returns value from env OS_REGION_NAME
func GetRegion(authOpts golangsdk.AuthOptions) string {
	name := authOpts.TenantName
	if name == "" {
		name = authOpts.DelegatedProject
	}
	region := strings.Split(name, "_")[0]
	return getenv("OS_REGION_NAME", region)
}

// GetRegionFromAKSK returns the region that was specified in the auth options.
// If a region was not set it returns value from env OS_REGION_NAME
func GetRegionFromAKSK(authOpts golangsdk.AKSKAuthOptions) string {
	name := authOpts.ProjectName
	if name == "" {
		name = authOpts.DelegatedProject
	}
	region := strings.Split(name, "_")[0]
	return getenv("OS_REGION_NAME", region)
}

// getenv returns value from env is present or default value
func getenv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
