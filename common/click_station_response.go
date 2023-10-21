package common

import "github.com/google/uuid"

// ClickStationResponse represents the response returned by the API when a user clicks on a station.
type ClickStationResponse struct {
	// Ok indicates whether the request was successful or not.
	Ok bool `json:"ok"`

	// Message contains an optional message returned by the server.
	Message string `json:"message"`

	// StationUuid is the unique identifier of the station.
	StationUuid uuid.UUID `json:"stationuuid"`

	// Name is the name of the station.
	Name string `json:"name"`

	// Url is the URL of the station's stream.
	Url RadioGoGoURL `json:"url"`
}
