package domain

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"strconv"
	"strings"
	"time"
	"unicode"
)

var (
	ErrFileTooLarge       = errors.New("file too large")
	ErrUnsupportedFile    = errors.New("unsupported file type")
	ErrFileNotFound       = errors.New("file not found")
	ErrInvalidSlot        = errors.New("invalid slot")
	ErrSlotSizeMismatch   = errors.New("unexpected size")
	ErrStorageUnavailable = errors.New("file storage unavailable")
)

const (
	SlotTimeout      = 5 * time.Hour
	slotSecretLength = 40
	alphanum         = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
)

type UploadSlot struct {
	UserStr  string
	Secret   string
	FileName string
	Size     int64
	Expires  int64
}

func (s UploadSlot) Path() string {
	return s.UserStr + "/" + s.Secret + "/" + s.FileName
}

func (s UploadSlot) StorageKey() string {
	return "ejabberd:" + s.UserStr + "/" + s.Secret
}

type StoredFile struct {
	Name        string
	ContentType string
	Size        int64
	Body        io.ReadCloser
}

func NewUploadSlot(owner JID, fileName string, size int64, now time.Time) (UploadSlot, error) {
	secret, err := randomAlphanum(slotSecretLength)
	if err != nil {
		return UploadSlot{}, err
	}
	sum := sha1.Sum([]byte(owner.LUser + "@" + owner.LServer))
	return UploadSlot{
		UserStr:  hex.EncodeToString(sum[:]),
		Secret:   secret,
		FileName: ReplaceSpecialChars(fileName),
		Size:     size,
		Expires:  now.Add(SlotTimeout).Unix(),
	}, nil
}

func ReplaceSpecialChars(s string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' || r == '.' || r == '-' {
			return r
		}
		return '_'
	}, s)
}

func randomAlphanum(n int) (string, error) {
	buf := make([]byte, n)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	for i, b := range buf {
		buf[i] = alphanum[int(b)%len(alphanum)]
	}
	return string(buf), nil
}

func SignSlot(key []byte, path string, size, expires int64) string {
	m := hmac.New(sha256.New, key)
	m.Write([]byte(path + " " + strconv.FormatInt(size, 10) + " " + strconv.FormatInt(expires, 10)))
	return hex.EncodeToString(m.Sum(nil))
}

func VerifySlot(key []byte, path string, size, expires int64, sig string, now time.Time) bool {
	if expires < now.Unix() {
		return false
	}
	want := SignSlot(key, path, size, expires)
	return hmac.Equal([]byte(want), []byte(sig))
}

func FileTooLarge(maxSize int64, lang string) StanzaError {
	e := ErrNotAcceptable("File larger than " + strconv.FormatInt(maxSize, 10) + " bytes")
	e.Lang = lang
	tl := NewElement("file-too-large", NSHTTPUpload)
	tl.Append(NewElement("max-file-size", "").AppendText(strconv.FormatInt(maxSize, 10)))
	e.Extra = []*Element{tl}
	return e
}
