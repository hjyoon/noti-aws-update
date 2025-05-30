package internal

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
	DBUser       string
	DBPassword   string
	DBHost       string
	DBPort       string
	DBName       string
	AppPort      string
}

const (
	envFile         = ".env"
	defaultTestdata = "./testdata"
)

func LoadConfig() Config {
	// .env 존재 시 IMAP 모드, 없으면 테스트 모드
	if err := godotenv.Load(envFile); err != nil {
		log.Printf("No %s; fallback to testdata mode", envFile)
		return Config{
			Mode:        ModeTestdata,
			TestdataDir: defaultTestdata,
		}
	}
	appPort := os.Getenv("APP_PORT")
	if appPort == "" {
		appPort = "8000" // 환경 변수 미설정 시 디폴트 포트번호 지정
	}
	return Config{
		Mode:         ModeIMAP,
		ImapServer:   os.Getenv("IMAP_SERVER"),
		ImapUser:     os.Getenv("IMAP_USER"),
		ImapPassword: os.Getenv("IMAP_PASSWORD"),
		TestdataDir:  defaultTestdata,
		DBUser:       os.Getenv("DATABASE_USER"),
		DBPassword:   os.Getenv("DATABASE_PASSWORD"),
		DBHost:       os.Getenv("DATABASE_HOST"),
		DBPort:       os.Getenv("DATABASE_PORT"),
		DBName:       os.Getenv("DATABASE_DB"),
		AppPort:      appPort,
	}
}
