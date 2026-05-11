#!/bin/bash

# Dogfooding Root Folder
DOGFOODING_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
PROJECT_ROOT="$( cd "$DOGFOODING_DIR/../.." && pwd )"

# Global Data Folder
export EKKENDATA_DIR="$DOGFOODING_DIR/data"
export EKKENAPI_PORT=3033

echo "==========================================="
echo "   STARTING EKKEN DOGFOODING SUITE"
echo "==========================================="

# 1. Build Production Binary (Once)
echo "[1/3] Running 'make build'..."
cd "$PROJECT_ROOT" || exit
make build

# 2. Data & Server Preparation
echo "[2/3] Starting Ekken Dogfooding Server..."
# Always clear old data before test suite runs
rm -rf "$EKKENDATA_DIR"
mkdir -p "$EKKENDATA_DIR"

./dist/ekken > "$EKKENDATA_DIR/server.log" 2>&1 &
SERVER_PID=$!

# Cleanup function to ensure server is killed and port is cleared
cleanup() {
    echo "Stopping Ekken Dogfooding server..."
    if [ -n "$SERVER_PID" ]; then
        kill "$SERVER_PID" 2>/dev/null
        # Wait a bit for graceful shutdown, then force if still alive
        sleep 1
        kill -9 "$SERVER_PID" 2>/dev/null
    fi
    
    # Final check on the port just in case of zombies
    ZOMBIE_PID=$(lsof -t -i:"$EKKENAPI_PORT")
    if [ -n "$ZOMBIE_PID" ]; then
        echo "Clearing zombie process on port $EKKENAPI_PORT..."
        kill -9 "$ZOMBIE_PID" 2>/dev/null
    fi
}

trap cleanup EXIT

echo "Waiting for server to be ready on port $EKKENAPI_PORT..."
sleep 3

SERVER_URL="http://localhost:$EKKENAPI_PORT"

# 3. Loop through all usecases
echo "[3/3] Executing Usecases..."
echo "-------------------------------------------"

# Find all folders starting with "usecase" inside the dogfooding folder
for USECASE_DIR in "$DOGFOODING_DIR"/usecase*; do
    # Skip if no usecase folder exists
    [ -e "$USECASE_DIR" ] || continue
    
    USECASE_NAME=$(basename "$USECASE_DIR")
    WORKFLOW_FILE="$USECASE_DIR/workflow.json"
    
    if [ ! -f "$WORKFLOW_FILE" ]; then
        echo "⚠️  Skip $USECASE_NAME: workflow.json not found."
        continue
    fi
    
    echo "▶ Running $USECASE_NAME..."
    
    # Extract ID from workflow.json for polling purposes
    WORKFLOW_ID=$(grep -o '"id": *"[^"]*"' "$WORKFLOW_FILE" | head -1 | sed 's/"id": *"//' | sed 's/"//')
    
    # Run workflow
    RESPONSE=$(curl -s -X POST "$SERVER_URL/api/workflows/run" \
      -H "Content-Type: application/json" \
      -d @"$WORKFLOW_FILE")
    
    IS_OK=$(echo "$RESPONSE" | grep -o '"ok":true')
    if [ -z "$IS_OK" ]; then
        echo "   ❌ FAILED to start: $RESPONSE"
        continue
    fi
    
    # Polling workflow status
    while true; do
        STATUS_RESP=$(curl -s "$SERVER_URL/api/workflows/$WORKFLOW_ID/status")
        STATUS=$(echo "$STATUS_RESP" | sed -n 's/.*"status":"\([^"]*\)".*/\1/p')
        
        if [ "$STATUS" == "idle" ]; then
            # If status returns to idle, execution is complete.
            # Check logs for any errors.
            LOGS=$(curl -s "$SERVER_URL/api/workflows/$WORKFLOW_ID/logs")
            HAS_ERROR=$(echo "$LOGS" | grep -o '"level":"error"')
            
            if [ -z "$HAS_ERROR" ]; then
                echo "   ✅ PASSED (No errors)"
            else
                echo "   ❌ FAILED (Errors found in logs)"
            fi
            break
        elif [ "$STATUS" == "error" ] || [ "$STATUS" == "stopped" ]; then
            echo "   ❌ FAILED (Final status: $STATUS)"
            break
        fi
        sleep 2
    done
done

echo "==========================================="
echo "   DOGFOODING SUITE FINISHED"
echo "==========================================="
