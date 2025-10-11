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
	"path/filepath"

	"github.com/spf13/cobra"
)

var getCmd = &cobra.Command{Use: "get <blueprint>", Args: cobra.ExactArgs(1), Short: "Fetch and render a blueprint (remote)",
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		out := filepath.Join(".", name)
		genName = name
		genOut = out
		genRemote = true
		fmt.Printf("Generating %s into %s using remote asset...\n", name, out)
		return genCmd.RunE(cmd, []string{})
	},
}

func init() { rootCmd.AddCommand(getCmd) }
