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
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	corereg "github.com/getDragon-dev/dragon-core/registry"
)

func resolveRegistry() (string, error) {
	if registryPath != "" {
		return registryPath, nil
	}
	cfg, err := readConfig()
	if err != nil {
		return "", err
	}
	if cfg.Default == "" || len(cfg.Registries) == 0 {
		return "", fmt.Errorf("no default registry configured")
	}
	for _, r := range cfg.Registries {
		if r.Name == cfg.Default {
			return r.URL, nil
		}
	}
	return "", fmt.Errorf("default registry %q not found", cfg.Default)
}

func resolveOrder() ([]string, error) {
	if registryPath != "" {
		return []string{registryPath}, nil
	}
	cfg, err := readConfig()
	if err != nil {
		return nil, err
	}
	urls := []string{}
	if len(cfg.Order) > 0 {
		for _, name := range cfg.Order {
			for _, r := range cfg.Registries {
				if r.Name == name {
					urls = append(urls, r.URL)
					break
				}
			}
		}
		return urls, nil
	}
	for _, r := range cfg.Registries {
		if r.Name == cfg.Default {
			urls = append(urls, r.URL)
			break
		}
	}
	for _, r := range cfg.Registries {
		if r.Name != cfg.Default {
			urls = append(urls, r.URL)
		}
	}
	return urls, nil
}

func loadRegistry() (corereg.Database, error) {
	loc, err := resolveRegistry()
	if err != nil {
		return corereg.Database{}, err
	}

	if strings.HasPrefix(loc, "http://") || strings.HasPrefix(loc, "https://") {
		resp, err := http.Get(loc)
		if err != nil {
			return corereg.Database{}, err
		}
		defer resp.Body.Close()
		if resp.StatusCode/100 != 2 {
			b, _ := io.ReadAll(resp.Body)
			return corereg.Database{}, fmt.Errorf("GET %s: %d: %s", loc, resp.StatusCode, string(b))
		}
		data, err := io.ReadAll(resp.Body)
		if err != nil {
			return corereg.Database{}, err
		}
		var db corereg.Database
		if err := json.Unmarshal(data, &db); err != nil {
			return corereg.Database{}, fmt.Errorf("parse registry %s: %w", loc, err)
		}
		return db, nil
	}

	// fallback to local file
	return corereg.Load(loc)
}

func loadAllRegistries() ([]struct {
	URL string
	DB  corereg.Database
}, error) {
	urls, err := resolveOrder()
	if err != nil {
		return nil, err
	}
	res := make([]struct {
		URL string
		DB  corereg.Database
	}, 0, len(urls))

	for _, u := range urls {
		var db corereg.Database
		if strings.HasPrefix(u, "http://") || strings.HasPrefix(u, "https://") {
			resp, err := http.Get(u)
			if err != nil {
				return nil, fmt.Errorf("fetch %s: %w", u, err)
			}
			defer resp.Body.Close()
			if resp.StatusCode/100 != 2 {
				b, _ := io.ReadAll(resp.Body)
				return nil, fmt.Errorf("GET %s: %d: %s", u, resp.StatusCode, string(b))
			}
			data, _ := io.ReadAll(resp.Body)
			if err := json.Unmarshal(data, &db); err != nil {
				return nil, fmt.Errorf("decode %s: %w", u, err)
			}
		} else {
			db, err = corereg.Load(u)
			if err != nil {
				return nil, fmt.Errorf("load %s: %w", u, err)
			}
		}
		res = append(res, struct {
			URL string
			DB  corereg.Database
		}{URL: u, DB: db})
	}
	return res, nil
}

// Simple semver check for forms like 1.2.3
func parseSemver(v string) (int, int, int) {
	v = strings.TrimSpace(strings.TrimPrefix(v, "v"))
	parts := strings.SplitN(v, ".", 3)
	to := func(s string) int { n, _ := strconv.Atoi(s); return n }
	maj, min, patch := 0, 0, 0
	if len(parts) > 0 {
		maj = to(parts[0])
	}
	if len(parts) > 1 {
		min = to(parts[1])
	}
	if len(parts) > 2 {
		patch = to(parts[2])
	}
	return maj, min, patch
}

func satisfies(v string, constraint string) bool {
	if constraint == "" {
		return true
	}
	vM, vm, vp := parseSemver(v)
	// support basic forms: ">=1.2.3", "<=1.2.3", ">1.2.3", "<1.2.3", "=1.2.3", "^1.2", "~1.2" (caret/tilde minor handling)
	op := ""
	ver := constraint
	for _, o := range []string{">=", "<=", ">", "<", "=", "^", "~"} {
		if strings.HasPrefix(constraint, o) {
			op = o
			ver = strings.TrimSpace(strings.TrimPrefix(constraint, o))
			break
		}
	}
	tM, tm, tp := parseSemver(ver)
	cmp := func(aM, aN, aP, bM, bN, bP int) int {
		if aM != bM {
			if aM < bM {
				return -1
			}
			return 1
		}
		if aN != bN {
			if aN < bN {
				return -1
			}
			return 1
		}
		if aP != bP {
			if aP < bP {
				return -1
			}
			return 1
		}
		return 0
	}
	switch op {
	case ">=":
		return cmp(vM, vm, vp, tM, tm, tp) >= 0
	case "<=":
		return cmp(vM, vm, vp, tM, tm, tp) <= 0
	case ">":
		return cmp(vM, vm, vp, tM, tm, tp) > 0
	case "<":
		return cmp(vM, vm, vp, tM, tm, tp) < 0
	case "=":
		return cmp(vM, vm, vp, tM, tm, tp) == 0
	case "^": // compatible with same major; require >= t and < (t.major+1).0.0
		return cmp(vM, vm, vp, tM, tm, tp) >= 0 && vM == tM
	case "~": // same minor; require >= t and < t.major.(t.minor+1).0
		return cmp(vM, vm, vp, tM, tm, tp) >= 0 && vM == tM && vm == tm
	default:
		// plain version means equality
		return cmp(vM, vm, vp, tM, tm, tp) == 0
	}
}

// findBlueprint searches the active registry (if --registry is set) or, otherwise,
// all registries in the configured search order. It returns the blueprint and the
// source registry URL where it was found.
func findBlueprint(name string) (corereg.Blueprint, string, error) {
	if registryPath != "" {
		db, err := loadRegistry()
		if err != nil {
			return corereg.Blueprint{}, "", err
		}
		bp, err := corereg.Find(db, name)
		if err != nil {
			return corereg.Blueprint{}, "", err
		}
		loc, _ := resolveRegistry()
		return *bp, loc, nil
	}

	sets, err := loadAllRegistries()
	if err != nil {
		return corereg.Blueprint{}, "", err
	}
	for _, s := range sets {
		if bp, err := corereg.Find(s.DB, name); err == nil {
			return *bp, s.URL, nil
		}
	}
	return corereg.Blueprint{}, "", fmt.Errorf("blueprint %q not found in any configured registry", name)
}
