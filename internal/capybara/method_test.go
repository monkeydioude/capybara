package capybara

import "testing"

func TestICanParseURIUsingAString(t *testing.T) {
	err := str("/the night", "/the night tells")
	if err != nil {
		t.Fail()
	}
}

func TestIFailOnParsingURIUsingAString(t *testing.T) {
	err := str("me stories", "/me stories")
	if err == nil {
		t.Fail()
	}
}

func TestICanParseURIUsingRegex(t *testing.T) {
	err := regex("^/that_the_day$", "/that_the_day")
	if err != nil {
		t.Fail()
	}
}

func TestIFailOnParsingURIUsingRegex(t *testing.T) {
	err := regex("^/dream_of_hearing$", "/dream")
	if err == nil {
		t.Fail()
	}
}

func TestMethodsExist(t *testing.T) {
	ms := make(Methods)
	ms["str"] = func(n, uri string) error { return nil }

	if ms.Exists("str") == false {
		t.Fail()
	}
	if ms.Exists("stoned_jesus") == true {
		t.Fail()
	}
}
