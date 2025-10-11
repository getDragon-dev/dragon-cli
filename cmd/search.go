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
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

var searchAll bool
var searchQuery string
var searchTag string

var searchCmd = &cobra.Command{Use: "search", Short: "Search blueprints by name/description/tags",
	RunE: func(cmd *cobra.Command, args []string) error {
		q := strings.ToLower(strings.TrimSpace(searchQuery))
		if q == "" && searchTag == "" {
			return fmt.Errorf("--query or --tag required")
		}
		sets := []struct {
			URL   string
			Names []string
		}{}
		if searchAll {
			loaded, err := loadAllRegistries()
			if err != nil {
				return err
			}
			for _, s := range loaded {
				names := []string{}
				for _, bp := range s.DB.Blueprints {
					if match(bp.Name, bp.Description, bp.Tags, q, searchTag) {
						names = append(names, bp.Name)
					}
				}
				sets = append(sets, struct {
					URL   string
					Names []string
				}{s.URL, names})
			}
		} else {
			db, err := loadRegistry()
			if err != nil {
				return err
			}
			src, _ := resolveRegistry()
			names := []string{}
			for _, bp := range db.Blueprints {
				if match(bp.Name, bp.Description, bp.Tags, q, searchTag) {
					names = append(names, bp.Name)
				}
			}
			sets = append(sets, struct {
				URL   string
				Names []string
			}{src, names})
		}
		for _, s := range sets {
			fmt.Println(s.URL)
			for _, n := range s.Names {
				fmt.Printf("  - %s\n", n)
			}
		}
		return nil
	},
}

func match(name, desc string, tags []string, q, tag string) bool {
	if tag != "" {
		ok := false
		for _, t := range tags {
			if strings.EqualFold(t, tag) {
				ok = true
				break
			}
		}
		if !ok {
			return false
		}
	}
	if q == "" {
		return true
	}
	q = strings.ToLower(q)
	if strings.Contains(strings.ToLower(name), q) {
		return true
	}
	if strings.Contains(strings.ToLower(desc), q) {
		return true
	}
	for _, t := range tags {
		if strings.Contains(strings.ToLower(t), q) {
			return true
		}
	}
	return false
}

func init() {
	searchCmd.Flags().BoolVar(&searchAll, "all", true, "search across all registries by default")
	searchCmd.Flags().StringVar(&searchQuery, "query", "", "substring to search for")
	searchCmd.Flags().StringVar(&searchTag, "tag", "", "filter by exact tag")
	rootCmd.AddCommand(searchCmd)
}
