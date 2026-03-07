#!/usr/bin/env bash

set -e

BASE="http://localhost:8080"

CONCURRENCY=100
REQUESTS=100

LAT_FILE="/tmp/latencies.txt"
ERR_FILE="/tmp/errors.txt"

rm -f $LAT_FILE $ERR_FILE

request() {
  METHOD=$1
  URL=$2
  DATA=$3

  START=$(date +%s%3N)

  if [ -z "$DATA" ]; then
    CODE=$(curl -s -o /dev/null -w "%{http_code}" -X "$METHOD" "$URL")
  else
    CODE=$(curl -s -o /dev/null -w "%{http_code}" \
      -X "$METHOD" "$URL" \
      -H "Content-Type: application/json" \
      -d "$DATA")
  fi

  END=$(date +%s%3N)

  LAT=$((END - START))

  echo $LAT >> $LAT_FILE

  if [[ "$CODE" -ge 400 ]]; then
    echo "$CODE" >> $ERR_FILE
  fi
}

print_stats() {

  TOTAL=$(wc -l < $LAT_FILE)

  if [ -f "$ERR_FILE" ]; then
    ERRORS=$(wc -l < $ERR_FILE)
  else
    ERRORS=0
  fi

  AVG=$(awk '{s+=$1} END {print s/NR}' $LAT_FILE)
  MIN=$(sort -n $LAT_FILE | head -n1)
  MAX=$(sort -n $LAT_FILE | tail -n1)

  P50=$(sort -n $LAT_FILE | awk 'NR==int(NR*0.50)')
  P95=$(sort -n $LAT_FILE | awk 'NR==int(NR*0.95)')
  P99=$(sort -n $LAT_FILE | awk 'NR==int(NR*0.99)')

  RPS=$(awk "BEGIN {print $TOTAL / ($TOTAL_TIME / 1000)}")

  echo "Requests:        $TOTAL"
  echo "Errors:          $ERRORS"
  echo "Concurrency:     $CONCURRENCY"
  echo "Total time:      ${TOTAL_TIME} ms"
  echo "Requests/sec:    $RPS"
  echo

  echo "Latency:"
  echo "  avg: ${AVG} ms"
  echo "  min: ${MIN} ms"
  echo "  p50: ${P50} ms"
  echo "  p95: ${P95} ms"
  echo "  p99: ${P99} ms"
  echo "  max: ${MAX} ms"
}

benchmark() {

  NAME=$1
  METHOD=$2
  URL=$3
  DATA=$4

  echo
  echo "======================================"
  echo "$NAME"
  echo "======================================"

  rm -f $LAT_FILE $ERR_FILE

  START_TOTAL=$(date +%s%3N)

  for i in $(seq 1 $REQUESTS); do
  (
    request "$METHOD" "$URL" "$DATA"
  ) &

  if (( i % CONCURRENCY == 0 )); then
      wait
  fi
  done

  wait

  END_TOTAL=$(date +%s%3N)

  TOTAL_TIME=$((END_TOTAL - START_TOTAL))

  print_stats
}

echo "======================================"
echo "Functional Tests"
echo "======================================"

echo "Creating session"
SESSION=$(curl -s -X POST "$BASE/session/create")
SESSION_ID=$(echo "$SESSION" | jq -r '.session_id')

echo "Session: $SESSION_ID"

echo "Starting session"
curl -s -X POST "$BASE/session/$SESSION_ID/start" > /dev/null

sleep 1

echo "Checking status"
curl -s "$BASE/session/$SESSION_ID/status" | jq

echo "Executing test command"
EXEC=$(curl -s -X POST "$BASE/session/$SESSION_ID/exec" \
  -H "Content-Type: application/json" \
  -d '{"cmd":["echo","hello"]}')

JOB=$(echo "$EXEC" | jq -r '.job_id')

echo "Job ID: $JOB"

echo "Waiting for job"

while true; do
  STATUS=$(curl -s "$BASE/session/$SESSION_ID/job/$JOB")
  STATE=$(echo "$STATUS" | jq -r '.status')

  if [[ "$STATE" == "completed" || "$STATE" == "failed" ]]; then
    echo "$STATUS" | jq
    break
  fi

  sleep 1
done

echo
echo "======================================"
echo "Benchmark Tests"
echo "======================================"

benchmark \
"Create Session Endpoint" \
POST \
"$BASE/session/create"

benchmark \
"Session Status Endpoint" \
GET \
"$BASE/session/$SESSION_ID/status"

benchmark \
"Exec Endpoint" \
POST \
"$BASE/session/$SESSION_ID/exec" \
'{"cmd":["echo","bench"]}'

echo
echo "======================================"
echo "Mixed Workload Test"
echo "======================================"

rm -f $LAT_FILE $ERR_FILE

START_TOTAL=$(date +%s%3N)

for i in $(seq 1 $REQUESTS); do
(
  if (( i % 2 == 0 )); then
    request POST "$BASE/session/$SESSION_ID/exec" '{"cmd":["echo","mixed"]}'
  else
    request GET "$BASE/session/$SESSION_ID/status"
  fi
) &

if (( i % CONCURRENCY == 0 )); then
  wait
fi

done

wait

END_TOTAL=$(date +%s%3N)
TOTAL_TIME=$((END_TOTAL - START_TOTAL))

print_stats

echo
echo "======================================"
echo "Stress Test: Session Creation"
echo "======================================"

benchmark \
"Concurrent Session Creation" \
POST \
"$BASE/session/create"

echo
echo "Stopping session"
curl -s -X POST "$BASE/session/$SESSION_ID/stop" > /dev/null

echo "Deleting session"
curl -s -X DELETE "$BASE/session/$SESSION_ID/" > /dev/null

echo
echo "All tests completed"