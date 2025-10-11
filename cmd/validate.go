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
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

type vmanifest struct {
	Name        string   `yaml:"name"`
	Version     string   `yaml:"version"`
	Description string   `yaml:"description"`
	Tags        []string `yaml:"tags"`
}

var validateFile string

var validateCmd = &cobra.Command{Use: "validate", Short: "Validate a blueprint manifest.yaml",
	RunE: func(cmd *cobra.Command, args []string) error {
		b, err := os.ReadFile(validateFile)
		if err != nil {
			return err
		}
		var m vmanifest
		if err := yaml.Unmarshal(b, &m); err != nil {
			return err
		}
		if m.Name == "" {
			return errors.New("name is required")
		}
		if m.Version == "" {
			return errors.New("version is required")
		}
		fmt.Println("OK:", validateFile)
		return nil
	},
}

func init() {
	validateCmd.Flags().StringVar(&validateFile, "file", "manifest.yaml", "Path to manifest.yaml")
	rootCmd.AddCommand(validateCmd)
}
