package httpmock_test

import (
	"bytes"
	"io/ioutil"
	"testing"

	"gosrc.io/httpmock"
)

func TestHTTPMock(t *testing.T) {
	// Setup HTTP Mock
	mock := httpmock.NewMock("fixtures/")

	// Scenario generated with:
	// httprec add fixtures/ProcessOne -u https://www.process-one.net/
	fixtureName := "ProcessOne"
	if err := mock.LoadScenario(fixtureName); err != nil {
		t.Errorf("Cannot load fixture scenario %s: %s", fixtureName, err)
		return
	}

	HTTPClient := mock.Client
	resp, err := HTTPClient.Get("https://www.process-one.net/")
	if err != nil {
		t.Errorf("Cannot get page: %s", err)
		return
	}
	defer resp.Body.Close()

	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Errorf("Cannot read page body: %s", err)
		return
	}
	if bytes.Contains(content, []byte("ProcessOne")) == false {
		t.Errorf("'ProcessOne' not found on page")
	}
}
