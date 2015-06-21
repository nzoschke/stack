package stack

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"testing"
)

type Cases []struct {
	got, want interface{}
}

func TestVersion(t *testing.T) {
	tmpl := _template(t, nil)

	cases := Cases{
		{tmpl.AWSTemplateFormatVersion, "2010-09-09"},
	}

	_assert(t, cases)
}

func TestConditions(t *testing.T) {
	tmpl := _template(t, nil)
	conditions := tmpl.Conditions

	var keys sort.StringSlice = make([]string, 0, len(conditions))

	for k := range conditions {
		keys = append(keys, k)
	}
	keys.Sort()

	cases := Cases{
		{[]string(keys), []string{"BlankCluster"}},
	}

	_assert(t, cases)
}

func TestParameters(t *testing.T) {
	tmpl := _template(t, nil)
	p := tmpl.Parameters

	var keys sort.StringSlice = make([]string, 0, len(p))

	for k := range p {
		keys = append(keys, k)
	}
	keys.Sort()

	cases := Cases{
		{[]string(keys), []string{"Cluster", "Environment", "Kernel", "Key", "Release", "Repository", "Subnets", "VPC"}},
	}

	_assert(t, cases)
}

func TestResources(t *testing.T) {
	tmpl := _template(t, nil)
	resources := tmpl.Resources

	var keys sort.StringSlice = make([]string, 0, len(resources))

	for k := range resources {
		keys = append(keys, k)
	}
	keys.Sort()

	cases := Cases{
		{[]string(keys), []string{"DynamoBuilds", "DynamoChanges", "DynamoReleases", "ServiceRole", "Settings"}},

		{resources["DynamoBuilds"].Type, "AWS::DynamoDB::Table"},
		{resources["DynamoChanges"].Type, "AWS::DynamoDB::Table"},
		{resources["DynamoReleases"].Type, "AWS::DynamoDB::Table"},
		{resources["ServiceRole"].Type, "AWS::IAM::Role"},
		{resources["Settings"].Type, "AWS::S3::Bucket"},
	}

	_assert(t, cases)
}

func TestOutputs(t *testing.T) {
	tmpl := _template(t, nil)
	o := tmpl.Outputs

	var keys sort.StringSlice = make([]string, 0, len(o))

	for k := range o {
		keys = append(keys, k)
	}
	keys.Sort()

	cases := Cases{
		{[]string(keys), []string{"Settings"}},
	}

	_assert(t, cases)
}

func TestCelery(t *testing.T) {
}

func TestHttpd(t *testing.T) {
	manifest, _ := Import([]byte(`
web:
  image: httpd
  ports:
    - 80:80
`))

	tmpl := _template(t, manifest)
	j, _ := json.MarshalIndent(tmpl, "", "  ")
	fmt.Printf("%+v\n", string(j))
}

func TestProcfile(t *testing.T) {}

func TestDockerCompose(t *testing.T) {}

func _assert(t *testing.T, cases Cases) {
	for _, c := range cases {
		j1, err := json.Marshal(c.got)

		if err != nil {
			t.Errorf("Marshal %q, error %q", c.got, err)
		}

		j2, err := json.Marshal(c.want)

		if err != nil {
			t.Errorf("Marshal %q, error %q", c.want, err)
		}

		if !bytes.Equal(j1, j2) {
			t.Errorf("Got %q, want %q", c.got, c.want)
		}
	}
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
