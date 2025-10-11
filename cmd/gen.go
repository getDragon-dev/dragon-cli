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
)

var (
	genName   string
	genOut    string
	genRouter string
	genDB     string
	genRemote bool
)

var genCmd = &cobra.Command{
	Use:   "gen",
	Short: "Generate a project from a blueprint",
	Run: func(cmd *cobra.Command, args []string) {
		if genName == "" {
			log.Fatal("missing --blueprint/-b name")
		}
		bp, sourceURL, err := findBlueprint(genName)
		if err != nil {
			log.Fatal(err)
		}
		if err := applyVersionConstraint(bp.Version); err != nil {
			log.Fatal(err)
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

		ctx := coretempl.Context{
			"Name": genName,
		}
		if genName == "api-service" {
			ctx["Router"], ctx["DB"] = genRouter, genDB
		}
		user, err := loadUserVars()
		if err != nil {
			log.Fatal(err)
		}
		for k, v := range user {
			ctx[k] = v
		}
		promptIfNeeded(ctx, genName)

		if err := coretempl.RenderDir(src, genOut, ctx); err != nil {
			log.Fatal(err)
		}
		fmt.Println("Generated", genName, "into", genOut)
	},
}

func init() {
	genCmd.Flags().StringVarP(&genName, "blueprint", "b", "", "Blueprint name (required)")
	genCmd.Flags().StringVarP(&genOut, "out", "o", ".", "Output directory")
	genCmd.Flags().StringVar(&genRouter, "router", "chi", "Router: chi|gorilla|httprouter|servemux (api-service only)")
	genCmd.Flags().StringVar(&genDB, "db", "postgres-gorm", "DB: sqlite-native|sqlite-gorm|postgres-native|postgres-gorm|mysql-native|mysql-gorm (api-service only)")
	genCmd.Flags().BoolVar(&genRemote, "remote", false, "Download blueprint from release asset instead of local repo")
	_ = genCmd.MarkFlagRequired("blueprint")
	rootCmd.AddCommand(genCmd)
}

func completeBlueprints(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	sets, err := loadAllRegistries()
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}
	seen := map[string]bool{}
	suggestions := []string{}
	for _, s := range sets {
		for _, bp := range s.DB.Blueprints {
			if seen[bp.Name] {
				continue
			}
			if toComplete == "" || strings.HasPrefix(bp.Name, toComplete) {
				suggestions = append(suggestions, bp.Name)
			}
			seen[bp.Name] = true
		}
	}
	return suggestions, cobra.ShellCompDirectiveNoFileComp
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
