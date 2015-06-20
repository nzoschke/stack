package formation

import (
	"encoding/json"
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
