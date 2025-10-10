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
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var updateSrc string
var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update the local registry.json from a URL",
	RunE: func(cmd *cobra.Command, args []string) error {
		if updateSrc == "" {
			return fmt.Errorf("--from URL is required")
		}
		if !strings.HasPrefix(updateSrc, "http") {
			return fmt.Errorf("--from must be an HTTP(S) URL")
		}
		resp, err := http.Get(updateSrc)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		if resp.StatusCode != 200 {
			return fmt.Errorf("http %d", resp.StatusCode)
		}
		b, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		return os.WriteFile(registryPath, b, 0o644)
	},
}

func init() {
	updateCmd.Flags().StringVar(&updateSrc, "from", "", "URL to registry.json")
	rootCmd.AddCommand(updateCmd)
}
