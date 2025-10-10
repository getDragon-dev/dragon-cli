// Copyright 2025 getDragon-dev
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at http://www.apache.org/licenses/LICENSE-2.0
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	corereg "github.com/getDragon-dev/dragon-core/registry"
	coretempl "github.com/getDragon-dev/dragon-core/templates"
)

func main() {
	genCmd := flag.NewFlagSet("gen", flag.ExitOnError)
	blueprint := genCmd.String("b", "", "Blueprint name")
	out := genCmd.String("o", ".", "Output directory")
	regPath := genCmd.String("registry", "../dragon-registry/registry.json", "Path or URL to registry.json")

	relCmd := flag.NewFlagSet("release", flag.ExitOnError)
	bump := relCmd.String("bump", "patch", "Bump type: patch|minor|major")

	if len(os.Args) < 2 {
		fmt.Println("dragon <gen|release> [options]")
		os.Exit(1)
	}
	switch os.Args[1] {
	case "gen":
		genCmd.Parse(os.Args[2:])
		if *blueprint == "" {
			log.Fatal("missing -b <blueprint>")
		}
		db, err := corereg.Load(*regPath)
		if err != nil {
			log.Fatal(err)
		}
		bp, err := corereg.Find(db, *blueprint)
		if err != nil {
			log.Fatal(err)
		}
		// For demo/local use, assume blueprints are present in ../dragon-blueprints
		src := filepath.Join("../dragon-blueprints", bp.Path, "template")
		if err := coretempl.RenderDir(src, *out, coretempl.Context{"Name": *blueprint}); err != nil {
			log.Fatal(err)
		}
		fmt.Println("Generated", *blueprint, "into", *out)
	case "release":
		relCmd.Parse(os.Args[2:])
		fmt.Println("Release flow placeholder; bump:", *bump)
	default:
		fmt.Println("unknown command")
		os.Exit(1)
	}
}
