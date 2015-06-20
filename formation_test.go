package formation

import (
	"encoding/json"
	"reflect"
	"sort"
	"testing"
)

func TestVersion(t *testing.T) {
	tmpl := _template(t, nil)

	want := "2010-09-09"

	if tmpl.AWSTemplateFormatVersion != want {
		t.Errorf("TestVersion got %s, want %s", tmpl.AWSTemplateFormatVersion, want)
	}

}

func TestResources(t *testing.T) {
	tmpl := _template(t, nil)
	resources := tmpl.Resources

	var keys sort.StringSlice = make([]string, 0, len(resources))

	for k := range resources {
		keys = append(keys, k)
	}
	keys.Sort()

	var want sort.StringSlice = []string{"DynamoBuilds", "DynamoChanges", "DynamoReleases", "ServiceRole", "Settings"}

	if !reflect.DeepEqual(keys, want) {
		t.Errorf("TestResources got %s, want %s", keys, want)
	}

	s := (resources["Settings"]).(map[string]interface{})
	got := s["Type"]
	w := "AWS::S3::Bucket"

	if got != w {
		t.Errorf("TestResources got %s, want %s", keys, w)
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

func _template(t *testing.T, data interface{}) *Template {
	j, err := buildTemplate(data)

	if err != nil {
		t.Errorf("_template err: %s", err)
	}

	var tmpl Template
	err = json.Unmarshal([]byte(j), &tmpl)

	if err != nil {
		t.Errorf("_template err: %s", err)
	}

	return &tmpl
}
