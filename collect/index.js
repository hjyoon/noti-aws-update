import pkg from "pg";
const { Pool } = pkg;

const pool = new Pool({
  user: "ghost",
  host: "localhost",
  database: "appdb",
  password: "ghost",
  port: 5432,
});

function decorrelatedJitter(baseDelay, maxDelay, previousDelay) {
  if (!previousDelay) {
    previousDelay = baseDelay;
  }
  return Math.min(
    maxDelay,
    Math.random() * (previousDelay * 3 - baseDelay) + baseDelay,
  );
}

async function transactionCustom(item, tags, maxRetry = 3) {
  let attempts = 0;
  let delay = null;
  while (attempts < maxRetry) {
    const client = await pool.connect();
    try {
      await client.query("BEGIN");
      const whatsnews_res = await client.query(
        "INSERT INTO whatsnews(title, content, source_id, source_url, source_created_at, created_at, updated_at) VALUES($1, $2, $3, $4, $5, NOW(), NOW()) ON CONFLICT (source_id) DO NOTHING RETURNING id",
        [
          item.additionalFields.headline,
          item.additionalFields.postBody,
          item.id,
          item.additionalFields.headlineUrl,
          item.additionalFields.postDateTime,
        ],
      );

      let whatsnewsId;
      if (whatsnews_res.rows.length > 0) {
        whatsnewsId = whatsnews_res.rows[0].id;
      } else {
        const oldNews = await client.query(
          "SELECT id FROM whatsnews WHERE source_id = $1",
          [item.id],
        );
        whatsnewsId = oldNews.rows[0].id;
      }

      for (let j = 0; j < tags.length; j++) {
        const tags_res = await client.query(
          "INSERT INTO tags(name, created_at) VALUES($1, NOW()) ON CONFLICT (name) DO NOTHING RETURNING id",
          [tags[j].name],
        );

        let tagId;
        if (tags_res.rows.length > 0) {
          tagId = tags_res.rows[0].id;
        } else {
          const oldTag = await client.query(
            "SELECT id FROM tags WHERE name = $1",
            [tags[j].name],
          );
          tagId = oldTag.rows[0].id;
        }
        await client.query(
          "INSERT INTO whatsnews_tags(whatsnew_id, tag_id, created_at) VALUES($1, $2, NOW()) ON CONFLICT(whatsnew_id, tag_id) DO NOTHING",
          [whatsnewsId, tagId],
        );
      }
      await client.query("COMMIT");
      client.release();
      break;
    } catch (err) {
      console.error("Transaction rolled back", err);
      await client.query("ROLLBACK");
      client.release();
      if (err.code === "40P01") {
        attempts++;
        console.warn(`Deadlock detected. Retrying (${attempts}/${maxRetry})`);
        if (attempts >= maxRetry) throw err;
        await new Promise((resolve) => {
          {
            delay = decorrelatedJitter(100, 60000, delay);
            setTimeout(resolve, delay);
          }
        });
      } else {
        throw err;
      }
    }
  }
}

async function fetch_and_save(apiUrl) {
  const response = await fetch(apiUrl);
  const data = await response.json();
  if (data.metadata.count === 0) {
    return;
  }
  const promises = [];
  for (let i = 0; i < data.items.length; i++) {
    const { item, tags } = data.items[i];
    promises.push(transactionCustom(item, tags, 10));
  }
  await Promise.all(promises);
}

async function getTotalSize(apiUrl) {
  const response = await fetch(apiUrl);
  const data = await response.json();
  return data.metadata.totalHits;
}

async function fetch_group(directory_id, tags_id) {
  let apiUrl = new URL(BASE_URL);
  apiUrl.searchParams.append("item.directoryId", directory_id);
  apiUrl.searchParams.append("tags.id", tags_id);
  apiUrl.searchParams.append("size", "1");
  apiUrl.searchParams.append("item.locale", "en_US");

  const totalSize = await getTotalSize(apiUrl);
  const promises = [];
  for (let i = 0; i * BATCH_SIZE < totalSize; i++) {
    apiUrl = new URL(BASE_URL);
    apiUrl.searchParams.append("item.directoryId", directory_id);
    apiUrl.searchParams.append("tags.id", tags_id);
    apiUrl.searchParams.append("sort_by", "item.additionalFields.postDateTime");
    apiUrl.searchParams.append("sort_order", "desc");
    apiUrl.searchParams.append("size", "" + BATCH_SIZE);
    apiUrl.searchParams.append("page", "" + i);
    apiUrl.searchParams.append("item.locale", "en_US");
    promises.push(fetch_and_save(apiUrl));
    console.log(`Processing: ${i * BATCH_SIZE} of ${totalSize}`);
  }
  await Promise.all(promises);
  console.log(`OK\n` + `tags_id: ${tags_id}\n` + `rows inserted: ${totalSize}`);
}

const promises = [];

const BASE_URL = "https://aws.amazon.com/api/dirs/items/search";
const BATCH_SIZE = 100;

const groups = [
  ["whats-new-v2", "whats-new-v2#year#2025"],
  ["whats-new-v2", "whats-new-v2#year#2024"],
  ["whats-new-v2", "whats-new-v2#year#2023"],
  ["whats-new-v2", "whats-new-v2#year#2022"],
  ["whats-new-v2", "whats-new-v2#year#2021"],
  ["whats-new-v2", "whats-new-v2#year#2020"],
  ["whats-new-v2", "whats-new-v2#year#2019"],
  ["whats-new-v2", "whats-new-v2#year#2018"],
  ["whats-new-v2", "whats-new-v2#year#2017"],
  ["whats-new-v2", "whats-new-v2#year#2016"],
  ["whats-new-v2", "whats-new-v2#year#2015"],
  ["whats-new-v2", "whats-new-v2#year#2014"],
  ["whats-new-v2", "whats-new-v2#year#2013"],
  ["whats-new-v2", "whats-new-v2#year#2012"],
  ["whats-new-v2", "whats-new-v2#year#2011"],
  ["whats-new-v2", "whats-new-v2#year#2010"],
  ["whats-new-v2", "whats-new-v2#year#2009"],
  ["whats-new-v2", "whats-new-v2#year#2008"],
  ["whats-new-v2", "whats-new-v2#year#2007"],
  ["whats-new-v2", "whats-new-v2#year#2006"],
  ["whats-new-v2", "whats-new-v2#year#2005"],
  ["whats-new-v2", "whats-new-v2#year#2004"],
];

for (const [directory_id, tags_id] of groups) {
  promises.push(fetch_group(directory_id, tags_id));
}

await Promise.all(promises);
await pool.end();
console.log("All Completed");
