package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.krafton.com/ops2022/noti-aws-update/internal"
)

type AwsItem struct {
	Id               string         `json:"id"`
	AdditionalFields map[string]any `json:"additionalFields"`
}
type AwsTag struct {
	Name string `json:"name"`
}
type AwsApiItem struct {
	Item AwsItem  `json:"item"`
	Tags []AwsTag `json:"tags"`
}
type AwsApiResponse struct {
	Items    []AwsApiItem `json:"items"`
	Metadata struct {
		Count int `json:"count"`
	} `json:"metadata"`
}

func SourceIdExists(ctx context.Context, pool *pgxpool.Pool, sourceId string) (bool, error) {
	conn, err := pool.Acquire(ctx)
	if err != nil {
		return false, err
	}
	defer conn.Release()

	var exists bool
	err = conn.QueryRow(ctx, "SELECT EXISTS (SELECT 1 FROM whatsnews WHERE source_id = $1)", sourceId).Scan(&exists)
	return exists, err
}

func InsertAwsItem(ctx context.Context, pool *pgxpool.Pool, el AwsApiItem) error {
	conn, err := pool.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()

	tx, err := conn.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var (
		title, _         = el.Item.AdditionalFields["headline"].(string)
		body, _          = el.Item.AdditionalFields["postBody"].(string)
		url, _           = el.Item.AdditionalFields["headlineUrl"].(string)
		sourceTimeStr, _ = el.Item.AdditionalFields["postDateTime"].(string)
	)
	var sourceTime *time.Time
	if sourceTimeStr != "" {
		t, err := time.Parse(time.RFC3339, sourceTimeStr)
		if err == nil {
			sourceTime = &t
		}
	}

	var whatsnewsID int
	err = tx.QueryRow(ctx,
		`INSERT INTO whatsnews(title, content, source_id, source_url, source_created_at, created_at, updated_at)
         VALUES($1, $2, $3, $4, $5, NOW(), NOW())
         ON CONFLICT (source_id) DO NOTHING
         RETURNING id`,
		title, body, el.Item.Id, url, sourceTime,
	).Scan(&whatsnewsID)
	if errors.Is(err, pgx.ErrNoRows) {
		err = tx.QueryRow(ctx, "SELECT id FROM whatsnews WHERE source_id=$1", el.Item.Id).Scan(&whatsnewsID)
	}
	if err != nil {
		return fmt.Errorf("insert/select whatsnews: %w", err)
	}

	for _, tag := range el.Tags {
		var tagID int
		err = tx.QueryRow(ctx,
			`INSERT INTO tags(name, created_at) VALUES($1, NOW())
             ON CONFLICT (name) DO NOTHING RETURNING id`, tag.Name).Scan(&tagID)
		if errors.Is(err, pgx.ErrNoRows) {
			err = tx.QueryRow(ctx, "SELECT id FROM tags WHERE name=$1", tag.Name).Scan(&tagID)
		}
		if err != nil {
			return fmt.Errorf("insert/select tag: %w (name=%s)", err, tag.Name)
		}
		_, err = tx.Exec(ctx,
			`INSERT INTO whatsnews_tags(whatsnew_id, tag_id, created_at)
             VALUES($1, $2, NOW()) ON CONFLICT DO NOTHING`, whatsnewsID, tagID)
		if err != nil {
			return fmt.Errorf("insert whatsnews_tags: %w", err)
		}
	}
	return tx.Commit(ctx)
}

func ParseUntilExisting(ctx context.Context, pool *pgxpool.Pool, pageSize int) error {
	directoryID := "whats-new-v2"
	baseUrl := "https://aws.amazon.com/api/dirs/items/search"
	page := 0
	for {
		reqUrl := fmt.Sprintf("%s?item.directoryId=%s&sort_by=item.additionalFields.postDateTime&sort_order=desc&size=%d&page=%d&item.locale=en_US",
			baseUrl, directoryID, pageSize, page)

		req, err := http.NewRequestWithContext(ctx, "GET", reqUrl, nil)
		if err != nil {
			return err
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return err
		}
		if resp.StatusCode != 200 {
			b, _ := io.ReadAll(resp.Body)
			return fmt.Errorf("api error: %v, %s", resp.Status, string(b))
		}

		var apiResp AwsApiResponse
		if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
			resp.Body.Close()
			return err
		}
		resp.Body.Close()

		if len(apiResp.Items) == 0 {
			break
		}

		for _, el := range apiResp.Items {
			found, err := SourceIdExists(ctx, pool, el.Item.Id)
			if err != nil {
				return err
			}
			if found {
				log.Printf("source_id %s already exists; stop", el.Item.Id)
				return nil
			}
			if err := InsertAwsItem(ctx, pool, el); err != nil {
				log.Printf("Failed insert %s: %v", el.Item.Id, err)
				return err
			}
			log.Printf("Inserted source_id %s, headline='%s'", el.Item.Id, el.Item.AdditionalFields["headline"])
		}
		if len(apiResp.Items) < pageSize {
			break
		}
		page++
	}
	return nil
}

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

func SendToSlack(webhookURL, message string) error {
	payload := map[string]string{"text": message}
	body, _ := json.Marshal(payload)
	resp, err := http.Post(webhookURL, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("non-200 response from Slack: %d", resp.StatusCode)
	}
	return nil
}

func main() {
	cfg := internal.LoadConfig()
	ctx := context.Background()
	pool, err := internal.NewDBPool(cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()

	for {
		var readers []io.Reader

		if cfg.Mode == internal.ModeIMAP {
			c, err := internal.ConnectIMAP(cfg)
			if err != nil {
				log.Printf("IMAP connection error: %v", err)
				return
			}
			defer c.Logout()

			readers, err = internal.FetchMailReadersFromIMAP(c)
			if err != nil {
				log.Printf("Failed to fetch mail: %v", err)
				return
			}

		} else if cfg.Mode == internal.ModeTestdata {
			readers, err = fetchMailReadersFromDir(cfg.TestdataDir)
		}

		for _, r := range readers {
			newsItems, updates, subject := internal.ParseMail(r)
			printMailSummary(subject, newsItems, updates)

			var message strings.Builder
			// message.WriteString(fmt.Sprintf("*%s*\n", subject))
			// for _, item := range newsItems {
			// 	message.WriteString(fmt.Sprintf("- %s (%s)\n%s\n", item.Title, item.Date, item.Link))
			// }
			if len(updates) > 0 {
				message.WriteString("\nUpdates:\n" + strings.Join(updates, "\n"))
			}

			webhookURL := cfg.SlackWebHookUrl
			if webhookURL != "" {
				if err := SendToSlack(webhookURL, message.String()); err != nil {
					log.Printf("Slack notify failed: %v", err)
				}
			}
		}

		if err := ParseUntilExisting(ctx, pool, 100); err != nil {
			log.Printf("ParseUntilExisting 에러: %v", err)
		}
		<-ticker.C
	}
}
