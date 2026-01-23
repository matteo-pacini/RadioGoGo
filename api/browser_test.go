// Copyright (c) 2023-2026 Matteo Pacini
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
	"bytes"
	"io"
	"net/http"
	"testing"

	"github.com/zi0p4tch0/radiogogo/common"
	"github.com/zi0p4tch0/radiogogo/data"
	"github.com/zi0p4tch0/radiogogo/mocks"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestBrowserImplGetStations(t *testing.T) {

	// Note: Search term set to "searchTerm" in all test cases

	testCases := []struct {
		name             string
		queryType        common.StationQuery
		expectedEndpoint string
	}{
		{
			name:             "builds the correct URL for StationQueryAll",
			queryType:        common.StationQueryAll,
			expectedEndpoint: "/json/stations",
		},
		{
			name:             "builds the correct URL for StationQueryByUUID",
			queryType:        common.StationQueryByUuid,
			expectedEndpoint: "/json/stations/byuuid/searchTerm",
		},
		{
			name:             "builds the correct URL for StationQueryByName",
			queryType:        common.StationQueryByName,
			expectedEndpoint: "/json/stations/byname/searchTerm",
		},
		{
			name:             "builds the correct URL for StationQueryByNameExact",
			queryType:        common.StationQueryByNameExact,
			expectedEndpoint: "/json/stations/bynameexact/searchTerm",
		},
		{
			name:             "builds the correct URL for StationQueryByCodec",
			queryType:        common.StationQueryByCodec,
			expectedEndpoint: "/json/stations/bycodec/searchTerm",
		},
		{
			name:             "builds the correct URL for StationQueryByCodecExact",
			queryType:        common.StationQueryByCodecExact,
			expectedEndpoint: "/json/stations/bycodecexact/searchTerm",
		},
		{
			name:             "builds the correct URL for StationQueryByCountry",
			queryType:        common.StationQueryByCountry,
			expectedEndpoint: "/json/stations/bycountry/searchTerm",
		},
		{
			name:             "builds the correct URL for StationQueryByCountryExact",
			queryType:        common.StationQueryByCountryExact,
			expectedEndpoint: "/json/stations/bycountryexact/searchTerm",
		},
		{
			name:             "builds the correct URL for StationQueryByCountryCodeExact",
			queryType:        common.StationQueryByCountryCodeExact,
			expectedEndpoint: "/json/stations/bycountrycodeexact/searchTerm",
		},
		{
			name:             "builds the correct URL for StationQueryByState",
			queryType:        common.StationQueryByState,
			expectedEndpoint: "/json/stations/bystate/searchTerm",
		},
		{
			name:             "builds the correct URL for StationQueryByStateExact",
			queryType:        common.StationQueryByStateExact,
			expectedEndpoint: "/json/stations/bystateexact/searchTerm",
		},
		{
			name:             "builds the correct URL for StationQueryByLanguage",
			queryType:        common.StationQueryByLanguage,
			expectedEndpoint: "/json/stations/bylanguage/searchTerm",
		},
		{
			name:             "builds the correct URL for StationQueryByLanguageExact",
			queryType:        common.StationQueryByLanguageExact,
			expectedEndpoint: "/json/stations/bylanguageexact/searchTerm",
		},
		{
			name:             "builds the correct URL for StationQueryByTag",
			queryType:        common.StationQueryByTag,
			expectedEndpoint: "/json/stations/bytag/searchTerm",
		},
		{
			name:             "builds the correct URL for StationQueryByTagExact",
			queryType:        common.StationQueryByTagExact,
			expectedEndpoint: "/json/stations/bytagexact/searchTerm",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			mockHttpClient := mocks.MockHttpClient{
				DoFunc: func(req *http.Request) (*http.Response, error) {
					assert.Equal(t, tc.expectedEndpoint, req.URL.Path)
					assert.Equal(t, "GET", req.Method)
					assert.Equal(t, "application/json", req.Header.Get("Accept"))
					assert.Equal(t, data.UserAgent, req.Header.Get("User-Agent"))
					responseBody := io.NopCloser(bytes.NewReader([]byte(`[]`)))
					return &http.Response{
						StatusCode: 200,
						Body:       responseBody,
					}, nil
				},
			}

			browser, err := NewRadioBrowserWithDependencies(&mockHttpClient)

			assert.NoError(t, err)

			_, err = browser.GetStations(tc.queryType, "searchTerm", "name", false, 0, 10, true)

			assert.NoError(t, err)

		})
	}
}
func TestBrowserImplClickStation(t *testing.T) {

	station := common.Station{
		StationUuid: uuid.MustParse("941ef6f1-0699-4821-95b1-2b678e3ff62e"),
	}

	mockHttpClient := mocks.MockHttpClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			expectedUrl := "https://all.api.radio-browser.info/json/url/941ef6f1-0699-4821-95b1-2b678e3ff62e"
			assert.Equal(t, "POST", req.Method)
			assert.Equal(t, expectedUrl, req.URL.String())
			assert.Equal(t, "application/json", req.Header.Get("Accept"))
			assert.Equal(t, data.UserAgent, req.Header.Get("User-Agent"))

			responseBody := io.NopCloser(bytes.NewReader([]byte(`
			{
				"ok": true,
				"message": "retrieved station url",
				"stationuuid": "9617a958-0601-11e8-ae97-52543be04c81",
				"name": "Station name",
				"url": "http://this.is.an.url"
			}
			`)))
			return &http.Response{
				StatusCode: 200,
				Body:       responseBody,
			}, nil
		},
	}

	radioBrowser, err := NewRadioBrowserWithDependencies(&mockHttpClient)
	assert.NoError(t, err)

	response, err := radioBrowser.ClickStation(station)
	assert.NoError(t, err)

	assert.Equal(t, true, response.Ok)
}

func TestBrowserImplGetStationsByUUIDs(t *testing.T) {
	t.Run("returns empty slice for empty input", func(t *testing.T) {
		mockHttpClient := mocks.MockHttpClient{
			DoFunc: func(req *http.Request) (*http.Response, error) {
				t.Error("HTTP client should not be called for empty UUIDs")
				return nil, nil
			},
		}

		browser, err := NewRadioBrowserWithDependencies(&mockHttpClient)
		assert.NoError(t, err)

		stations, err := browser.GetStationsByUUIDs([]uuid.UUID{})
		assert.NoError(t, err)
		assert.Empty(t, stations)
	})

	t.Run("builds correct URL with multiple UUIDs", func(t *testing.T) {
		uuid1 := uuid.MustParse("941ef6f1-0699-4821-95b1-2b678e3ff62e")
		uuid2 := uuid.MustParse("16a73a57-5dba-11e8-b0ce-52543be04c81")

		mockHttpClient := mocks.MockHttpClient{
			DoFunc: func(req *http.Request) (*http.Response, error) {
				assert.Equal(t, "/json/stations/byuuid", req.URL.Path)
				assert.Equal(t, "GET", req.Method)
				assert.Contains(t, req.URL.RawQuery, "uuids=")
				assert.Contains(t, req.URL.RawQuery, uuid1.String())
				assert.Contains(t, req.URL.RawQuery, uuid2.String())
				assert.Equal(t, "application/json", req.Header.Get("Accept"))
				assert.Equal(t, data.UserAgent, req.Header.Get("User-Agent"))

				responseBody := io.NopCloser(bytes.NewReader([]byte(`[
					{"stationuuid": "941ef6f1-0699-4821-95b1-2b678e3ff62e", "name": "Station 1"},
					{"stationuuid": "16a73a57-5dba-11e8-b0ce-52543be04c81", "name": "Station 2"}
				]`)))
				return &http.Response{
					StatusCode: 200,
					Body:       responseBody,
				}, nil
			},
		}

		browser, err := NewRadioBrowserWithDependencies(&mockHttpClient)
		assert.NoError(t, err)

		stations, err := browser.GetStationsByUUIDs([]uuid.UUID{uuid1, uuid2})
		assert.NoError(t, err)
		assert.Len(t, stations, 2)
		assert.Equal(t, "Station 1", stations[0].Name)
		assert.Equal(t, "Station 2", stations[1].Name)
	})

	t.Run("handles API error", func(t *testing.T) {
		mockHttpClient := mocks.MockHttpClient{
			DoFunc: func(req *http.Request) (*http.Response, error) {
				responseBody := io.NopCloser(bytes.NewReader([]byte(`error`)))
				return &http.Response{
					StatusCode: 500,
					Body:       responseBody,
				}, nil
			},
		}

		browser, err := NewRadioBrowserWithDependencies(&mockHttpClient)
		assert.NoError(t, err)

		_, err = browser.GetStationsByUUIDs([]uuid.UUID{uuid.New()})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "500")
	})
}

func TestBrowserImpl_ErrorHandling(t *testing.T) {
	t.Run("GetStations handles network error", func(t *testing.T) {
		mockHttpClient := mocks.MockHttpClient{
			DoFunc: func(req *http.Request) (*http.Response, error) {
				return nil, &networkError{message: "connection refused"}
			},
		}

		browser, err := NewRadioBrowserWithDependencies(&mockHttpClient)
		assert.NoError(t, err)

		_, err = browser.GetStations(common.StationQueryByName, "test", "name", false, 0, 10, true)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "connection refused")
	})

	t.Run("GetStations handles HTTP 400 Bad Request", func(t *testing.T) {
		mockHttpClient := mocks.MockHttpClient{
			DoFunc: func(req *http.Request) (*http.Response, error) {
				responseBody := io.NopCloser(bytes.NewReader([]byte(`{"error": "bad request"}`)))
				return &http.Response{
					StatusCode: 400,
					Body:       responseBody,
				}, nil
			},
		}

		browser, err := NewRadioBrowserWithDependencies(&mockHttpClient)
		assert.NoError(t, err)

		_, err = browser.GetStations(common.StationQueryByName, "test", "name", false, 0, 10, true)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "400")
	})

	t.Run("GetStations handles HTTP 500 Internal Server Error", func(t *testing.T) {
		mockHttpClient := mocks.MockHttpClient{
			DoFunc: func(req *http.Request) (*http.Response, error) {
				responseBody := io.NopCloser(bytes.NewReader([]byte(`Internal Server Error`)))
				return &http.Response{
					StatusCode: 500,
					Body:       responseBody,
				}, nil
			},
		}

		browser, err := NewRadioBrowserWithDependencies(&mockHttpClient)
		assert.NoError(t, err)

		_, err = browser.GetStations(common.StationQueryByName, "test", "name", false, 0, 10, true)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "500")
	})

	t.Run("GetStations handles HTTP 503 Service Unavailable", func(t *testing.T) {
		mockHttpClient := mocks.MockHttpClient{
			DoFunc: func(req *http.Request) (*http.Response, error) {
				responseBody := io.NopCloser(bytes.NewReader([]byte(`Service Unavailable`)))
				return &http.Response{
					StatusCode: 503,
					Body:       responseBody,
				}, nil
			},
		}

		browser, err := NewRadioBrowserWithDependencies(&mockHttpClient)
		assert.NoError(t, err)

		_, err = browser.GetStations(common.StationQueryByName, "test", "name", false, 0, 10, true)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "503")
	})

	t.Run("GetStations handles malformed JSON response", func(t *testing.T) {
		mockHttpClient := mocks.MockHttpClient{
			DoFunc: func(req *http.Request) (*http.Response, error) {
				responseBody := io.NopCloser(bytes.NewReader([]byte(`{invalid json`)))
				return &http.Response{
					StatusCode: 200,
					Body:       responseBody,
				}, nil
			},
		}

		browser, err := NewRadioBrowserWithDependencies(&mockHttpClient)
		assert.NoError(t, err)

		_, err = browser.GetStations(common.StationQueryByName, "test", "name", false, 0, 10, true)
		assert.Error(t, err)
	})

	t.Run("GetStations handles empty response", func(t *testing.T) {
		mockHttpClient := mocks.MockHttpClient{
			DoFunc: func(req *http.Request) (*http.Response, error) {
				responseBody := io.NopCloser(bytes.NewReader([]byte(`[]`)))
				return &http.Response{
					StatusCode: 200,
					Body:       responseBody,
				}, nil
			},
		}

		browser, err := NewRadioBrowserWithDependencies(&mockHttpClient)
		assert.NoError(t, err)

		stations, err := browser.GetStations(common.StationQueryByName, "test", "name", false, 0, 10, true)
		assert.NoError(t, err)
		assert.Empty(t, stations)
	})

	t.Run("ClickStation handles network error", func(t *testing.T) {
		mockHttpClient := mocks.MockHttpClient{
			DoFunc: func(req *http.Request) (*http.Response, error) {
				return nil, &networkError{message: "timeout"}
			},
		}

		browser, err := NewRadioBrowserWithDependencies(&mockHttpClient)
		assert.NoError(t, err)

		station := common.Station{
			StationUuid: uuid.MustParse("941ef6f1-0699-4821-95b1-2b678e3ff62e"),
		}
		_, err = browser.ClickStation(station)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "timeout")
	})

	t.Run("ClickStation handles HTTP error", func(t *testing.T) {
		mockHttpClient := mocks.MockHttpClient{
			DoFunc: func(req *http.Request) (*http.Response, error) {
				responseBody := io.NopCloser(bytes.NewReader([]byte(`Not Found`)))
				return &http.Response{
					StatusCode: 404,
					Body:       responseBody,
				}, nil
			},
		}

		browser, err := NewRadioBrowserWithDependencies(&mockHttpClient)
		assert.NoError(t, err)

		station := common.Station{
			StationUuid: uuid.MustParse("941ef6f1-0699-4821-95b1-2b678e3ff62e"),
		}
		_, err = browser.ClickStation(station)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "404")
	})

	t.Run("ClickStation handles malformed JSON response", func(t *testing.T) {
		mockHttpClient := mocks.MockHttpClient{
			DoFunc: func(req *http.Request) (*http.Response, error) {
				responseBody := io.NopCloser(bytes.NewReader([]byte(`not json`)))
				return &http.Response{
					StatusCode: 200,
					Body:       responseBody,
				}, nil
			},
		}

		browser, err := NewRadioBrowserWithDependencies(&mockHttpClient)
		assert.NoError(t, err)

		station := common.Station{
			StationUuid: uuid.MustParse("941ef6f1-0699-4821-95b1-2b678e3ff62e"),
		}
		_, err = browser.ClickStation(station)
		assert.Error(t, err)
	})

	t.Run("GetStationsByUUIDs handles network error", func(t *testing.T) {
		mockHttpClient := mocks.MockHttpClient{
			DoFunc: func(req *http.Request) (*http.Response, error) {
				return nil, &networkError{message: "dns lookup failed"}
			},
		}

		browser, err := NewRadioBrowserWithDependencies(&mockHttpClient)
		assert.NoError(t, err)

		_, err = browser.GetStationsByUUIDs([]uuid.UUID{uuid.New()})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "dns lookup failed")
	})

	t.Run("GetStationsByUUIDs handles malformed JSON response", func(t *testing.T) {
		mockHttpClient := mocks.MockHttpClient{
			DoFunc: func(req *http.Request) (*http.Response, error) {
				responseBody := io.NopCloser(bytes.NewReader([]byte(`[{broken}`)))
				return &http.Response{
					StatusCode: 200,
					Body:       responseBody,
				}, nil
			},
		}

		browser, err := NewRadioBrowserWithDependencies(&mockHttpClient)
		assert.NoError(t, err)

		_, err = browser.GetStationsByUUIDs([]uuid.UUID{uuid.New()})
		assert.Error(t, err)
	})
}

// networkError is a test helper for simulating network errors
type networkError struct {
	message string
}

func (e *networkError) Error() string {
	return e.message
}
