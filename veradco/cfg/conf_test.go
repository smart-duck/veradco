package conf

import (
	"testing"
)

func Test_MatchRegex(t *testing.T) {

	regex := "(!~)^toto$"
	value := "toto"

	pbool, err := matchRegex(regex, value)

	if *pbool == true {
		t.Errorf("\"matchRegex('%s', '%s')\" failed, expected -> true, got -> %t", regex, value, *pbool)
	} else if err == nil {
		t.Logf("\"matchRegex('%s', '%s')\" succeeded, expected -> true, got -> %t", regex, value, *pbool)
	} else {
		t.Errorf("\"matchRegex('%s', '%s')\" failed with error %v", regex, value, err)
	}
}