package policy

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"

	"gopkg.in/yaml.v3"
)

// LoadDir reads every *.yaml/*.yml file anywhere under dir (recursively)
// and parses each as a single Policy manifest. This lets policies for the
// same tool be grouped into a subdirectory (e.g. .guard/policies/git/*.yaml)
// instead of all living flat in one directory. Files are returned in
// sorted path order so results are deterministic.
func LoadDir(dir string) ([]Policy, error) {
	var paths []string
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		ext := filepath.Ext(d.Name())
		if ext != ".yaml" && ext != ".yml" {
			return nil
		}
		paths = append(paths, path)
		return nil
	})
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read policies dir %q: %w", dir, err)
	}
	sort.Strings(paths)

	policies := make([]Policy, 0, len(paths))
	for _, p := range paths {
		pol, err := LoadFile(p)
		if err != nil {
			return nil, err
		}
		policies = append(policies, pol)
	}
	return policies, nil
}

// LoadFile parses a single policy manifest file.
func LoadFile(path string) (Policy, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Policy{}, fmt.Errorf("read policy file %q: %w", path, err)
	}

	var pol Policy
	if err := yaml.Unmarshal(data, &pol); err != nil {
		return Policy{}, fmt.Errorf("parse policy file %q: %w", path, err)
	}
	pol.SourcePath = path
	return pol, nil
}
