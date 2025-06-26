#!/bin/sh

DEFAULT_WORKERS=1
DEFAULT_HOST=http://localhost:8089

WORKERS="${1:-$DEFAULT_WORKERS}"
HOST="${2:-$DEFAULT_HOST}"

cd "$(dirname "$0")" || exit

if [ -d "venv" ]; then
    . ./venv/bin/activate
    echo "🚀  Locust master 시작 (UI → http://localhost:8089)"
    locust -f ./run.py --master --expect-workers "$WORKERS" --host="$HOST" &
    MASTER_PID=$!
    echo "🚀  Worker $WORKERS개 기동"
    for _ in $(seq "$WORKERS"); do
        locust -f ./run.py --worker --master-host localhost &
    done
    trap 'echo -e "\n🛑  클러스터 종료"; kill $MASTER_PID $(jobs -pr) 2>/dev/null' INT TERM
    wait $MASTER_PID
fi
