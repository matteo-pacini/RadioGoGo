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

// Go client for the Radio Browser API.
//
// # Usage
//
// Create a new instance of RadioBrowser:
//
//	browser, err := radiogogo.NewRadioBrowser()
//	if err != nil {
//	    // Handle error
//	}
package api

import (
	"math/rand"
	"net"
	"net/http"
	"net/url"
)

type RadioBrowser struct {
	// The base URL for the Radio Browser API.)
	BaseUrl url.URL
}

// NewRadioBrowser creates a new instance of RadioBrowser and returns a pointer to it.
// It performs a DNS lookup to get a random IP address of the radio browser API and sets the base URL of the browser.
// Returns an error if the DNS lookup fails.
func NewRadioBrowser() (*RadioBrowser, error) {
	browser := &RadioBrowser{}
	ips, err := dnsLookup("all.api.radio-browser.info")
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
	browser.BaseUrl = *url
	return browser, nil
}

func init() {
	Client = &http.Client{}
}
