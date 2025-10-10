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

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	sem "github.com/getDragon-dev/dragon-core/semver"
)

var bumpKind, bumpManifest string

type manifest struct {
	Name        string   `yaml:"name"`
	Version     string   `yaml:"version"`
	Description string   `yaml:"description"`
	Tags        []string `yaml:"tags"`
}

var bumpCmd = &cobra.Command{
	Use:   "bump",
	Short: "Bump version in manifest.yaml (patch|minor|major)",
	RunE: func(cmd *cobra.Command, args []string) error {
		b, err := os.ReadFile(bumpManifest)
		if err != nil {
			return err
		}
		var m manifest
		if err := yaml.Unmarshal(b, &m); err != nil {
			return err
		}
		v, err := sem.Parse(m.Version)
		if err != nil {
			return err
		}
		m.Version = v.Bump(bumpKind).String()
		out, err := yaml.Marshal(&m)
		if err != nil {
			return err
		}
		if err := os.WriteFile(bumpManifest, out, 0o644); err != nil {
			return err
		}
		fmt.Println("bumped to", m.Version)
		return nil
	},
}

func init() {
	bumpCmd.Flags().StringVar(&bumpKind, "kind", "patch", "Bump kind: patch|minor|major")
	bumpCmd.Flags().StringVar(&bumpManifest, "file", "manifest.yaml", "Path to manifest.yaml")
	rootCmd.AddCommand(bumpCmd)
}
