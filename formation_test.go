package formation

import (
	"encoding/json"
	"reflect"
	"sort"
	"testing"
)

func TestEmptyVersion(t *testing.T) {
  tmpl := _template(t, nil)

	if tmpl.AWSTemplateFormatVersion != "2010-09-09" {
		t.Errorf("AWSTemplateFormatVersion got %s, want %s", tmpl.AWSTemplateFormatVersion, "2010-09-09")
	}

}

func TestEmptyResources(t *testing.T) {
  tmpl := _template(t, nil)

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

func _template(t *testing.T, data interface{}) (*Template) {
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