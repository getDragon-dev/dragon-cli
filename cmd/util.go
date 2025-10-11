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

func loadLocation(loc string) (corereg.Database, error) {
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
		var db corereg.Database
		data, _ := io.ReadAll(resp.Body)
		if err := json.Unmarshal(data, &db); err != nil {
			return corereg.Database{}, err
		}
		if db.Blueprints == nil {
			db.Blueprints = []corereg.Blueprint{}
		}
		return db, nil
	}
	return corereg.Load(loc)
}

func loadRegistry() (corereg.Database, error) {
	loc, err := resolveRegistry()
	if err != nil {
		return corereg.Database{}, err
	}
	return loadLocation(loc)
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
		db, err := loadLocation(u)
		if err != nil {
			return nil, fmt.Errorf("load %s: %w", u, err)
		}
		res = append(res, struct {
			URL string
			DB  corereg.Database
		}{URL: u, DB: db})
	}
	return res, nil
}

// semver
func parseSemver(v string) (int, int, int) {
	v = strings.TrimSpace(strings.TrimPrefix(v, "v"))
	p := strings.SplitN(v, ".", 3)
	to := func(s string) int { n, _ := strconv.Atoi(s); return n }
	a, b, c := 0, 0, 0
	if len(p) > 0 {
		a = to(p[0])
	}
	if len(p) > 1 {
		b = to(p[1])
	}
	if len(p) > 2 {
		c = to(p[2])
	}
	return a, b, c
}
func satisfies(v, constraint string) bool {
	if constraint == "" {
		return true
	}
	vM, vm, vp := parseSemver(v)
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
	case "^":
		return cmp(vM, vm, vp, tM, tm, tp) >= 0 && vM == tM
	case "~":
		return cmp(vM, vm, vp, tM, tm, tp) >= 0 && vM == tM && vm == tm
	default:
		return cmp(vM, vm, vp, tM, tm, tp) == 0
	}
}

func findBlueprint(name string) (corereg.Blueprint, string, error) {
	sets, err := loadAllRegistries()
	if err != nil {
		return corereg.Blueprint{}, "", err
	}
	for _, s := range sets {
		if bp, err := corereg.Find(s.DB, name); err == nil {
			return *bp, s.URL, nil
		}
	}
	return corereg.Blueprint{}, "", fmt.Errorf("blueprint %q not found", name)
}
