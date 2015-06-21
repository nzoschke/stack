package stack

import (
	"math/rand"
	"strconv"

	yaml "gopkg.in/yaml.v2"
)

type ManifestEntry struct {
	Command interface{} `yaml:"command"`
	Links   []string    `yaml:"links"`
	Ports   []string    `yaml:"ports"`
	Volumes []string    `yaml:"volumes"`

	Randoms []string
}

type Manifest map[string]ManifestEntry

func Import(y []byte) (Manifest, error) {
	var manifest Manifest

	err := yaml.Unmarshal(y, &manifest)

	if err != nil {
		return manifest, err
	}

	for i, e := range manifest {
		for _ = range e.Ports {
			e.Randoms = append(e.Randoms, randomPort())
		}
		manifest[i] = e
	}

	return manifest, nil
}

func (m Manifest) HasPorts() bool {
	for _, me := range m {
		if len(me.Ports) > 0 {
			return true
		}
	}

	return false
}

func (me ManifestEntry) HasPorts() bool {
	return len(me.Ports) > 0
}

func (m Manifest) HasProcesses() bool {
	return len(m) > 0
}

func randomPort() string {
	return strconv.Itoa(rand.Intn(50000) + 5000)
}
