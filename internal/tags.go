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

func GetTags(ctx context.Context, pool *pgxpool.Pool,
	limit, offset int, nameFilter string) (TagsResult, error) {

	const q = `
SELECT t.id,
       t.name,
       COALESCE(s.news_cnt, 0)        AS news_cnt,
       COUNT(*) OVER()                AS total_rows
FROM   tags        AS t
LEFT   JOIN tag_stats AS s ON s.tag_id = t.id
WHERE  ($1 = '' OR t.name ILIKE $1)
ORDER  BY news_cnt DESC, t.name
LIMIT  $2 OFFSET $3;
`

	like := "%"
	if nameFilter != "" {
		like = "%" + nameFilter + "%"
	}

	rows, err := pool.Query(ctx, q, like, limit, offset)
	if err != nil {
		return TagsResult{}, err
	}
	defer rows.Close()

	res := TagsResult{
		Limit:  limit,
		Offset: offset,
		Page:   1,
	}

	for rows.Next() {
		var (
			t         Tag
			totalRows int
		)
		if err := rows.Scan(&t.Id, &t.Name, &t.NewsCount, &totalRows); err != nil {
			return TagsResult{}, err
		}
		if res.Total == 0 { // 첫 행에서 totalRows 확보
			res.Total = totalRows
		}
		res.Items = append(res.Items, t)
	}

	if res.Total > 0 && res.Limit > 0 {
		res.Page = res.Offset/res.Limit + 1
		res.TotalPage = (res.Total + res.Limit - 1) / res.Limit
	} else {
		res.TotalPage = 1
	}

	return res, rows.Err()
}
