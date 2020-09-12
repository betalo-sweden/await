// Copyright (C) 2016-2018 Betalo AB
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package main

import (
	"context"
	"crypto/tls"
	"errors"
	"net/http"
	"net/url"
)

type httpResource struct {
	url.URL
}

func (r *httpResource) Await(ctx context.Context) error {
	var client *http.Client

	if skipTLSVerification(r) {
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		client = &http.Client{Transport: tr}
	} else {
		client = &http.Client{}
	}

	// IDEA(uwe): Use fragment to set method

	req, err := http.NewRequest("GET", r.URL.String(), nil)
	if err != nil {
		return err
	}

	// IDEA(uwe): Use k/v pairs in fragment to set headers

	req = req.WithContext(ctx)

	req.Header.Set("User-Agent", "await/"+version)

	resp, err := client.Do(req)
	if err != nil {
		return &unavailabilityError{err}
	}
	defer func() { _ = resp.Body.Close() }()

	// IDEA(uwe): Use fragment to set tolerated status code

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}

	return &unavailabilityError{errors.New(resp.Status)}
}

func skipTLSVerification(r *httpResource) bool {
	opts := parseFragment(r.URL.Fragment)
	vals, ok := opts["tls"]
	return ok && r.URL.Scheme == "https" && len(vals) == 1 && vals[0] == "skip-verify"
}
