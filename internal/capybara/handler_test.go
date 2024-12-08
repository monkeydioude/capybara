package capybara

import "testing"

var dh *DummyHTTPHandler
var dr *DummyResponseWriter

func TestIcanBuildURL(t *testing.T) {
	if defaultLocalhost+":9090" != buildURL(9090) {
		t.Fail()
	}
}

func init() {
	dh = &DummyHTTPHandler{}
	dr = &DummyResponseWriter{}
}
