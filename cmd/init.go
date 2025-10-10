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
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var initDir, initName, initDesc string
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new blueprint skeleton in a directory",
	RunE: func(cmd *cobra.Command, args []string) error {
		dir := initDir
		if dir == "" {
			dir = "."
		}
		if initName == "" {
			return fmt.Errorf("--name is required")
		}
		bpDir := filepath.Join(dir, initName)
		if err := os.MkdirAll(filepath.Join(bpDir, "template"), 0o755); err != nil {
			return err
		}
		manifest := fmt.Sprintf("name: %s\nversion: 0.1.0\ndescription: %s\ntags: []\n", initName, initDesc)
		if err := os.WriteFile(filepath.Join(bpDir, "manifest.yaml"), []byte(manifest), 0o644); err != nil {
			return err
		}
		return nil
	},
}

func init() {
	initCmd.Flags().StringVar(&initDir, "dir", ".", "Directory to create the blueprint in")
	initCmd.Flags().StringVar(&initName, "name", "", "Blueprint name (required)")
	initCmd.Flags().StringVar(&initDesc, "desc", "", "Description")
	rootCmd.AddCommand(initCmd)
}
