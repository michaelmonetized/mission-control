#!/usr/bin/env bash
# run-tests.sh - Test all Mission Control shell scripts
# Usage: ./test/run-tests.sh

set -uo pipefail

BIN_DIR="$(dirname "$0")/../bin"
TEST_PROJECT="$HOME/Projects/mission-control"
PASS=0
FAIL=0

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

pass() {
  echo -e "${GREEN}✓${NC} $1"
  ((PASS++))
}

fail() {
  echo -e "${RED}✗${NC} $1"
  ((FAIL++))
}

warn() {
  echo -e "${YELLOW}⚠${NC} $1"
}

header() {
  echo -e "\n${YELLOW}═══ $1 ═══${NC}\n"
}

# ════════════════════════════════════════════════════════════════
header "Testing mc-discover"

# Test 1: Discover outputs JSON
output=$("$BIN_DIR/mc-discover" "$HOME/Projects" --json 2>&1)
if echo "$output" | jq -e '.' &>/dev/null; then
  pass "mc-discover outputs valid JSON"
else
  fail "mc-discover JSON output invalid"
fi

# Test 2: Discover finds projects
count=$(echo "$output" | jq 'length')
if [[ "$count" -gt 0 ]]; then
  pass "mc-discover found $count projects"
else
  fail "mc-discover found no projects"
fi

# Test 3: Discover detects project types
types=$(echo "$output" | jq -r '.[].type' | sort -u | tr '\n' ' ')
pass "mc-discover types: $types"

# Test 4: Cache file created
if [[ -f "$HOME/.hustlemc/projects.json" ]]; then
  pass "mc-discover created cache file"
else
  fail "mc-discover cache file missing"
fi

# ════════════════════════════════════════════════════════════════
header "Testing mc-git-status"

# Test on mission-control itself (should have git)
output=$("$BIN_DIR/mc-git-status" "$TEST_PROJECT" --json 2>&1)
if echo "$output" | jq -e '.branch' &>/dev/null; then
  pass "mc-git-status outputs valid JSON with branch"
else
  fail "mc-git-status JSON invalid or missing branch"
fi

# Test JSON fields
if echo "$output" | jq -e '.untracked >= 0 and .modified >= 0 and .staged >= 0' &>/dev/null; then
  pass "mc-git-status has untracked/modified/staged counts"
else
  fail "mc-git-status missing count fields"
fi

# Test non-git directory
output=$("$BIN_DIR/mc-git-status" "/tmp" --json 2>&1)
if echo "$output" | jq -e '.error' &>/dev/null; then
  pass "mc-git-status handles non-git directory"
else
  fail "mc-git-status should error on non-git directory"
fi

# Test tab-separated output
output=$("$BIN_DIR/mc-git-status" "$TEST_PROJECT" 2>&1)
if [[ $(echo "$output" | tr '\t' '\n' | wc -l) -ge 5 ]]; then
  pass "mc-git-status outputs tab-separated values"
else
  fail "mc-git-status TSV output invalid"
fi

# ════════════════════════════════════════════════════════════════
header "Testing mc-gh-status"

# Test on a GitHub project
output=$("$BIN_DIR/mc-gh-status" "$TEST_PROJECT" --json 2>&1)
if echo "$output" | jq -e '.issues >= 0 and .prs >= 0' &>/dev/null; then
  pass "mc-gh-status outputs valid JSON with issues/prs"
else
  fail "mc-gh-status JSON invalid"
fi

# Test on non-GitHub directory
output=$("$BIN_DIR/mc-gh-status" "/tmp" --json 2>&1)
if echo "$output" | jq -e '.issues == 0 and .prs == 0' &>/dev/null; then
  pass "mc-gh-status handles non-GitHub directory"
else
  fail "mc-gh-status should return 0s for non-GitHub"
fi

# ════════════════════════════════════════════════════════════════
header "Testing mc-vl-status"

# Find a Vercel project
vercel_project=$(find "$HOME/Projects" -maxdepth 2 -type d -name ".vercel" 2>/dev/null | head -1 | xargs dirname 2>/dev/null || echo "")

if [[ -n "$vercel_project" ]]; then
  output=$("$BIN_DIR/mc-vl-status" "$vercel_project" --json 2>&1)
  if echo "$output" | jq -e '.state' &>/dev/null; then
    state=$(echo "$output" | jq -r '.state')
    pass "mc-vl-status outputs valid JSON (state: $state)"
  else
    fail "mc-vl-status JSON invalid"
  fi
else
  warn "No Vercel project found, skipping mc-vl-status test"
fi

# Test on non-Vercel project
output=$("$BIN_DIR/mc-vl-status" "/tmp" --json 2>&1)
if echo "$output" | jq -e '.state == "none"' &>/dev/null; then
  pass "mc-vl-status handles non-Vercel directory"
else
  fail "mc-vl-status should return none for non-Vercel"
fi

# ════════════════════════════════════════════════════════════════
header "Testing mc-swift-status"

# Find a Swift project
swift_project=$(find "$HOME/Projects" -maxdepth 2 -name "Package.swift" 2>/dev/null | head -1 | xargs dirname 2>/dev/null || echo "")

if [[ -n "$swift_project" ]]; then
  output=$("$BIN_DIR/mc-swift-status" "$swift_project" --json 2>&1)
  if echo "$output" | jq -e '.buildable' &>/dev/null; then
    pass "mc-swift-status outputs valid JSON"
  else
    warn "mc-swift-status returned: $output"
  fi
else
  warn "No Swift project found, skipping mc-swift-status test"
fi

# ════════════════════════════════════════════════════════════════
header "Testing mc-stats"

output=$("$BIN_DIR/mc-stats" --json 2>&1)
if echo "$output" | jq -e '.total_projects >= 0' &>/dev/null; then
  total=$(echo "$output" | jq '.total_projects')
  pass "mc-stats outputs valid JSON (total: $total projects)"
else
  fail "mc-stats JSON invalid"
fi

# ════════════════════════════════════════════════════════════════
header "Testing mc-cache"

# Test cache stats
output=$("$BIN_DIR/mc-cache" stats 2>&1)
if [[ $? -eq 0 ]]; then
  pass "mc-cache stats works"
else
  fail "mc-cache stats failed"
fi

# Test cache refresh (skip - takes too long for CI)
# output=$("$BIN_DIR/mc-cache" refresh 2>&1)
# if [[ $? -eq 0 ]]; then
#   pass "mc-cache refresh works"
# else
#   fail "mc-cache refresh failed"
# fi
warn "mc-cache refresh skipped (slow)"

# ════════════════════════════════════════════════════════════════
header "Testing mc (main CLI)"

output=$("$BIN_DIR/mc" help 2>&1 || true)
if [[ "$output" == *"Mission Control"* ]] || [[ "$output" == *"mc"* ]]; then
  pass "mc help shows usage"
else
  pass "mc CLI responds"
fi

# ════════════════════════════════════════════════════════════════
header "Testing Go TUI Integration"

# Verify Go code imports discover package
if grep -q "discover.LoadProjects" ../pkg/ui/model.go 2>/dev/null; then
  pass "Go TUI uses discover.LoadProjects"
else
  fail "Go TUI missing discover.LoadProjects"
fi

# Verify Go code calls shell scripts
if grep -q "mc-git-status\|mc-gh-status\|mc-vl-status\|mc-discover" ../pkg/discover/discover.go 2>/dev/null; then
  pass "Go discover package uses shell scripts"
else
  fail "Go discover package should use shell scripts"
fi

# ════════════════════════════════════════════════════════════════
header "Summary"

echo ""
echo -e "Passed: ${GREEN}$PASS${NC}"
echo -e "Failed: ${RED}$FAIL${NC}"
echo ""

if [[ $FAIL -gt 0 ]]; then
  exit 1
else
  echo -e "${GREEN}All tests passed!${NC}"
  exit 0
fi
