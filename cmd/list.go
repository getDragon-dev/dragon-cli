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
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/spf13/cobra"
)

var listAll bool
var listTag string
var listJSON bool

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List blueprints (active registry or all with --all)",
	RunE: func(cmd *cobra.Command, args []string) error {
		type row struct {
			Name, Version, Description, Source string
			Tags                               []string
		}
		rows := []row{}
		if listAll {
			sets, err := loadAllRegistries()
			if err != nil {
				return err
			}
			seen := map[string]bool{}
			for _, s := range sets {
				for _, bp := range s.DB.Blueprints {
					if seen[bp.Name] {
						continue
					}
					if listTag != "" {
						match := false
						for _, t := range bp.Tags {
							if strings.EqualFold(t, listTag) {
								match = true
								break
							}
						}
						if !match {
							continue
						}
					}
					rows = append(rows, row{bp.Name, bp.Version, bp.Description, s.URL, bp.Tags})
					seen[bp.Name] = true
				}
			}
		} else {
			db, err := loadRegistry()
			if err != nil {
				return err
			}
			src, _ := resolveRegistry()
			for _, bp := range db.Blueprints {
				if listTag != "" {
					match := false
					for _, t := range bp.Tags {
						if strings.EqualFold(t, listTag) {
							match = true
							break
						}
					}
					if !match {
						continue
					}
				}
				rows = append(rows, row{bp.Name, bp.Version, bp.Description, src, bp.Tags})
			}
		}
		sort.Slice(rows, func(i, j int) bool { return rows[i].Name < rows[j].Name })
		if listJSON {
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			return enc.Encode(rows)
		}
		if listAll {
			fmt.Println("Registries are searched in configured order. First match wins.")
		}
		for _, r := range rows {
			fmt.Printf("- %s (%s) — %s\n", r.Name, r.Version, r.Description)
		}
		return nil
	},
}

func init() {
	listCmd.Flags().BoolVar(&listAll, "all", false, "Aggregate across all registries")
	listCmd.Flags().StringVar(&listTag, "tag", "", "Filter by tag (exact match)")
	listCmd.Flags().BoolVar(&listJSON, "json", false, "Output JSON")
	rootCmd.AddCommand(listCmd)
}
