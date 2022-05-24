/*
 * Copyright 2019 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 *
 */

package utils

import (
	"strings"
)

// IsEmptyString checks whether a string is empty
func IsEmptyString(s *string) bool {
	return s == nil || *s == ""
}

// IsStringPtrValueEqual checks whether the value of string pointer `a` is equal to value of string `b`.
func IsStringPtrValueEqual(a *string, b string) bool {
	return a != nil && *a == b
}

// StringValue returns an empty string (value/ptr)
func StringValue(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// StringEqual compares to strings
func StringEqual(a, b *string) bool {
	return a == b || (a != nil && b != nil && *a == *b)
}

// SetStringValue sets an optional string value in a string map
// if the value is defined and not empty
func SetStringValue(values map[string]interface{}, key string, value *string) {
	if !IsEmptyString(value) {
		values[key] = *value
	}
}

// SimpleMatch returns whether the given pattern matches the given text.
// It also returns a score indicating the match between `pattern` and `text`. The higher the score the higher the match.
// Only simple wildcard patterns are supposed to be passed, e.g. '*', 'tex*'.
func SimpleMatch(pattern, text string) (bool, int) {
	const wildcard = "*"
	if pattern == wildcard {
		return true, 0
	}
	if pattern == text {
		return true, len(text)
	}
	if strings.HasSuffix(pattern, wildcard) && strings.HasPrefix(text, pattern[:len(pattern)-1]) {
		s := strings.SplitAfterN(text, pattern[:len(pattern)-1], 2)
		return true, len(s[0])
	}
	if strings.HasPrefix(pattern, wildcard) && strings.HasSuffix(text, pattern[1:]) {
		i := strings.LastIndex(text, pattern[1:])
		return true, len(text) - i
	}

	return false, 0
}
