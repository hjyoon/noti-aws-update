package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.krafton.com/ops2022/noti-aws-update/internal"
)

func fetchMailReadersFromDir(dir string) ([]io.Reader, error) {
	files, err := os.ReadDir(dir)
	if err != nil {
		log.Println("Failed to read directory:", err)
		return nil, err
	}
	var readers []io.Reader
	for _, f := range files {
		if !f.IsDir() && strings.HasSuffix(f.Name(), internal.MIMEFileExtension) {
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

func printMailSummary(subject string, newsItems []internal.NewsItem, updates []string) {
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
	cfg := internal.LoadConfig()

	pool, err := internal.NewDBPool(cfg)
	defer pool.Close()

	var readers []io.Reader

	switch cfg.Mode {
	case internal.ModeIMAP:
		c, err := internal.ConnectIMAP(cfg)
		if err != nil {
			log.Fatal(err)
		}
		defer c.Logout()
		readers, err = internal.FetchMailReadersFromIMAP(c)
	case internal.ModeTestdata:
		readers, err = fetchMailReadersFromDir(cfg.TestdataDir)
	}

	if err != nil {
		log.Fatal(err)
	}
	for _, r := range readers {
		newsItems, updates, subject := internal.ParseMail(r)
		printMailSummary(subject, newsItems, updates)
	}

	internal.StartHTTPServer(pool, cfg.AppPort)
}
