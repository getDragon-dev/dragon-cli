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

	"github.com/spf13/cobra"
)

var infoCmd = &cobra.Command{Use: "info <blueprint>", Args: cobra.ExactArgs(1), Short: "Show detailed info about a blueprint",
	RunE: func(cmd *cobra.Command, args []string) error {
		bp, src, err := findBlueprint(args[0])
		if err != nil {
			return err
		}
		if err := applyVersionConstraint(bp.Version); err != nil {
			return err
		}
		fmt.Printf("Name: %s\nVersion: %s\nDescription: %s\nTags: %v\nDownload: %s\nRepo: %s\nPath: %s\nSource Registry: %s\n",
			bp.Name, bp.Version, bp.Description, bp.Tags, bp.DownloadURL, bp.Repo, bp.Path, src)
		return nil
	},
}

func init() { rootCmd.AddCommand(infoCmd) }
