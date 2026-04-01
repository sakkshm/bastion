
#!/bin/bash

BASE_URL="http://localhost:8080"

echo "-----------------------------"
echo "Creating session..."
echo "-----------------------------"

SESSION_ID=$(curl -s -X POST $BASE_URL/session/create | jq -r '.session_id')

echo "Session ID: $SESSION_ID"

echo
echo "-----------------------------"
echo "Starting session..."
echo "-----------------------------"

curl -s -X POST $BASE_URL/session/$SESSION_ID/start | jq

echo
echo "-----------------------------"
echo "Submitting job..."
echo "-----------------------------"

JOB_ID=$(curl -s -X POST $BASE_URL/session/$SESSION_ID/exec \
  -H "Content-Type: application/json" \
  -d '{"cmd": ["sh", "-c", "sleep 100"]}' \
  | jq -r '.job_id')

echo "Job ID: $JOB_ID"

echo
echo "-----------------------------"
echo "Polling job status..."
echo "-----------------------------"

for i in {1..5}
do
  RESPONSE=$(curl -s -X GET $BASE_URL/session/$SESSION_ID/job/$JOB_ID)
  STATUS=$(echo $RESPONSE | jq -r '.status')

  echo "Status: $STATUS"

  sleep 1
done


echo
echo "-----------------------------"
echo "Checking job status..."
echo "-----------------------------"



curl -s -X GET $BASE_URL/session/$SESSION_ID/job/$JOB_ID | jq

echo
echo "-----------------------------"
echo "Test completed"
echo "-----------------------------"
