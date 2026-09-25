package http

import (
	"encoding/base64"
	"errors"
	"reflect"
	"testing"

	"github.com/JIeerioSst/rent-house-service/internal/domain"
)

func TestCursorRoundTrip(t *testing.T) {
	for _, c := range []*domain.HomestayCursor{
		{Sort: "newest", ID: 12},
		{Sort: "rating", ID: 3, Rating: 4.333333333333333, Count: 17},
		{Sort: "price_asc", ID: 5, Price: "5000000.00"},
		{Sort: "price_desc", ID: 5}, // a homestay without the rate: empty price
		{Sort: "recommended", ID: 8, Count: 4, Score: 4.6923076923076925},
	} {
		got, err := decodeCursor(encodeCursor(c))
		if err != nil || !reflect.DeepEqual(got, c) {
			t.Errorf("%+v -> %+v, %v", c, got, err)
		}
	}
	if encodeCursor(nil) != "" {
		t.Error("no cursor, no token")
	}
}

func TestDecodeCursorRejectsGarbage(t *testing.T) {
	enc := func(s string) string { return base64.RawURLEncoding.EncodeToString([]byte(s)) }
	for name, tok := range map[string]string{
		"not base64":     "%%%",
		"not json":       enc("hello"),
		"unknown sort":   enc(`{"s":"oldest","i":1}`),
		"missing id":     enc(`{"s":"newest"}`),
		"negative id":    enc(`{"s":"newest","i":-4}`),
		"unknown field":  enc(`{"s":"newest","i":1,"z":1}`),
		"negative count": enc(`{"s":"rating","i":1,"c":-1}`),
		"empty":          "",
	} {
		if _, err := decodeCursor(tok); !errors.Is(err, domain.ErrInvalid) {
			t.Errorf("%s: err = %v", name, err)
		}
	}
}
