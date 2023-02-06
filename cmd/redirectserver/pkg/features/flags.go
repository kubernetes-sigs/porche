// Copyright 2023 The Kubernetes Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package features

import (
	"os"
	"strings"
	"sync"
)

// AllowRegionToBeSpecified controls whether users can force a region with the `region` query-parameter.
var AllowRegionToBeSpecified = Feature{Key: "AllowRegionToBeSpecified"}

// Feature is the type for a feature, that can be activated via the FEATURE_FLAGS env var.
type Feature struct {
	Key string

	initOnce sync.Once
	enabled  bool
}

// IsEnabled returns true if the feature is enabled.
func (f *Feature) IsEnabled() bool {
	f.initOnce.Do(f.init)
	return f.enabled
}

// init populates the enabled value for the feature-flag, parsing the FEATURE_FLAGS env var
func (f *Feature) init() {
	featureFlags := os.Getenv("FEATURE_FLAGS")
	fields := strings.FieldsFunc(featureFlags, func(r rune) bool {
		return r == ','
	})
	for _, field := range fields {
		if strings.EqualFold(field, f.Key) {
			f.enabled = true
		}
	}
}
