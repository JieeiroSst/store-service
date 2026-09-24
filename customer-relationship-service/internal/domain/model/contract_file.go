package model

import (
	"strings"
	"time"

	"gorm.io/gorm"
)

// Kinds of contract file. Every file attached to a contract has exactly one.
const (
	FileKindContract   = "contract"   // the contract itself (hợp đồng)
	FileKindAppendix   = "appendix"   // phụ lục
	FileKindRenewal    = "renewal"    // the signed renewal appendix, kept by the service
	FileKindAttachment = "attachment" // tài liệu đính kèm
	FileKindScan       = "scan"       // bản scan hợp đồng giấy đã ký
	FileKindLegal      = "legal"      // giấy tờ pháp lý: đăng ký kinh doanh, giấy ủy quyền, ...
	FileKindAcceptance = "acceptance" // biên bản nghiệm thu, thanh lý
	FileKindPayment    = "payment"    // chứng từ thanh toán, hóa đơn
	FileKindOther      = "other"
)

// Where a file came from.
const (
	FileSourceUpload = "upload"
	FileSourceSystem = "system"
)

// File formats, by the content we recognise (not by what the client claims).
const (
	FormatPDF  = "pdf"
	FormatDOCX = "docx"
	FormatXLSX = "xlsx"
	FormatDOC  = "doc"
	FormatXLS  = "xls"
	FormatJPEG = "jpeg"
	FormatPNG  = "png"
)

// FileFormat describes an accepted format.
type FileFormat struct {
	Name        string
	ContentType string
	Extensions  []string
}

// FileFormats lists every format contract files may have.
var FileFormats = []FileFormat{
	{FormatPDF, "application/pdf", []string{".pdf"}},
	{FormatDOCX, "application/vnd.openxmlformats-officedocument.wordprocessingml.document", []string{".docx"}},
	{FormatXLSX, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", []string{".xlsx"}},
	{FormatDOC, "application/msword", []string{".doc"}},
	{FormatXLS, "application/vnd.ms-excel", []string{".xls"}},
	{FormatJPEG, "image/jpeg", []string{".jpg", ".jpeg"}},
	{FormatPNG, "image/png", []string{".png"}},
}

// FileKind is the rule set of one kind.
type FileKind struct {
	Name  string `json:"kind"`
	Label string `json:"label"`
	// Formats accepted for this kind.
	Formats []string `json:"formats"`
	// System kinds are written by the service only: they cannot be uploaded
	// or deleted through the API.
	System bool `json:"system"`
}

var (
	docs    = []string{FormatPDF, FormatDOCX, FormatDOC}
	images  = []string{FormatJPEG, FormatPNG}
	sheets  = []string{FormatXLSX, FormatXLS}
	general = []string{FormatPDF, FormatDOCX, FormatDOC, FormatXLSX, FormatXLS, FormatJPEG, FormatPNG}
)

// FileKinds lists the kinds in display order.
var FileKinds = []FileKind{
	{FileKindContract, "Hợp đồng", docs, false},
	{FileKindAppendix, "Phụ lục", docs, false},
	{FileKindRenewal, "Phụ lục gia hạn đã ký số", []string{FormatPDF}, true},
	{FileKindScan, "Bản scan hợp đồng đã ký", append([]string{FormatPDF}, images...), false},
	{FileKindLegal, "Giấy tờ pháp lý", append([]string{FormatPDF}, images...), false},
	{FileKindAcceptance, "Biên bản nghiệm thu, thanh lý", append(append([]string{}, docs...), images...), false},
	{FileKindPayment, "Chứng từ thanh toán", append(append([]string{FormatPDF}, sheets...), images...), false},
	{FileKindAttachment, "Tài liệu đính kèm", general, false},
	{FileKindOther, "Khác", general, false},
}

// LookupFileKind returns the rules of a kind.
func LookupFileKind(name string) (FileKind, bool) {
	for _, k := range FileKinds {
		if k.Name == name {
			return k, true
		}
	}
	return FileKind{}, false
}

// Allows reports whether the kind accepts the format.
func (k FileKind) Allows(format string) bool {
	for _, f := range k.Formats {
		if f == format {
			return true
		}
	}
	return false
}

// LookupFormat finds a format by name.
func LookupFormat(name string) (FileFormat, bool) {
	for _, f := range FileFormats {
		if f.Name == name {
			return f, true
		}
	}
	return FileFormat{}, false
}

// FormatForExtension maps a file extension (with the dot, any case) to a format.
func FormatForExtension(ext string) (FileFormat, bool) {
	ext = strings.ToLower(ext)
	for _, f := range FileFormats {
		for _, e := range f.Extensions {
			if e == ext {
				return f, true
			}
		}
	}
	return FileFormat{}, false
}

// ContractFile is a file attached to a contract: the contract itself, its
// appendices, scans, legal papers, the signed renewal and so on. The bytes
// live in the object store (ObjectKey) or, when none is configured, in Data.
// Deleting is soft: the row is hidden and the stored bytes are kept.
type ContractFile struct {
	Base
	DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"column:deleted_at;index"`

	ContractID  uint   `json:"contract_id" gorm:"column:contract_id;index"`
	Kind        string `json:"kind" gorm:"column:kind;size:32;index"`
	Name        string `json:"name" gorm:"column:name;size:255"`
	ContentType string `json:"content_type" gorm:"column:content_type;size:128"`
	Format      string `json:"format" gorm:"column:format;size:16"`
	Size        int64  `json:"size" gorm:"column:size"`
	SHA256      string `json:"sha256" gorm:"column:sha256;size:64"`
	Description string `json:"description" gorm:"column:description;size:1024"`
	// UploadedBy is the uploader's display name, UploaderID their subject.
	UploadedBy string `json:"uploaded_by" gorm:"column:uploaded_by;size:255"`
	UploaderID string `json:"uploader_id" gorm:"column:uploader_id;size:255;index"`
	Source     string `json:"source" gorm:"column:source;size:16"`

	// Versions: a new upload that replaces a file joins its lineage with the
	// next version number and becomes the latest; the older ones stay.
	// LineageID is the id of the first version (0 for the first version itself).
	LineageID  uint  `json:"lineage_id" gorm:"column:lineage_id;index"`
	Version    int   `json:"version" gorm:"column:version"`
	Latest     bool  `json:"latest" gorm:"column:latest;index"`
	ReplacesID *uint `json:"replaces_id" gorm:"column:replaces_id"`

	// What the upload checks found.
	ScanStatus string     `json:"scan_status" gorm:"column:scan_status;size:16"`
	ScanEngine string     `json:"scan_engine" gorm:"column:scan_engine;size:64"`
	ScannedAt  *time.Time `json:"scanned_at" gorm:"column:scanned_at"`

	ObjectKey string `json:"-" gorm:"column:object_key;size:512"`
	Data      []byte `json:"-" gorm:"column:data;type:mediumblob"`
}

func (ContractFile) TableName() string { return "crm_contract_file" }

// Root is the id that names the file's lineage.
func (f ContractFile) Root() uint {
	if f.LineageID != 0 {
		return f.LineageID
	}
	return f.ID
}

func (ContractFile) SearchColumns() []string { return []string{"name", "description"} }

// Scan outcomes.
const (
	ScanClean   = "clean"
	ScanSkipped = "skipped" // no scanner reachable and scanning is not required
	ScanSystem  = "system"  // produced by the service, not scanned
)

// What happened to a file.
const (
	FileEventUploaded   = "uploaded"
	FileEventNewVersion = "new_version"
	FileEventDownloaded = "downloaded"
	FileEventDeleted    = "deleted"
	FileEventRejected   = "rejected" // an upload the checks refused
	FileEventDenied     = "denied"   // an action the caller may not do
)

// ContractFileEvent is one entry of the audit trail of contract files. Entries
// are only ever added.
type ContractFileEvent struct {
	ID         uint      `json:"id" gorm:"column:id;primaryKey"`
	At         time.Time `json:"at" gorm:"column:at;index"`
	ContractID uint      `json:"contract_id" gorm:"column:contract_id;index"`
	// FileID is 0 for an upload that was rejected before it became a file;
	// LineageID is the file's lineage root (see ContractFile.Root).
	FileID    uint   `json:"file_id" gorm:"column:file_id;index"`
	LineageID uint   `json:"lineage_id" gorm:"column:lineage_id;index"`
	Version   int    `json:"version" gorm:"column:version"`
	FileName  string `json:"file_name" gorm:"column:file_name;size:255"`
	Kind      string `json:"kind" gorm:"column:kind;size:32"`
	Action    string `json:"action" gorm:"column:action;size:32;index"`

	ActorID      string `json:"actor_id" gorm:"column:actor_id;size:255;index"`
	ActorName    string `json:"actor_name" gorm:"column:actor_name;size:255"`
	ActorRoles   string `json:"actor_roles" gorm:"column:actor_roles;size:255"`
	ActorService bool   `json:"actor_service" gorm:"column:actor_service"`
	IP           string `json:"ip" gorm:"column:ip;size:64"`

	Detail string `json:"detail" gorm:"column:detail;size:1024"`
}

func (ContractFileEvent) TableName() string { return "crm_contract_file_event" }
