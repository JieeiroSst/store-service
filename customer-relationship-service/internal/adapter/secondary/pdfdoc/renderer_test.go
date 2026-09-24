package pdfdoc

import (
	"bytes"
	"testing"
	"time"

	"github.com/JIeeiroSst/customer-relationship-service/config"
	"github.com/JIeeiroSst/customer-relationship-service/internal/domain/model"
	"github.com/JIeeiroSst/customer-relationship-service/internal/domain/port"
)

func doc() port.RenewalDocument {
	at := time.Date(2026, 3, 5, 8, 30, 0, 0, time.UTC)
	return port.RenewalDocument{
		Contract:    &model.Contract{Base: model.Base{ID: 12}, AccountID: 3, Number: "HĐ-2025/001", Title: "Dịch vụ phần mềm quản lý"},
		Account:     &model.Account{Name: "Công ty TNHH Thương mại Ánh Dương", BillingAddress: "12 Nguyễn Huệ, Quận 1, TP. Hồ Chí Minh", TaxCode: "0312345678", Phone: "0901234567", RepresentativeTitle: "Giám đốc"},
		SignedBy:    "Nguyễn Thị Hồng Nhung",
		EndDate:     at.AddDate(1, 0, 0),
		SigningTime: at,
	}
}

func newRenderer() port.ContractRenderer {
	return NewRenderer(&config.Config{Company: config.CompanyConfig{
		Name: "Công ty Cổ phần Giải pháp Số Việt", Address: "99 Trần Hưng Đạo, Hoàn Kiếm, Hà Nội", TaxCode: "0109876543",
		Phone: "02412345678", Representative: "Trần Văn Bình", Title: "Tổng giám đốc", Place: "Hà Nội",
	}})
}

func TestRender_IsDeterministic(t *testing.T) {
	r := newRenderer()
	first, err := r.Render(doc())
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 20; i++ {
		again, err := r.Render(doc())
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(first, again) {
			t.Fatalf("render %d differs from the first", i)
		}
	}

	other := doc()
	other.SignedBy = "Lê Văn Cường"
	if changed, _ := r.Render(other); bytes.Equal(first, changed) {
		t.Fatal("a different signer produced the same bytes")
	}
	other = doc()
	other.Account = &model.Account{Name: "Công ty khác"}
	if changed, _ := r.Render(other); bytes.Equal(first, changed) {
		t.Fatal("a different account produced the same bytes")
	}
}

func TestRender_IsAPDFWithTheExpectedShape(t *testing.T) {
	out, err := newRenderer().Render(doc())
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.HasPrefix(out, []byte("%PDF-")) || !bytes.HasSuffix(bytes.TrimSpace(out), []byte("%%EOF")) {
		t.Fatalf("not a well-formed PDF: %.20q ... %.20q", out, out[len(out)-20:])
	}
	// Text is stored as glyph ids, so check that the TrueType font is embedded.
	if !bytes.Contains(out, []byte("/FontFile2")) {
		t.Error("the font is not embedded")
	}
}

func TestRender_HandlesMissingPartyDetails(t *testing.T) {
	d := doc()
	d.Account = &model.Account{}
	d.Contract.Number, d.Contract.Title = "", ""
	d.Contract.EndDate = nil
	if _, err := NewRenderer(&config.Config{}).Render(d); err != nil {
		t.Fatal(err)
	}
}

func TestVietnameseDate(t *testing.T) {
	// 20:00 UTC on the 31st is already the 1st in Vietnam.
	got := vietnameseDate(time.Date(2026, 1, 31, 20, 0, 0, 0, time.UTC))
	if got != "ngày 01 tháng 02 năm 2026" {
		t.Fatalf("got %q", got)
	}
}
