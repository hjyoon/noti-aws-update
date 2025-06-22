package internal

import (
	"strings"
	"testing"
)

func TestExtractWhatsNewTable(t *testing.T) {
	const mailBody = `
What's New
Title A <https://example.com/a>
2024년 06월 01일

Title B <https://example.com/b>
2024년 06월 02일

Upcoming Launches

주요 업데이트
* 업데이트 1
* 업데이트 2
`
	items := extractWhatsNewTable(mailBody)
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}
	if items[0].Title != "Title A" {
		t.Errorf("items[0].Title: got %q, want %q", items[0].Title, "Title A")
	}
	if items[0].Link != "https://example.com/a" {
		t.Errorf("items[0].Link: got %q, want %q", items[0].Link, "https://example.com/a")
	}
	if items[0].Date != "2024년 06월 01일" {
		t.Errorf("items[0].Date: got %q, want %q", items[0].Date, "2024년 06월 01일")
	}
	if items[1].Title != "Title B" {
		t.Errorf("items[1].Title: got %q, want %q", items[1].Title, "Title B")
	}
	if items[1].Link != "https://example.com/b" {
		t.Errorf("items[1].Link: got %q, want %q", items[1].Link, "https://example.com/b")
	}
	if items[1].Date != "2024년 06월 02일" {
		t.Errorf("items[1].Date: got %q, want %q", items[1].Date, "2024년 06월 02일")
	}
}

func TestExtractMainUpdates(t *testing.T) {
	const mailBody = `
주요 업데이트
* 신규 서비스 오픈
* 정책 변경
제목
`
	updates := extractMainUpdates(mailBody)
	if len(updates) != 2 {
		t.Fatalf("expected 2 updates, got %d", len(updates))
	}
	if updates[0] != "* 신규 서비스 오픈" {
		t.Errorf("updates[0]: got %q, want %q", updates[0], "* 신규 서비스 오픈")
	}
	if updates[1] != "* 정책 변경" {
		t.Errorf("updates[1]: got %q, want %q", updates[1], "* 정책 변경")
	}
}

func TestParseMail(t *testing.T) {
	rawMail := `Subject: AWS Weekly Update (AWS Confidential)
MIME-Version: 1.0
Content-Type: text/plain; charset=UTF-8

What's New
Title X <https://ex.com/x>
2024년 06월 03일

Upcoming Launches

주요 업데이트
* 업데이트 X
`
	r := strings.NewReader(rawMail)
	items, updates, subject := ParseMail(r)
	if subject != "AWS Weekly Update (AWS Confidential)" {
		t.Errorf("subject: got %q, want %q", subject, "AWS Weekly Update (AWS Confidential)")
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	if items[0].Title != "Title X" {
		t.Errorf("items[0].Title: got %q, want %q", items[0].Title, "Title X")
	}
	if items[0].Link != "https://ex.com/x" {
		t.Errorf("items[0].Link: got %q, want %q", items[0].Link, "https://ex.com/x")
	}
	if items[0].Date != "2024년 06월 03일" {
		t.Errorf("items[0].Date: got %q, want %q", items[0].Date, "2024년 06월 03일")
	}
	if len(updates) != 1 {
		t.Fatalf("expected 1 update, got %d", len(updates))
	}
	if updates[0] != "* 업데이트 X" {
		t.Errorf("updates[0]: got %q, want %q", updates[0], "* 업데이트 X")
	}
}
