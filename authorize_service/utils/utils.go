package utils

import (
	"encoding/base64"
	"strings"

	"github.com/JIeeiroSst/utils/logger"
)

func DecodeBase(msg, decode string) bool {
	msgDecode, err := base64.StdEncoding.DecodeString(msg)
	if err != nil {
		logger.ConfigZap().Error(err)
		return false
	}
	if !strings.EqualFold(string(msgDecode), decode) {
		logger.ConfigZap().Error("Decode base failed")
		return false
	}
	logger.ConfigZap().Info("Decode base success")
	return true
}

func DecodeByte(msg string) []byte {
	sDec, err := base64.StdEncoding.DecodeString(msg)
	if err != nil {
		logger.ConfigZap().Error(err)
		return nil
	}
	return sDec
}
