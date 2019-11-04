package capybara

import "testing"

func TestIcanBuildURL(t *testing.T) {
	if defaultLocalhost+":9090" != buildURL(9090) {
		t.Fail()
	}
}
