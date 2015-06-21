package formation

import (
	"encoding/json"
	"testing"
)

func TestRef(t *testing.T) {
	var f interface{}
	b := []byte(`{ "Name": { "Ref": "AWS::StackName" } }`)
	err := json.Unmarshal(b, &f)

	if err != nil {
		t.Errorf("Error %q", err)
	}

	cases := Cases{
		{translate(f), map[string]string{"Name": "teststack"}},
	}

	_assert(t, cases)
}

func TestJoin(t *testing.T) {
	var f interface{}
	b := []byte(`{ "TableName": { "Fn::Join": [ "-", [ "myapp", "builds" ] ] } }`)
	err := json.Unmarshal(b, &f)

	if err != nil {
		t.Errorf("Error %q", err)
	}

	cases := Cases{
		{translate(f), map[string]string{"TableName": "myapp-builds"}},
	}

	_assert(t, cases)
}

func TestJoinRef(t *testing.T) {
	var f1, f2 interface{}

	_ = json.Unmarshal(
		[]byte(`{ "TableName": { "Fn::Join": [ "-", [ { "Ref": "AWS::StackName" }, "builds" ] ] } }`),
		&f1,
	)

	_ = json.Unmarshal(
		[]byte(`{ "Resource": [ { "Fn::Join": [ "", [ "arn:aws:kinesis:*:*:stream/", { "Ref": "AWS::StackName" }, "-*" ] ] } ] }`),
		&f2,
	)

	cases := Cases{
		{translate(f1), map[string]string{"TableName": "teststack-builds"}},
		{translate(f2), map[string][]string{"Resource": []string{"arn:aws:kinesis:*:*:stream/teststack-*"}}},
	}

	_assert(t, cases)
}
