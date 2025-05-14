package main

import (
	"fmt"
	"io"
	"log"
	"mime"
	"os"
	"regexp"
	"strings"

	"github.com/joho/godotenv"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"github.com/emersion/go-message/mail"
)

const (
	SubjectFilter             = "AWS Weekly Update (AWS Confidential)"
	SectionWhatsNew           = "What's New"
	SectionUpcomingLaunches   = "Upcoming Launches"
	SectionMainUpdates        = "주요 업데이트"
	TableHeaderTitle          = "제목"
	MIMETextPlain             = "text/plain"
	DateFormatPattern         = `^\d{4}년 \d{2}월 \d{2}일$`
	URLPattern                = `^(.*?)<(https?://[^>]+)>`
	MIMEFileExtension         = ".mime"
	TestdataDirectoryFallback = "./testdata"
	EnvFilePath               = ".env"
)

type NewsItem struct {
	Title string
	Link  string
	Date  string
}

func extractWhatsNewTable(body string) []NewsItem {
	lines := strings.Split(body, "\n")
	start := -1

	// "What's New" 이후부터 시작
	for i, line := range lines {
		if strings.Contains(line, SectionWhatsNew) {
			start = i
			break
		}
	}
	if start == -1 {
		log.Printf("Could not find section: %q", SectionWhatsNew)
		return nil
	}

	// 표 존재 구간(빈줄 2개 또는 다른 섹션 시작 전까지)만 검사
	re := regexp.MustCompile(URLPattern)
	datePattern := regexp.MustCompile(DateFormatPattern)
	var items []NewsItem

	for i := start + 1; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])
		if strings.Contains(line, SectionUpcomingLaunches) {
			break
		}
		m := re.FindStringSubmatch(line)
		if len(m) != 3 {
			continue
		}

		// 날짜 줄 탐색 (빈 줄 skip)
		date := ""
		for j := i + 1; j < len(lines); j++ {
			dateCandidate := strings.TrimSpace(lines[j])
			if dateCandidate == "" {
				continue
			}
			if datePattern.MatchString(dateCandidate) {
				date = dateCandidate
				i = j
				break
			}
			break
		}

		if date != "" {
			items = append(items, NewsItem{
				Title: m[1],
				Link:  m[2],
				Date:  date,
			})
		}

	}
	return items
}

func extractMainUpdates(body string) []string {
	lines := strings.Split(body, "\n")
	start := -1

	// "주요 업데이트" 이후부터 시작
	for i := start + 1; i < len(lines); i++ {
		if strings.Contains(lines[i], SectionMainUpdates) {
			start = i
			break
		}
	}
	if start == -1 {
		log.Printf("Could not find section: %q", SectionMainUpdates)
		return nil
	}

	re := regexp.MustCompile(`^\s*\*`)
	var updates []string
	for i := start + 1; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])
		if strings.Contains(line, TableHeaderTitle) {
			break
		}
		if re.MatchString(line) {
			update := re.ReplaceAllString(line, "")
			updates = append(updates, strings.TrimSpace(update))
		}
	}
	return updates
}

func fetchMailReadersFromIMAP(c *client.Client) ([]io.Reader, error) {
	mbox, err := c.Select("INBOX", false)
	if err != nil {
		log.Println("Failed to select INBOX:", err)
		return nil, err
	}
	if mbox.Messages == 0 {
		log.Println("No messages in mailbox")
		return nil, nil
	}

	criteria := imap.NewSearchCriteria()
	criteria.WithoutFlags = []string{imap.SeenFlag}
	criteria.Header.Add("Subject", SubjectFilter)
	ids, err := c.Search(criteria)
	if err != nil {
		log.Println("Search failed:", err)
		return nil, err
	}
	if len(ids) == 0 {
		log.Println("No unseen messages")
		return nil, nil
	}

	seqSet := new(imap.SeqSet)
	seqSet.AddNum(ids...)

	section := &imap.BodySectionName{}
	items := []imap.FetchItem{section.FetchItem()}

	messages := make(chan *imap.Message, len(ids))
	done := make(chan error, 1)

	// Fetch with error capture
	go func() {
		done <- c.Fetch(seqSet, items, messages)
	}()

	var readers []io.Reader
	for msg := range messages {
		if msg == nil {
			continue
		}
		r := msg.GetBody(section)
		if r == nil {
			log.Println("Message body is nil")
			continue
		}
		raw, err := io.ReadAll(r)
		if err != nil {
			log.Println("Failed to read message body:", err)
			continue
		}
		readers = append(readers, strings.NewReader(string(raw)))
	}

	if fetchErr := <-done; fetchErr != nil {
		fmt.Println("Failed to fetch messages:", fetchErr)
		return nil, err
	}

	return readers, nil
}

func parseMail(src io.Reader) ([]NewsItem, []string, string) {
	mr, err := mail.CreateReader(src)
	if err != nil {
		log.Println("CreateReader failed:", err)
		return nil, nil, ""
	}
	header := mr.Header
	subject, _ := header.Subject()

	printed := false
	var newsItems []NewsItem
	var updates []string

	for {
		p, err := mr.NextPart()
		if err == io.EOF {
			break
		} else if err != nil {
			log.Println("NextPart error", err)
			break
		}
		if h, ok := p.Header.(*mail.InlineHeader); ok {
			ct, _, _ := h.ContentType()
			mediaType, _, err := mime.ParseMediaType(ct)
			if err != nil {
				log.Println("Failed to parse media type:", err)
				break
			}
			if mediaType == MIMETextPlain {
				body, err := io.ReadAll(p.Body)
				if err != nil {
					log.Println("Failed to read message body:", err)
					continue
				}
				newsItems = extractWhatsNewTable(string(body))
				updates = extractMainUpdates(string(body))
				printed = true
				break
			}
		}
	}

	if !printed {
		log.Println("No text/plain body found")
	}

	return newsItems, updates, subject
}

// 디렉터리 내 모든 .mime 파일에서 처리
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
	err := godotenv.Load(EnvFilePath)
	if err != nil {
		log.Printf("Could not load .env file (%s): %v", EnvFilePath, err)
		log.Println("falling back to testdata mode")

		// Failover: 디렉토리에서 메일 읽기
		readers, err := fetchMailReadersFromDir(TestdataDirectoryFallback)
		if err != nil {
			log.Fatalf("Failed to read from testdata: %v", err)
		}
		for _, r := range readers {
			newsItems, updates, subject := parseMail(r)
			printMailSummary(subject, newsItems, updates)
		}
		return
	}

	// IMAP 처리 (.env 로드 성공시)
	imap_server := os.Getenv("IMAP_SERVER")
	imap_user := os.Getenv("IMAP_USER")
	imap_password := os.Getenv("IMAP_PASSWORD")

	log.Println("Connecting to server:", imap_server)

	c, err := client.DialTLS(imap_server, nil)
	if err != nil {
		log.Fatalf("failed to connect to IMAP server: %v", err)
	}
	defer c.Logout()
	log.Println("Connected")

	// Login
	if err := c.Login(imap_user, imap_password); err != nil {
		log.Fatalf("Login failed: %v", err)
	}
	log.Println("Logged in")

	readers, err := fetchMailReadersFromIMAP(c)
	if err != nil {
		log.Fatalf("Failed to fetch mail: %v", err)
		log.Fatal(err)
	}
	for _, r := range readers {
		newsItems, updates, subject := parseMail(r)
		printMailSummary(subject, newsItems, updates)
	}
}
