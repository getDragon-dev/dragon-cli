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
	"archive/zip"
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	coretempl "github.com/getDragon-dev/dragon-core/templates"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var (
	genName        string
	genOut         string
	genRouter      string
	genDB          string
	genRemote      bool
	genVersion     string
	genVarsFile    string
	genSets        []string
	genInteractive bool
)

var genCmd = &cobra.Command{Use: "gen", Short: "Generate a project from a blueprint", ValidArgsFunction: completeBlueprints,
	Run: func(cmd *cobra.Command, args []string) {
		if genName == "" {
			log.Fatal("missing --blueprint/-b name")
		}
		bp, sourceURL, err := findBlueprint(genName)
		if err != nil {
			log.Fatal(err)
		}
		if genVersion != "" && !satisfies(bp.Version, genVersion) {
			log.Fatalf("blueprint version %s does not satisfy constraint %s", bp.Version, genVersion)
		}

		var src string
		if genRemote {
			tmp, err := downloadAndExtractTemplate(bp.DownloadURL)
			if err != nil {
				log.Fatal(err)
			}
			src = tmp
		} else {
			src = filepath.Join("../dragon-blueprints", bp.Path, "template")
			if _, err := os.Stat(src); err != nil {
				log.Fatalf("template not found locally: %s (use --remote to download from %s)", src, sourceURL)
			}
		}

		ctx := coretempl.Context{"Name": genName}
		if genName == "api-service" {
			ctx["Router"], ctx["DB"] = genRouter, genDB
		}
		vars, err := loadUserVars(genVarsFile, genSets)
		if err != nil {
			log.Fatal(err)
		}
		for k, v := range vars {
			ctx[k] = v
		}
		if genInteractive {
			promptAPI(ctx)
		}

		if err := coretempl.RenderDir(src, genOut, ctx); err != nil {
			log.Fatal(err)
		}
		fmt.Println("Generated", genName, "into", genOut)
	},
}

func init() {
	genCmd.Flags().StringVarP(&genName, "blueprint", "b", "", "Blueprint name (required)")
	genCmd.Flags().StringVarP(&genOut, "out", "o", ".", "Output directory")
	genCmd.Flags().StringVar(&genRouter, "router", "servemux", "Router: chi|gorilla|httprouter|servemux (api-service only)")
	genCmd.Flags().StringVar(&genDB, "db", "sqlite-native", "DB: sqlite-native|sqlite-gorm|postgres-native|postgres-gorm|mysql-native|mysql-gorm (api-service only)")
	genCmd.Flags().BoolVar(&genRemote, "remote", false, "Download blueprint from release asset instead of local repo")
	genCmd.Flags().StringVar(&genVersion, "version", "", "Version constraint (e.g. ^1.0, >=1.2.3)")
	genCmd.Flags().StringVar(&genVarsFile, "vars", "", "YAML/JSON file with template variables")
	genCmd.Flags().StringSliceVar(&genSets, "set", nil, "Set template var (key=value), repeatable")
	genCmd.Flags().BoolVar(&genInteractive, "interactive", false, "Prompt for common variables when missing")
	_ = genCmd.MarkFlagRequired("blueprint")
	rootCmd.AddCommand(genCmd)
}

func completeBlueprints(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	sets, err := loadAllRegistries()
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}
	seen := map[string]bool{}
	out := []string{}
	for _, s := range sets {
		for _, bp := range s.DB.Blueprints {
			if seen[bp.Name] {
				continue
			}
			if toComplete == "" || strings.HasPrefix(bp.Name, toComplete) {
				out = append(out, bp.Name)
			}
			seen[bp.Name] = true
		}
	}
	return out, cobra.ShellCompDirectiveNoFileComp
}

func downloadAndExtractTemplate(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode/100 != 2 {
		b, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("GET %s: %d: %s", url, resp.StatusCode, string(b))
	}
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	zr, err := zip.NewReader(bytes.NewReader(b), int64(len(b)))
	if err != nil {
		return "", err
	}
	dst, err := os.MkdirTemp("", "dragon-tpl-")
	if err != nil {
		return "", err
	}
	for _, f := range zr.File {
		if f.FileInfo().IsDir() {
			continue
		}
		if !strings.HasPrefix(f.Name, "template/") {
			continue
		}
		out := filepath.Join(dst, f.Name)
		if err := os.MkdirAll(filepath.Dir(out), 0o755); err != nil {
			return "", err
		}
		rc, err := f.Open()
		if err != nil {
			return "", err
		}
		data, _ := io.ReadAll(rc)
		rc.Close()
		if err := os.WriteFile(out, data, 0o644); err != nil {
			return "", err
		}
	}
	return filepath.Join(dst, "template"), nil
}

func loadUserVars(file string, sets []string) (map[string]any, error) {
	vars := map[string]any{}
	if file != "" {
		b, err := os.ReadFile(file)
		if err != nil {
			return nil, err
		}
		if err := yaml.Unmarshal(b, &vars); err != nil {
			return nil, err
		}
	}
	for _, kv := range sets {
		parts := strings.SplitN(kv, "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("bad --set %q, want key=value", kv)
		}
		vars[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
	}
	return vars, nil
}

func promptAPI(ctx map[string]any) {
	in := bufio.NewReader(os.Stdin)
	ask := func(k, def string) string {
		if v, ok := ctx[k]; ok && fmt.Sprint(v) != "" {
			return fmt.Sprint(v)
		}
		fmt.Printf("%s [%s]: ", k, def)
		s, _ := in.ReadString('\n')
		s = strings.TrimSpace(s)
		if s == "" {
			return def
		}
		return s
	}
	ctx["Module"] = ask("Module", fmt.Sprint(ctx["Module"]))
	ctx["Router"] = ask("Router", fmt.Sprint(ctx["Router"]))
	ctx["DB"] = ask("DB", fmt.Sprint(ctx["DB"]))
}
