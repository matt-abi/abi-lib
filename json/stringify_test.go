package json

import "testing"

func TestStringify(t *testing.T) {

	s, err := Stringify(map[string]interface{}{
		"appid":   "7nq179154935",
		"ver":     "1.0.0-37",
		"ability": "cloud",
		"env":     "unit",
	})

	if err != nil {
		t.Fatal(err)
	} else {
		t.Log(s)
	}
}
