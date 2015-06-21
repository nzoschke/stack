package formation

import (
	"encoding/json"
	"testing"
)

func TestEqual(t *testing.T) {
	var f1, f2 interface{}

	p := map[string]string{
		"Cluster": "convox",
	}

	_ = json.Unmarshal(
		[]byte(`{ "BlankCluster": { "Fn::Equals": [ { "Ref": "Cluster" }, "" ] } }`),
		&f1,
	)

	_ = json.Unmarshal(
		[]byte(`{ "BlankPostgresService": { "Fn::Equals": [ { "Ref": "PostgresService" }, "" ] } }`),
		&f2,
	)

	cases := Cases{
		{translate(f1, p), map[string]bool{"BlankCluster": false}},
		{translate(f2, p), map[string]bool{"BlankPostgresService": true}},
	}

	_assert(t, cases)
}

func TestRef(t *testing.T) {
	var f1, f2 interface{}

	p := map[string]string{
		"Cluster": "convox",
	}

	_ = json.Unmarshal(
		[]byte(`{ "Name": { "Ref": "AWS::StackName" } }`),
		&f1,
	)

	_ = json.Unmarshal(
		[]byte(`{ "Cluster": { "Ref": "Cluster" } }`),
		&f2,
	)

	cases := Cases{
		{translate(f1, p), map[string]string{"Name": "teststack"}},
		{translate(f2, p), map[string]string{"Cluster": "convox"}},
	}

	_assert(t, cases)
}

func TestJoin(t *testing.T) {
	var f interface{}
	var p map[string]string

	b := []byte(`{ "TableName": { "Fn::Join": [ "-", [ "myapp", "builds" ] ] } }`)
	err := json.Unmarshal(b, &f)

	if err != nil {
		t.Errorf("Error %q", err)
	}

	cases := Cases{
		{translate(f, p), map[string]string{"TableName": "myapp-builds"}},
	}

	_assert(t, cases)
}

func TestJoinRef(t *testing.T) {
	var f1, f2 interface{}
	var p map[string]string

	_ = json.Unmarshal(
		[]byte(`{ "TableName": { "Fn::Join": [ "-", [ { "Ref": "AWS::StackName" }, "builds" ] ] } }`),
		&f1,
	)

	_ = json.Unmarshal(
		[]byte(`{ "Resource": [ { "Fn::Join": [ "", [ "arn:aws:kinesis:*:*:stream/", { "Ref": "AWS::StackName" }, "-*" ] ] } ] }`),
		&f2,
	)

	cases := Cases{
		{translate(f1, p), map[string]string{"TableName": "teststack-builds"}},
		{translate(f2, p), map[string][]string{"Resource": []string{"arn:aws:kinesis:*:*:stream/teststack-*"}}},
	}

	_assert(t, cases)
}

func TestAll(t *testing.T) {
	params := map[string]string{
		"Cluster": "convox",
	}

	tmpl, ok := translate(_template(t, nil), params).(*Template)

	if !ok {
		t.Errorf("Error %q", ok)
	}

	cases := Cases{
		{tmpl.Resources["DynamoBuilds"].Properties["TableName"], "teststack-builds"},
		{tmpl.Resources["DynamoChanges"].Properties["TableName"], "teststack-changes"},
		{tmpl.Resources["DynamoReleases"].Properties["TableName"], "teststack-releases"},
		{tmpl.Resources["Settings"].Properties["Tags"], []map[string]string{{"Key": "system", "Value": "convox"}, {"Key": "app", "Value": "teststack"}}},
	}

	_assert(t, cases)
}
