#!/bin/bash

# Test script for e line editor
# This script tests the basic functionality after bug fixes

echo "Testing e line editor functionality..."

# Create a test file
cat > test_input.txt << 'EOF'
line one
line two
line three
line four
line five
EOF

# Test 1: Basic commands functionality
echo "Test 1: Testing basic commands"
echo -e "commands\nquit" | timeout 5 ./e test_input.txt 2>/dev/null || echo "✗ Commands test failed"

# Test 2: Print functionality
echo "Test 2: Testing print functionality"
echo -e "print\nquit" | timeout 5 ./e test_input.txt 2>/dev/null || echo "✗ Print test failed"

# Test 3: Delete functionality (testing the fixed delete command)
echo "Test 3: Testing delete functionality"
echo -e "d\nprint\nquit" | timeout 5 ./e test_input.txt 2>/dev/null || echo "✗ Delete test failed"

# Test 4: Insert functionality
echo "Test 4: Testing insert functionality"
echo -e "new line\ni\nprint\nquit" | timeout 5 ./e test_input.txt 2>/dev/null || echo "✗ Insert test failed"

# Test 5: oops functionality (testing the fixed oops command)
echo "Test 5: Testing oops functionality"
echo -e "line to remove\noops\ncb\nquit" | timeout 5 ./e test_input.txt 2>/dev/null || echo "✗ Oops test failed"

# Test 6: Search functionality
echo "Test 6: Testing search functionality"
echo -e "search:line\nquit" | timeout 5 ./e test_input.txt 2>/dev/null || echo "✗ Search test failed"

# Test 7: Line number navigation
echo "Test 7: Testing line number navigation"
echo -e ":2\nprint:1\nquit" | timeout 5 ./e test_input.txt 2>/dev/null || echo "✗ Line navigation test failed"

# Test 8: Edge case - negative line numbers (testing the fixed printLines function)
echo "Test 8: Testing negative line numbers"
echo -e "print:-1:2\nquit" | timeout 5 ./e test_input.txt 2>/dev/null || echo "✗ Negative line test failed"

# Clean up
rm -f test_input.txt

echo "All tests completed!"
echo ""
echo "To run the tests manually:"
echo "1. Build the program: go build -o e main.go commands.go"
echo "2. Run the test script: chmod +x test_script.sh && ./test_script.sh"
echo "3. Or run automated tests: go test -v"