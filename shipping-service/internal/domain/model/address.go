package model

import (
	"errors"
	"regexp"
	"strings"
)

type Address struct {
	Name       string `json:"name"`
	Phone      string `json:"phone"`
	Street     string `json:"street"`
	WardCode   string `json:"ward_code"`
	DistrictID int    `json:"district_id"`
	ProvinceID int    `json:"province_id"`
}

var (
	vnMobile   = regexp.MustCompile(`^0[35789][0-9]{8}$`)
	vnLandline = regexp.MustCompile(`^02[0-9]{9}$`)

	ErrInvalidPhone = errors.New("phone must be a Vietnamese number, e.g. 0912345678 or +84912345678")
)

func NormalizeVNPhone(raw string) (string, error) {
	p := strings.NewReplacer(" ", "", ".", "", "-", "", "(", "", ")", "").Replace(strings.TrimSpace(raw))
	switch {
	case strings.HasPrefix(p, "+84"):
		p = "0" + p[3:]
	case strings.HasPrefix(p, "84") && len(p) == 11:
		p = "0" + p[2:]
	}
	if !vnMobile.MatchString(p) && !vnLandline.MatchString(p) {
		return "", ErrInvalidPhone
	}
	return p, nil
}

func (a Address) Normalize() (Address, error) {
	a.Name = strings.TrimSpace(a.Name)
	a.Street = strings.TrimSpace(a.Street)
	a.WardCode = strings.TrimSpace(a.WardCode)
	switch {
	case a.Name == "":
		return a, errors.New("name is required")
	case a.Street == "":
		return a, errors.New("street is required")
	case a.WardCode == "" || a.DistrictID <= 0:
		return a, errors.New("ward_code and district_id are required")
	}
	phone, err := NormalizeVNPhone(a.Phone)
	if err != nil {
		return a, err
	}
	a.Phone = phone
	return a, nil
}
