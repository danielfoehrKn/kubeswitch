package utils

import "os"

// Ptr Returns the pointer to any type T
func Ptr[T any](v T) *T {
	return &v
}

func Contains[T comparable](slice []T, element T) bool {
	for _, item := range slice {
		if item == element {
			return true
		}
	}
	return false
}

// EnumSliceToStringSlice is a generic function to convert a slice of any type T
// that has the underlying type 'string' to a slice of string.
// The constraint ~string allows T to be any type whose
// underlying type is string (like the enum types from the generated STACKIT SDK modules).
func EnumSliceToStringSlice[T ~string](inputSlice []T) []string {
	result := make([]string, len(inputSlice))

	for i, element := range inputSlice {
		result[i] = string(element)
	}

	return result
}

func GetEnvOrDefault(envVar, defaultValue string) string {
	if value := os.Getenv(envVar); value != "" {
		return value
	}
	return defaultValue
}
