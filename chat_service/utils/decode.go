package utils

import (
	"crypto/subtle"
	"encoding/base64"

	"github.com/JIeeiroSst/chat-service/pkg/log"
)

func DecodeBase(msg, decode string) bool {
	msgDecode, err := base64.StdEncoding.DecodeString(msg)
	if err != nil {
		log.Error(err.Error())
		return false
	}
	if decode == "" || subtle.ConstantTimeCompare(msgDecode, []byte(decode)) != 1 {
		log.Error("Decode base failed")
		return false
	}
	log.Info("DEcode base success")
	return true
}

func DecodeByte(msg string) []byte {
	sDec, err := base64.StdEncoding.DecodeString(msg)
	if err != nil {
		return nil
	}
	return sDec
}
