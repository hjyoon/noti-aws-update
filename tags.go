package main

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Tag struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
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
		total int
		args  []any
	)
	queryCount := `SELECT COUNT(*) FROM tags`
	queryData := `SELECT id, name FROM tags`
	where := ""
	if nameFilter != "" {
		where = " WHERE name ILIKE $1"
		args = append(args, "%"+nameFilter+"%")
	}
	queryCount += where
	queryData += where + ` ORDER BY name LIMIT $2 OFFSET $3`

	argsForData := args
	if nameFilter != "" {
		argsForData = append(argsForData, limit, offset)
	} else {
		argsForData = append(argsForData, limit, offset)
	}

	if nameFilter != "" {
		if err := pool.QueryRow(ctx, queryCount, args[0]).Scan(&total); err != nil {
			return TagsResult{}, err
		}
	} else {
		if err := pool.QueryRow(ctx, queryCount).Scan(&total); err != nil {
			return TagsResult{}, err
		}
	}

	rows, err := pool.Query(ctx, queryData, argsForData...)
	if err != nil {
		return TagsResult{}, err
	}
	defer rows.Close()

	tags := []Tag{}
	for rows.Next() {
		var t Tag
		if err := rows.Scan(&t.Id, &t.Name); err != nil {
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
