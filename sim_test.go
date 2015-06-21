package formation

import (
	"encoding/json"
	"testing"
)

func TestRef(t *testing.T) {
	var f interface{}
	b := []byte(`{ "Ref": "AWS::StackName" }`)
	err := json.Unmarshal(b, &f)

	if err != nil {
		t.Errorf("Error %q", err)
	}

	cases := Cases{
		{translate(f), map[string]string{"Ref": "teststack"}},
	}

	_assert(t, cases)
}
