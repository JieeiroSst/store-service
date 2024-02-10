package utils

import (
	"encoding/base64"
	"strings"

	"github.com/JIeeiroSst/search-service/pkg/logger"
)

func DecodeBase(msg, decode string) bool {
	msgDecode, err := base64.StdEncoding.DecodeString(msg)
	if err != nil {
		logger.Logger().Error(err.Error())
		return false
	}
	if !strings.EqualFold(string(msgDecode), decode) {
		logger.Logger().Error("Decode base failed")
		return false
	}
	logger.Logger().Info("DEcode base success")
	return true
}

func DecodeByte(msg string) []byte {
	sDec, err := base64.StdEncoding.DecodeString(msg)
	if err != nil {
		return nil
	}
	return sDec
}
