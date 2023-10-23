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

package common

// StationQuery represents the type of query that can be performed on a radio station.
type StationQuery string

// The following constants represent the different types of queries that can be performed on a radio station.
const (
	StationQueryAll                StationQuery = ""                   // Returns all radio stations.
	StationQueryByUuid             StationQuery = "byuuid"             // Returns radio stations by UUID.
	StationQueryByName             StationQuery = "byname"             // Returns radio stations by name.
	StationQueryByNameExact        StationQuery = "bynameexact"        // Returns radio stations by exact name.
	StationQueryByCodec            StationQuery = "bycodec"            // Returns radio stations by codec.
	StationQueryByCodecExact       StationQuery = "bycodecexact"       // Returns radio stations by exact codec.
	StationQueryByCountry          StationQuery = "bycountry"          // Returns radio stations by country.
	StationQueryByCountryExact     StationQuery = "bycountryexact"     // Returns radio stations by exact country.
	StationQueryByCountryCodeExact StationQuery = "bycountrycodeexact" // Returns radio stations by exact country code.
	StationQueryByState            StationQuery = "bystate"            // Returns radio stations by state.
	StationQueryByStateExact       StationQuery = "bystateexact"       // Returns radio stations by exact state.
	StationQueryByLanguage         StationQuery = "bylanguage"         // Returns radio stations by language.
	StationQueryByLanguageExact    StationQuery = "bylanguageexact"    // Returns radio stations by exact language.
	StationQueryByTag              StationQuery = "bytag"              // Returns radio stations by tag.
	StationQueryByTagExact         StationQuery = "bytagexact"         // Returns radio stations by exact tag.
)

func (m StationQuery) Render() string {
	switch m {
	case StationQueryByUuid:
		return "By UUID"
	case StationQueryByName:
		return "By Name"
	case StationQueryByNameExact:
		return "By Exact Name"
	case StationQueryByCodec:
		return "By Codec"
	case StationQueryByCodecExact:
		return "By Exact Codec"
	case StationQueryByCountry:
		return "By Country"
	case StationQueryByCountryExact:
		return "By Exact Country"
	case StationQueryByCountryCodeExact:
		return "By Exact Country Code"
	case StationQueryByState:
		return "By State"
	case StationQueryByStateExact:
		return "By Exact State"
	case StationQueryByLanguage:
		return "By Language"
	case StationQueryByLanguageExact:
		return "By Exact Language"
	case StationQueryByTag:
		return "By Tag"
	case StationQueryByTagExact:
		return "By Exact Tag"
	}
	return "None"
}

func (m StationQuery) ExampleString() string {
	switch m {
	case StationQueryByName:
		return `
Examples:
- "BBC Radio" will return all stations with "BBC Radio" in their name
- "Italia" will return all stations with "Italia" in their name
- "Romance" will return all stations with "Romance" in their name
`
	case StationQueryByNameExact:
		return `
Examples:
- "BBC Radio 1" will return all stations with "BBC Radio 1" as their name
- "Radio Italia" will return all stations with "Radio Italia" as their name
- "Radio Romance" will return all stations with "Radio Romance" as their name
`
	case StationQueryByCodec:
		return `
Examples:
- "mp3" will return all stations with "mp3" in their codec
- "aac" will return all stations with "aac" in their codec
- "ogg" will return all stations with "ogg" in their codec
`
	case StationQueryByCodecExact:
		return `
Examples:
- "mp3" will return all stations with "mp3" as their codec
- "aac" will return all stations with "aac" as their codec
- "ogg" will return all stations with "ogg" as their codec
`
	case StationQueryByCountry:
		return `
Examples:
- "Italy" will return all stations with "Italy" in their country name.
- "United" will return all stations with "United" in their country name.
- "Republic" will return all stations with "Republic" in their country name.
`
	case StationQueryByCountryExact:
		return `
Examples:
- "Italy" will return all stations with "Italy" as their country
- "United States" will return all stations with "United States" as their country
- "United Kingdom" will return all stations with "United Kingdom" as their country
`
	case StationQueryByCountryCodeExact:
		return `
Examples:
- "IT" will return all stations with "IT" as their country code
- "US" will return all stations with "US" as their country code
- "UK" will return all stations with "UK" as their country code
`
	case StationQueryByState:
		return `
Examples:
- "Lombardy" will return all stations with "Lombardy" in their state
- "California" will return all stations with "California" in their state
- "New York" will return all stations with "New York" in their state
`
	case StationQueryByStateExact:
		return `
Examples:
- "Lombardy" will return all stations with "Lombardy" as their state
- "California" will return all stations with "California" as their state
- "New York" will return all stations with "New York" as their state
`
	case StationQueryByLanguage:
		return `
Examples:
- "Italian" will return all stations with "Italian" in their language
- "English" will return all stations with "English" in their language
- "Spanish" will return all stations with "Spanish" in their language
`
	case StationQueryByLanguageExact:
		return `
Examples:
- "Italian" will return all stations with "Italian" as their language
- "English" will return all stations with "English" as their language
- "Spanish" will return all stations with "Spanish" as their language
`
	case StationQueryByTag:
		return `
Examples:
- "rock" will return all stations with "rock" in their tags
- "jazz" will return all stations with "jazz" in their tags
- "pop" will return all stations with "pop" in their tags
`
	case StationQueryByTagExact:
		return `
Examples:
- "rock" will return all stations with "rock" as their tags
- "jazz" will return all stations with "jazz" as their tags
- "pop" will return all stations with "pop" as their tags
`
	}
	return ""
}
