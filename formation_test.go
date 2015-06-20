package formation

import (
	"encoding/json"
	"reflect"
	"sort"
	"testing"
)

func TestEmptyManifest(t *testing.T) {
	var manifest Manifest

	j, err := buildTemplate(manifest)

	if err != nil {
		t.Errorf("TestEmptyManifest buildTemplate err: %s", err)
	}

	var tmpl Template
	err = json.Unmarshal([]byte(j), &tmpl)

	if err != nil {
		t.Errorf("TestEmptyManifest json.Unmarshal err: %s", err)
	}

	if tmpl.AWSTemplateFormatVersion != "2010-09-09" {
		t.Errorf("AWSTemplateFormatVersion got %s, want %s", tmpl.AWSTemplateFormatVersion, "2010-09-09")
	}

	var keys sort.StringSlice = make([]string, 0, len(tmpl.Resources))

	for k := range tmpl.Resources {
		keys = append(keys, k)
	}
	keys.Sort()

	var want sort.StringSlice = []string{"DynamoBuilds", "DynamoChanges", "DynamoReleases", "ServiceRole", "Settings"}

	if !reflect.DeepEqual(keys, want) {
		t.Errorf("Resources got %s, want %s", keys, want)
	}
}

func TestCelery(t *testing.T) {
}

func TestHttpd(t *testing.T) {

}

func TestProcfile(t *testing.T) {

}

func TestDockerCompose(t *testing.T) {

}

func TestJSONEvaluation(t *testing.T) {

}
