# noti-aws-update

## Getting Started

This project lets you collect, track, and query AWS "What's New" updates automatically.
A scheduler periodically fetches new announcements and can also send notifications to Slack.

## Prerequisites

- **Go** (v1.23 or higher)
- **PostgreSQL** (accessible locally or via Docker)
- **Docker & Docker Compose** (if running containerized)
- **IMAP account credentials** (for email updates)
- **Slack webhook URL** (if you want update notifications delivered to Slack)
- (Optional) **Node.js** (`collect/` ETL scripts)
- (Recommended) Linux/Unix-like environment for `run.sh`

## Installation

1. **Clone the repository:**
   ```bash
   git clone https://github.krafton.com/ops2022/noti-aws-update.git
   cd noti-aws-update
   ```

2. **Prepare configuration:**
   Copy `.env.example` to `.env` and fill out the following:
   ```
   IMAP_SERVER=imap.example.com:993
   IMAP_USER=exampleuser@example.com
   IMAP_PASSWORD=your_imap_password
   DATABASE_USER=user
   DATABASE_PASSWORD=password
   DATABASE_DB=dbname
   DATABASE_HOST=localhost
   DATABASE_PORT=5432
   PGADMIN_DEFAULT_EMAIL=admin@admin.com
   PGADMIN_DEFAULT_PASSWORD=admin
   APP_PORT=8000
   ```

3. **Install Go dependencies:**
   ```bash
   go mod tidy
   ```

4. **(Optional) If you use collect scripts:**
   ```bash
   cd collect
   npm install
   cd ..
   ```

## Usage

### 1. Interactive Launch Script

Run:
```bash
./run.sh
```
And select:
1. Go build/run (scheduler + API locally)
2. Docker Compose Up (all containers: scheduler, HTTP server, DB, pgAdmin)
3. Go build/run + init DB (local run after DB init)
4. Docker Compose Up + init DB (fresh DB & all services)

### 2. Manual Operation

- **Build and run:**
  ```bash
  go build -o ./build/myapp ./cmd/httpserver
  go build -o ./build/scheduler ./cmd/scheduler
  ./build/scheduler &
  ./build/myapp
  ```
- **(Optional) Initialize DB:**
  ```bash
  psql -h <db_host> -U <db_user> -d <db_name> -f ./initdb/init.sql
  ```
- **(Optional) Run ETL scripts:**
  ```bash
  (cd collect && npm run start)
  ```

### 3. Docker Compose

```bash
docker-compose down -v
docker-compose --env-file ./.env up --build
```
- API: [http://localhost:8000](http://localhost:8000)

## API Endpoints

- `GET /health` — Health check
- `GET /api/tags` — List tags (with pagination/name filter)
- `GET /api/whatsnews` — List news (filter by tag IDs: `?tags=1,2`)

## Database Schema

See [`initdb/init.sql`](./initdb/init.sql).
Main tables: `whatsnews`, `tags`, `whatsnews_tags`.

## Branching & Git Workflow

### Branches

- **main**: The main, production-ready branch.
- **develop**: The primary branch for new development, features, and testing.

### Typical Workflow

1. **Clone the repository:**
   ```bash
   git clone https://github.krafton.com/ops2022/noti-aws-update.git
   ```
2. **Switch to the develop branch:**
   ```bash
   git pull origin develop
   git switch develop
   ```
3. Do your work and commit/push to `develop`.

4. **Merge to `main` when ready for production:**
   ```bash
   git switch main
   git merge develop
   git push origin main
   ```
