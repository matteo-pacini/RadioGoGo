// Copyright (c) 2023 Matteo Pacini
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

package api

import (
	"math/rand"
	"net"
	"net/http"
	"net/url"
)

type RadioBrowser struct {
	// The HTTP client used to make requests to the Radio Browser API.
	httpClient HTTPClient
	// The base URL for the Radio Browser API.)
	baseUrl url.URL
}

// DefaultRadioBrowser returns a new instance of RadioBrowser with the default DNS lookup service and HTTP client.
func NewDefaultRadioBrowser() (*RadioBrowser, error) {
	return NewRadioBrowser(
		&DefaultDNSLookupService{},
		http.DefaultClient,
	)
}

// NewRadioBrowser creates a new instance of RadioBrowser struct with the provided DNSLookupService and HTTPClient.
// It returns a pointer to the created instance and an error if any.
// The function performs a DNS lookup for "all.api.radio-browser.info" and selects a random IP address from the returned list.
// It then constructs a base URL using the selected IP address and sets it as the baseUrl of the created instance.
func NewRadioBrowser(
	dnsLookupService DNSLookupService,
	httpClient HTTPClient,
) (*RadioBrowser, error) {
	browser := &RadioBrowser{
		httpClient: httpClient,
	}
	ips, err := dnsLookupService.LookupIP("all.api.radio-browser.info")
	if err != nil {
		return nil, err
	}

	randomIp := ips[rand.Intn(len(ips))]

	if net.ParseIP(randomIp).To4() == nil {
		randomIp = "[" + randomIp + "]"
	}

	url, err := url.Parse("http://" + randomIp + "/json")
	if err != nil {
		return nil, err
	}
	browser.baseUrl = *url
	return browser, nil
}
