package domain

// AdminLevel identifies a tier in the administrative hierarchy.
type AdminLevel int

const (
	LevelProvince AdminLevel = 1 // tỉnh / thành phố
	LevelDistrict AdminLevel = 2 // quận / huyện / thị xã
	LevelWard     AdminLevel = 3 // phường / xã / thị trấn
)

// Province is a level-1 unit (tỉnh/thành phố). Code = mã GSO 2 chữ số.
type Province struct {
	Code     string  `json:"code"`
	Name     string  `json:"name"`
	NameFull string  `json:"name_full,omitempty"`
	NameEn   string  `json:"name_en,omitempty"`
	Type     string  `json:"type,omitempty"`
	Region   string  `json:"region,omitempty"`
	Slug     string  `json:"slug,omitempty"`
	Gid      string  `json:"gid,omitempty"` // GADM GID_1
	AreaKm2  float64 `json:"area_km2,omitempty"`
}

// District is a level-2 unit (quận/huyện). Code = mã GSO 3 chữ số.
type District struct {
	Code         string  `json:"code"`
	ProvinceCode string  `json:"province_code"`
	Name         string  `json:"name"`
	NameFull     string  `json:"name_full,omitempty"`
	NameEn       string  `json:"name_en,omitempty"`
	Type         string  `json:"type,omitempty"`
	Slug         string  `json:"slug,omitempty"`
	Gid          string  `json:"gid,omitempty"` // GADM GID_2
	AreaKm2      float64 `json:"area_km2,omitempty"`
}

// Ward is a level-3 unit (phường/xã). Code = mã GSO 5 chữ số.
type Ward struct {
	Code         string  `json:"code"`
	DistrictCode string  `json:"district_code"`
	ProvinceCode string  `json:"province_code"`
	Name         string  `json:"name"`
	NameFull     string  `json:"name_full,omitempty"`
	NameEn       string  `json:"name_en,omitempty"`
	Type         string  `json:"type,omitempty"`
	Slug         string  `json:"slug,omitempty"`
	Gid          string  `json:"gid,omitempty"` // GADM GID_3
	AreaKm2      float64 `json:"area_km2,omitempty"`
}

// Point is a WGS84 geographic coordinate.
type Point struct {
	Lng float64 `json:"lng"`
	Lat float64 `json:"lat"`
}

// Street is a named road/street within a province.
type Street struct {
	ID           int     `json:"id"`
	ProvinceCode string  `json:"province_code"`
	DistrictCode string  `json:"district_code,omitempty"`
	Name         string  `json:"name"`
	NameFull     string  `json:"name_full,omitempty"`
	NameEn       string  `json:"name_en,omitempty"`
	Slug         string  `json:"slug,omitempty"`
	Gid          string  `json:"gid,omitempty"`
	LengthM      float64 `json:"length_m,omitempty"`
	DistanceM    float64 `json:"distance_m,omitempty"`
}

// LocationResult is the full administrative chain a point resolves to.
type LocationResult struct {
	Found     bool      `json:"found"`
	Province  *Province `json:"province,omitempty"`
	District  *District `json:"district,omitempty"`
	Ward      *Ward     `json:"ward,omitempty"`
	Street    *Street   `json:"street,omitempty"`
	Point     Point     `json:"point"`
	DistanceM float64   `json:"distance_m,omitempty"`
}

// WhitelistCity represents a province/city that is in the priority whitelist.
type WhitelistCity struct {
	ProvinceCode string `json:"province_code"`
	Note         string `json:"note,omitempty"`
}

// WhitelistResult is returned by the whitelist-check API.
type WhitelistResult struct {
	Whitelisted bool      `json:"whitelisted"`
	Province    *Province `json:"province,omitempty"`
	District    *District `json:"district,omitempty"`
	Ward        *Ward     `json:"ward,omitempty"`
	Street      *Street   `json:"street,omitempty"`
	Point       Point     `json:"point"`
}

// GeoFeature is a generic unit parsed from GeoJSON for the importer. The
// importer matches it to a seeded row by GSO code (preferred) or GADM gid.
type GeoFeature struct {
	Level       AdminLevel
	Gid         string // GADM GID_*
	Name        string
	ProvinceGid string
	DistrictGid string
	GeoJSON     string
}
