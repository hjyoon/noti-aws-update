package main

import (
	"context"
	"encoding/json"
	"time"

	"github.com/jackc/pgx/v5"
)

type WhatsNews struct {
	Id              int        `json:"id"`
	Title           string     `json:"title"`
	Content         string     `json:"content"`
	SourceUrl       string     `json:"source_url"`
	SourceCreatedAt *time.Time `json:"source_created_at"`
	Tags            []Tag      `json:"tags"`
}

type WhatsNewsResult struct {
	Items     []WhatsNews `json:"items"`
	Total     int         `json:"total"`
	Limit     int         `json:"limit"`
	Offset    int         `json:"offset"`
	Page      int         `json:"page"`
	TotalPage int         `json:"total_page"`
}

func GetWhatsnews(ctx context.Context, conn *pgx.Conn, limit, offset int, tagIDs []int) (WhatsNewsResult, error) {
	var (
		total int
		args  []any
	)
	filterWhere := ""
	filterHaving := ""

	if len(tagIDs) > 0 {
		filterWhere = "WHERE wnt.tag_id = ANY($1)"
		filterHaving = "HAVING COUNT(DISTINCT wnt.tag_id) = $2"
		args = append(args, tagIDs, len(tagIDs))
	}
	queryCount := `
SELECT COUNT(*) FROM (
  SELECT wn.id
  FROM whatsnews wn
  JOIN whatsnews_tags wnt ON wn.id = wnt.whatsnew_id
  ` + filterWhere + `
  GROUP BY wn.id
  ` + filterHaving + `
) sub
`
	queryData := `
SELECT wn.id, wn.title, wn.content, wn.source_url, wn.source_created_at,
  COALESCE(
    json_agg(
      DISTINCT jsonb_build_object('id', t2.id, 'name', t2.name)
    ) FILTER (WHERE t2.id IS NOT NULL), '[]'
  ) AS tags
FROM (
    SELECT wn.id
    FROM whatsnews wn
    JOIN whatsnews_tags wnt ON wn.id = wnt.whatsnew_id
    ` + filterWhere + `
    GROUP BY wn.id
    ` + filterHaving + `
    ORDER BY MAX(wn.source_created_at) DESC, wn.title
    LIMIT $3 OFFSET $4
) filtered
JOIN whatsnews wn ON wn.id = filtered.id
LEFT JOIN whatsnews_tags wnt2 ON wn.id = wnt2.whatsnew_id
LEFT JOIN tags t2 ON t2.id = wnt2.tag_id
GROUP BY wn.id, wn.title, wn.content, wn.source_url, wn.source_created_at
ORDER BY wn.source_created_at DESC, wn.title
`
	argsForData := append(args, limit, offset)
	argsForCount := args

	if err := conn.QueryRow(ctx, queryCount, argsForCount...).Scan(&total); err != nil {
		return WhatsNewsResult{}, err
	}

	rows, err := conn.Query(ctx, queryData, argsForData...)
	if err != nil {
		return WhatsNewsResult{}, err
	}
	defer rows.Close()

	var items []WhatsNews
	for rows.Next() {
		var item WhatsNews
		var tagsJson []byte

		err := rows.Scan(
			&item.Id, &item.Title, &item.Content,
			&item.SourceUrl,
			&item.SourceCreatedAt,
			&tagsJson,
		)
		if err != nil {
			return WhatsNewsResult{}, err
		}
		if err := json.Unmarshal(tagsJson, &item.Tags); err != nil {
			return WhatsNewsResult{}, err
		}
		items = append(items, item)
	}

	page := 1
	if limit > 0 {
		page = offset/limit + 1
	}
	totalPage := 1
	if limit > 0 {
		totalPage = (total + limit - 1) / limit
	}
	return WhatsNewsResult{
		Items:     items,
		Total:     total,
		Limit:     limit,
		Offset:    offset,
		Page:      page,
		TotalPage: totalPage,
	}, rows.Err()
}
