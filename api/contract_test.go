//go:build contract

// Contract tests against the real RadioBrowser API.
// Excluded from normal test runs; execute with:
//
//	go test -tags contract -count=1 ./api/
//
// A failure here signals a network problem or a change in the API contract.
package api

import (
	"encoding/json"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zi0p4tch0/radiogogo/common"
	"github.com/zi0p4tch0/radiogogo/data"
	"uuid"
)

const contractTimeout = 15 * time.Second

func newContractBrowser(t *testing.T) RadioBrowserService {
	t.Helper()
	browser, err := NewRadioBrowserWithDependencies(&http.Client{Timeout: contractTimeout})
	require.NoError(t, err)
	return browser
}

func searchStations(t *testing.T, browser RadioBrowserService) []common.Station {
	t.Helper()
	stations, err := browser.GetStations(common.StationQueryByName, "bbc", "votes", true, 0, 5, true)
	require.NoError(t, err, "live GetStations request failed")
	require.NotEmpty(t, stations, "expected at least one station for name query 'bbc'")
	return stations
}

func TestContract_GetStations(t *testing.T) {
	browser := newContractBrowser(t)
	stations := searchStations(t, browser)

	first := stations[0]
	assert.NotEqual(t, uuid.Nil(), first.StationUuid, "StationUuid must not be zero")
	assert.NotEmpty(t, first.Name, "Name must not be empty")
	assert.NotEmpty(t, first.Url.URL.String(), "Url must not be empty")
}

func TestContract_GetStationsByUUIDs(t *testing.T) {
	browser := newContractBrowser(t)
	stations := searchStations(t, browser)
	want := stations[0].StationUuid

	byUuid, err := browser.GetStationsByUUIDs([]uuid.UUID{want})
	require.NoError(t, err, "live GetStationsByUUIDs request failed")
	require.NotEmpty(t, byUuid, "expected station for UUID %s", want)
	assert.Equal(t, want, byUuid[0].StationUuid)
}

func TestContract_ClickStation(t *testing.T) {
	browser := newContractBrowser(t)
	stations := searchStations(t, browser)
	station := stations[0]

	response, err := browser.ClickStation(station)
	require.NoError(t, err, "live ClickStation request failed")
	assert.True(t, response.Ok, "expected ok=true, message: %q", response.Message)
	assert.Equal(t, station.StationUuid, response.StationUuid)
	assert.NotEmpty(t, response.Name)
	assert.NotEmpty(t, response.Url.URL.String())
}

func TestContract_VoteStation(t *testing.T) {
	browser := newContractBrowser(t)
	stations := searchStations(t, browser)

	// Votes are rate-limited per IP (one per station every 10 minutes), so
	// ok=false is acceptable; the contract is HTTP 200 plus a decodable
	// response carrying a message.
	response, err := browser.VoteStation(stations[0])
	require.NoError(t, err, "live VoteStation request failed")
	assert.NotEmpty(t, response.Message, "expected a message regardless of vote outcome (ok=%v)", response.Ok)
}

// TestContract_StationSchemaDrift decodes the raw JSON as generic maps and
// asserts every key the Station struct tags depend on is present. Renamed or
// removed fields would otherwise be silently zeroed by struct decoding.
func TestContract_StationSchemaDrift(t *testing.T) {
	req, err := http.NewRequest("GET", "https://all.api.radio-browser.info/json/stations/byname/bbc?limit=5&hidebroken=true", nil)
	require.NoError(t, err)
	req.Header.Set("User-Agent", data.UserAgent)
	req.Header.Set("Accept", "application/json")

	client := &http.Client{Timeout: contractTimeout}
	resp, err := client.Do(req)
	require.NoError(t, err, "raw stations request failed")
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var raw []map[string]any
	require.NoError(t, json.Unmarshal(body, &raw), "stations response is not a JSON array of objects")
	require.NotEmpty(t, raw, "expected at least one station in raw response")

	// Every JSON tag from common.Station.
	requiredKeys := []string{
		"changeuuid",
		"stationuuid",
		"name",
		"url",
		"url_resolved",
		"favicon",
		"tags",
		"countrycode",
		"state",
		"language",
		"languagecodes",
		"votes",
		"lastchangetime_iso8601",
		"codec",
		"bitrate",
		"hls",
		"lastcheckok",
		"lastchecktime_iso8601",
		"lastcheckoktime_iso8601",
		"lastlocalchecktime_iso8601",
		"clicktimestamp_iso8601",
		"clickcount",
		"clicktrend",
		"ssl_error",
		"geo_lat",
		"geo_long",
		"has_extended_info",
	}

	station := raw[0]
	for _, key := range requiredKeys {
		_, present := station[key]
		assert.True(t, present, "key %q missing from raw station JSON (schema drift)", key)
	}
}
