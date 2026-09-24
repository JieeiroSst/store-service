package pdfdoc

import (
	"bytes"
	"crypto/sha256"
	_ "embed"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/JIeeiroSst/customer-relationship-service/config"
	"github.com/JIeeiroSst/customer-relationship-service/internal/domain/model"
	"github.com/JIeeiroSst/customer-relationship-service/internal/domain/port"
	"github.com/go-pdf/fpdf"
	"go.uber.org/fx"
)

//go:embed fonts/Tinos-Regular.ttf
var fontRegular []byte

//go:embed fonts/Tinos-Bold.ttf
var fontBold []byte

//go:embed fonts/Tinos-Italic.ttf
var fontItalic []byte

//go:embed fonts/Tinos-BoldItalic.ttf
var fontBoldItalic []byte

const (
	family = "Tinos"

	marginLeft   = 30.0
	marginRight  = 20.0
	marginTop    = 20.0
	marginBottom = 20.0

	bodySize = 13.0
	lineH    = 6.4
	blank    = "………………………………"
)

var vietnam = time.FixedZone("ICT", 7*3600)

type renderer struct{ partyA config.CompanyConfig }

func NewRenderer(cfg *config.Config) port.ContractRenderer {
	return renderer{partyA: cfg.Company}
}

func (r renderer) Render(d port.RenewalDocument) ([]byte, error) {
	c, b := d.Contract, d.Account
	signingTime := d.SigningTime.UTC().Truncate(time.Second)
	endDate := d.EndDate.UTC().Truncate(time.Second)
	payload := model.SigningPayload(c, d.SignedBy, endDate, signingTime)
	sum := sha256.Sum256(payload)

	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.SetCompression(false)
	pdf.SetCatalogSort(true)
	pdf.SetCreationDate(signingTime)
	pdf.SetModificationDate(signingTime)
	pdf.SetProducer("customer-relationship-service", true)
	pdf.SetCreator("customer-relationship-service", true)
	pdf.SetTitle("Phụ lục gia hạn hợp đồng "+contractNumber(c), true)
	pdf.SetLang("vi-VN")
	pdf.SetMargins(marginLeft, marginTop, marginRight)
	pdf.SetAutoPageBreak(true, marginBottom)
	pdf.AddUTF8FontFromBytes(family, "", fontRegular)
	pdf.AddUTF8FontFromBytes(family, "B", fontBold)
	pdf.AddUTF8FontFromBytes(family, "I", fontItalic)
	pdf.AddUTF8FontFromBytes(family, "BI", fontBoldItalic)
	pdf.AliasNbPages("{nb}")

	pdf.SetFooterFunc(func() {
		pdf.SetY(-14)
		pdf.SetFont(family, "I", 9)
		pdf.SetTextColor(90, 90, 90)
		pdf.CellFormat(0, 5, fmt.Sprintf("Phụ lục gia hạn hợp đồng %s  ·  Mã xác thực %s  ·  Trang %d/{nb}",
			contractNumber(c), hex.EncodeToString(sum[:6]), pdf.PageNo()), "T", 0, "C", false, 0, "")
	})

	pdf.AddPage()
	w := &writer{pdf: pdf}

	// Quốc hiệu, tiêu ngữ.
	w.center("CỘNG HÒA XÃ HỘI CHỦ NGHĨA VIỆT NAM", "B", 13, 6)
	w.center("Độc lập – Tự do – Hạnh phúc", "B", 13.5, 6)
	y := pdf.GetY() + 0.5
	pdf.SetLineWidth(0.3)
	pdf.Line(105-27, y, 105+27, y)
	pdf.SetY(y + 6)

	// Tên văn bản, số.
	w.center("PHỤ LỤC GIA HẠN HỢP ĐỒNG", "B", 16, 8)
	w.center("Số: "+contractNumber(c)+"/PLGH-"+fmt.Sprint(signingTime.In(vietnam).Year()), "", bodySize, lineH)
	if c.Title != "" {
		w.center("("+c.Title+")", "I", bodySize, lineH)
	}
	pdf.Ln(3)

	// Căn cứ.
	for _, line := range []string{
		"Căn cứ Bộ luật Dân sự số 91/2015/QH13 ngày 24 tháng 11 năm 2015;",
		"Căn cứ Luật Thương mại số 36/2005/QH11 ngày 14 tháng 6 năm 2005;",
		"Căn cứ Luật Giao dịch điện tử số 20/2023/QH15 ngày 22 tháng 6 năm 2023;",
		fmt.Sprintf("Căn cứ Hợp đồng số %s giữa hai bên và thỏa thuận của các bên.", contractNumber(c)),
	} {
		w.para("I", line, "J")
	}
	pdf.Ln(2)

	w.para("", fmt.Sprintf("Hôm nay, %s, tại %s, chúng tôi gồm:", vietnameseDate(signingTime), or(r.partyA.Place)), "J")
	pdf.Ln(1)

	// Các bên.
	w.heading("BÊN A", "(sau đây gọi là “Bên A”)")
	w.field("Tên đơn vị", r.partyA.Name, true)
	w.field("Địa chỉ", r.partyA.Address, false)
	w.field("Mã số thuế", r.partyA.TaxCode, false)
	w.field("Điện thoại", r.partyA.Phone, false)
	w.field("Đại diện", r.partyA.Representative, false)
	w.field("Chức vụ", r.partyA.Title, false)
	pdf.Ln(2)

	w.heading("BÊN B", "(sau đây gọi là “Bên B”)")
	w.field("Tên đơn vị", b.Name, true)
	w.field("Địa chỉ", b.BillingAddress, false)
	w.field("Mã số thuế", b.TaxCode, false)
	w.field("Điện thoại", b.Phone, false)
	w.field("Đại diện", d.SignedBy, false)
	w.field("Chức vụ", b.RepresentativeTitle, false)
	pdf.Ln(2)

	w.para("", "Hai bên thống nhất ký Phụ lục gia hạn hợp đồng này với các điều khoản như sau:", "J")
	pdf.Ln(1)

	// Điều khoản.
	w.article("Điều 1. Gia hạn hợp đồng")
	oldEnd := blank
	if c.EndDate != nil {
		oldEnd = vietnameseDate(c.EndDate.In(vietnam))
	}
	w.para("", fmt.Sprintf("1.1. Hợp đồng số %s hết hiệu lực vào %s.", contractNumber(c), oldEnd), "J")
	w.para("", fmt.Sprintf("1.2. Hai bên đồng ý gia hạn Hợp đồng đến hết %s.", vietnameseDate(endDate.In(vietnam))), "J")

	w.article("Điều 2. Các nội dung khác")
	w.para("", "Các điều khoản khác của Hợp đồng không trái với Phụ lục này vẫn giữ nguyên giá trị pháp lý. Phụ lục này là bộ phận không tách rời của Hợp đồng.", "J")

	w.article("Điều 3. Hình thức và hiệu lực")
	w.para("", "3.1. Phụ lục này được lập dưới dạng thông điệp dữ liệu và được ký bằng chữ ký số theo quy định của Luật Giao dịch điện tử; các bên thừa nhận giá trị pháp lý của Phụ lục và chữ ký số trên đó.", "J")
	if r.partyA.SignKeyFile != "" {
		w.para("", "3.2. Phụ lục do Bên B ký số trước, sau đó Bên A ký số đối ứng; Phụ lục có hiệu lực kể từ thời điểm Bên A ký số đối ứng (thời điểm ghi trên chữ ký số của Bên A).", "J")
	} else {
		w.para("", fmt.Sprintf("3.2. Phụ lục có hiệu lực kể từ thời điểm Bên B ký số (dự kiến %s) và được Bên A ghi nhận trên hệ thống quản lý hợp đồng.", signingTime.In(vietnam).Format("15:04 02/01/2006")), "J")
	}

	w.article("Điều 4. Giải quyết tranh chấp")
	w.para("", "Mọi tranh chấp phát sinh từ Phụ lục này được giải quyết trước hết bằng thương lượng, hòa giải giữa hai bên; nếu không thành, một trong hai bên có quyền khởi kiện tại Tòa án có thẩm quyền theo quy định của pháp luật Việt Nam.", "J")

	// Chữ ký.
	w.ensureSpace(58)
	pdf.Ln(6)
	colW := (210 - marginLeft - marginRight) / 2
	top := pdf.GetY()
	w.signBlock(marginLeft, top, colW, "ĐẠI DIỆN BÊN A", r.partyA.Representative)
	w.signBlock(marginLeft+colW, top, colW, "ĐẠI DIỆN BÊN B", d.SignedBy)
	pdf.SetY(top + 46)

	// Machine-readable binding, kept small.
	pdf.Ln(4)
	pdf.SetFont(family, "", 7)
	pdf.SetTextColor(120, 120, 120)
	pdf.MultiCell(0, 3.4, "Dữ liệu ký số (payload): "+string(payload), "", "L", false)

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

type writer struct{ pdf *fpdf.Fpdf }

func (w *writer) center(text, style string, size, h float64) {
	w.pdf.SetFont(family, style, size)
	w.pdf.SetTextColor(0, 0, 0)
	w.pdf.CellFormat(0, h, text, "", 1, "C", false, 0, "")
}

func (w *writer) para(style, text, align string) {
	w.pdf.SetFont(family, style, bodySize)
	w.pdf.SetTextColor(0, 0, 0)
	w.pdf.MultiCell(0, lineH, text, "", align, false)
}

func (w *writer) heading(title, note string) {
	w.pdf.SetTextColor(0, 0, 0)
	w.pdf.SetFont(family, "B", bodySize)
	w.pdf.Write(lineH, title+" ")
	w.pdf.SetFont(family, "I", bodySize)
	w.pdf.Write(lineH, note)
	w.pdf.Ln(lineH)
}

func (w *writer) article(title string) {
	w.ensureSpace(24)
	w.pdf.Ln(1)
	w.pdf.SetFont(family, "B", bodySize)
	w.pdf.SetTextColor(0, 0, 0)
	w.pdf.MultiCell(0, lineH, title, "", "L", false)
}

// field prints "- Label: value", with dotted blanks for missing values.
func (w *writer) field(label, value string, bold bool) {
	w.pdf.SetTextColor(0, 0, 0)
	w.pdf.SetFont(family, "", bodySize)
	w.pdf.Write(lineH, "-  "+label+": ")
	style := ""
	if bold {
		style = "B"
	}
	w.pdf.SetFont(family, style, bodySize)
	w.pdf.Write(lineH, or(value))
	w.pdf.Ln(lineH)
}

func (w *writer) ensureSpace(mm float64) {
	_, pageH := w.pdf.GetPageSize()
	if w.pdf.GetY()+mm > pageH-marginBottom {
		w.pdf.AddPage()
	}
}

// signBlock draws one column of the signature area at (x, y).
func (w *writer) signBlock(x, y, width float64, title, name string) {
	p := w.pdf
	p.SetTextColor(0, 0, 0)
	p.SetXY(x, y)
	p.SetFont(family, "B", bodySize)
	p.CellFormat(width, lineH, title, "", 1, "C", false, 0, "")
	p.SetX(x)
	p.SetFont(family, "I", 11)
	p.CellFormat(width, 5.5, "(Ký số, ghi rõ họ tên, đóng dấu)", "", 1, "C", false, 0, "")
	p.SetXY(x, y+36)
	p.SetFont(family, "B", bodySize)
	p.CellFormat(width, lineH, name, "", 1, "C", false, 0, "")
}

func or(s string) string {
	if strings.TrimSpace(s) == "" {
		return blank
	}
	return s
}

func contractNumber(c *model.Contract) string {
	if strings.TrimSpace(c.Number) != "" {
		return c.Number
	}
	return fmt.Sprintf("HĐ-%06d", c.ID)
}

// vietnameseDate formats t as "ngày 05 tháng 03 năm 2026".
func vietnameseDate(t time.Time) string {
	t = t.In(vietnam)
	return fmt.Sprintf("ngày %02d tháng %02d năm %d", t.Day(), int(t.Month()), t.Year())
}

var Module = fx.Options(fx.Provide(NewRenderer))
