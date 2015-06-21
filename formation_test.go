package formation

import (
	"bytes"
	"encoding/json"
	"reflect"
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

// Golang reflection: traversing arbitrary structures
// https://gist.github.com/hvoecking/10772475

func translate(obj interface{}) interface{} {
	// Wrap the original in a reflect.Value
	original := reflect.ValueOf(obj)

	copy := reflect.New(original.Type()).Elem()
	translateRecursive(copy, original)

	// Remove the reflection wrapper
	return copy.Interface()
}

func translateRecursive(copy, original reflect.Value) {
	pseudoParams := map[string]string{
		"AWS::AccountId":        "123456789012",
		"AWS::NotificationARNs": "arn1, arn2", // []string{"arn1, arn2"},
		"AWS::NoValue":          "",
		"AWS::Region":           "us-west-2",
		"AWS::StackId":          "arn:aws:cloudformation:us-west-2:123456789012:stack/teststack/51af3dc0-da77-11e4-872e-1234567db123",
		"AWS::StackName":        "teststack",
	}

	switch original.Kind() {
	// The first cases handle nested structures and translate them recursively

	// If it is a pointer we need to unwrap and call once again
	case reflect.Ptr:
		// To get the actual value of the original we have to call Elem()
		// At the same time this unwraps the pointer so we don't end up in
		// an infinite recursion
		originalValue := original.Elem()
		// Check if the pointer is nil
		if !originalValue.IsValid() {
			return
		}
		// Allocate a new object and set the pointer to it
		copy.Set(reflect.New(originalValue.Type()))
		// Unwrap the newly created pointer
		translateRecursive(copy.Elem(), originalValue)

	// If it is an interface (which is very similar to a pointer), do basically the
	// same as for the pointer. Though a pointer is not the same as an interface so
	// note that we have to call Elem() after creating a new object because otherwise
	// we would end up with an actual pointer
	case reflect.Interface:
		// Get rid of the wrapping interface
		originalValue := original.Elem()

		if originalValue.Type() == reflect.TypeOf(make(map[string]string)) && len(originalValue.MapKeys()) == 1 && originalValue.MapKeys()[0].String() == "Ref" {
			k := originalValue.MapKeys()[0]
			v := originalValue.MapIndex(k)

			copyValue := reflect.New(reflect.TypeOf("")).Elem()
			copyValue.SetString(pseudoParams[v.String()])
			translateRecursive(copyValue, copyValue)
			copy.Set(copyValue)
		} else {
			// Create a new object. Now new gives us a pointer, but we want the value it
			// points to, so we have to call Elem() to unwrap it
			copyValue := reflect.New(originalValue.Type()).Elem()
			translateRecursive(copyValue, originalValue)
			copy.Set(copyValue)
		}

	// If it is a struct we translate each field
	case reflect.Struct:
		for i := 0; i < original.NumField(); i += 1 {
			translateRecursive(copy.Field(i), original.Field(i))
		}

	// If it is a slice we create a new slice and translate each element
	case reflect.Slice:
		copy.Set(reflect.MakeSlice(original.Type(), original.Len(), original.Cap()))
		for i := 0; i < original.Len(); i += 1 {
			translateRecursive(copy.Index(i), original.Index(i))
		}

	// If it is a map we create a new map and translate each value
	case reflect.Map:
		copy.Set(reflect.MakeMap(original.Type()))
		for _, key := range original.MapKeys() {
			originalValue := original.MapIndex(key)
			// New gives us a pointer, but again we want the value
			copyValue := reflect.New(originalValue.Type()).Elem()
			translateRecursive(copyValue, originalValue)
			copy.SetMapIndex(key, copyValue)
		}

	// Otherwise we cannot traverse anywhere so this finishes the the recursion

	// If it is a string translate it (yay finally we're doing what we came for)
	case reflect.String:
		// translatedString := dict[original.Interface().(string)]
		translatedString := original.Interface().(string)
		copy.SetString(translatedString)

	// And everything else will simply be taken from the original
	default:
		copy.Set(original)
	}

}

func TestJSONEvaluation(t *testing.T) {
	tmpl := _template(t, nil)
	p := tmpl.Resources["DynamoBuilds"].Properties

	join := map[string][]interface{}{
		"Fn::Join": []interface{}{
			"-",
			[]interface{}{
				map[string]string{"Ref": "AWS::StackName"},
				"builds",
			},
		},
	}

	cases := Cases{
		{p["TableName"], join},
		{translate(join), "teststack-builds"},
	}

	_assert(t, cases)
}

func TestCelery(t *testing.T) {}

func TestHttpd(t *testing.T) {}

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
