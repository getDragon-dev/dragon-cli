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
	"os/exec"

	"github.com/spf13/cobra"
)

var relAuto bool

var releaseCmd = &cobra.Command{Use: "release", Short: "Run GoReleaser if available", RunE: func(cmd *cobra.Command, args []string) error {
	if _, err := exec.LookPath("goreleaser"); err != nil {
		fmt.Println("goreleaser not found in PATH; run: go install github.com/goreleaser/goreleaser/v2@latest")
		return nil
	}
	c := exec.Command("goreleaser", "release", "--clean")
	c.Stdout, c.Stderr = cmd.OutOrStdout(), cmd.ErrOrStderr()
	return c.Run()
}}

func init() {
	releaseCmd.Flags().BoolVar(&relAuto, "auto", true, "Auto mode (noop flag for now)")
	rootCmd.AddCommand(releaseCmd)
}
