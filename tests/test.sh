#!/usr/bin/env bash

set -e

BASE="http://localhost:8080"

request() {
  METHOD=$1
  URL=$2
  DATA=$3

  START=$(date +%s%3N)

  if [ -z "$DATA" ]; then
    RESP=$(curl -s -w "\n%{http_code}" -X "$METHOD" "$URL")
  else
    RESP=$(curl -s -w "\n%{http_code}" -X "$METHOD" "$URL" \
      -H "Content-Type: application/json" \
      -d "$DATA")
  fi

  END=$(date +%s%3N)
  DURATION=$((END-START))

  BODY=$(echo "$RESP" | head -n -1)
  CODE=$(echo "$RESP" | tail -n1)

  echo "HTTP $CODE (${DURATION}ms)" >&2

  if [ "$CODE" -ge 400 ]; then
    echo "Request failed ($CODE)" >&2
  fi

  echo "$BODY"
}

echo "Creating session"
CREATE=$(request POST "$BASE/session/create")
echo "$CREATE"

SESSION_ID=$(echo "$CREATE" | jq -r '.session_id')

echo "Session ID: $SESSION_ID"
echo

echo "Starting session"
request POST "$BASE/session/$SESSION_ID/start" | jq
echo

echo "Checking session status"
request GET "$BASE/session/$SESSION_ID/status" | jq
echo

echo "Executing multiple commands"

EXEC1=$(request POST "$BASE/session/$SESSION_ID/exec" '{"cmd":["echo","hello"]}')
echo "$EXEC1"
JOB1=$(echo "$EXEC1" | jq -r '.job_id')

EXEC2=$(request POST "$BASE/session/$SESSION_ID/exec" '{"cmd":["uname","-a"]}')
echo "$EXEC2"
JOB2=$(echo "$EXEC2" | jq -r '.job_id')

echo "Job1: $JOB1"
echo "Job2: $JOB2"
echo

echo "Polling job status"

for JOB in $JOB1 $JOB2; do
  while true; do
    STATUS=$(curl -s "$BASE/session/$SESSION_ID/job/$JOB")
    STATE=$(echo "$STATUS" | jq -r '.status')

    echo "$STATUS" | jq

    if [[ "$STATE" == "completed" || "$STATE" == "failed" ]]; then
      break
    fi

    sleep 1
  done
done

echo
echo "Testing invalid request (missing cmd)"
request POST "$BASE/session/$SESSION_ID/exec" '{}' | jq
echo

echo "Concurrent sessions test"

for i in {1..3}; do
  curl -s -X POST "$BASE/session/create" | jq &
done

wait

echo
echo "Parallel job test"

for i in {1..5}; do
  curl -s -X POST "$BASE/session/$SESSION_ID/exec" \
    -H "Content-Type: application/json" \
    -d '{"cmd":["echo","parallel"]}' | jq &
done

wait

echo
echo "Stress test: creating many sessions"

for i in {1..100}; do
  curl -s -X POST "$BASE/session/create" > /dev/null &
done

wait

echo "Stopping session"
request POST "$BASE/session/$SESSION_ID/stop" | jq
echo

echo "Deleting session"
request DELETE "$BASE/session/$SESSION_ID/" | jq
echo

echo "Test script completed"