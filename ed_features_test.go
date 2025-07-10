package main

import (
	"io"
	"os"
	"reflect"
	"strings"
	"testing"
)

// Test Line Addressing System
func TestLineAddressing(t *testing.T) {
	tests := []struct {
		name           string
		content        []string
		currentLine    int
		address        string
		expectedLine   int
		expectedError  bool
	}{
		{
			name:          "current line (.)",
			content:       []string{"line1", "line2", "line3"},
			currentLine:   1,
			address:       ".",
			expectedLine:  1,
			expectedError: false,
		},
		{
			name:          "last line ($)",
			content:       []string{"line1", "line2", "line3"},
			currentLine:   1,
			address:       "$",
			expectedLine:  2,
			expectedError: false,
		},
		{
			name:          "specific line number",
			content:       []string{"line1", "line2", "line3"},
			currentLine:   1,
			address:       "2",
			expectedLine:  1,
			expectedError: false,
		},
		{
			name:          "relative positive (+)",
			content:       []string{"line1", "line2", "line3"},
			currentLine:   0,
			address:       "+1",
			expectedLine:  1,
			expectedError: false,
		},
		{
			name:          "relative negative (-)",
			content:       []string{"line1", "line2", "line3"},
			currentLine:   2,
			address:       "-1",
			expectedLine:  1,
			expectedError: false,
		},
		{
			name:          "out of bounds positive",
			content:       []string{"line1", "line2", "line3"},
			currentLine:   2,
			address:       "+5",
			expectedLine:  2,
			expectedError: true,
		},
		{
			name:          "out of bounds negative",
			content:       []string{"line1", "line2", "line3"},
			currentLine:   0,
			address:       "-5",
			expectedLine:  0,
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			FB = FileBuffer{
				Pos:      tt.currentLine,
				Length:   len(tt.content),
				Contents: tt.content,
				Meta:     make(map[string]string),
			}
			Line = tt.currentLine

			// Test address parsing (this will need to be implemented)
			line, err := parseAddress(tt.address)
			
			if tt.expectedError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if line != tt.expectedLine {
					t.Errorf("Expected line %d, got %d", tt.expectedLine, line)
				}
			}
		})
	}
}

// Test Range Addressing
func TestRangeAddressing(t *testing.T) {
	tests := []struct {
		name          string
		content       []string
		currentLine   int
		rangeStr      string
		expectedStart int
		expectedEnd   int
		expectedError bool
	}{
		{
			name:          "comma range (1,3)",
			content:       []string{"line1", "line2", "line3", "line4"},
			currentLine:   1,
			rangeStr:      "1,3",
			expectedStart: 0,
			expectedEnd:   2,
			expectedError: false,
		},
		{
			name:          "percent range (%)",
			content:       []string{"line1", "line2", "line3"},
			currentLine:   1,
			rangeStr:      "%",
			expectedStart: 0,
			expectedEnd:   2,
			expectedError: false,
		},
		{
			name:          "semicolon range (.,;)",
			content:       []string{"line1", "line2", "line3"},
			currentLine:   1,
			rangeStr:      ".,$",
			expectedStart: 1,
			expectedEnd:   2,
			expectedError: false,
		},
		{
			name:          "single line range",
			content:       []string{"line1", "line2", "line3"},
			currentLine:   1,
			rangeStr:      "2",
			expectedStart: 1,
			expectedEnd:   1,
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			FB = FileBuffer{
				Pos:      tt.currentLine,
				Length:   len(tt.content),
				Contents: tt.content,
				Meta:     make(map[string]string),
			}
			Line = tt.currentLine

			// Test range parsing (this will need to be implemented)
			start, end, err := parseRange(tt.rangeStr)
			
			if tt.expectedError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if start != tt.expectedStart || end != tt.expectedEnd {
					t.Errorf("Expected range [%d,%d], got [%d,%d]", tt.expectedStart, tt.expectedEnd, start, end)
				}
			}
		})
	}
}

// Test Substitute Command
func TestSubstituteCommand(t *testing.T) {
	tests := []struct {
		name            string
		content         []string
		currentLine     int
		command         string
		expectedContent []string
		expectedError   bool
	}{
		{
			name:            "simple substitute",
			content:         []string{"hello world", "hello there"},
			currentLine:     0,
			command:         "s/hello/hi/",
			expectedContent: []string{"hi world", "hello there"},
			expectedError:   false,
		},
		{
			name:            "global substitute",
			content:         []string{"hello hello world", "hello there"},
			currentLine:     0,
			command:         "s/hello/hi/g",
			expectedContent: []string{"hi hi world", "hello there"},
			expectedError:   false,
		},
		{
			name:            "substitute with range",
			content:         []string{"hello world", "hello there", "hello again"},
			currentLine:     0,
			command:         "1,2s/hello/hi/",
			expectedContent: []string{"hi world", "hi there", "hello again"},
			expectedError:   false,
		},
		{
			name:            "substitute with backreference",
			content:         []string{"hello world", "goodbye world"},
			currentLine:     0,
			command:         "s/\\(hello\\|goodbye\\)/[\\1]/",
			expectedContent: []string{"[hello] world", "goodbye world"},
			expectedError:   false,
		},
		{
			name:          "substitute no match",
			content:       []string{"hello world", "goodbye world"},
			currentLine:   0,
			command:       "s/xyz/abc/",
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			FB = FileBuffer{
				Pos:      tt.currentLine,
				Length:   len(tt.content),
				Contents: make([]string, len(tt.content)),
				Meta:     make(map[string]string),
			}
			copy(FB.Contents, tt.content)
			Line = tt.currentLine

			// Test substitute command (this will need to be implemented)
			err := executeSubstitute(tt.command)
			
			if tt.expectedError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if !reflect.DeepEqual(FB.Contents, tt.expectedContent) {
					t.Errorf("Expected content %v, got %v", tt.expectedContent, FB.Contents)
				}
			}
		})
	}
}

// Test Global Commands
func TestGlobalCommand(t *testing.T) {
	tests := []struct {
		name            string
		content         []string
		command         string
		expectedContent []string
		expectedError   bool
	}{
		{
			name:            "global delete",
			content:         []string{"line1", "delete this", "line3", "delete this too"},
			command:         "g/delete/d",
			expectedContent: []string{"line1", "line3"},
			expectedError:   false,
		},
		{
			name:            "global print",
			content:         []string{"apple", "banana", "cherry"},
			command:         "g/a/p",
			expectedContent: []string{"apple", "banana", "cherry"}, // content unchanged
			expectedError:   false,
		},
		{
			name:            "global substitute",
			content:         []string{"hello world", "hello there", "goodbye world"},
			command:         "g/hello/s/hello/hi/",
			expectedContent: []string{"hi world", "hi there", "goodbye world"},
			expectedError:   false,
		},
		{
			name:          "global with no matches",
			content:       []string{"line1", "line2", "line3"},
			command:       "g/xyz/d",
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			FB = FileBuffer{
				Pos:      0,
				Length:   len(tt.content),
				Contents: make([]string, len(tt.content)),
				Meta:     make(map[string]string),
			}
			copy(FB.Contents, tt.content)
			Line = 0

			// Test global command (this will need to be implemented)
			err := executeGlobal(tt.command)
			
			if tt.expectedError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if !reflect.DeepEqual(FB.Contents, tt.expectedContent) {
					t.Errorf("Expected content %v, got %v", tt.expectedContent, FB.Contents)
				}
			}
		})
	}
}

// Test Inverse Global Command
func TestInverseGlobalCommand(t *testing.T) {
	tests := []struct {
		name            string
		content         []string
		command         string
		expectedContent []string
		expectedError   bool
	}{
		{
			name:            "inverse global delete",
			content:         []string{"line1", "delete this", "line3", "delete this too"},
			command:         "v/delete/d",
			expectedContent: []string{"delete this", "delete this too"},
			expectedError:   false,
		},
		{
			name:            "inverse global print",
			content:         []string{"apple", "banana", "cherry"},
			command:         "v/a/p",
			expectedContent: []string{"apple", "banana", "cherry"}, // content unchanged
			expectedError:   false,
		},
		{
			name:            "inverse global substitute",
			content:         []string{"hello world", "hello there", "goodbye world"},
			command:         "v/hello/s/goodbye/farewell/",
			expectedContent: []string{"hello world", "hello there", "farewell world"},
			expectedError:   false,
		},
		{
			name:          "inverse global with no matches",
			content:       []string{"xyz", "xyz", "xyz"},
			command:       "v/xyz/d",
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			FB = FileBuffer{
				Pos:      0,
				Length:   len(tt.content),
				Contents: make([]string, len(tt.content)),
				Meta:     make(map[string]string),
			}
			copy(FB.Contents, tt.content)
			Line = 0

			// Test inverse global command
			err := executeInverseGlobal(tt.command)
			
			if tt.expectedError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if !reflect.DeepEqual(FB.Contents, tt.expectedContent) {
					t.Errorf("Expected content %v, got %v", tt.expectedContent, FB.Contents)
				}
			}
		})
	}
}

// Test Filename Command
func TestFilenameCommand(t *testing.T) {
	tests := []struct {
		name            string
		command         string
		initialFilename string
		expectedOutput  string
		expectedError   bool
	}{
		{
			name:            "show current filename",
			command:         "f",
			initialFilename: "test.txt",
			expectedOutput:  "test.txt",
			expectedError:   false,
		},
		{
			name:            "show filename when none set",
			command:         "f",
			initialFilename: "",
			expectedOutput:  "no filename set",
			expectedError:   false,
		},
		{
			name:            "set new filename",
			command:         "f newfile.txt",
			initialFilename: "old.txt",
			expectedOutput:  "newfile.txt",
			expectedError:   false,
		},
		{
			name:            "set filename with spaces",
			command:         "f my file.txt",
			initialFilename: "",
			expectedOutput:  "my file.txt",
			expectedError:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			FB = FileBuffer{
				Pos:      0,
				Length:   0,
				Contents: []string{},
				Meta:     make(map[string]string),
			}
			if tt.initialFilename != "" {
				FB.Meta["filename"] = tt.initialFilename
			}

			// Capture output
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			// Test filename command
			err := executeFilename(tt.command)

			// Restore stdout and get output
			w.Close()
			os.Stdout = oldStdout
			output, _ := io.ReadAll(r)
			actualOutput := strings.TrimSpace(string(output))

			if tt.expectedError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if actualOutput != tt.expectedOutput {
					t.Errorf("Expected output %q, got %q", tt.expectedOutput, actualOutput)
				}
			}
		})
	}
}

// Test Help Command
func TestHelpCommand(t *testing.T) {
	tests := []struct {
		name          string
		verbose       bool
		expectedError bool
	}{
		{
			name:          "brief help",
			verbose:       false,
			expectedError: false,
		},
		{
			name:          "detailed help",
			verbose:       true,
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture output
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			// Test help command
			err := executeHelp(tt.verbose)

			// Restore stdout and get output
			w.Close()
			os.Stdout = oldStdout
			output, _ := io.ReadAll(r)
			actualOutput := strings.TrimSpace(string(output))

			if tt.expectedError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if actualOutput == "" {
					t.Errorf("Expected some output but got none")
				}
				if tt.verbose {
					if !strings.Contains(actualOutput, "Ed-compatible line editor help") {
						t.Errorf("Expected detailed help but got: %s", actualOutput)
					}
				} else {
					if !strings.Contains(actualOutput, "Available commands") {
						t.Errorf("Expected brief help but got: %s", actualOutput)
					}
				}
			}
		})
	}
}

// Test Edit Commands
func TestEditCommand(t *testing.T) {
	tests := []struct {
		name            string
		command         string
		filename        string
		createFile      bool
		fileContent     string
		expectedError   bool
	}{
		{
			name:            "edit existing file",
			command:         "e testfile.txt",
			filename:        "testfile.txt",
			createFile:      true,
			fileContent:     "line1\nline2\nline3",
			expectedError:   false,
		},
		{
			name:            "edit with just e command",
			command:         "e",
			filename:        "",
			createFile:      false,
			fileContent:     "",
			expectedError:   true,
		},
		{
			name:            "edit nonexistent file",
			command:         "e nonexistent.txt",
			filename:        "nonexistent.txt",
			createFile:      false,
			fileContent:     "",
			expectedError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			FB = FileBuffer{
				Pos:      0,
				Length:   0,
				Contents: []string{},
				Meta:     make(map[string]string),
			}

			// Create test file if needed
			if tt.createFile {
				if err := os.WriteFile(tt.filename, []byte(tt.fileContent), 0644); err != nil {
					t.Fatalf("Failed to create test file: %v", err)
				}
				defer os.Remove(tt.filename)
			}

			// Test edit command
			err := executeEdit(tt.command)
			
			if tt.expectedError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				// Check if file was loaded
				if tt.createFile {
					expectedLines := strings.Split(tt.fileContent, "\n")
					if !reflect.DeepEqual(FB.Contents, expectedLines) {
						t.Errorf("Expected content %v, got %v", expectedLines, FB.Contents)
					}
				}
			}
		})
	}
}

// Test Edit Unconditional Command
func TestEditUnconditionalCommand(t *testing.T) {
	tests := []struct {
		name            string
		command         string
		filename        string
		createFile      bool
		fileContent     string
		expectedError   bool
	}{
		{
			name:            "edit unconditional existing file",
			command:         "E testfile.txt",
			filename:        "testfile.txt",
			createFile:      true,
			fileContent:     "line1\nline2\nline3",
			expectedError:   false,
		},
		{
			name:            "edit unconditional with just E command",
			command:         "E",
			filename:        "",
			createFile:      false,
			fileContent:     "",
			expectedError:   true,
		},
		{
			name:            "edit unconditional nonexistent file",
			command:         "E nonexistent.txt",
			filename:        "nonexistent.txt",
			createFile:      false,
			fileContent:     "",
			expectedError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			FB = FileBuffer{
				Pos:      0,
				Length:   0,
				Contents: []string{},
				Meta:     make(map[string]string),
			}

			// Create test file if needed
			if tt.createFile {
				if err := os.WriteFile(tt.filename, []byte(tt.fileContent), 0644); err != nil {
					t.Fatalf("Failed to create test file: %v", err)
				}
				defer os.Remove(tt.filename)
			}

			// Test edit unconditional command
			err := executeEditUnconditional(tt.command)
			
			if tt.expectedError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				// Check if file was loaded
				if tt.createFile {
					expectedLines := strings.Split(tt.fileContent, "\n")
					if !reflect.DeepEqual(FB.Contents, expectedLines) {
						t.Errorf("Expected content %v, got %v", expectedLines, FB.Contents)
					}
				}
			}
		})
	}
}

// Test Change Command
func TestChangeCommand(t *testing.T) {
	tests := []struct {
		name            string
		content         []string
		currentLine     int
		rangeStr        string
		newLines        []string
		expectedContent []string
		expectedError   bool
	}{
		{
			name:            "change single line",
			content:         []string{"line1", "line2", "line3"},
			currentLine:     1,
			rangeStr:        "2",
			newLines:        []string{"new line"},
			expectedContent: []string{"line1", "new line", "line3"},
			expectedError:   false,
		},
		{
			name:            "change multiple lines",
			content:         []string{"line1", "line2", "line3", "line4"},
			currentLine:     0,
			rangeStr:        "2,3",
			newLines:        []string{"replacement1", "replacement2"},
			expectedContent: []string{"line1", "replacement1", "replacement2", "line4"},
			expectedError:   false,
		},
		{
			name:            "change with fewer replacement lines",
			content:         []string{"line1", "line2", "line3", "line4"},
			currentLine:     0,
			rangeStr:        "2,4",
			newLines:        []string{"single replacement"},
			expectedContent: []string{"line1", "single replacement"},
			expectedError:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			FB = FileBuffer{
				Pos:      tt.currentLine,
				Length:   len(tt.content),
				Contents: make([]string, len(tt.content)),
				Meta:     make(map[string]string),
			}
			copy(FB.Contents, tt.content)
			Line = tt.currentLine

			// Test change command (this will need to be implemented)
			err := executeChange(tt.rangeStr, tt.newLines)
			
			if tt.expectedError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if !reflect.DeepEqual(FB.Contents, tt.expectedContent) {
					t.Errorf("Expected content %v, got %v", tt.expectedContent, FB.Contents)
				}
			}
		})
	}
}

// Test Move Command
func TestMoveCommand(t *testing.T) {
	tests := []struct {
		name            string
		content         []string
		sourceRange     string
		destAddress     string
		expectedContent []string
		expectedError   bool
	}{
		{
			name:            "move single line",
			content:         []string{"line1", "line2", "line3", "line4"},
			sourceRange:     "2",
			destAddress:     "4",
			expectedContent: []string{"line1", "line3", "line4", "line2"},
			expectedError:   false,
		},
		{
			name:            "move multiple lines",
			content:         []string{"line1", "line2", "line3", "line4", "line5"},
			sourceRange:     "2,3",
			destAddress:     "5",
			expectedContent: []string{"line1", "line4", "line5", "line2", "line3"},
			expectedError:   false,
		},
		{
			name:            "move to beginning",
			content:         []string{"line1", "line2", "line3", "line4"},
			sourceRange:     "4",
			destAddress:     "0",
			expectedContent: []string{"line4", "line1", "line2", "line3"},
			expectedError:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			FB = FileBuffer{
				Pos:      0,
				Length:   len(tt.content),
				Contents: make([]string, len(tt.content)),
				Meta:     make(map[string]string),
			}
			copy(FB.Contents, tt.content)
			Line = 0

			// Test move command (this will need to be implemented)
			err := executeMove(tt.sourceRange, tt.destAddress)
			
			if tt.expectedError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if !reflect.DeepEqual(FB.Contents, tt.expectedContent) {
					t.Errorf("Expected content %v, got %v", tt.expectedContent, FB.Contents)
				}
			}
		})
	}
}

// Test Copy/Transfer Command
func TestCopyCommand(t *testing.T) {
	tests := []struct {
		name            string
		content         []string
		sourceRange     string
		destAddress     string
		expectedContent []string
		expectedError   bool
	}{
		{
			name:            "copy single line",
			content:         []string{"line1", "line2", "line3"},
			sourceRange:     "1",
			destAddress:     "3",
			expectedContent: []string{"line1", "line2", "line3", "line1"},
			expectedError:   false,
		},
		{
			name:            "copy multiple lines",
			content:         []string{"line1", "line2", "line3", "line4"},
			sourceRange:     "1,2",
			destAddress:     "4",
			expectedContent: []string{"line1", "line2", "line3", "line4", "line1", "line2"},
			expectedError:   false,
		},
		{
			name:            "copy to beginning",
			content:         []string{"line1", "line2", "line3"},
			sourceRange:     "3",
			destAddress:     "0",
			expectedContent: []string{"line3", "line1", "line2", "line3"},
			expectedError:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			FB = FileBuffer{
				Pos:      0,
				Length:   len(tt.content),
				Contents: make([]string, len(tt.content)),
				Meta:     make(map[string]string),
			}
			copy(FB.Contents, tt.content)
			Line = 0

			// Test copy command (this will need to be implemented)
			err := executeCopy(tt.sourceRange, tt.destAddress)
			
			if tt.expectedError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if !reflect.DeepEqual(FB.Contents, tt.expectedContent) {
					t.Errorf("Expected content %v, got %v", tt.expectedContent, FB.Contents)
				}
			}
		})
	}
}

// Test Join Command
func TestJoinCommand(t *testing.T) {
	tests := []struct {
		name            string
		content         []string
		rangeStr        string
		expectedContent []string
		expectedError   bool
	}{
		{
			name:            "join two lines",
			content:         []string{"line1", "line2", "line3"},
			rangeStr:        "1,2",
			expectedContent: []string{"line1line2", "line3"},
			expectedError:   false,
		},
		{
			name:            "join multiple lines",
			content:         []string{"line1", "line2", "line3", "line4"},
			rangeStr:        "1,3",
			expectedContent: []string{"line1line2line3", "line4"},
			expectedError:   false,
		},
		{
			name:            "join with current line",
			content:         []string{"line1", "line2", "line3"},
			rangeStr:        ".,+1",
			expectedContent: []string{"line1line2", "line3"},
			expectedError:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			FB = FileBuffer{
				Pos:      0,
				Length:   len(tt.content),
				Contents: make([]string, len(tt.content)),
				Meta:     make(map[string]string),
			}
			copy(FB.Contents, tt.content)
			Line = 0

			// Test join command (this will need to be implemented)
			err := executeJoin(tt.rangeStr)
			
			if tt.expectedError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if !reflect.DeepEqual(FB.Contents, tt.expectedContent) {
					t.Errorf("Expected content %v, got %v", tt.expectedContent, FB.Contents)
				}
			}
		})
	}
}

// Test Mark and Goto Mark Commands
func TestMarkCommands(t *testing.T) {
	tests := []struct {
		name         string
		content      []string
		currentLine  int
		markLine     int
		markChar     string
		expectedMark map[string]int
		expectedError bool
	}{
		{
			name:         "mark single line",
			content:      []string{"line1", "line2", "line3"},
			currentLine:  1,
			markLine:     1,
			markChar:     "a",
			expectedMark: map[string]int{"a": 1},
			expectedError: false,
		},
		{
			name:         "mark multiple lines",
			content:      []string{"line1", "line2", "line3", "line4"},
			currentLine:  0,
			markLine:     2,
			markChar:     "b",
			expectedMark: map[string]int{"b": 2},
			expectedError: false,
		},
		{
			name:          "invalid mark character",
			content:       []string{"line1", "line2", "line3"},
			currentLine:   1,
			markLine:      1,
			markChar:      "1",
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			FB = FileBuffer{
				Pos:      tt.currentLine,
				Length:   len(tt.content),
				Contents: tt.content,
				Meta:     make(map[string]string),
			}
			Line = tt.currentLine

			// Test mark command (this will need to be implemented)
			err := executeMark(tt.markLine, tt.markChar)
			
			if tt.expectedError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				// Check if mark was set correctly
				if !checkMark(tt.markChar, tt.expectedMark[tt.markChar]) {
					t.Errorf("Mark %s not set correctly", tt.markChar)
				}
			}
		})
	}
}

// Test Undo Command
func TestUndoCommand(t *testing.T) {
	tests := []struct {
		name            string
		content         []string
		operations      []string
		expectedContent []string
		expectedError   bool
	}{
		{
			name:            "undo delete",
			content:         []string{"line1", "line2", "line3"},
			operations:      []string{"2d", "u"},
			expectedContent: []string{"line1", "line2", "line3"},
			expectedError:   false,
		},
		{
			name:            "undo insert",
			content:         []string{"line1", "line3"},
			operations:      []string{"1i", "inserted line", ".", "u"},
			expectedContent: []string{"line1", "line3"},
			expectedError:   false,
		},
		{
			name:            "undo substitute",
			content:         []string{"hello world", "goodbye world"},
			operations:      []string{"s/hello/hi/", "u"},
			expectedContent: []string{"hello world", "goodbye world"},
			expectedError:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			FB = FileBuffer{
				Pos:      0,
				Length:   len(tt.content),
				Contents: make([]string, len(tt.content)),
				Meta:     make(map[string]string),
			}
			copy(FB.Contents, tt.content)
			Line = 0

			// Test undo command (this will need to be implemented)
			err := executeUndo()
			
			if tt.expectedError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if !reflect.DeepEqual(FB.Contents, tt.expectedContent) {
					t.Errorf("Expected content %v, got %v", tt.expectedContent, FB.Contents)
				}
			}
		})
	}
}

// Test Cut and Paste (Yank and Put) Commands
func TestCutPasteCommands(t *testing.T) {
	tests := []struct {
		name            string
		content         []string
		yankRange       string
		putAddress      string
		expectedContent []string
		expectedError   bool
	}{
		{
			name:            "yank and put single line",
			content:         []string{"line1", "line2", "line3"},
			yankRange:       "1",
			putAddress:      "3",
			expectedContent: []string{"line1", "line2", "line3", "line1"},
			expectedError:   false,
		},
		{
			name:            "yank and put multiple lines",
			content:         []string{"line1", "line2", "line3", "line4"},
			yankRange:       "1,2",
			putAddress:      "4",
			expectedContent: []string{"line1", "line2", "line3", "line4", "line1", "line2"},
			expectedError:   false,
		},
		{
			name:            "yank and put at beginning",
			content:         []string{"line1", "line2", "line3"},
			yankRange:       "3",
			putAddress:      "0",
			expectedContent: []string{"line3", "line1", "line2", "line3"},
			expectedError:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			FB = FileBuffer{
				Pos:      0,
				Length:   len(tt.content),
				Contents: make([]string, len(tt.content)),
				Meta:     make(map[string]string),
			}
			copy(FB.Contents, tt.content)
			Line = 0

			// Test yank command (this will need to be implemented)
			err := executeYank(tt.yankRange)
			if err != nil {
				t.Errorf("Yank failed: %v", err)
				return
			}

			// Test put command (this will need to be implemented)
			err = executePut(tt.putAddress)
			
			if tt.expectedError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if !reflect.DeepEqual(FB.Contents, tt.expectedContent) {
					t.Errorf("Expected content %v, got %v", tt.expectedContent, FB.Contents)
				}
			}
		})
	}
}

// Test Regular Expression Search
func TestRegexSearch(t *testing.T) {
	tests := []struct {
		name          string
		content       []string
		currentLine   int
		pattern       string
		forward       bool
		expectedLine  int
		expectedError bool
	}{
		{
			name:          "forward search",
			content:       []string{"apple", "banana", "cherry", "date"},
			currentLine:   0,
			pattern:       "ch",
			forward:       true,
			expectedLine:  2,
			expectedError: false,
		},
		{
			name:          "backward search",
			content:       []string{"apple", "banana", "cherry", "date"},
			currentLine:   3,
			pattern:       "ba",
			forward:       false,
			expectedLine:  1,
			expectedError: false,
		},
		{
			name:          "search with wrap around",
			content:       []string{"apple", "banana", "cherry", "date"},
			currentLine:   3,
			pattern:       "app",
			forward:       true,
			expectedLine:  0,
			expectedError: false,
		},
		{
			name:          "search not found",
			content:       []string{"apple", "banana", "cherry", "date"},
			currentLine:   0,
			pattern:       "xyz",
			forward:       true,
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			FB = FileBuffer{
				Pos:      tt.currentLine,
				Length:   len(tt.content),
				Contents: tt.content,
				Meta:     make(map[string]string),
			}
			Line = tt.currentLine

			// Test regex search (this will need to be implemented)
			line, err := executeRegexSearch(tt.pattern, tt.forward)
			
			if tt.expectedError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if line != tt.expectedLine {
					t.Errorf("Expected line %d, got %d", tt.expectedLine, line)
				}
			}
		})
	}
}

// Test Shell Command Execution
func TestShellCommand(t *testing.T) {
	tests := []struct {
		name          string
		command       string
		expectOutput  bool
		expectedError bool
	}{
		{
			name:          "simple echo command",
			command:       "echo hello",
			expectOutput:  true,
			expectedError: false,
		},
		{
			name:          "ls command",
			command:       "ls /tmp",
			expectOutput:  true,
			expectedError: false,
		},
		{
			name:          "invalid command",
			command:       "nonexistentcommand",
			expectOutput:  false,
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test shell command (this will need to be implemented)
			output, err := executeShellCommand(tt.command)
			
			if tt.expectedError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if tt.expectOutput && len(output) == 0 {
					t.Errorf("Expected output but got none")
				}
			}
		})
	}
}



// Test Write Append Command
func TestWriteAppendCommand(t *testing.T) {
	tests := []struct {
		name          string
		content       []string
		rangeStr      string
		filename      string
		existingFile  string
		expectedError bool
	}{
		{
			name:          "write append all lines",
			content:       []string{"line1", "line2", "line3"},
			rangeStr:      "%",
			filename:      "test_append.txt",
			existingFile:  "existing content\n",
			expectedError: false,
		},
		{
			name:          "write append specific range",
			content:       []string{"line1", "line2", "line3"},
			rangeStr:      "1,2",
			filename:      "test_append2.txt",
			existingFile:  "initial\n",
			expectedError: false,
		},
		{
			name:          "write append to new file",
			content:       []string{"line1", "line2"},
			rangeStr:      "%",
			filename:      "test_new.txt",
			existingFile:  "",
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			FB = FileBuffer{
				Pos:      0,
				Length:   len(tt.content),
				Contents: tt.content,
				Meta:     make(map[string]string),
			}
			Line = 0

			// Create existing file if needed
			if tt.existingFile != "" {
				if err := os.WriteFile(tt.filename, []byte(tt.existingFile), 0644); err != nil {
					t.Fatalf("Failed to create existing file: %v", err)
				}
			}
			defer os.Remove(tt.filename)

			// Test write append command
			err := executeWriteAppend(tt.rangeStr, tt.filename)
			
			if tt.expectedError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				// Check if file was created/appended
				if _, err := os.Stat(tt.filename); os.IsNotExist(err) {
					t.Errorf("File was not created")
				}
			}
		})
	}
}

// Test Write Command
func TestWriteCommand(t *testing.T) {
	tests := []struct {
		name          string
		content       []string
		rangeStr      string
		filename      string
		expectedError bool
	}{
		{
			name:          "write all lines",
			content:       []string{"line1", "line2", "line3"},
			rangeStr:      "%",
			filename:      "test_output.txt",
			expectedError: false,
		},
		{
			name:          "write specific range",
			content:       []string{"line1", "line2", "line3", "line4"},
			rangeStr:      "2,3",
			filename:      "test_range.txt",
			expectedError: false,
		},
		{
			name:          "write to existing file",
			content:       []string{"new content"},
			rangeStr:      "1",
			filename:      "existing.txt",
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			FB = FileBuffer{
				Pos:      0,
				Length:   len(tt.content),
				Contents: tt.content,
				Meta:     make(map[string]string),
			}
			Line = 0

			// Test write command (this will need to be implemented)
			err := executeWrite(tt.rangeStr, tt.filename)
			
			if tt.expectedError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				// Clean up test file
				os.Remove(tt.filename)
			}
		})
	}
}

// Test Read Command
func TestReadCommand(t *testing.T) {
	tests := []struct {
		name            string
		content         []string
		address         string
		filename        string
		readContent     []string
		expectedContent []string
		expectedError   bool
	}{
		{
			name:            "read into empty buffer",
			content:         []string{},
			address:         "0",
			filename:        "test_read.txt",
			readContent:     []string{"read line 1", "read line 2"},
			expectedContent: []string{"read line 1", "read line 2"},
			expectedError:   false,
		},
		{
			name:            "read after specific line",
			content:         []string{"line1", "line2"},
			address:         "1",
			filename:        "test_read.txt",
			readContent:     []string{"inserted line"},
			expectedContent: []string{"line1", "inserted line", "line2"},
			expectedError:   false,
		},
		{
			name:          "read nonexistent file",
			content:       []string{"line1"},
			address:       "1",
			filename:      "nonexistent.txt",
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			FB = FileBuffer{
				Pos:      0,
				Length:   len(tt.content),
				Contents: make([]string, len(tt.content)),
				Meta:     make(map[string]string),
			}
			copy(FB.Contents, tt.content)
			Line = 0

			// Create test file if readContent is provided
			if tt.readContent != nil {
				content := strings.Join(tt.readContent, "\n")
				if err := os.WriteFile(tt.filename, []byte(content), 0644); err != nil {
					t.Fatalf("Failed to create test file: %v", err)
				}
				defer os.Remove(tt.filename)
			}

			// Test read command (this will need to be implemented)
			err := executeRead(tt.address, tt.filename)
			
			if tt.expectedError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if !reflect.DeepEqual(FB.Contents, tt.expectedContent) {
					t.Errorf("Expected content %v, got %v", tt.expectedContent, FB.Contents)
				}
			}
		})
	}
}

// Test List Command (unambiguous printing)
func TestListCommand(t *testing.T) {
	tests := []struct {
		name            string
		content         []string
		rangeStr        string
		expectedOutput  []string
		expectedError   bool
	}{
		{
			name:            "list single line",
			content:         []string{"hello\tworld"},
			rangeStr:        "1",
			expectedOutput:  []string{"hello\\tworld$"},
			expectedError:   false,
		},
		{
			name:            "list multiple lines",
			content:         []string{"line1", "line2\t", "line3"},
			rangeStr:        "1,3",
			expectedOutput:  []string{"line1$", "line2\\t$", "line3$"},
			expectedError:   false,
		},
		{
			name:            "list with special characters",
			content:         []string{"hello\nworld", "tab\there"},
			rangeStr:        "%",
			expectedOutput:  []string{"hello\\nworld$", "tab\\there$"},
			expectedError:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			FB = FileBuffer{
				Pos:      0,
				Length:   len(tt.content),
				Contents: tt.content,
				Meta:     make(map[string]string),
			}
			Line = 0

			// Test list command (this will need to be implemented)
			output, err := executeList(tt.rangeStr)
			
			if tt.expectedError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if !reflect.DeepEqual(output, tt.expectedOutput) {
					t.Errorf("Expected output %v, got %v", tt.expectedOutput, output)
				}
			}
		})
	}
}

// Test Number Command (print with line numbers)
func TestNumberCommand(t *testing.T) {
	tests := []struct {
		name            string
		content         []string
		rangeStr        string
		expectedOutput  []string
		expectedError   bool
	}{
		{
			name:            "number single line",
			content:         []string{"hello world"},
			rangeStr:        "1",
			expectedOutput:  []string{"1\thello world"},
			expectedError:   false,
		},
		{
			name:            "number multiple lines",
			content:         []string{"line1", "line2", "line3"},
			rangeStr:        "1,3",
			expectedOutput:  []string{"1\tline1", "2\tline2", "3\tline3"},
			expectedError:   false,
		},
		{
			name:            "number with range",
			content:         []string{"line1", "line2", "line3", "line4"},
			rangeStr:        "2,4",
			expectedOutput:  []string{"2\tline2", "3\tline3", "4\tline4"},
			expectedError:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			FB = FileBuffer{
				Pos:      0,
				Length:   len(tt.content),
				Contents: tt.content,
				Meta:     make(map[string]string),
			}
			Line = 0

			// Test number command (this will need to be implemented)
			output, err := executeNumber(tt.rangeStr)
			
			if tt.expectedError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if !reflect.DeepEqual(output, tt.expectedOutput) {
					t.Errorf("Expected output %v, got %v", tt.expectedOutput, output)
				}
			}
		})
	}
}

// Test Equals Command (print line number)
func TestEqualsCommand(t *testing.T) {
	tests := []struct {
		name          string
		content       []string
		address       string
		expectedNum   int
		expectedError bool
	}{
		{
			name:          "equals current line",
			content:       []string{"line1", "line2", "line3"},
			address:       ".",
			expectedNum:   1,
			expectedError: false,
		},
		{
			name:          "equals last line",
			content:       []string{"line1", "line2", "line3"},
			address:       "$",
			expectedNum:   3,
			expectedError: false,
		},
		{
			name:          "equals specific line",
			content:       []string{"line1", "line2", "line3"},
			address:       "2",
			expectedNum:   2,
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			FB = FileBuffer{
				Pos:      0,
				Length:   len(tt.content),
				Contents: tt.content,
				Meta:     make(map[string]string),
			}
			Line = 0

			// Test equals command (this will need to be implemented)
			num, err := executeEquals(tt.address)
			
			if tt.expectedError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if num != tt.expectedNum {
					t.Errorf("Expected line number %d, got %d", tt.expectedNum, num)
				}
			}
		})
	}
}

