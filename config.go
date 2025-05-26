package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Mode int

const (
	ModeTestdata Mode = iota
	ModeIMAP
)

type Config struct {
	Mode         Mode
	ImapServer   string
	ImapUser     string
	ImapPassword string
	TestdataDir  string
	DBUrl        string
}

const (
	envFile         = ".env"
	defaultTestdata = "./testdata"
)

func loadConfig() Config {
	// .env 존재 시 IMAP 모드, 없으면 테스트 모드
	if err := godotenv.Load(envFile); err != nil {
		log.Printf("No %s; fallback to testdata mode", envFile)
		return Config{
			Mode:        ModeTestdata,
			TestdataDir: defaultTestdata,
		}
	}
	return Config{
		Mode:         ModeIMAP,
		ImapServer:   os.Getenv("IMAP_SERVER"),
		ImapUser:     os.Getenv("IMAP_USER"),
		ImapPassword: os.Getenv("IMAP_PASSWORD"),
		TestdataDir:  defaultTestdata,
		DBUrl:        os.Getenv("DATABASE_URL"),
	}
}
