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
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"runtime"

	"github.com/spf13/cobra"
)

var (
	registryPath string
)

var rootCmd = &cobra.Command{
	Use:   "dragon",
	Short: "Dragon - project generator and blueprint manager",
}

func init() {
	rootCmd.PersistentFlags().StringVar(&registryPath, "registry", "", "Path or URL to registry.json")
	rootCmd.CompletionOptions.DisableDefaultCmd = true

}

func Execute() {
	ensureConfigDefaults()
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

type Config struct {
	Registries []Registry `json:"registries"`
	Default    string     `json:"default"`
	Order      []string   `json:"order,omitempty"`
}

type Registry struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

func configDir() string {
	if x := os.Getenv("XDG_CONFIG_HOME"); x != "" {
		return filepath.Join(x, "dragon")
	}
	if runtime.GOOS == "windows" {
		if app := os.Getenv("AppData"); app != "" {
			return filepath.Join(app, "dragon")
		}
	}
	return filepath.Join(os.Getenv("HOME"), ".config", "dragon")
}

func configPath() string {
	return filepath.Join(configDir(), "config.json")
}

func readConfig() (Config, error) {
	p := configPath()
	b, err := os.ReadFile(p)
	if err != nil {
		return Config{}, err
	}
	var cfg Config
	if err := json.Unmarshal(b, &cfg); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func writeConfig(cfg Config) error {
	_ = os.MkdirAll(configDir(), 0o755)
	b, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(configPath(), b, 0o644)
}

func ensureConfigDefaults() {
	p := configPath()
	if _, err := os.Stat(p); err == nil {
		return
	}
	_ = os.MkdirAll(configDir(), 0o755)
	cfg := Config{
		Registries: []Registry{{
			Name: "public",
			URL:  "https://getdragon.dev/registry.json",
		}},
		Default: "public",
		Order:   []string{"public"},
	}
	_ = writeConfig(cfg)
}
