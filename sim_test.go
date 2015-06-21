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
