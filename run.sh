#!/bin/sh

set -e

echo "실행 방식 선택:"
echo "1) Go build/run"
echo "2) Docker Compose Up"
echo "3) Go build/run + init DB"
echo "4) Docker Compose Up + init DB"
printf "번호 입력: "
read -r opt

case "$opt" in
  1)
    go mod tidy
    go build -o ./build/myapp
    ./build/myapp
    ;;
  2)
    docker-compose down
    docker-compose --env-file ./.env up myapp pgadmin --build
    ;;
  3)
    . ./.env
    PGPASSWORD="$DATABASE_PASSWORD" psql -h "$DATABASE_HOST" -U "$DATABASE_USER" -d "$DATABASE_DB" -f ./initdb/init.sql
    (cd collect/ && npm run start)
    go mod tidy
    go build -o ./build/myapp
    ./build/myapp
    ;;
  4)
    docker-compose down -v
    docker-compose --env-file ./.env up --build
    ;;
  *)
    echo "올바른 번호를 선택하세요."
    ;;
esac
