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
	"net/url"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var registryCmd = &cobra.Command{Use: "registry", Short: "Manage registries"}

var registryListCmd = &cobra.Command{Use: "list", RunE: func(cmd *cobra.Command, args []string) error {
	cfg, err := readConfig()
	if err != nil {
		return err
	}
	fmt.Println("Registries:")
	for _, r := range cfg.Registries {
		mark := " "
		if r.Name == cfg.Default {
			mark = "*"
		}
		fmt.Printf("%s %s -> %s\n", mark, r.Name, r.URL)
	}
	if len(cfg.Order) > 0 {
		fmt.Printf("Order: %s\n", strings.Join(cfg.Order, ", "))
	}
	return nil
}}

var (
	regName string
	regURL  string
)

var registryAddCmd = &cobra.Command{Use: "add", RunE: func(cmd *cobra.Command, args []string) error {
	if regName == "" || regURL == "" {
		return errors.New("--name and --url required")
	}
	cfg, _ := readConfig()
	found := false
	for i := range cfg.Registries {
		if cfg.Registries[i].Name == regName {
			cfg.Registries[i].URL = regURL
			found = true
			break
		}
	}
	if !found {
		cfg.Registries = append(cfg.Registries, Registry{Name: regName, URL: regURL})
	}
	if cfg.Default == "" {
		cfg.Default = regName
	}
	if len(cfg.Order) == 0 {
		cfg.Order = []string{cfg.Default}
	}
	return writeConfig(cfg)
}}

var registryRemoveCmd = &cobra.Command{Use: "remove <name>", Args: cobra.ExactArgs(1), RunE: func(cmd *cobra.Command, args []string) error {
	name := args[0]
	cfg, err := readConfig()
	if err != nil {
		return err
	}
	out := make([]Registry, 0, len(cfg.Registries))
	for _, r := range cfg.Registries {
		if r.Name != name {
			out = append(out, r)
		}
	}
	cfg.Registries = out
	if cfg.Default == name {
		cfg.Default = ""
	}
	ord := []string{}
	for _, s := range cfg.Order {
		if s != name {
			ord = append(ord, s)
		}
	}
	cfg.Order = ord
	return writeConfig(cfg)
}}

var registryDefaultCmd = &cobra.Command{Use: "set-default <name>", Args: cobra.ExactArgs(1), RunE: func(cmd *cobra.Command, args []string) error {
	name := args[0]
	cfg, err := readConfig()
	if err != nil {
		return err
	}
	ok := false
	for _, r := range cfg.Registries {
		if r.Name == name {
			ok = true
			break
		}
	}
	if !ok {
		return fmt.Errorf("registry %q not found", name)
	}
	cfg.Default = name
	ord := []string{name}
	for _, s := range cfg.Order {
		if s != name {
			ord = append(ord, s)
		}
	}
	cfg.Order = ord
	return writeConfig(cfg)
}}

var registryUseCmd = &cobra.Command{Use: "use <url-or-path>", Args: cobra.ExactArgs(1), RunE: func(cmd *cobra.Command, args []string) error {
	target := args[0]
	name := autoName(target)
	cfg, _ := readConfig()
	found := false
	for i := range cfg.Registries {
		if cfg.Registries[i].Name == name {
			cfg.Registries[i].URL = target
			found = true
			break
		}
	}
	if !found {
		cfg.Registries = append(cfg.Registries, Registry{Name: name, URL: target})
	}
	cfg.Default = name
	ord := []string{name}
	for _, s := range cfg.Order {
		if s != name {
			ord = append(ord, s)
		}
	}
	cfg.Order = ord
	return writeConfig(cfg)
}}

var orderSetInput string
var registryOrderSetCmd = &cobra.Command{Use: "order set", RunE: func(cmd *cobra.Command, args []string) error {
	if orderSetInput == "" {
		return errors.New("--names a,b,c required")
	}
	cfg, err := readConfig()
	if err != nil {
		return err
	}
	names := strings.Split(orderSetInput, ",")
	set := map[string]bool{}
	for _, n := range names {
		n = strings.TrimSpace(n)
		if n != "" {
			set[n] = true
		}
	}
	for n := range set {
		ok := false
		for _, r := range cfg.Registries {
			if r.Name == n {
				ok = true
				break
			}
		}
		if !ok {
			return fmt.Errorf("unknown registry in order: %s", n)
		}
	}
	cfg.Order = []string{}
	for _, n := range names {
		n = strings.TrimSpace(n)
		if n != "" {
			cfg.Order = append(cfg.Order, n)
		}
	}
	return writeConfig(cfg)
}}

func autoName(u string) string {
	if strings.HasPrefix(u, "http://") || strings.HasPrefix(u, "https://") {
		if pu, err := url.Parse(u); err == nil && pu.Host != "" {
			return pu.Host
		}
	}
	base := filepath.Base(u)
	if base == "." || base == ".." || base == "" {
		base = "local"
	}
	return base
}

func init() {
	registryAddCmd.Flags().StringVar(&regName, "name", "", "Registry name")
	registryAddCmd.Flags().StringVar(&regURL, "url", "", "Registry URL or local path")
	registryOrderSetCmd.Flags().StringVar(&orderSetInput, "names", "", "Comma-separated registry names in desired order")
	registryCmd.AddCommand(registryListCmd, registryAddCmd, registryRemoveCmd, registryDefaultCmd, registryUseCmd, registryOrderSetCmd)
	rootCmd.AddCommand(registryCmd)
}
