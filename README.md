# HTTPMock

[![Codeship Status for processone/httpmock](https://app.codeship.com/projects/cf2e6700-1b1a-0137-02c9-72d9af1082b6/status?branch=master)](https://app.codeship.com/projects/328623)

If you need to write tests for code involving lot of API and web page scrapping, you often end up saving pages as
fixtures and loading those fixtures to try injecting them in your code to simulate the query.

However, it is time consuming to manage your test scenario and difficult to mock the HTTP calls as they can be done
very deep in your code.

This HTTPMock library intend to make those HTTP requests heavy test easier by allowing to record HTTP scenarii and
replay them easily in your tests.

## Overview

HTTPMock is composed of:

- a Go HTTP Mock library for writing tests,
- an HTTP scenario recording tool.

This library is used to record scenario that can be replayed locally, exactly as recorded.
It is helpful to prepare tests for your Go code when there is underlying HTTP requests. You can thus be sure
the test will replay real traffic consistently. Those recorded tests are also useful to document the behaviour 
of your code against a specific version of API or content. When the target HTTP endpoint changes and breaks your
code, you can thus now easily generate a diff of the HTTP content to understand the change in behaviour and
adapt your code accordingly.

## Usage

### Recording scenario

Recording a scenario is done by adding URL endpoint your want to support.

When you add an endpoint to a scenario, the instrument HTTP client will be able to reply properly with the recorded
data.

The recorder stores:

- Replies to that request, including header and body;
- Redirect sequence if any. This is important to exercice some behaviour in your HTTP crawler code.

You can install HTTP recorder with the following command:

```bash
go get -u gosrc.io/httpmock/httprec
```

To create a scenario file in your fixtures directory, you can then use the following command:

```bash
httprec add fixtures/scenario1 -u https://www.process-one.net/
```

### Using the scenario in a test case

To use the scenario in a test case, you need to:

- Create an HTTPMock instance
- Load the scenario you would like to use (set of endpoints and replies)
- Pass the HTTPClient provided by HTTPMock to your code / library so that replies received by your code will be
  the one from the scenario.

For example:

```go
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
```

### Writing custom HTTP mocks

You can configure HTTP mocks to reply with custom data that does not come from a scenario.

Here is an example on how to use this feature. In this example, we test that we can follow redirect and return the
final of a sequence. This can be used to check that we can properly resolve short URLs:

```go
package httpmock_test

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"testing"

	"gosrc.io/httpmock"

	"github.com/processone/dpk/pkg/semweb"
)

// TODO: Rewrite, based on new mock package.
func TestFollowRedirect(t *testing.T) {
	targetSite := "https://process-one.net"
	responder := func(req *http.Request) (*http.Response, error) {
		if req.URL.Host == "t.co" {
			resp := RedirectResponse(targetSite)
			resp.Request = req
			return resp, nil
		}
		if req.URL.Host == "process-one.net" {
			resp := SimplePageResponse("Target Page Title")
			resp.Request = req
			return resp, nil
		}
		t.Errorf("unknown host: %s", req.Host)
		return nil, errors.New("unknown host")
	}

	mock := httpmock.NewMock("")
	mock.SetResponder(responder)
	c := semweb.NewClient()
	c.HTTPClient = mock.Client
	uri := c.FollowRedirect("https://t.co/shortURL")
	if uri != targetSite {
		t.Errorf("unexpected uri: %s", uri)
	}
}

// Simple basic HTTP responses

func RedirectResponse(location string) *http.Response {
	status := 301
	reader := bytes.NewReader([]byte{})
	header := http.Header{}
	header.Add("Location", location)
	response := http.Response{
		Status:     strconv.Itoa(status),
		StatusCode: status,
		Body:       ioutil.NopCloser(reader),
		Header:     header,
	}
	return &response
}

func SimplePageResponse(title string) *http.Response {
	status := 200
	template := `<html>
<head><title>%s</title></head>
<body><h2>%s</h2></body>
</html>`
	page := fmt.Sprintf(template, title, title)
	reader := strings.NewReader(page)
	response := http.Response{
		Status:     strconv.Itoa(status),
		StatusCode: status,
		Body:       ioutil.NopCloser(reader),
		Header:     http.Header{},
	}
	return &response
}
```
