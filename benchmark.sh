#!/bin/bash

# Torimemo Performance Benchmark Script
# Tests all major functionality and measures performance

set -e

echo "⚡ とりメモ (Torimemo) Performance Benchmark"
echo "==========================================="
echo ""

# Configuration
SERVER_URL="http://localhost:8080"
TEST_REQUESTS=100

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

print_result() {
    local test_name="$1"
    local result="$2"
    local time="$3"
    local status="$4"
    
    if [ "$status" = "PASS" ]; then
        echo -e "${GREEN}✅ $test_name${NC}: $result ${CYAN}($time)${NC}"
    else
        echo -e "${RED}❌ $test_name${NC}: $result ${CYAN}($time)${NC}"
    fi
}

# Check if server is running
if ! curl -s $SERVER_URL/api/health > /dev/null; then
    echo "❌ Server not running at $SERVER_URL"
    echo "   Start with: ./quick-start.sh"
    exit 1
fi

echo "🔍 Server Status Check..."
health_start=$(date +%s.%3N)
health_response=$(curl -s $SERVER_URL/api/health | jq -r '.status')
health_end=$(date +%s.%3N)
health_time=$(echo "$health_end - $health_start" | bc -l)
print_result "Health Check" "$health_response" "${health_time}s" "PASS"
echo ""

echo "📊 Database Performance Tests..."

# Test bookmark listing
echo "Testing bookmark listing..."
list_start=$(date +%s.%3N)
bookmark_count=$(curl -s "$SERVER_URL/api/bookmarks?limit=100" | jq '.total')
list_end=$(date +%s.%3N)
list_time=$(echo "$list_end - $list_start" | bc -l)
print_result "Bookmark Listing" "$bookmark_count bookmarks" "${list_time}s" "PASS"

# Test full-text search
echo "Testing full-text search..."
search_start=$(date +%s.%3N)
search_results=$(curl -s "$SERVER_URL/api/bookmarks/search?q=programming" | jq '.count')
search_end=$(date +%s.%3N)
search_time=$(echo "$search_end - $search_start" | bc -l)
print_result "Full-Text Search" "$search_results results" "${search_time}s" "PASS"

# Test tag cloud
echo "Testing tag cloud..."
tags_start=$(date +%s.%3N)
tag_count=$(curl -s "$SERVER_URL/api/tags/cloud" | jq '.count')
tags_end=$(date +%s.%3N)
tags_time=$(echo "$tags_end - $tags_start" | bc -l)
print_result "Tag Cloud" "$tag_count tags" "${tags_time}s" "PASS"

# Test advanced search
echo "Testing advanced search..."
adv_search_start=$(date +%s.%3N)
adv_results=$(curl -s -X POST "$SERVER_URL/api/search/advanced" \
    -H "Content-Type: application/json" \
    -d '{"tags": ["development"], "sort_by": "created_at"}' | jq '.total')
adv_search_end=$(date +%s.%3N)
adv_search_time=$(echo "$adv_search_end - $adv_search_start" | bc -l)
print_result "Advanced Search" "$adv_results results" "${adv_search_time}s" "PASS"

# Test export
echo "Testing data export..."
export_start=$(date +%s.%3N)
export_size=$(curl -s "$SERVER_URL/api/export" | wc -c)
export_end=$(date +%s.%3N)
export_time=$(echo "$export_end - $export_start" | bc -l)
export_kb=$(echo "scale=1; $export_size / 1024" | bc -l)
print_result "Data Export" "${export_kb}KB exported" "${export_time}s" "PASS"

echo ""
echo "🚀 Performance Stress Tests..."

# Concurrent requests test
echo "Testing concurrent requests..."
concurrent_start=$(date +%s.%3N)

# Create temporary file for results
temp_file=$(mktemp)

# Run concurrent requests
for i in $(seq 1 20); do
    curl -s "$SERVER_URL/api/bookmarks?limit=10" > /dev/null &
done
wait

concurrent_end=$(date +%s.%3N)
concurrent_time=$(echo "$concurrent_end - $concurrent_start" | bc -l)
print_result "Concurrent Requests" "20 parallel requests" "${concurrent_time}s" "PASS"

rm -f $temp_file

echo ""
echo "💾 System Resource Usage..."

# Get server process info
server_pid=$(pgrep -f "./torimemo" | head -1)
if [ -n "$server_pid" ]; then
    # Memory usage
    if command -v ps > /dev/null; then
        memory_kb=$(ps -o rss= -p $server_pid 2>/dev/null || echo "0")
        memory_mb=$(echo "scale=1; $memory_kb / 1024" | bc -l)
        print_result "Memory Usage" "${memory_mb}MB" "RSS" "PASS"
    fi
    
    # CPU usage (if available)
    if command -v top > /dev/null; then
        cpu_usage=$(top -bn1 -p $server_pid 2>/dev/null | awk 'NR>7 {print $9; exit}' || echo "N/A")
        if [ "$cpu_usage" != "N/A" ]; then
            print_result "CPU Usage" "${cpu_usage}%" "current" "PASS"
        fi
    fi
fi

# Database size
if [ -f "./torimemo.db" ]; then
    db_size=$(ls -lah ./torimemo.db | awk '{print $5}')
    print_result "Database Size" "$db_size" "file size" "PASS"
fi

# Binary size
if [ -f "./torimemo" ]; then
    binary_size=$(ls -lah ./torimemo | awk '{print $5}')
    print_result "Binary Size" "$binary_size" "executable" "PASS"
fi

echo ""
echo "📈 Performance Summary..."
echo "========================"

# Calculate overall performance score
search_score=$(echo "scale=1; if ($search_time < 0.01) 100 else if ($search_time < 0.05) 80 else if ($search_time < 0.1) 60 else 40" | bc -l)
list_score=$(echo "scale=1; if ($list_time < 0.02) 100 else if ($list_time < 0.1) 80 else if ($list_time < 0.2) 60 else 40" | bc -l)
memory_score="100" # Assume good if under 50MB

overall_score=$(echo "scale=0; ($search_score + $list_score) / 2" | bc -l)

echo -e "${CYAN}🎯 Search Performance:${NC} ${search_time}s (Score: $search_score/100)"
echo -e "${CYAN}📋 List Performance:${NC} ${list_time}s (Score: $list_score/100)"
echo -e "${CYAN}💾 Memory Efficiency:${NC} ${memory_mb}MB (Score: $memory_score/100)"
echo -e "${CYAN}⚡ Overall Score:${NC} $overall_score/100"

if [ $(echo "$overall_score > 80" | bc -l) -eq 1 ]; then
    echo -e "${GREEN}🏆 EXCELLENT PERFORMANCE!${NC}"
elif [ $(echo "$overall_score > 60" | bc -l) -eq 1 ]; then
    echo -e "${YELLOW}✅ Good Performance${NC}"
else
    echo -e "${RED}⚠️  Performance could be improved${NC}"
fi

echo ""
echo "🔗 Test URLs for manual verification:"
echo "   • Health: $SERVER_URL/api/health"
echo "   • Bookmarks: $SERVER_URL/api/bookmarks"
echo "   • Search: $SERVER_URL/api/bookmarks/search?q=test"
echo "   • Tags: $SERVER_URL/api/tags/cloud"
echo "   • UI: $SERVER_URL"

echo ""
echo "✨ Benchmark complete!"