package api

import (
	"bytes"
	"io"
	"net/http"
	"radiogogo/data"
	"radiogogo/mocks"
	"radiogogo/models"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestBrowserImplEscapesIPV6(t *testing.T) {

	mockDNSLookupService := mocks.MockDNSLookupService{
		LookupIPFunc: func(host string) ([]string, error) {
			return []string{"2001:db8::1"}, nil
		},
	}

	mockHttpClient := mocks.MockHttpClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			return nil, io.EOF
		},
	}

	browser, err := NewRadioBrowserWithDependencies(&mockDNSLookupService, &mockHttpClient)
	assert.NoError(t, err)

	assert.Equal(t, "http://[2001:db8::1]/json", browser.baseUrl.String())

	assert.NoError(t, err)

}

func TestBrowserImplGetStations(t *testing.T) {

	// Note: Search term set to "searchTerm" in all test cases

	testCases := []struct {
		name             string
		queryType        StationQuery
		expectedEndpoint string
	}{
		{
			name:             "builds the correct URL for StationQueryAll",
			queryType:        StationQueryAll,
			expectedEndpoint: "/json/stations",
		},
		{
			name:             "builds the correct URL for StationQueryByUUID",
			queryType:        StationQueryByUuid,
			expectedEndpoint: "/json/stations/byuuid/searchTerm",
		},
		{
			name:             "builds the correct URL for StationQueryByName",
			queryType:        StationQueryByName,
			expectedEndpoint: "/json/stations/byname/searchTerm",
		},
		{
			name:             "builds the correct URL for StationQueryByNameExact",
			queryType:        StationQueryByNameExact,
			expectedEndpoint: "/json/stations/bynameexact/searchTerm",
		},
		{
			name:             "builds the correct URL for StationQueryByCodec",
			queryType:        StationQueryByCodec,
			expectedEndpoint: "/json/stations/bycodec/searchTerm",
		},
		{
			name:             "builds the correct URL for StationQueryByCodecExact",
			queryType:        StationQueryByCodecExact,
			expectedEndpoint: "/json/stations/bycodecexact/searchTerm",
		},
		{
			name:             "builds the correct URL for StationQueryByCountry",
			queryType:        StationQueryByCountry,
			expectedEndpoint: "/json/stations/bycountry/searchTerm",
		},
		{
			name:             "builds the correct URL for StationQueryByCountryExact",
			queryType:        StationQueryByCountryExact,
			expectedEndpoint: "/json/stations/bycountryexact/searchTerm",
		},
		{
			name:             "builds the correct URL for StationQueryByCountryCodeExact",
			queryType:        StationQueryByCountryCodeExact,
			expectedEndpoint: "/json/stations/bycountrycodeexact/searchTerm",
		},
		{
			name:             "builds the correct URL for StationQueryByState",
			queryType:        StationQueryByState,
			expectedEndpoint: "/json/stations/bystate/searchTerm",
		},
		{
			name:             "builds the correct URL for StationQueryByStateExact",
			queryType:        StationQueryByStateExact,
			expectedEndpoint: "/json/stations/bystateexact/searchTerm",
		},
		{
			name:             "builds the correct URL for StationQueryByLanguage",
			queryType:        StationQueryByLanguage,
			expectedEndpoint: "/json/stations/bylanguage/searchTerm",
		},
		{
			name:             "builds the correct URL for StationQueryByLanguageExact",
			queryType:        StationQueryByLanguageExact,
			expectedEndpoint: "/json/stations/bylanguageexact/searchTerm",
		},
		{
			name:             "builds the correct URL for StationQueryByTag",
			queryType:        StationQueryByTag,
			expectedEndpoint: "/json/stations/bytag/searchTerm",
		},
		{
			name:             "builds the correct URL for StationQueryByTagExact",
			queryType:        StationQueryByTagExact,
			expectedEndpoint: "/json/stations/bytagexact/searchTerm",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			mockDNSLookupService := mocks.MockDNSLookupService{
				LookupIPFunc: func(host string) ([]string, error) {
					return []string{"127.0.0.1"}, nil
				},
			}

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

			browser, err := NewRadioBrowserWithDependencies(&mockDNSLookupService, &mockHttpClient)

			assert.NoError(t, err)

			_, err = browser.GetStations(tc.queryType, "searchTerm", "name", false, 0, 10, true)

			assert.NoError(t, err)

		})
	}
}
func TestBrowserImplClickStation(t *testing.T) {

	station := models.Station{
		StationUuid: uuid.MustParse("941ef6f1-0699-4821-95b1-2b678e3ff62e"),
	}

	mockDNSLookupService := mocks.MockDNSLookupService{
		LookupIPFunc: func(host string) ([]string, error) {
			return []string{"127.0.0.1"}, nil
		},
	}

	mockHttpClient := mocks.MockHttpClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			expectedUrl := "http://127.0.0.1/json/url/941ef6f1-0699-4821-95b1-2b678e3ff62e"
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

	radioBrowser, err := NewRadioBrowserWithDependencies(&mockDNSLookupService, &mockHttpClient)
	assert.NoError(t, err)

	response, err := radioBrowser.ClickStation(station)
	assert.NoError(t, err)

	assert.Equal(t, true, response.Ok)
}
