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
	"fmt"
	"net/url"
	"os"
)

type resource interface {
	fmt.Stringer
	Await(context.Context) error
}

type unavailabilityError struct {
	Reason error
}

// Error implements the error interface.
func (e *unavailabilityError) Error() string {
	return e.Reason.Error()
}

func parseResources(urlArgs []string) ([]resource, error) {
	var resources []resource
	for _, urlArg := range urlArgs {
		expandedUrl := os.ExpandEnv(urlArg)

		// Leveraging the fact the Go's URL parser matches e.g. `curl -s
		// http://example.com` as url.Path instead of throwing an error.
		u, err := url.Parse(expandedUrl)
		if err != nil {
			return nil, err
		}
		res, err := identifyResource(*u)
		if err != nil {
			return nil, err
		}
		resources = append(resources, res)
	}
	return resources, nil
}

func identifyResource(u url.URL) (resource, error) {
	switch u.Scheme {
	case "http", "https":
		return &httpResource{u}, nil
	case "ws", "wss":
		return &websocketResource{u}, nil
	case "tcp", "tcp4", "tcp6":
		return &tcpResource{u}, nil
	case "file":
		return &fileResource{u}, nil
	case "postgres":
		return &postgresqlResource{u}, nil
	case "mysql":
		return &mysqlResource{u}, nil
	case "":
		return &commandResource{u}, nil
	default:
		return nil, fmt.Errorf("unsupported resource scheme: %v", u.Scheme)
	}
}

func parseFragment(fragment string) url.Values {
	// Skip encountered decoding errors on invalid format for now
	v, _ := url.ParseQuery(fragment)
	// Maintain backwards-compatibility for the case that key is empty
	delete(v, "")

	return v
}
