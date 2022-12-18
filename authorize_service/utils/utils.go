package utils

import (
	"encoding/base64"
	"strings"

	"github.com/JieeiroSst/authorize-service/pkg/log"
)

func DecodeBase(msg, decode string) bool {
	msgDecode, err := base64.StdEncoding.DecodeString(msg)
	if err != nil {
		log.Error(err)
		return false
	}
	if !strings.EqualFold(string(msgDecode), decode) {
		log.Error("Decode base failed")
		return false
	}
	log.Info("DEcode base success")
	return true
}
