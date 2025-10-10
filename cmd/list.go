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
	"log"

	corereg "github.com/getDragon-dev/dragon-core/registry"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List available blueprints from registry.json",
	Run: func(cmd *cobra.Command, args []string) {
		db, err := corereg.Load(registryPath)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("Available Blueprints:")
		for _, bp := range db.Blueprints {
			fmt.Printf("- %s (%s) — %s\n", bp.Name, bp.Version, bp.Description)
		}
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
