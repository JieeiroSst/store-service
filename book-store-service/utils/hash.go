package utils

import (
	"crypto/subtle"
	"encoding/base64"
)

func DecodeBase(msg, decode string) bool {
	msgDecode, err := base64.StdEncoding.DecodeString(msg)
	if err != nil {

		return false
	}
	if decode == "" || subtle.ConstantTimeCompare(msgDecode, []byte(decode)) != 1 {

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
