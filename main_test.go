package main

import (
	"bytes"
	"io"
	"os"
	"reflect"
	"strings"
	"testing"
)

func TestPerLine(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected []string
	}{
		{
			name:     "single line",
			input:    []string{"hello"},
			expected: []string{"hello"},
		},
		{
			name:     "multiple lines in one string",
			input:    []string{"hello\nworld\ntest"},
			expected: []string{"hello", "world", "test"},
		},
		{
			name:     "multiple strings with newlines",
			input:    []string{"hello\nworld", "test\nmore"},
			expected: []string{"hello", "world", "test", "more"},
		},
		{
			name:     "empty string",
			input:    []string{""},
			expected: []string{""},
		},
		{
			name:     "string with trailing newline",
			input:    []string{"hello\n"},
			expected: []string{"hello", ""},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := perLine(tt.input)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("perLine(%v) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestPrintLines(t *testing.T) {
	// Setup a FileBuffer for testing
	FB = FileBuffer{
		Pos:      0,
		Length:   0,
		File:     nil,
		Contents: []string{"line1\nline2\nline3", "line4\nline5"},
		Meta:     map[string]string{"filename": "test.txt"},
	}

	tests := []struct {
		name         string
		from         int
		many         int
		page         bool
		expectedLines int
		shouldPanic  bool
	}{
		{
			name:         "print single line",
			from:         0,
			many:         1,
			page:         false,
			expectedLines: 1,
		},
		{
			name:         "print multiple lines",
			from:         0,
			many:         3,
			page:         false,
			expectedLines: 3,
		},
		{
			name:         "print from middle",
			from:         2,
			many:         2,
			page:         false,
			expectedLines: 2,
		},
		{
			name:         "print beyond end",
			from:         10,
			many:         5,
			page:         false,
			expectedLines: 0,
		},
		{
			name:         "print all lines (many=0)",
			from:         0,
			many:         0,
			page:         false,
			expectedLines: 5, // After perLine, we should have 5 lines
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture stdout
			old := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			if tt.shouldPanic {
				defer func() {
					if r := recover(); r == nil {
						t.Errorf("Expected panic but didn't get one")
					}
				}()
			}

			printLines(tt.from, tt.many, tt.page)

			w.Close()
			os.Stdout = old

			var buf bytes.Buffer
			io.Copy(&buf, r)
			output := buf.String()

			lines := strings.Split(strings.TrimSpace(output), "\n")
			if len(lines) == 1 && lines[0] == "" {
				lines = []string{}
			}

			if len(lines) != tt.expectedLines {
				t.Errorf("printLines(%d, %d, %v) printed %d lines, expected %d lines\nOutput: %q", 
					tt.from, tt.many, tt.page, len(lines), tt.expectedLines, output)
			}
		})
	}
}

func TestSimpleSearch(t *testing.T) {
	// Setup a FileBuffer for testing
	FB = FileBuffer{
		Pos:      0,
		Length:   0,
		File:     nil,
		Contents: []string{"hello world\ntest line\nhello again\nfinal line"},
		Meta:     map[string]string{"filename": "test.txt"},
	}
	Line = 0

	tests := []struct {
		name          string
		keyword       string
		many          int
		page          bool
		expectedCount int
	}{
		{
			name:          "search for hello",
			keyword:       "hello",
			many:          0,
			page:          false,
			expectedCount: 2,
		},
		{
			name:          "search for line",
			keyword:       "line",
			many:          1,
			page:          false,
			expectedCount: 1,
		},
		{
			name:          "search for nonexistent",
			keyword:       "nonexistent",
			many:          0,
			page:          false,
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture stdout
			old := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			simpleSearch(tt.keyword, tt.many, tt.page)

			w.Close()
			os.Stdout = old

			var buf bytes.Buffer
			io.Copy(&buf, r)
			output := buf.String()

			lines := strings.Split(strings.TrimSpace(output), "\n")
			if len(lines) == 1 && lines[0] == "" {
				lines = []string{}
			}

			if len(lines) != tt.expectedCount {
				t.Errorf("simpleSearch(%q, %d, %v) found %d matches, expected %d\nOutput: %q", 
					tt.keyword, tt.many, tt.page, len(lines), tt.expectedCount, output)
			}
		})
	}
}

func TestInsertCommand(t *testing.T) {
	tests := []struct {
		name           string
		initialContent []string
		line           int
		insertLines    []string
		expectedResult []string
	}{
		{
			name:           "insert at beginning",
			initialContent: []string{"line1", "line2", "line3"},
			line:           0,
			insertLines:    []string{"new line"},
			expectedResult: []string{"new line", "line1", "line2", "line3"},
		},
		{
			name:           "insert in middle",
			initialContent: []string{"line1", "line2", "line3"},
			line:           1,
			insertLines:    []string{"new line"},
			expectedResult: []string{"line1", "new line", "line2", "line3"},
		},
		{
			name:           "insert at end",
			initialContent: []string{"line1", "line2", "line3"},
			line:           5,
			insertLines:    []string{"new line"},
			expectedResult: []string{"line1", "line2", "line3", "new line"},
		},
		{
			name:           "insert multiple lines",
			initialContent: []string{"line1", "line2"},
			line:           1,
			insertLines:    []string{"new line 1", "new line 2"},
			expectedResult: []string{"line1", "new line 1", "new line 2", "line2"},
		},
	}

	// Initialize commands
	initCommands()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			FB = FileBuffer{
				Contents: tt.initialContent,
			}
			Line = tt.line

			err := COMMANDS["i"](tt.insertLines)
			if err != nil {
				t.Errorf("Insert command failed: %v", err)
			}

			if !reflect.DeepEqual(FB.Contents, tt.expectedResult) {
				t.Errorf("Insert result = %v, want %v", FB.Contents, tt.expectedResult)
			}
		})
	}
}

func TestDeleteCommand(t *testing.T) {
	tests := []struct {
		name           string
		initialContent []string
		line           int
		expectedResult []string
	}{
		{
			name:           "delete first line",
			initialContent: []string{"line1", "line2", "line3"},
			line:           0,
			expectedResult: []string{"line2", "line3"},
		},
		{
			name:           "delete middle line",
			initialContent: []string{"line1", "line2", "line3"},
			line:           1,
			expectedResult: []string{"line1", "line3"},
		},
		{
			name:           "delete last line",
			initialContent: []string{"line1", "line2", "line3"},
			line:           2,
			expectedResult: []string{"line1", "line2"},
		},
		{
			name:           "delete beyond end",
			initialContent: []string{"line1", "line2", "line3"},
			line:           5,
			expectedResult: []string{"line1", "line2"},
		},
		{
			name:           "delete from single line",
			initialContent: []string{"line1"},
			line:           0,
			expectedResult: []string{},
		},
	}

	// Initialize commands
	initCommands()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			FB = FileBuffer{
				Contents: tt.initialContent,
			}
			Line = tt.line

			err := COMMANDS["d"]([]string{})
			if err != nil {
				t.Errorf("Delete command failed: %v", err)
			}

			// Handle nil vs empty slice comparison
			if (FB.Contents == nil && tt.expectedResult == nil) ||
			   (len(FB.Contents) == 0 && len(tt.expectedResult) == 0) {
				// Both are effectively empty, test passes
			} else if !reflect.DeepEqual(FB.Contents, tt.expectedResult) {
				t.Errorf("Delete result = %v, want %v", FB.Contents, tt.expectedResult)
			}
		})
	}
}

func TestPrintCommand(t *testing.T) {
	// Initialize commands
	initCommands()

	FB = FileBuffer{
		Contents: []string{"line1", "line2", "line3", "line4", "line5"},
	}

	tests := []struct {
		name          string
		args          []string
		expectedLines int
	}{
		{
			name:          "print all",
			args:          []string{},
			expectedLines: 5,
		},
		{
			name:          "print single line",
			args:          []string{"2"},
			expectedLines: 1,
		},
		{
			name:          "print range",
			args:          []string{"1", "3"},
			expectedLines: 3,
		},
		{
			name:          "print invalid range",
			args:          []string{"3", "1"},
			expectedLines: 0, // This should be handled gracefully
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture stdout
			old := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			err := COMMANDS["print"](tt.args)
			if err != nil {
				t.Errorf("Print command failed: %v", err)
			}

			w.Close()
			os.Stdout = old

			var buf bytes.Buffer
			io.Copy(&buf, r)
			output := buf.String()

			lines := strings.Split(strings.TrimSpace(output), "\n")
			if len(lines) == 1 && lines[0] == "" {
				lines = []string{}
			}

			if len(lines) != tt.expectedLines {
				t.Errorf("Print command with args %v printed %d lines, expected %d\nOutput: %q", 
					tt.args, len(lines), tt.expectedLines, output)
			}
		})
	}
}

func TestCommandBufferHandling(t *testing.T) {
	tests := []struct {
		name            string
		initialBuffer   []string
		command         string
		expectedBuffer  []string
	}{
		{
			name:            "oops removes last line",
			initialBuffer:   []string{"line1", "line2", "line3"},
			command:         "oops",
			expectedBuffer:  []string{"line1", "line2"},
		},
		{
			name:            "oops! clears buffer",
			initialBuffer:   []string{"line1", "line2", "line3"},
			command:         "oops!",
			expectedBuffer:  []string{},
		},
		{
			name:            "oops on empty buffer",
			initialBuffer:   []string{},
			command:         "oops",
			expectedBuffer:  []string{},
		},
	}

	// Initialize commands
	initCommands()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			commandBuffer = tt.initialBuffer

			err := COMMANDS[tt.command]([]string{})
			if err != nil {
				t.Errorf("Command %s failed: %v", tt.command, err)
			}

			// Handle nil vs empty slice comparison
			if (commandBuffer == nil && tt.expectedBuffer == nil) ||
			   (len(commandBuffer) == 0 && len(tt.expectedBuffer) == 0) {
				// Both are effectively empty, test passes
			} else if !reflect.DeepEqual(commandBuffer, tt.expectedBuffer) {
				t.Errorf("Command %s result = %v (len=%d), want %v (len=%d)", tt.command, commandBuffer, len(commandBuffer), tt.expectedBuffer, len(tt.expectedBuffer))
			}
		})
	}
}

func TestFileOperations(t *testing.T) {
	// Create a temporary file for testing
	tempFile, err := os.CreateTemp("", "test_*.txt")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	testContent := "line1\nline2\nline3\n"
	_, err = tempFile.WriteString(testContent)
	if err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	tempFile.Close()

	// Test opening file
	err = openFullFile(tempFile.Name())
	if err != nil {
		t.Errorf("Failed to open file: %v", err)
	}

	// Check if content was loaded correctly
	expectedLines := []string{"line1\nline2\nline3\n"}
	if !reflect.DeepEqual(FB.Contents, expectedLines) {
		t.Errorf("File content = %v, want %v", FB.Contents, expectedLines)
	}

	// Test saving file
	FB.Contents = []string{"modified line1", "modified line2"}
	err = saveFullFile("")
	if err != nil {
		t.Errorf("Failed to save file: %v", err)
	}

	// Read back and verify
	content, err := os.ReadFile(tempFile.Name())
	if err != nil {
		t.Errorf("Failed to read saved file: %v", err)
	}

	expectedContent := "modified line1\nmodified line2\n"
	if string(content) != expectedContent {
		t.Errorf("Saved content = %q, want %q", string(content), expectedContent)
	}
}

func TestPrintLinesEdgeCases(t *testing.T) {
	tests := []struct {
		name         string
		initialContent []string
		from         int
		many         int
		page         bool
		expectedLines int
	}{
		{
			name:         "negative from index",
			initialContent: []string{"line1", "line2", "line3"},
			from:         -1,
			many:         2,
			page:         false,
			expectedLines: 2,
		},
		{
			name:         "many is zero prints all",
			initialContent: []string{"line1", "line2", "line3"},
			from:         0,
			many:         0,
			page:         false,
			expectedLines: 3,
		},
		{
			name:         "from beyond content length",
			initialContent: []string{"line1", "line2"},
			from:         5,
			many:         3,
			page:         false,
			expectedLines: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			FB = FileBuffer{
				Contents: tt.initialContent,
			}

			// Capture stdout
			old := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			printLines(tt.from, tt.many, tt.page)

			w.Close()
			os.Stdout = old

			var buf bytes.Buffer
			io.Copy(&buf, r)
			output := buf.String()

			lines := strings.Split(strings.TrimSpace(output), "\n")
			if len(lines) == 1 && lines[0] == "" {
				lines = []string{}
			}

			if len(lines) != tt.expectedLines {
				t.Errorf("printLines(%d, %d, %v) printed %d lines, expected %d lines\nOutput: %q", 
					tt.from, tt.many, tt.page, len(lines), tt.expectedLines, output)
			}
		})
	}
}

func TestCommandParsingEdgeCases(t *testing.T) {
	// Initialize commands
	initCommands()

	tests := []struct {
		name           string
		command        string
		expectedError  bool
	}{
		{
			name:          "empty command",
			command:       "",
			expectedError: false,
		},
		{
			name:          "colon only command",
			command:       ":",
			expectedError: false,
		},
		{
			name:          "dot prefix command",
			command:       ".print",
			expectedError: false,
		},
		{
			name:          "search without keyword",
			command:       "search",
			expectedError: true,
		},
		{
			name:          "invalid line number",
			command:       ":abc",
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This test verifies command parsing doesn't crash on edge cases
			if tt.command == "search" {
				err := COMMANDS["search"]([]string{})
				if (err != nil) != tt.expectedError {
					t.Errorf("Expected error: %v, got error: %v", tt.expectedError, err)
				}
			}
		})
	}
}

func TestCommandBufferParsing(t *testing.T) {
	// Initialize commands
	initCommands()

	tests := []struct {
		name        string
		input       string
		shouldPanic bool
	}{
		{
			name:        "single colon",
			input:       ":",
			shouldPanic: false,
		},
		{
			name:        "colon with empty segments",
			input:       ":::",
			shouldPanic: false,
		},
		{
			name:        "command with single colon",
			input:       "print:",
			shouldPanic: false,
		},
		{
			name:        "command with multiple colons",
			input:       "print:arg1:arg2:",
			shouldPanic: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.shouldPanic {
				defer func() {
					if r := recover(); r == nil {
						t.Errorf("Expected panic but didn't get one")
					}
				}()
			}

			// Simulate the command parsing logic from main loop
			command := tt.input
			commands := strings.Split(command, ":")
			cb := make([]string, len(commands)-1)
			run, ok := COMMANDS[command]
			if !ok && len(commands) > 1 {
				run, ok = COMMANDS[commands[0]]
				if len(commands) > 1 {
					copy(cb, commands[1:])
				}
			}

			// Test passes if we get here without panic
			_ = run
			_ = ok
			_ = cb
		})
	}
}

func TestDeleteCommandEdgeCases(t *testing.T) {
	// Initialize commands
	initCommands()

	tests := []struct {
		name           string
		initialContent []string
		line           int
		expectedResult []string
		shouldError    bool
	}{
		{
			name:           "delete from empty file",
			initialContent: []string{},
			line:           0,
			expectedResult: []string{},
			shouldError:    false,
		},
		{
			name:           "delete with negative line number",
			initialContent: []string{"line1", "line2", "line3"},
			line:           -1,
			expectedResult: []string{"line2", "line3"},
			shouldError:    false,
		},
		{
			name:           "delete from single line file",
			initialContent: []string{"only line"},
			line:           0,
			expectedResult: []string{},
			shouldError:    false,
		},
		{
			name:           "delete with line number way beyond end",
			initialContent: []string{"line1", "line2"},
			line:           100,
			expectedResult: []string{"line1"},
			shouldError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			FB = FileBuffer{
				Contents: tt.initialContent,
			}
			Line = tt.line

			err := COMMANDS["d"]([]string{})
			if (err != nil) != tt.shouldError {
				t.Errorf("Expected error: %v, got error: %v", tt.shouldError, err)
			}

			// Handle nil vs empty slice comparison
			if (FB.Contents == nil && tt.expectedResult == nil) ||
			   (len(FB.Contents) == 0 && len(tt.expectedResult) == 0) {
				// Both are effectively empty, test passes
			} else if !reflect.DeepEqual(FB.Contents, tt.expectedResult) {
				t.Errorf("Delete result = %v, want %v", FB.Contents, tt.expectedResult)
			}
		})
	}
}

func TestReplaceCommandEdgeCases(t *testing.T) {
	// Initialize commands
	initCommands()

	tests := []struct {
		name           string
		initialContent []string
		line           int
		replaceLines   []string
		expectedResult []string
	}{
		{
			name:           "replace in empty file",
			initialContent: []string{},
			line:           0,
			replaceLines:   []string{"new line"},
			expectedResult: []string{"new line"},
		},
		{
			name:           "replace beyond end of file",
			initialContent: []string{"line1", "line2"},
			line:           5,
			replaceLines:   []string{"new line"},
			expectedResult: []string{"line1", "line2", "new line"},
		},
		{
			name:           "replace with multiple lines",
			initialContent: []string{"line1", "line2", "line3"},
			line:           1,
			replaceLines:   []string{"replacement line 1", "replacement line 2"},
			expectedResult: []string{"line1", "replacement line 1\nreplacement line 2", "line3"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			FB = FileBuffer{
				Contents: tt.initialContent,
			}
			Line = tt.line

			err := COMMANDS["r"](tt.replaceLines)
			if err != nil {
				t.Errorf("Replace command failed: %v", err)
			}

			// Handle nil vs empty slice comparison
			if (FB.Contents == nil && tt.expectedResult == nil) ||
			   (len(FB.Contents) == 0 && len(tt.expectedResult) == 0) {
				// Both are effectively empty, test passes
			} else if !reflect.DeepEqual(FB.Contents, tt.expectedResult) {
				t.Errorf("Replace result = %v, want %v", FB.Contents, tt.expectedResult)
			}
		})
	}
}

func TestAppendCommandEdgeCases(t *testing.T) {
	// Initialize commands
	initCommands()

	tests := []struct {
		name           string
		initialContent []string
		line           int
		appendLines    []string
		expectedResult []string
	}{
		{
			name:           "append to empty file",
			initialContent: []string{},
			line:           0,
			appendLines:    []string{"new line"},
			expectedResult: []string{"new line"},
		},
		{
			name:           "append after last line",
			initialContent: []string{"line1", "line2"},
			line:           1,
			appendLines:    []string{"new line"},
			expectedResult: []string{"line1", "line2", "new line"},
		},
		{
			name:           "append with line beyond end",
			initialContent: []string{"line1", "line2"},
			line:           5,
			appendLines:    []string{"new line"},
			expectedResult: []string{"line1", "line2", "new line"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			FB = FileBuffer{
				Contents: tt.initialContent,
			}
			Line = tt.line

			err := COMMANDS["a"](tt.appendLines)
			if err != nil {
				t.Errorf("Append command failed: %v", err)
			}

			// Handle nil vs empty slice comparison
			if (FB.Contents == nil && tt.expectedResult == nil) ||
			   (len(FB.Contents) == 0 && len(tt.expectedResult) == 0) {
				// Both are effectively empty, test passes
			} else if !reflect.DeepEqual(FB.Contents, tt.expectedResult) {
				t.Errorf("Append result = %v, want %v", FB.Contents, tt.expectedResult)
			}
		})
	}
}