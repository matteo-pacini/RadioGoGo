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
	"net"
	"testing"
)

func TestDnsLookup(t *testing.T) {

	t.Run("lookup valid IP address returns IP address", func(t *testing.T) {

		lookupIPFunc = func(host string) ([]net.IP, error) {
			return []net.IP{net.ParseIP("127.0.0.1")}, nil
		}

		ips, err := dnsLookup("8.8.8.8")
		if err != nil {
			t.Errorf("Expected no error, but got %v", err)
		}
		if len(ips) != 1 {
			t.Errorf("Expected one IP address, but got %v", ips)
		}
		if ips[0] != "8.8.8.8" {
			t.Errorf("Expected IP address to be 8.8.8.8, but got %v", ips[0])
		}
	})

	t.Run("lookup valid hostname returns IP address if lookup succeeds", func(t *testing.T) {

		lookupIPFunc = func(host string) ([]net.IP, error) {
			return []net.IP{net.ParseIP("127.0.0.1")}, nil
		}

		ips, err := dnsLookup("all.api.radio-browser.info")
		if err != nil {
			t.Errorf("Expected no error, but got %v", err)
		}
		if len(ips) == 0 {
			t.Errorf("Expected at least one IP address, but got %v", ips)
		}
		if ips[0] != "127.0.0.1" {
			t.Errorf("Expected IP address to be 127.0.0.1")
		}
	})

	t.Run("lookup valid hostname returns error if lookup fails", func(t *testing.T) {

		lookupIPFunc = func(host string) ([]net.IP, error) {
			return []net.IP{}, &net.DNSError{}
		}

		ips, err := dnsLookup("all.api.radio-browser.info")
		if err == nil {
			t.Errorf("Expected error, but got none")
		}
		if len(ips) != 0 {
			t.Errorf("Expected no IP addresses, but got %v", ips)
		}
	})

}
