/*
Mail format assumptions
-----------------------
The plain-text body is expected to look like:

What's New
<title> <URL>
YYYY년 MM월 DD일

... (repeats)

Upcoming Launches <- marks end of table

주요 업데이트
* <bullet text> <- collect until "제목" header

Rules
  • "What's New" section ends when "Upcoming Launches" appears.
  • Each table row = one line "Title <URL>" followed by a non-blank date line.
  • Dates must match `^\d{4}년 \d{2}월 \d{2}일$`.
  • "주요 업데이트" collects bullet lines (`^\s*\*`), stops at "제목" header.

이 포맷만 만족하면 정상적으로 파싱이 된다.
testdata/ 폴더안에 예시 email 파일이 첨부되어 있음.
*/

package main

import (
	"io"
	"log"
	"mime"
	"regexp"
	"strings"

	"github.com/emersion/go-message/mail"
)

type NewsItem struct {
	Title string
	Link  string
	Date  string
}

func parseMail(src io.Reader) ([]NewsItem, []string, string) {
	mr, err := mail.CreateReader(src)
	if err != nil {
		log.Println("CreateReader failed:", err)
		return nil, nil, ""
	}
	subject, _ := mr.Header.Subject()

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
				txt := string(body)
				newsItems = extractWhatsNewTable(txt)
				updates = extractMainUpdates(txt)
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

// Whats New 표 추출
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

	n := len(lines)
	for i := start + 1; i < n; i++ {
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
		for j := i + 1; j < n; j++ {
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

	n := len(lines)

	// "주요 업데이트" 이후부터 시작
	for i := start + 1; i < n; i++ {
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
	for i := start + 1; i < n; i++ {
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
