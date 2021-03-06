package httpmock

import (
	"fmt"
	"log"
	"mime"
	"net"
	"net/http"
	"net/url"
	"path/filepath"
	"time"
)

type Recorder struct {
	Logger *log.Logger
}

func (rec Recorder) Record(uri string, scnName string) error {
	client := newHTTPClient()

	var seq Sequence
	scn, created, err := InitScenario(scnName + ".json")
	if err != nil {
		return fmt.Errorf("cannot open scenario file: %s", err)
	}
	if !created {
		rec.logf("Updating existing scenario: it contains %d sequences.\n", scn.Count())
	}

	// Record a new sequence
Loop:
	for redirect := 0; redirect <= 10; redirect++ {
		resp, err := client.Get(uri)

		// on error
		if err != nil {
			step := Step{
				RequestURL: uri,
				Response:   NewResponse(resp),
				Err:        err.Error(),
			}
			seq.Steps = append(seq.Steps, step)
			break Loop
		}

		// Add step to sequence and record body file
		r := NewResponse(resp)
		// On success
		switch {
		// Record and follow redirects steps:
		case resp.StatusCode >= 300 && resp.StatusCode < 400:
			// Redirect
			location := resp.Header.Get("Location")
			rec.logf("Recording redirect to: %s", location)
			step := Step{
				Method:     "GET",
				RequestURL: uri,
				Response:   r,
			}
			seq.Steps = append(seq.Steps, step)
			_ = resp.Body.Close()
			// Retry resolving the next link, with new discovered location
			uri, err = formatRedirectUrl(uri, location)
			if err != nil {
				// Not a valid URL, do not redirect further
				break Loop
			}
		default:
			rec.logf("Recording response: %d", resp.StatusCode)
			prefix := filepath.Base(scnName)
			r.BodyFilename = fmt.Sprintf("%s-%d-%d%s", prefix, scn.Count()+1, redirect+1, extension(r.Header))
			step := Step{
				Method:     "GET",
				RequestURL: uri,
				Response:   r,
			}
			if err := step.SaveBody(resp.Body, filepath.Join(filepath.Dir(scnName), r.BodyFilename)); err != nil {
				rec.logf("Cannot save body file %s: %s\n", r.BodyFilename, err)
			}
			seq.Steps = append(seq.Steps, step)
			_ = resp.Body.Close()
			break Loop
		}
	}
	// Add Sequence to scenario
	if err = scn.AddSequence(seq); err != nil {
		return fmt.Errorf("cannot add sequence to scenario: %s", err)
	}
	// Save file
	filename := scnName + ".json"
	if err := scn.SaveTo(filename); err != nil {
		return fmt.Errorf("cannot save sequence to file %s: %s", filename, err)
	}
	// Save basic .url files
	filename = scnName + ".url"
	if err := scn.SaveAsURLList(filename); err != nil {
		return fmt.Errorf("cannot save url list to file %s: %s", filename, err)
	}

	return nil
}

func (rec Recorder) logf(format string, a ...interface{}) {
	rec.Logger.Printf(format, a...)
}

//=============================================================================
// Helpers

// newHTTPClient returns an http.Client that does not automatically follow redirects.
func newHTTPClient() *http.Client {
	transport := &http.Transport{
		DialContext: (&net.Dialer{
			Timeout: 5 * time.Second,
		}).DialContext,
		TLSHandshakeTimeout: 5 * time.Second,
	}
	client := http.Client{
		Timeout:   time.Second * 15,
		Transport: transport,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	return &client
}

// formatRedirectUrl returns a valid full URL from an original URL and a "Location" Header.
// It support local redirection on same host.
func formatRedirectUrl(originalUrl, locationHeader string) (string, error) {
	newUrl, err := url.Parse(locationHeader)
	if err != nil {
		return "", err
	}

	// This is a relative URL, we need to use the host from original URL
	if newUrl.Host == "" && newUrl.Scheme == "" {
		oldUrl, err := url.Parse(originalUrl)
		if err != nil {
			return "", err
		}
		newUrl.Host = oldUrl.Host
		newUrl.Scheme = oldUrl.Scheme
	}
	return newUrl.String(), nil
}

// extension is used to get the extension to record HTTP response body.
func extension(header http.Header) string {
	content := header["Content-Type"]
	for _, typ := range content {
		if ext, err := mime.ExtensionsByType(typ); err == nil {
			// We return the longest known extension
			bestExtension := ""
			for _, e := range ext {
				if len(e) > len(bestExtension) {
					bestExtension = e
				}
			}
			if bestExtension != "" {
				return bestExtension
			}
		}
	}
	return ".data"
}
