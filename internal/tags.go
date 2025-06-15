package internal

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Tag struct {
	Id        int    `json:"id"`
	Name      string `json:"name"`
	NewsCount int    `json:"news_count"`
}

type TagsResult struct {
	Items     []Tag `json:"items"`
	Total     int   `json:"total"`
	Limit     int   `json:"limit"`
	Offset    int   `json:"offset"`
	Page      int   `json:"page"`
	TotalPage int   `json:"total_page"`
}

func GetTags(ctx context.Context, pool *pgxpool.Pool, limit, offset int, nameFilter string) (TagsResult, error) {
	var (
		total      int
		queryCount string
		queryData  string
		argsCount  []any
		argsData   []any
	)

	if nameFilter != "" {
		queryCount = `SELECT COUNT(*) FROM tags WHERE name ILIKE $1`
		queryData = `
            SELECT tags.id, tags.name, COUNT(wnt.whatsnew_id) as news_count
            FROM tags
            LEFT JOIN whatsnews_tags wnt ON tags.id = wnt.tag_id
            WHERE tags.name ILIKE $1
            GROUP BY tags.id, tags.name
            ORDER BY news_count DESC, tags.name
            LIMIT $2 OFFSET $3
        `
		likeName := "%" + nameFilter + "%"
		argsCount = []any{likeName}
		argsData = []any{likeName, limit, offset}
	} else {
		queryCount = `SELECT COUNT(*) FROM tags`
		queryData = `
            SELECT tags.id, tags.name, COUNT(wnt.whatsnew_id) as news_count
            FROM tags
            LEFT JOIN whatsnews_tags wnt ON tags.id = wnt.tag_id
            GROUP BY tags.id, tags.name
            ORDER BY news_count DESC, tags.name
            LIMIT $1 OFFSET $2
        `
		argsCount = []any{}
		argsData = []any{limit, offset}
	}

	if err := pool.QueryRow(ctx, queryCount, argsCount...).Scan(&total); err != nil {
		return TagsResult{}, err
	}

	rows, err := pool.Query(ctx, queryData, argsData...)
	if err != nil {
		return TagsResult{}, err
	}
	defer rows.Close()

	tags := []Tag{}
	for rows.Next() {
		var t Tag
		if err := rows.Scan(&t.Id, &t.Name, &t.NewsCount); err != nil {
			return TagsResult{}, err
		}
		tags = append(tags, t)
	}

	page := 1
	if limit > 0 {
		page = offset/limit + 1
	}
	totalPage := 1
	if limit > 0 {
		totalPage = (total + limit - 1) / limit
	}

	return TagsResult{
		Items:     tags,
		Total:     total,
		Limit:     limit,
		Offset:    offset,
		Page:      page,
		TotalPage: totalPage,
	}, rows.Err()
}
