DROP TABLE IF EXISTS whatsnews_tags CASCADE;
DROP TABLE IF EXISTS tags CASCADE;
DROP TABLE IF EXISTS whatsnews CASCADE;

CREATE TABLE IF NOT EXISTS tags (
  id SERIAL PRIMARY KEY,
  name VARCHAR(32) UNIQUE NOT NULL,
  created_at TIMESTAMP
);

CREATE TABLE IF NOT EXISTS whatsnews (
  id SERIAL PRIMARY KEY,
  title VARCHAR(128) NOT NULL,
  content TEXT,
  source_id VARCHAR(64),
  source_url VARCHAR(1024),
  source_created_at TIMESTAMP,
  created_at TIMESTAMP,
  updated_at TIMESTAMP
);

CREATE TABLE IF NOT EXISTS whatsnews_tags (
  whatsnew_id INTEGER NOT NULL REFERENCES whatsnews(id),
  tag_id INTEGER NOT NULL REFERENCES tags(id),
  created_at TIMESTAMP,
  PRIMARY KEY (whatsnew_id, tag_id)
);
