package infrastructure

import "log"

func initLogger() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}
