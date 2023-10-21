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

// DNSLookupService defines the behavior for looking up IP addresses for a given host.
type DNSLookupService interface {
	// LookupIP returns the IP address of a given host. If the host is already an IP address, it returns the same IP address.
	// Otherwise, it performs a DNS lookup and returns the IP addresses associated with the host.
	// The function uses the default resolver and the "ip4" network to perform the lookup.
	// If the lookup fails, it returns an empty slice and the error encountered.
	LookupIP(host string) ([]string, error)
}

// DNSLookupServiceImpl provides a default implementation of the DNSLookupService interface.
type DNSLookupServiceImpl struct{}

func NewDNSLookupService() DNSLookupService {
	return &DNSLookupServiceImpl{}
}

// LookupIP performs a DNS lookup to retrieve IP addresses for the given host.
func (s *DNSLookupServiceImpl) LookupIP(host string) ([]string, error) {

	if net.ParseIP(host) != nil {
		return []string{host}, nil
	}

	resolver := net.DefaultResolver

	ips, err := resolver.LookupIP(context.Background(), "ip4", host)
	if err != nil {
		return []string{}, err
	}

	ipStrings := make([]string, len(ips))
	for i, ip := range ips {
		ipStrings[i] = ip.String()
	}

	return ipStrings, nil

}
