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
	"os"
	"path/filepath"

	corereg "github.com/getDragon-dev/dragon-core/registry"
	coretempl "github.com/getDragon-dev/dragon-core/templates"
	"github.com/spf13/cobra"
)

var genName, genOut string
var genCmd = &cobra.Command{
	Use:   "gen",
	Short: "Generate a project from a blueprint",
	Run: func(cmd *cobra.Command, args []string) {
		if genName == "" {
			log.Fatal("missing --blueprint/-b name")
		}
		db, err := corereg.Load(registryPath)
		if err != nil {
			log.Fatal(err)
		}
		bp, err := corereg.Find(db, genName)
		if err != nil {
			log.Fatal(err)
		}

		src := filepath.Join("../dragon-blueprints", bp.Path, "template")
		if _, err := os.Stat(src); err != nil {
			log.Fatalf("template not found: %s", src)
		}
		if err := coretempl.RenderDir(src, genOut, coretempl.Context{"Name": genName}); err != nil {
			log.Fatal(err)
		}
		fmt.Println("Generated", genName, "into", genOut)
	},
}

func init() {
	genCmd.Flags().StringVarP(&genName, "blueprint", "b", "", "Blueprint name (required)")
	genCmd.Flags().StringVarP(&genOut, "out", "o", ".", "Output directory")
	_ = genCmd.MarkFlagRequired("blueprint")
	rootCmd.AddCommand(genCmd)
}
