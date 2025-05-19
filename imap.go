package main

import (
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
)

func connectIMAP(cfg Config) (*client.Client, error) {
	log.Println("Connecting:", cfg.ImapServer)

	c, err := client.DialTLS(cfg.ImapServer, nil)
	if err != nil {
		return nil, fmt.Errorf("dial: %w", err)
	}
	if err := c.Login(cfg.ImapUser, cfg.ImapPassword); err != nil {
		_ = c.Logout()
		return nil, fmt.Errorf("login: %w", err)
	}
	return c, nil
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

	crit := imap.NewSearchCriteria()
	crit.WithoutFlags = []string{imap.SeenFlag}
	crit.Header.Add("Subject", SubjectFilter)

	ids, err := c.Search(crit)
	if err != nil {
		log.Println("Search failed:", err)
		return nil, err
	}
	if len(ids) == 0 {
		log.Println("No unseen messages")
		return nil, nil
	}

	seq := new(imap.SeqSet)
	seq.AddNum(ids...)
	section := &imap.BodySectionName{}
	items := []imap.FetchItem{section.FetchItem()}

	msgCh := make(chan *imap.Message, len(ids))
	done := make(chan error, 1)
	go func() { done <- c.Fetch(seq, items, msgCh) }()

	var readers []io.Reader
	for msg := range msgCh {
		if msg == nil {
			continue
		}
		body := msg.GetBody(section)
		if body == nil {
			log.Println("Message body is nil")
			continue
		}
		raw, err := io.ReadAll(body)
		if err != nil {
			log.Println("Failed to read message body:", err)
			continue
		}
		readers = append(readers, strings.NewReader(string(raw)))
	}

	if err := <-done; err != nil {
		fmt.Println("Failed to fetch messages:", err)
		return nil, err
	}

	return readers, nil
}
