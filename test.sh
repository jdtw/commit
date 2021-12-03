#! /bin/bash
set -euxo pipefail

PORT=9090
ADDR="http://localhost:${PORT}"

killall -u ${USER} commit || true

go build -o . ./...

./commit --port "${PORT}" &
until curl -s -X POST "${ADDR}"; do
    echo "Waiting for server to start..."
    sleep 1
done

commit=$(curl -s -X POST "${ADDR}" -d "I know a thing!")
verified=$(curl -s -X POST "${ADDR}/verify" --data-binary "${commit}")
test "${verified}" = "true"

killall -u ${USER} commit
