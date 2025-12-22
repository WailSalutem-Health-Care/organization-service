package auth

import (
	"io/ioutil"

	"gopkg.in/yaml.v3"
)

// Permissions maps role -> []permission
type Permissions map[string][]string

type permissionsFile struct {
	Roles map[string][]string `yaml:"roles"`
}

// LoadPermissions loads a permissions.yml file and returns a role->permissions map.
func LoadPermissions(path string) (Permissions, error) {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var pf permissionsFile
	if err := yaml.Unmarshal(b, &pf); err != nil {
		return nil, err
	}
	return Permissions(pf.Roles), nil
}
