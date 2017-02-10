// Copyright (c) 2016 Betalo AB
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
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"

	"context"
)

type httpResource struct {
	url.URL
}

func (r *httpResource) Await(ctx context.Context) error {
	client := &http.Client{
		Transport: &http.Transport{
			DisableKeepAlives: true,
		},
	}

	// IDEA(uwe): Use fragment to set method

	req, err := http.NewRequest("GET", r.URL.String(), nil)
	if err != nil {
		return err
	}

	// IDEA(uwe): Use k/v pairs in fragment to set headers

	req = req.WithContext(ctx)
	req.Close = true

	req.Header.Set("User-Agent", "await/"+version)

	resp, err := client.Do(req)
	if err != nil {
		return &unavailabilityError{err}
	}
	defer resp.Body.Close()

	// IDEA(uwe): Use fragment to set tolerated status code

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return &unavailabilityError{errors.New(resp.Status)}
	}

	if _, err := io.Copy(ioutil.Discard, resp.Body); err != nil {
		return &unavailabilityError{err}
	}

	return nil
}
