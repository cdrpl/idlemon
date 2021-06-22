package main

import "flag"

func ParseCliFlags() (envFile string) {
	flag.StringVar(&envFile, "e", ENV_FILE, "path to the .env file. use -e nil to prevent .env file from being loaded")
	flag.Parse()
	return
}
