package common

type StationQuery string

const (
	StationQueryAll                StationQuery = ""
	StationQueryByUuid             StationQuery = "byuuid"
	StationQueryByName             StationQuery = "byname"
	StationQueryByNameExact        StationQuery = "bynameexact"
	StationQueryByCodec            StationQuery = "bycodec"
	StationQueryByCodecExact       StationQuery = "bycodecexact"
	StationQueryByCountry          StationQuery = "bycountry"
	StationQueryByCountryExact     StationQuery = "bycountryexact"
	StationQueryByCountryCodeExact StationQuery = "bycountrycodeexact"
	StationQueryByState            StationQuery = "bystate"
	StationQueryByStateExact       StationQuery = "bystateexact"
	StationQueryByLanguage         StationQuery = "bylanguage"
	StationQueryByLanguageExact    StationQuery = "bylanguageexact"
	StationQueryByTag              StationQuery = "bytag"
	StationQueryByTagExact         StationQuery = "bytagexact"
)
