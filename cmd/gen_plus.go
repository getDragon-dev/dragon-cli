/*
 * // Copyright 2025 getDragon-dev
 * // Licensed under the Apache License, Version 2.0 (the "License");
 * // you may not use this file except in compliance with the License.
 * // You may obtain a copy of the License at http://www.apache.org/licenses/LICENSE-2.0
 * // Unless required by applicable law or agreed to in writing, software
 * // distributed under the License is distributed on an "AS IS" BASIS,
 * // WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * // See the License for the specific language governing permissions and
 * // limitations under the License.
 */

package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

var genVersion string
var genVarsFile string
var genSets []string
var genInteractive bool

func init() {
	genCmd.Flags().StringVar(&genVersion, "version", "", "Version constraint (e.g. ^1.0, >=1.2.3)")
	genCmd.Flags().StringSliceVar(&genSets, "set", nil, "Set template var (key=value), repeatable")
	genCmd.Flags().StringVar(&genVarsFile, "vars", "", "YAML/JSON file with template variables")
	genCmd.Flags().BoolVar(&genInteractive, "interactive", false, "Prompt for common variables when missing")
}

func applyVersionConstraint(bpVersion string) error {
	if genVersion == "" {
		return nil
	}
	if !satisfies(bpVersion, genVersion) {
		return fmt.Errorf("blueprint version %s does not satisfy constraint %s", bpVersion, genVersion)
	}
	return nil
}

func loadUserVars() (map[string]any, error) {
	vars := map[string]any{}
	if genVarsFile != "" {
		b, err := os.ReadFile(genVarsFile)
		if err != nil {
			return nil, err
		}
		if strings.HasSuffix(strings.ToLower(genVarsFile), ".json") {
			if err := yaml.Unmarshal(b, &vars); err != nil {
				return nil, err
			}
		} else {
			if err := yaml.Unmarshal(b, &vars); err != nil {
				return nil, err
			}
		}
	}
	for _, kv := range genSets {
		parts := strings.SplitN(kv, "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("bad --set %q, want key=value", kv)
		}
		vars[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
	}
	return vars, nil
}

func promptIfNeeded(ctx map[string]any, name string) {
	if !genInteractive {
		return
	}
	in := bufio.NewReader(os.Stdin)
	req := func(k, def string) string {
		if v, ok := ctx[k]; ok && fmt.Sprint(v) != "" {
			return fmt.Sprint(v)
		}
		fmt.Printf("%s [%s]: ", k, def)
		s, _ := in.ReadString('\n')
		s = strings.TrimSpace(s)
		if s == "" {
			return def
		}
		return s
	}
	if name == "api-service" {
		ctx["Router"] = req("Router", fmt.Sprint(ctx["Router"]))
		ctx["DB"] = req("DB", fmt.Sprint(ctx["DB"]))
	}
}
