package geoip

import (
	"net"

	"github.com/JIeeiroSst/shortlink-service/internal/domain"
	"github.com/oschwald/geoip2-golang"
	"go.uber.org/zap"
)

type Lookup struct {
	reader *geoip2.Reader
	log    *zap.Logger
}

func New(path string, log *zap.Logger) (*Lookup, error) {
	if path == "" {
		return &Lookup{log: log}, nil
	}
	reader, err := geoip2.Open(path)
	if err != nil {
		return nil, err
	}
	return &Lookup{reader: reader, log: log}, nil
}

func (l *Lookup) Close() error {
	if l.reader == nil {
		return nil
	}
	return l.reader.Close()
}

func (l *Lookup) Lookup(ip string) domain.GeoLocation {
	if l.reader == nil {
		return domain.GeoLocation{}
	}
	parsed := net.ParseIP(ip)
	if parsed == nil {
		return domain.GeoLocation{}
	}
	record, err := l.reader.City(parsed)
	if err != nil || record.Country.IsoCode == "" {
		return domain.GeoLocation{}
	}

	code := record.Country.IsoCode
	name := record.Country.Names["en"]
	if name == "" {
		name = domain.CountryName(code)
	}
	loc := domain.GeoLocation{
		CountryCode: &code,
		CountryName: &name,
	}
	if len(record.Subdivisions) > 0 {
		region := record.Subdivisions[0].Names["en"]
		if region != "" {
			loc.Region = &region
		}
	}
	if city := record.City.Names["en"]; city != "" {
		loc.City = &city
	}
	if record.Location.Latitude != 0 || record.Location.Longitude != 0 {
		lat, lon := record.Location.Latitude, record.Location.Longitude
		loc.Latitude = &lat
		loc.Longitude = &lon
	}
	if tz := record.Location.TimeZone; tz != "" {
		loc.Timezone = &tz
	}
	return loc
}
