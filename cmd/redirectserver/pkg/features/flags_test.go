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
	"fmt"
	"os"
	"testing"
)

func TestFeatureFlags(t *testing.T) {
	grid := []struct {
		Key         string
		EnvVar      string
		WantEnabled bool
	}{
		{Key: "Foo", EnvVar: "", WantEnabled: false},
		{Key: "Foo", EnvVar: ",", WantEnabled: false},
		{Key: "Foo", EnvVar: "Foo", WantEnabled: true},
		{Key: "Bar", EnvVar: "Foo", WantEnabled: false},
		{Key: "Bar", EnvVar: "Bar,Foo", WantEnabled: true},
		{Key: "Foo", EnvVar: ",Bar,,Foo,,,", WantEnabled: true},
		{Key: "Bar", EnvVar: "bar", WantEnabled: true},
		{Key: "Bar", EnvVar: "bart", WantEnabled: false},
		{Key: "Bar", EnvVar: "rebar", WantEnabled: false},
	}

	for _, g := range grid {
		g := g
		name := fmt.Sprintf("Key=%s,EnvVar=%s", g.Key, g.EnvVar)
		t.Run(name, func(t *testing.T) {
			oldEnvVar := os.Getenv("FEATURE_FLAGS")
			defer func() { os.Setenv("FEATURE_FLAGS", oldEnvVar) }()
			os.Setenv("FEATURE_FLAGS", g.EnvVar)

			ff := Feature{Key: g.Key}
			got := ff.IsEnabled()
			if got != g.WantEnabled {
				t.Errorf("IsEnabled did not give expected value; got=%v, want=%v", got, g.WantEnabled)
			}
		})
	}
}
