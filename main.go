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

type NewsItem struct {
	Title string
	Link  string
	Date  string
}

func extractWhatsNewTable(body string) []NewsItem {
	// fmt.Print(body)
	lines := strings.Split(body, "\n")
	start := -1

	// 1. "What's New" 이후부터 시작
	for i, line := range lines {
		if strings.Contains(line, "What's New") {
			start = i
			break
		}
	}
	if start == -1 {
		return nil
	}

	// 2. 표 존재 구간(빈줄 2개 또는 다른 섹션 시작 전까지)만 검사
	re := regexp.MustCompile(`^(.*?)<(https?://[^>]+)>`)
	datePattern := regexp.MustCompile(`^\d{4}년 \d{2}월 \d{2}일$`)
	var items []NewsItem

	for i := start + 1; i < len(lines)-1; i++ {
		line := strings.TrimSpace(lines[i])
		if strings.Contains(line, "Upcoming Launches") {
			break
		}
		m := re.FindStringSubmatch(line)
		if len(m) == 3 {
			// 날짜 줄 탐색 (빈 줄 skip)
			date := ""
			for j := i + 1; j < len(lines); j++ {
				dateCandidate := strings.TrimSpace(lines[j])
				if dateCandidate == "" {
					continue
				}
				date = dateCandidate
				break
			}
			if datePattern.MatchString(date) {
				items = append(items, NewsItem{
					Title: m[1],
					Link:  m[2],
					Date:  date,
				})
			}
		}
	}
	return items
}

func extractMainUpdates(body string) []string {
	lines := strings.Split(body, "\n")
	start, end := -1, -1

	// 1. "What's New" 찾기
	for i, line := range lines {
		if strings.Contains(line, "What's New") {
			start = i
			break
		}
	}
	if start == -1 {
		return nil
	}
	// 2. "주요 업데이트" 찾기
	for i := start + 1; i < len(lines); i++ {
		if strings.Contains(lines[i], "주요 업데이트") {
			start = i
			break
		}
	}
	if start == -1 {
		return nil
	}
	// 3. "제목" 혹은 표 시작 이전까지 끝 인덱스 찾기
	for i := start + 1; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == "제목" {
			end = i
			break
		}
	}
	if end == -1 {
		end = len(lines)
	}
	// 4. 주요 업데이트 "* ..." 줄만 추출
	re := regexp.MustCompile(`^\s*\*`)
	var updates []string
	for i := start + 1; i < end; i++ {
		line := lines[i]
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
		return nil, err
	}
	if mbox.Messages == 0 {
		log.Println("No messages in mailbox")
		return nil, nil
	}

	criteria := imap.NewSearchCriteria()
	criteria.WithoutFlags = []string{imap.SeenFlag}
	criteria.Header.Add("Subject", "AWS Weekly Update (AWS Confidential)")
	ids, err := c.Search(criteria)
	if err != nil {
		return nil, err
	}
	if len(ids) == 0 {
		log.Println("No unseen messages")
		return nil, nil
	}

	seqSet := new(imap.SeqSet)
	seqSet.AddNum(ids...)

	section := &imap.BodySectionName{}
	items := []imap.FetchItem{section.FetchItem(), imap.FetchEnvelope}

	messages := make(chan *imap.Message, len(ids))
	var readers []io.Reader
	go func() {
		if err := c.Fetch(seqSet, items, messages); err != nil {
			log.Fatal(err)
		}
	}()
	for msg := range messages {
		if msg == nil {
			continue
		}
		r := msg.GetBody(section)
		if r == nil {
			continue
		}
		raw, err := io.ReadAll(r)
		if err != nil {
			continue
		}
		readers = append(readers, strings.NewReader(string(raw)))
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
			log.Println("Error reading part:", err)
			break
		}
		if h, ok := p.Header.(*mail.InlineHeader); ok {
			ct, _, _ := h.ContentType()
			mediaType, _, err := mime.ParseMediaType(ct)
			if err != nil {
				mediaType = strings.ToLower(ct)
			}
			if mediaType == "text/plain" {
				body, _ := io.ReadAll(p.Body)
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

// .mime 파일 하나에서 읽어오기 (io.Reader 반환)
func openMimeFile(path string) (io.Reader, error) {
	return os.Open(path)
}

// 디렉터리 내 모든 .mime 파일에서 처리
func fetchMailReadersFromDir(dir string) ([]io.Reader, error) {
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var readers []io.Reader
	for _, f := range files {
		if !f.IsDir() && strings.HasSuffix(f.Name(), ".mime") {
			path := dir + "/" + f.Name()
			src, err := openMimeFile(path)
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
	err := godotenv.Load(".env")
	if err != nil {
		log.Println("Error loading .env file; falling back to ./testdata mode")

		// Failover: 디렉토리에서 메일 읽기
		readers, err := fetchMailReadersFromDir("./testdata")
		if err != nil {
			log.Fatal(err)
		}
		for _, r := range readers {
			newsItems, updates, subject := parseMail(r)
			printMailSummary(subject, newsItems, updates)
		}
		return
	}

	// 2. IMAP 처리 (.env 로드 성공시)
	imap_server := os.Getenv("IMAP_SERVER")
	imap_user := os.Getenv("IMAP_USER")
	imap_password := os.Getenv("IMAP_PASSWORD")

	log.Println(imap_server)
	log.Println("Connecting to server...")

	c, err := client.DialTLS(imap_server, nil)
	if err != nil {
		log.Fatal("err")
	}
	defer c.Logout()
	log.Println("Connected")

	// Login
	if err := c.Login(imap_user, imap_password); err != nil {
		log.Fatal(err)
	}
	log.Println("Logged in")

	readers, err := fetchMailReadersFromIMAP(c)
	if err != nil {
		log.Fatal(err)
	}
	for _, r := range readers {
		newsItems, updates, subject := parseMail(r)
		printMailSummary(subject, newsItems, updates)
	}
}
