package internal

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
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

func GetWhatsnews(
	ctx context.Context,
	pool *pgxpool.Pool,
	limit, offset int,
	tagIDs []int,
	search string,
) (WhatsNewsResult, error) {

	if limit <= 0 {
		limit = 30
	}
	if offset < 0 {
		offset = 0
	}

	var (
		conds   []string
		args    []any
		paramNo = 1
	)

	// 검색어
	if strings.TrimSpace(search) != "" {
		conds = append(
			conds,
			"(wn.title ILIKE $"+strconv.Itoa(paramNo)+" OR wn.content ILIKE $"+strconv.Itoa(paramNo)+")",
		)
		args = append(args, "%"+search+"%")
		paramNo++
	}

	// 태그 배열
	hasTags := len(tagIDs) > 0
	tagParamNo := 0
	if hasTags {
		tagParamNo = paramNo
		args = append(args, tagIDs) // $tagParamNo
		paramNo++
	}

	buildWhere := func(extra string) string {
		all := make([]string, len(conds))
		copy(all, conds)
		if extra != "" {
			all = append(all, extra)
		}
		if len(all) == 0 {
			return ""
		}
		return "WHERE " + strings.Join(all, " AND ")
	}

	var (
		countSQL string
		dataSQL  string
	)

	limitParam := paramNo
	offsetParam := paramNo + 1
	args = append(args, limit, offset)

	if hasTags {
		candidateCond := "wn.id IN (SELECT whatsnew_id FROM candidates)"

		countSQL = `
WITH wanted AS (
  SELECT unnest($` + strconv.Itoa(tagParamNo) + `::int[]) AS tag_id
), candidates AS (
  SELECT wnt.whatsnew_id
  FROM   whatsnews_tags wnt
  JOIN   wanted w ON w.tag_id = wnt.tag_id
  GROUP  BY wnt.whatsnew_id
  HAVING COUNT(DISTINCT w.tag_id) = (SELECT COUNT(*) FROM wanted)
)
SELECT COUNT(*)
FROM   whatsnews wn
` + buildWhere(candidateCond) + `;
`

		dataSQL = `
WITH wanted AS (
  SELECT unnest($` + strconv.Itoa(tagParamNo) + `::int[]) AS tag_id
), candidates AS (
  SELECT wnt.whatsnew_id
  FROM   whatsnews_tags wnt
  JOIN   wanted w ON w.tag_id = wnt.tag_id
  GROUP  BY wnt.whatsnew_id
  HAVING COUNT(DISTINCT w.tag_id) = (SELECT COUNT(*) FROM wanted)
), filtered AS (
  SELECT  wn.id, wn.title, wn.content, wn.source_url, wn.source_created_at
  FROM    whatsnews wn
  ` + buildWhere(candidateCond) + `
  ORDER BY wn.source_created_at DESC, wn.id
  LIMIT   $` + strconv.Itoa(limitParam) + ` OFFSET $` + strconv.Itoa(offsetParam) + `
)
SELECT f.id, f.title, f.content, f.source_url, f.source_created_at,
       COALESCE(t.tags,'[]') AS tags
FROM   filtered f
LEFT JOIN LATERAL (
  SELECT json_agg(
           jsonb_build_object('id',t.id,'name',t.name)
           ORDER BY t.name
         ) AS tags
  FROM   whatsnews_tags wnt2
  JOIN   tags t ON t.id = wnt2.tag_id
  WHERE  wnt2.whatsnew_id = f.id
) t ON TRUE
ORDER BY f.source_created_at DESC, f.id;
`
	} else { // 태그 선택이 없을 때
		countSQL = `
SELECT COUNT(*) FROM whatsnews wn
` + buildWhere("") + `;
`

		dataSQL = `
WITH filtered AS (
  SELECT  wn.id, wn.title, wn.content, wn.source_url, wn.source_created_at
  FROM    whatsnews wn
  ` + buildWhere("") + `
  ORDER BY wn.source_created_at DESC, wn.id
  LIMIT   $` + strconv.Itoa(limitParam) + ` OFFSET $` + strconv.Itoa(offsetParam) + `
)
SELECT f.id, f.title, f.content, f.source_url, f.source_created_at,
       COALESCE(t.tags,'[]') AS tags
FROM   filtered f
LEFT JOIN LATERAL (
  SELECT json_agg(
           jsonb_build_object('id',t.id,'name',t.name)
           ORDER BY t.name
         ) AS tags
  FROM   whatsnews_tags wnt2
  JOIN   tags t ON t.id = wnt2.tag_id
  WHERE  wnt2.whatsnew_id = f.id
) t ON TRUE
ORDER BY f.source_created_at DESC, f.id;
`
	}

	var total int
	if err := pool.QueryRow(ctx, countSQL, args[:limitParam-1]...).Scan(&total); err != nil {
		return WhatsNewsResult{}, err
	}

	rows, err := pool.Query(ctx, dataSQL, args...)
	if err != nil {
		return WhatsNewsResult{}, err
	}
	defer rows.Close()

	var items []WhatsNews
	for rows.Next() {
		var it WhatsNews
		var tagsJSON []byte
		if err := rows.Scan(&it.Id, &it.Title, &it.Content, &it.SourceUrl, &it.SourceCreatedAt, &tagsJSON); err != nil {
			return WhatsNewsResult{}, err
		}
		if err := json.Unmarshal(tagsJSON, &it.Tags); err != nil {
			return WhatsNewsResult{}, err
		}
		items = append(items, it)
	}
	if err := rows.Err(); err != nil {
		return WhatsNewsResult{}, err
	}

	page := offset/limit + 1
	totalPage := (total + limit - 1) / limit
	if totalPage == 0 {
		totalPage = 1
	}

	return WhatsNewsResult{
		Items:     items,
		Total:     total,
		Limit:     limit,
		Offset:    offset,
		Page:      page,
		TotalPage: totalPage,
	}, nil
}
