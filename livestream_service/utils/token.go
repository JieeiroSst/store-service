package main

import (
	"fmt"
	"time"

	"github.com/dgrijalva/jwt-go"
)

func Token() string {
	var VIDEOSDK_API_KEY = ""
	var VIDEOSDK_SECRET_KEY = ""

	var permissions [2]string
	permissions[0] = "allow_join"
	permissions[1] = "allow_mod"

	atClaims := jwt.MapClaims{}
	atClaims["apikey"] = VIDEOSDK_API_KEY
	atClaims["permissions"] = permissions
	atClaims["iat"] = time.Now().Unix()
	atClaims["exp"] = time.Now().Add(time.Minute * 120).Unix()
	at := jwt.NewWithClaims(jwt.SigningMethodHS256, atClaims)

	token, err := at.SignedString([]byte(VIDEOSDK_SECRET_KEY))
	if err != nil {
		fmt.Printf("%v\n", err)
	}
	return token
}
