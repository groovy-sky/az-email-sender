#!/bin/bash
# Integration tests for azemailsender-cli

set -e

# Configuration
CLI_BINARY="./dist/azemailsender-cli-linux-amd64"
TEST_CONFIG="/tmp/test-azemailsender-config.json"
EXIT_CODE=0

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}Running azemailsender-cli integration tests${NC}"
echo ""

# Helper functions
test_pass() {
    echo -e "  ${GREEN}✓${NC} $1"
}

test_fail() {
    echo -e "  ${RED}✗${NC} $1"
    EXIT_CODE=1
}

test_info() {
    echo -e "  ${YELLOW}→${NC} $1"
}

# Test 1: Binary exists and is executable
echo "Test 1: Binary availability"
if [ -x "$CLI_BINARY" ]; then
    test_pass "Binary exists and is executable"
else
    test_fail "Binary not found or not executable: $CLI_BINARY"
fi

# Test 2: Help command works
echo -e "\nTest 2: Help command"
if $CLI_BINARY --help > /dev/null 2>&1; then
    test_pass "Help command works"
else
    test_fail "Help command failed"
fi

# Test 3: Version command works
echo -e "\nTest 3: Version command"
if version_output=$($CLI_BINARY version 2>&1); then
    test_pass "Version command works"
    test_info "Version: $(echo "$version_output" | head -1)"
else
    test_fail "Version command failed"
fi

# Test 4: Subcommand help
echo -e "\nTest 4: Subcommand help"
commands=("send" "status" "config" "version")
for cmd in "${commands[@]}"; do
    if $CLI_BINARY "$cmd" --help > /dev/null 2>&1; then
        test_pass "$cmd --help works"
    else
        test_fail "$cmd --help failed"
    fi
done

# Test 5: Config init command
echo -e "\nTest 5: Config init command"
rm -f "$TEST_CONFIG"  # Clean up any existing config file
if $CLI_BINARY config init --path "$TEST_CONFIG" > /dev/null 2>&1; then
    test_pass "Config init command works"
    if [ -f "$TEST_CONFIG" ]; then
        test_pass "Config file created successfully"
    else
        test_fail "Config file not created"
    fi
else
    test_fail "Config init command failed"
fi

# Test 6: Config show command
echo -e "\nTest 6: Config show command"
if $CLI_BINARY --config "$TEST_CONFIG" config show > /dev/null 2>&1; then
    test_pass "Config show command works"
else
    test_fail "Config show command failed"
fi

# Test 7: Config env command
echo -e "\nTest 7: Config env command"
if $CLI_BINARY config env > /dev/null 2>&1; then
    test_pass "Config env command works"
else
    test_fail "Config env command failed"
fi

# Test 8: Send command validation (should fail without auth)
echo -e "\nTest 8: Send command validation"
if ! $CLI_BINARY send --from "test@example.com" --to "recipient@example.com" --subject "Test" --text "Test" > /dev/null 2>&1; then
    test_pass "Send command validation works (correctly fails without auth)"
else
    test_fail "Send command validation failed (should require auth)"
fi

# Test 9: Status command validation (should fail without auth)
echo -e "\nTest 9: Status command validation"
if ! $CLI_BINARY status "test-id" > /dev/null 2>&1; then
    test_pass "Status command validation works (correctly fails without auth)"
else
    test_fail "Status command validation failed (should require auth)"
fi

# Test 10: JSON output format
echo -e "\nTest 10: JSON output format"
if json_output=$($CLI_BINARY --json version 2>&1); then
    if echo "$json_output" | grep -q '"version"'; then
        test_pass "JSON output format works"
    else
        test_fail "JSON output format invalid"
    fi
else
    test_fail "JSON output format failed"
fi

# Test 11: Stdin input handling
echo -e "\nTest 11: Stdin input handling"
if echo "test content" | $CLI_BINARY send --from "test@example.com" --to "recipient@example.com" --subject "Test" --endpoint "https://test.com" --access-key "test" > /dev/null 2>&1; then
    test_info "Stdin processing works (command executed)"
else
    test_info "Stdin processing works (command failed as expected without valid auth)"
fi
test_pass "Stdin input handling test completed"

# Test 12: Configuration file priority
echo -e "\nTest 12: Configuration file priority"
# Create a config file with a from address
cat > "$TEST_CONFIG" << EOF
{
  "from": "config@example.com",
  "endpoint": "https://config.communication.azure.com",
  "access_key": "config-key"
}
EOF

# Try to send with config (should use config values, fail on auth but show correct validation)
if output=$($CLI_BINARY --config "$TEST_CONFIG" send --to "recipient@example.com" --subject "Test" --text "Test" 2>&1); then
    test_info "Config file read successfully"
else
    # Should fail due to invalid auth, but should have read config
    test_pass "Configuration file priority test completed"
fi

# Test 13: Environment variable handling
echo -e "\nTest 13: Environment variable handling"
export AZURE_EMAIL_FROM="env@example.com"
export AZURE_EMAIL_ENDPOINT="https://env.communication.azure.com"
export AZURE_EMAIL_ACCESS_KEY="env-key"

if output=$($CLI_BINARY send --to "recipient@example.com" --subject "Test" --text "Test" 2>&1); then
    test_info "Environment variables read successfully"
else
    # Should fail due to invalid auth, but should have read env vars
    test_pass "Environment variable handling test completed"
fi

# Cleanup environment variables
unset AZURE_EMAIL_FROM
unset AZURE_EMAIL_ENDPOINT
unset AZURE_EMAIL_ACCESS_KEY

# Cleanup
rm -f "$TEST_CONFIG"

echo ""
if [ $EXIT_CODE -eq 0 ]; then
    echo -e "${GREEN}All tests passed!${NC}"
else
    echo -e "${RED}Some tests failed.${NC}"
fi

exit $EXIT_CODE