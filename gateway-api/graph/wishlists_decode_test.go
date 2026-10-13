package graph

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"

	"github.com/JIeeiroSst/gateway-api/graph/model"
)

func TestDecodeWishlistsServiceResponses(t *testing.T) {
	decode := func(body string) any {
		var raw any
		dec := json.NewDecoder(bytes.NewBufferString(body))
		dec.UseNumber()
		if err := dec.Decode(&raw); err != nil {
			t.Fatal(err)
		}
		return raw
	}
	ctx := context.Background()

	w, err := restDecode[*model.Wishlist](ctx, "wishlists-service", "/", dig(decode(`{"data":{"id":10,"name":"Party","key":"","visibility":"Shared","date_created":"2026-10-13T15:57:05.888292+07:00","owner_id":"1","owner":{"id":"1","username":"alice","name":"Alice","active":true},"lines":[{"id":8,"wishlist_id":10,"product_id":9,"title":"Balloon","quantity":3}]}}`), "data"))
	if err != nil {
		t.Fatal(err)
	}
	if w.ID != 10 || w.Owner == nil || w.Owner.Username != "alice" || len(w.Lines) != 1 || w.Lines[0].ProductID != 9 || w.Lines[0].Quantity != 3 || w.DateCreated == "" {
		t.Fatalf("unexpected wishlist %+v", w)
	}

	page, err := restDecode[*model.WishlistPage](ctx, "wishlists-service", "/", dig(decode(`{"data":{"items":[],"total":5,"limit":50,"offset":0}}`), "data"))
	if err != nil || page.Total != 5 || page.Limit != 50 {
		t.Fatalf("page: %+v %v", page, err)
	}

	lookup, err := restDecode[*model.WishlistProductLookup](ctx, "wishlists-service", "/", dig(decode(`{"data":{"in_wishlist":true,"product_id":501,"wishlists":[{"wishlist_id":5,"wishlist_name":"Default","line_id":7,"quantity":2}]}}`), "data"))
	if err != nil || !lookup.InWishlist || lookup.Wishlists[0].LineID != 7 {
		t.Fatalf("lookup: %+v %v", lookup, err)
	}
}
