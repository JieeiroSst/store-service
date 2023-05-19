package main

import (
	"log"

	"golift.io/ffmpeg"
)

func main() {
	securitypsy := "rtsp://user:pass@127.0.0.1:8000/++stream?cameraNum=1"
	output := "/tmp/securitypsy_captured_file.mov"

	c := &ffmpeg.Config{
		FFMPEG: "/usr/local/bin/ffmpeg",
		Copy:   true, // do not transcode
		Audio:  true, // retain audio stream
		Time:   10,   // 10 seconds
	}

	encode := ffmpeg.Get(c)
	cmd, out, err := encode.SaveVideo(securitypsy, output, "SecuritySpyVideoTitle")

	log.Println("Command Used:", cmd)
	log.Println("Command Output:", out)

	if err != nil {
		log.Fatalln(err)
	}

	log.Println("Saved file from", securitypsy, "to", output)

	dahua := "rtsp://admin:password@192.168.1.12/live"
	output = "/tmp/dahua_captured_file.m4v"

	f := ffmpeg.Get(&ffmpeg.Config{
		Audio:  true, // retain audio stream
		Time:   10,   // 10 seconds
		Width:  1920,
		Height: 1080,
		CRF:    23,
		Level:  "4.0",
		Rate:   5,
		Prof:   "baseline",
	})

	cmd, out, err = f.SaveVideo(dahua, output, "DahuaVideoTitle")

	log.Println("Command Used:", cmd)
	log.Println("Command Output:", out)

	if err != nil {
		log.Fatalln(err)
	}

	log.Println("Saved file from", dahua, "to", output)
}