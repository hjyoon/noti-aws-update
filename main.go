package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

func fetchMailReadersFromDir(dir string) ([]io.Reader, error) {
	files, err := os.ReadDir(dir)
	if err != nil {
		log.Println("Failed to read directory:", err)
		return nil, err
	}
	var readers []io.Reader
	for _, f := range files {
		if !f.IsDir() && strings.HasSuffix(f.Name(), MIMEFileExtension) {
			path := dir + "/" + f.Name()
			src, err := os.Open(path)
			if err != nil {
				log.Printf("Failed to open %s: %v", path, err)
				continue
			}
			readers = append(readers, src)
		}
	}
	return readers, nil
}

func printMailSummary(subject string, newsItems []NewsItem, updates []string) {
	fmt.Println("Subject:", subject)
	fmt.Println("--- WhatsNewTable ---")
	for i, item := range newsItems {
		fmt.Println("========================================")
		fmt.Printf("%d.\n", i+1)
		fmt.Printf("제목: %s\n", strings.TrimSpace(item.Title))
		fmt.Printf("링크: %s\n", item.Link)
		fmt.Printf("날짜: %s\n", item.Date)
	}
	fmt.Println("--- MainUpdates ---")
	for _, u := range updates {
		fmt.Println(u)
	}
}

func main() {
	cfg := loadConfig()

	conn, err := NewDB(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close(context.Background())

	var readers []io.Reader

	switch cfg.Mode {
	case ModeIMAP:
		c, err := connectIMAP(cfg)
		if err != nil {
			log.Fatal(err)
		}
		defer c.Logout()
		readers, err = fetchMailReadersFromIMAP(c)
	case ModeTestdata:
		readers, err = fetchMailReadersFromDir(cfg.TestdataDir)
	}

	if err != nil {
		log.Fatal(err)
	}
	for _, r := range readers {
		newsItems, updates, subject := parseMail(r)
		printMailSummary(subject, newsItems, updates)
	}

	StartHTTPServer(conn)
}
