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

var infoCmd = &cobra.Command{
	Use:   "info <blueprint>",
	Short: "Show detailed info about a blueprint",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		db, err := corereg.Load(registryPath)
		if err != nil {
			log.Fatal(err)
		}
		bp, err := corereg.Find(db, args[0])
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("Name: %s\nVersion: %s\nDescription: %s\nTags: %v\nDownload: %s\nRepo: %s\nPath: %s\n",
			bp.Name, bp.Version, bp.Description, bp.Tags, bp.DownloadURL, bp.Repo, bp.Path)
	},
}

func init() {
	rootCmd.AddCommand(infoCmd)
}
