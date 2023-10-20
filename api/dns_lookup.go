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
	"context"
	"net"
)

var lookupIPFunc = func(host string) ([]net.IP, error) {
	resolver := net.DefaultResolver
	return resolver.LookupIP(context.Background(), "ip4", host)
}

// dnsLookup performs a DNS lookup for the given hostname and returns a slice of IP addresses as strings.
// If the hostname is already an IP address, it returns a slice containing only that IP address.
// If the lookup fails, it returns an empty slice and the error encountered.
func dnsLookup(hostname string) ([]string, error) {

	if net.ParseIP(hostname) != nil {
		return []string{hostname}, nil
	}

	ips, err := lookupIPFunc(hostname)
	if err != nil {
		return []string{}, err
	}

	ipStrings := make([]string, len(ips))
	for i, ip := range ips {
		ipStrings[i] = ip.String()
	}

	return ipStrings, nil

}
