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

You can install HTTP recorder with the following command:

```bash
go get -u gosrc.io/httpmock/httprec
```

To create a scenario file in your fixtures directory, you can then use the following command:

```bash
httprec add fixtures/scenario1 -u https://www.process-one.net/
```