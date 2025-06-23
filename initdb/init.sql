DROP TABLE IF EXISTS whatsnews_tags CASCADE;
DROP TABLE IF EXISTS tags CASCADE;
DROP TABLE IF EXISTS whatsnews CASCADE;

CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE TABLE IF NOT EXISTS tags (
  id SERIAL PRIMARY KEY,
  name VARCHAR(128) UNIQUE NOT NULL,
  created_at TIMESTAMP
);

CREATE TABLE IF NOT EXISTS whatsnews (
  id SERIAL PRIMARY KEY,
  title VARCHAR(512) NOT NULL,
  content TEXT,
  source_id VARCHAR(256) UNIQUE NOT NULL,
  source_url VARCHAR(1024),
  source_created_at TIMESTAMP,
  created_at TIMESTAMP NOT NULL DEFAULT now(),
  updated_at TIMESTAMP NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS whatsnews_tags (
  whatsnew_id INTEGER NOT NULL REFERENCES whatsnews(id) ON DELETE CASCADE,
  tag_id INTEGER NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
  created_at TIMESTAMP NOT NULL DEFAULT now(),
  PRIMARY KEY (whatsnew_id, tag_id)
);


CREATE INDEX IF NOT EXISTS idx_whatsnews_source_created_at ON whatsnews (source_created_at DESC);

CREATE INDEX IF NOT EXISTS idx_whatsnews_tags_tag_id ON whatsnews_tags (tag_id);
CREATE INDEX IF NOT EXISTS idx_whatsnews_tags_whatsnew_id ON whatsnews_tags (whatsnew_id);
CREATE INDEX IF NOT EXISTS idx_whatsnews_tags_tag_whatsnew ON whatsnews_tags (tag_id, whatsnew_id);

CREATE INDEX IF NOT EXISTS idx_tags_name_trgm ON tags USING gin (name gin_trgm_ops);
CREATE INDEX IF NOT EXISTS idx_whatsnews_title_trgm ON whatsnews USING gin (title gin_trgm_ops);
CREATE INDEX IF NOT EXISTS idx_whatsnews_content_trgm ON whatsnews USING gin (content gin_trgm_ops);
