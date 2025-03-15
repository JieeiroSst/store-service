package utils

import (
	"encoding/base64"
	"regexp"
	"strings"

	"github.com/JIeeiroSst/user-service/common"
)

func DecodeBase(msg, decode string) bool {
	msgDecode, err := base64.StdEncoding.DecodeString(msg)
	if err != nil {
		return false
	}
	if !strings.EqualFold(string(msgDecode), decode) {
		return false
	}
	return true
}

func DecodeByte(msg string) []byte {
	sDec, err := base64.StdEncoding.DecodeString(msg)
	if err != nil {
		return nil
	}
	return sDec
}

func CheckPassword(password string) error {
	regex := `([A-Z])\w+`
	matched, err := regexp.MatchString(regex, password)
	if !matched {
		return common.PasswordFailed
	}
	if err != nil {
		return err
	}
	return nil
}

func CheckEmail(email string) error {
	regex := `^[a-z][a-z0-9_\.]{5,32}@[a-z0-9]{2,}(\.[a-z0-9]{2,4}){1,2}$`
	matched, err := regexp.MatchString(regex, email)
	if !matched {
		return common.EmailFailed
	}
	if err != nil {
		return nil
	}
	return nil
}

func CheckIP(ip string) error {
	regex := `/^(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$/`
	matched, err := regexp.MatchString(regex, ip)
	if !matched {
		return common.IPFailed
	}
	if err != nil {
		return err
	}
	return nil
}
