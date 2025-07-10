package main

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

// Global variables for ed functionality
var (
	marks      = make(map[string]int) // Line marks (a-z)
	cutBuffer  []string              // Cut/paste buffer
	undoBuffer []string              // Undo buffer
	undoLine   int                   // Undo line position
	lastSearch string               // Last search pattern
	lastSubst  string               // Last substitution pattern
)

// Address represents a line address in ed
type Address struct {
	Type  string // "line", "current", "last", "relative", "regex", "mark"
	Value string
	Line  int
}

// Range represents a range of lines
type Range struct {
	Start Address
	End   Address
}

// parseAddress parses a single address string and returns the line number
func parseAddress(address string) (int, error) {
	if address == "" {
		return Line, nil
	}

	switch {
	case address == ".":
		return Line, nil
	case address == "$":
		return len(FB.Contents) - 1, nil
	case address[0] == '+':
		offset := 1
		if len(address) > 1 {
			var err error
			offset, err = strconv.Atoi(address[1:])
			if err != nil {
				return 0, err
			}
		}
		result := Line + offset
		if result >= len(FB.Contents) {
			return 0, fmt.Errorf("invalid address")
		}
		return result, nil
	case address[0] == '-' || address[0] == '^':
		offset := 1
		if len(address) > 1 {
			var err error
			offset, err = strconv.Atoi(address[1:])
			if err != nil {
				return 0, err
			}
		}
		result := Line - offset
		if result < 0 {
			return 0, fmt.Errorf("invalid address")
		}
		return result, nil
	case address[0] == '/':
		// Forward regex search
		if len(address) < 2 || address[len(address)-1] != '/' {
			return 0, fmt.Errorf("invalid regex")
		}
		pattern := address[1 : len(address)-1]
		return regexSearch(pattern, true)
	case address[0] == '?':
		// Backward regex search
		if len(address) < 2 || address[len(address)-1] != '?' {
			return 0, fmt.Errorf("invalid regex")
		}
		pattern := address[1 : len(address)-1]
		return regexSearch(pattern, false)
	case address[0] == '\'':
		// Mark reference
		if len(address) != 2 {
			return 0, fmt.Errorf("invalid mark")
		}
		mark := address[1:]
		if line, ok := marks[mark]; ok {
			return line, nil
		}
		return 0, fmt.Errorf("undefined mark")
	default:
		// Numeric address
		line, err := strconv.Atoi(address)
		if err != nil {
			return 0, err
		}
		if line < 0 || line > len(FB.Contents) {
			return 0, fmt.Errorf("invalid address")
		}
		if line == 0 {
			return -1, nil // Special case: address 0 means before first line
		}
		return line - 1, nil // Convert to 0-based indexing
	}
}

// parseRange parses a range string and returns start and end line numbers
func parseRange(rangeStr string) (int, int, error) {
	if rangeStr == "" {
		return Line, Line, nil
	}

	if rangeStr == "%" {
		return 0, len(FB.Contents) - 1, nil
	}

	// Handle comma-separated ranges
	if strings.Contains(rangeStr, ",") {
		parts := strings.SplitN(rangeStr, ",", 2)
		start, err := parseAddress(parts[0])
		if err != nil {
			return 0, 0, err
		}
		end, err := parseAddress(parts[1])
		if err != nil {
			return 0, 0, err
		}
		return start, end, nil
	}

	// Handle semicolon-separated ranges
	if strings.Contains(rangeStr, ";") {
		parts := strings.SplitN(rangeStr, ";", 2)
		start, err := parseAddress(parts[0])
		if err != nil {
			return 0, 0, err
		}
		Line = start // Set current line for second address
		end, err := parseAddress(parts[1])
		if err != nil {
			return 0, 0, err
		}
		return start, end, nil
	}

	// Single address
	line, err := parseAddress(rangeStr)
	if err != nil {
		return 0, 0, err
	}
	return line, line, nil
}

// regexSearch performs forward or backward regex search
func regexSearch(pattern string, forward bool) (int, error) {
	if pattern == "" {
		pattern = lastSearch
	}
	lastSearch = pattern

	regex, err := regexp.Compile(pattern)
	if err != nil {
		return 0, err
	}

	start := Line
	if forward {
		start++
	} else {
		start--
	}

	// Search in the specified direction
	for i := 0; i < len(FB.Contents); i++ {
		var index int
		if forward {
			index = (start + i) % len(FB.Contents)
		} else {
			index = (start - i + len(FB.Contents)) % len(FB.Contents)
		}

		if regex.MatchString(FB.Contents[index]) {
			return index, nil
		}
	}

	return 0, fmt.Errorf("pattern not found")
}

// executeRegexSearch implements regex search functionality
func executeRegexSearch(pattern string, forward bool) (int, error) {
	line, err := regexSearch(pattern, forward)
	if err != nil {
		return 0, err
	}
	Line = line
	return line, nil
}

// executeSubstitute implements the substitute command
func executeSubstitute(command string) error {
	// Parse substitute command: [range]s/pattern/replacement/flags
	rangeStr := ""
	substCmd := command
	
	// Check if command has a range prefix
	if strings.Contains(command, "s/") && !strings.HasPrefix(command, "s/") {
		parts := strings.SplitN(command, "s/", 2)
		if len(parts) == 2 && parts[0] != "" {
			rangeStr = parts[0]
			substCmd = "s/" + parts[1]
		}
	}
	

	
	if len(substCmd) < 4 || substCmd[0] != 's' {
		return fmt.Errorf("invalid substitute command")
	}

	delimiter := substCmd[1]
	parts := strings.Split(substCmd[2:], string(delimiter))
	if len(parts) < 2 {
		return fmt.Errorf("invalid substitute command")
	}

	pattern := parts[0]
	replacement := parts[1]
	flags := ""
	if len(parts) > 2 {
		flags = parts[2]
	}

	// Convert ed-style regex to Go-style regex
	// \( and \) become ( and ) for groups
	// \| becomes | for alternation
	// \1, \2, etc. become $1, $2, etc. for backreferences
	goPattern := strings.ReplaceAll(pattern, "\\(", "(")
	goPattern = strings.ReplaceAll(goPattern, "\\)", ")")
	goPattern = strings.ReplaceAll(goPattern, "\\|", "|")
	
	goReplacement := replacement
	// Convert \1, \2, etc. to $1, $2, etc.
	for i := 1; i <= 9; i++ {
		goReplacement = strings.ReplaceAll(goReplacement, fmt.Sprintf("\\%d", i), fmt.Sprintf("$%d", i))
	}

	regex, err := regexp.Compile(goPattern)
	if err != nil {
		return err
	}

	global := strings.Contains(flags, "g")
	changed := false

	// Determine which lines to operate on
	var start, end int
	if rangeStr != "" {
		start, end, err = parseRange(rangeStr)
		if err != nil {
			return err
		}
	} else {
		start, end = Line, Line
	}

	// Perform substitution on the specified range
	for i := start; i <= end && i < len(FB.Contents); i++ {
		if regex.MatchString(FB.Contents[i]) {
			if global {
				FB.Contents[i] = regex.ReplaceAllString(FB.Contents[i], goReplacement)
			} else {
				// Replace only the first occurrence in this line
				firstMatch := true
				FB.Contents[i] = regex.ReplaceAllStringFunc(FB.Contents[i], func(match string) string {
					if !firstMatch {
						return match
					}
					firstMatch = false
					return regex.ReplaceAllString(match, goReplacement)
				})
			}
			changed = true
		}
	}

	if !changed {
		return fmt.Errorf("no substitution performed")
	}

	return nil
}

// executeGlobal implements the global command
func executeGlobal(command string) error {
	// Parse global command: g/pattern/command
	if len(command) < 4 || command[0] != 'g' {
		return fmt.Errorf("invalid global command")
	}

	delimiter := command[1]
	// Find the second delimiter to separate pattern from command
	firstDelimPos := 2
	secondDelimPos := strings.Index(command[firstDelimPos:], string(delimiter))
	if secondDelimPos == -1 {
		return fmt.Errorf("invalid global command")
	}
	secondDelimPos += firstDelimPos

	pattern := command[firstDelimPos:secondDelimPos]
	subCommand := command[secondDelimPos+1:]

	regex, err := regexp.Compile(pattern)
	if err != nil {
		return err
	}

	// Find matching lines
	var matchingLines []int
	for i := 0; i < len(FB.Contents); i++ {
		if regex.MatchString(FB.Contents[i]) {
			matchingLines = append(matchingLines, i)
		}
	}

	if len(matchingLines) == 0 {
		return fmt.Errorf("no matches found")
	}

	// Execute command on matching lines in reverse order (for deletions)
	for i := len(matchingLines) - 1; i >= 0; i-- {
		lineNum := matchingLines[i]
		Line = lineNum
		if err := executeCommand(subCommand); err != nil {
			return err
		}
	}

	return nil
}

// executeInverseGlobal executes inverse global command (v/pattern/command)
func executeInverseGlobal(command string) error {
	// Parse inverse global command: v/pattern/command
	if len(command) < 4 || command[0] != 'v' {
		return fmt.Errorf("invalid inverse global command")
	}

	delimiter := command[1]
	// Find the second delimiter to separate pattern from command
	firstDelimPos := 2
	secondDelimPos := strings.Index(command[firstDelimPos:], string(delimiter))
	if secondDelimPos == -1 {
		return fmt.Errorf("invalid inverse global command")
	}
	secondDelimPos += firstDelimPos

	pattern := command[firstDelimPos:secondDelimPos]
	subCommand := command[secondDelimPos+1:]

	regex, err := regexp.Compile(pattern)
	if err != nil {
		return err
	}

	// Find lines that DON'T match the pattern
	var matchingLines []int
	for i := 0; i < len(FB.Contents); i++ {
		if !regex.MatchString(FB.Contents[i]) {
			matchingLines = append(matchingLines, i)
		}
	}

	if len(matchingLines) == 0 {
		return fmt.Errorf("no matches found")
	}

	// Execute command on matching lines in reverse order (for deletions)
	for i := len(matchingLines) - 1; i >= 0; i-- {
		lineNum := matchingLines[i]
		Line = lineNum
		if err := executeCommand(subCommand); err != nil {
			return err
		}
	}

	return nil
}

// executeCommand executes a single ed command
func executeCommand(command string) error {
	if len(command) == 0 {
		return nil
	}

	switch command[0] {
	case 'd':
		return executeDelete("")
	case 'p':
		return executePrint("")
	case 's':
		return executeSubstitute(command)
	case 'c':
		return executeChange(".", []string{})
	case 'a':
		return nil // append - would need input mode
	case 'i':
		return nil // insert - would need input mode
	case 'l':
		output, err := executeList(".")
		if err != nil {
			return err
		}
		for _, line := range output {
			fmt.Println(line)
		}
		return nil
	case 'n':
		output, err := executeNumber(".")
		if err != nil {
			return err
		}
		for _, line := range output {
			fmt.Println(line)
		}
		return nil
	case 'y':
		return executeYank(".")
	case 'f':
		return executeFilename(command)
	case 'h':
		return executeHelp(false)
	case 'H':
		return executeHelp(true)
	case 'W':
		return executeWriteAppend("", command[1:])
	case 'e':
		return executeEdit(command)
	case 'E':
		return executeEditUnconditional(command)
	default:
		return fmt.Errorf("unknown command: %s", command)
	}
}

// executeChange implements the change command
func executeChange(rangeStr string, newLines []string) error {
	start, end, err := parseRange(rangeStr)
	if err != nil {
		return err
	}

	// Save for undo
	undoBuffer = make([]string, len(FB.Contents))
	copy(undoBuffer, FB.Contents)
	undoLine = Line

	// Remove old lines
	before := FB.Contents[:start]
	after := FB.Contents[end+1:]

	// Insert new lines
	FB.Contents = append(before, newLines...)
	FB.Contents = append(FB.Contents, after...)

	Line = start + len(newLines) - 1
	return nil
}

// executeMove implements the move command
func executeMove(sourceRange, destAddress string) error {
	start, end, err := parseRange(sourceRange)
	if err != nil {
		return err
	}

	dest, err := parseAddress(destAddress)
	if err != nil {
		return err
	}

	// Save for undo
	undoBuffer = make([]string, len(FB.Contents))
	copy(undoBuffer, FB.Contents)
	undoLine = Line

	// Extract lines to move (make a copy)
	linesToMove := make([]string, end-start+1)
	copy(linesToMove, FB.Contents[start:end+1])

	// Remove lines from source
	FB.Contents = append(FB.Contents[:start], FB.Contents[end+1:]...)

	// Adjust destination if it's after the moved lines
	if dest > start {
		dest -= (end - start + 1)
	}

	// Insert lines at destination
	var before, after []string
	if dest == -1 {
		// Special case: dest address 0 means before first line
		before = []string{}
		after = FB.Contents
	} else {
		before = FB.Contents[:dest+1]
		after = FB.Contents[dest+1:]
	}
	FB.Contents = append(before, linesToMove...)
	FB.Contents = append(FB.Contents, after...)

	if dest == -1 {
		Line = len(linesToMove) - 1
	} else {
		Line = dest + len(linesToMove)
	}
	return nil
}

// executeCopy implements the copy/transfer command
func executeCopy(sourceRange, destAddress string) error {
	start, end, err := parseRange(sourceRange)
	if err != nil {
		return err
	}

	dest, err := parseAddress(destAddress)
	if err != nil {
		return err
	}

	// Save for undo
	undoBuffer = make([]string, len(FB.Contents))
	copy(undoBuffer, FB.Contents)
	undoLine = Line

	// Extract lines to copy
	linesToCopy := make([]string, end-start+1)
	copy(linesToCopy, FB.Contents[start:end+1])

	// Insert lines at destination
	var before, after []string
	if dest == -1 {
		// Special case: dest address 0 means before first line
		before = []string{}
		after = FB.Contents
	} else {
		before = FB.Contents[:dest+1]
		after = FB.Contents[dest+1:]
	}
	FB.Contents = append(before, linesToCopy...)
	FB.Contents = append(FB.Contents, after...)

	if dest == -1 {
		Line = len(linesToCopy) - 1
	} else {
		Line = dest + len(linesToCopy)
	}
	return nil
}

// executeJoin implements the join command
func executeJoin(rangeStr string) error {
	start, end, err := parseRange(rangeStr)
	if err != nil {
		return err
	}

	if start >= end {
		return fmt.Errorf("invalid range for join")
	}

	// Save for undo
	undoBuffer = make([]string, len(FB.Contents))
	copy(undoBuffer, FB.Contents)
	undoLine = Line

	// Join lines
	joined := strings.Join(FB.Contents[start:end+1], "")

	// Replace range with joined line
	before := FB.Contents[:start]
	after := FB.Contents[end+1:]
	FB.Contents = append(before, joined)
	FB.Contents = append(FB.Contents, after...)

	Line = start
	return nil
}

// executeMark implements the mark command
func executeMark(line int, markChar string) error {
	if len(markChar) != 1 || markChar[0] < 'a' || markChar[0] > 'z' {
		return fmt.Errorf("invalid mark character")
	}

	if line < 0 || line >= len(FB.Contents) {
		return fmt.Errorf("invalid line number")
	}

	marks[markChar] = line
	return nil
}

// checkMark checks if a mark is set to the expected line
func checkMark(markChar string, expectedLine int) bool {
	line, ok := marks[markChar]
	return ok && line == expectedLine
}

// executeUndo implements the undo command
func executeUndo() error {
	if undoBuffer == nil {
		return fmt.Errorf("nothing to undo")
	}

	// Swap current buffer with undo buffer
	temp := make([]string, len(FB.Contents))
	copy(temp, FB.Contents)
	
	FB.Contents = undoBuffer
	undoBuffer = temp

	// Restore line position
	tempLine := Line
	Line = undoLine
	undoLine = tempLine

	return nil
}

// executeFilename implements the f command (show/set filename)
func executeFilename(command string) error {
	// If just 'f', show current filename
	if len(command) == 1 {
		if filename, ok := FB.Meta["filename"]; ok {
			fmt.Println(filename)
		} else {
			fmt.Println("no filename set")
		}
		return nil
	}

	// If 'f filename', set new filename
	if len(command) > 2 && command[1] == ' ' {
		newFilename := strings.TrimSpace(command[2:])
		if FB.Meta == nil {
			FB.Meta = make(map[string]string)
		}
		FB.Meta["filename"] = newFilename
		fmt.Println(newFilename)
		return nil
	}

	return fmt.Errorf("invalid f command syntax")
}

// executeHelp implements the h and H commands (help system)
func executeHelp(verbose bool) error {
	if verbose {
		fmt.Println("Ed-compatible line editor help:")
		fmt.Println("Line addressing:")
		fmt.Println("  .      - current line")
		fmt.Println("  $      - last line")
		fmt.Println("  n      - line number n")
		fmt.Println("  +n     - n lines forward")
		fmt.Println("  -n     - n lines backward")
		fmt.Println("  /pat/  - next line matching pattern")
		fmt.Println("  ?pat?  - previous line matching pattern")
		fmt.Println("  'x     - line marked with x")
		fmt.Println("")
		fmt.Println("Range addressing:")
		fmt.Println("  ,      - from first address to second")
		fmt.Println("  ;      - from current line to second address")
		fmt.Println("  %      - entire buffer (1,$)")
		fmt.Println("")
		fmt.Println("Commands:")
		fmt.Println("  a      - append text after line")
		fmt.Println("  c      - change (replace) lines")
		fmt.Println("  d      - delete lines")
		fmt.Println("  e      - edit file")
		fmt.Println("  E      - edit file unconditionally")
		fmt.Println("  f      - show/set filename")
		fmt.Println("  g/pat/cmd - execute cmd on lines matching pattern")
		fmt.Println("  h      - show brief help")
		fmt.Println("  H      - show detailed help")
		fmt.Println("  i      - insert text before line")
		fmt.Println("  j      - join lines")
		fmt.Println("  k      - mark line")
		fmt.Println("  l      - list lines with special chars visible")
		fmt.Println("  m      - move lines")
		fmt.Println("  n      - print lines with numbers")
		fmt.Println("  p      - print lines")
		fmt.Println("  q      - quit")
		fmt.Println("  r      - read file")
		fmt.Println("  s/pat/repl/ - substitute pattern with replacement")
		fmt.Println("  t      - copy lines")
		fmt.Println("  u      - undo last change")
		fmt.Println("  v/pat/cmd - execute cmd on lines NOT matching pattern")
		fmt.Println("  w      - write buffer to file")
		fmt.Println("  W      - write and append to file")
		fmt.Println("  x      - put from cut buffer")
		fmt.Println("  y      - yank lines to cut buffer")
		fmt.Println("  =      - print line number")
		fmt.Println("  !cmd   - execute shell command")
	} else {
		fmt.Println("Available commands: a c d e E f g h H i j k l m n p q r s t u v w W x y = !")
		fmt.Println("Use 'H' for detailed help")
	}
	return nil
}

// executeEdit implements the e command (edit file)
func executeEdit(command string) error {
	// TODO: Check if buffer has unsaved changes and warn
	filename := ""
	if len(command) > 2 && command[1] == ' ' {
		filename = strings.TrimSpace(command[2:])
	}
	
	if filename == "" {
		if fn, ok := FB.Meta["filename"]; ok {
			filename = fn
		} else {
			return fmt.Errorf("no filename specified")
		}
	}
	
	return openFullFile(filename)
}

// executeEditUnconditional implements the E command (edit file unconditionally)
func executeEditUnconditional(command string) error {
	filename := ""
	if len(command) > 2 && command[1] == ' ' {
		filename = strings.TrimSpace(command[2:])
	}
	
	if filename == "" {
		if fn, ok := FB.Meta["filename"]; ok {
			filename = fn
		} else {
			return fmt.Errorf("no filename specified")
		}
	}
	
	return openFullFile(filename)
}

// executeChange implements the change command
func executeYank(rangeStr string) error {
	start, end, err := parseRange(rangeStr)
	if err != nil {
		return err
	}

	// Copy lines to cut buffer
	cutBuffer = make([]string, end-start+1)
	copy(cutBuffer, FB.Contents[start:end+1])

	return nil
}

// executePut implements the put command
func executePut(address string) error {
	if cutBuffer == nil {
		return fmt.Errorf("nothing to put")
	}

	line, err := parseAddress(address)
	if err != nil {
		return err
	}

	// Save for undo
	undoBuffer = make([]string, len(FB.Contents))
	copy(undoBuffer, FB.Contents)
	undoLine = Line

	// Insert cut buffer contents
	var before, after []string
	if line == -1 {
		// Special case: address 0 means before first line
		before = []string{}
		after = FB.Contents
	} else {
		before = FB.Contents[:line+1]
		after = FB.Contents[line+1:]
	}
	FB.Contents = append(before, cutBuffer...)
	FB.Contents = append(FB.Contents, after...)

	if line == -1 {
		Line = len(cutBuffer) - 1
	} else {
		Line = line + len(cutBuffer)
	}
	return nil
}

// executeWrite implements the write command
func executeWrite(rangeStr, filename string) error {
	start, end, err := parseRange(rangeStr)
	if err != nil {
		return err
	}

	if filename == "" {
		if fn, ok := FB.Meta["filename"]; ok {
			filename = fn
		} else {
			return fmt.Errorf("no filename specified")
		}
	}

	// Write specified lines to file
	var content []string
	if rangeStr == "" || rangeStr == "%" {
		content = FB.Contents
	} else {
		content = FB.Contents[start : end+1]
	}

	text := strings.Join(content, "\n")
	if len(content) > 0 {
		text += "\n"
	}

	return os.WriteFile(filename, []byte(text), 0644)
}

// executeWriteAppend implements the W command (write and append)
func executeWriteAppend(rangeStr, filename string) error {
	start, end, err := parseRange(rangeStr)
	if err != nil {
		return err
	}

	if filename == "" {
		if fn, ok := FB.Meta["filename"]; ok {
			filename = fn
		} else {
			return fmt.Errorf("no filename specified")
		}
	}

	// Write specified lines to file (append mode)
	var content []string
	if rangeStr == "" || rangeStr == "%" {
		content = FB.Contents
	} else {
		content = FB.Contents[start : end+1]
	}

	text := strings.Join(content, "\n")
	if len(content) > 0 {
		text += "\n"
	}

	// Open file in append mode
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(text)
	return err
}

// executeRead implements the read command
func executeRead(address, filename string) error {
	line, err := parseAddress(address)
	if err != nil {
		return err
	}

	content, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	lines := strings.Split(string(content), "\n")
	// Remove empty last line if present
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}



	// Save for undo
	undoBuffer = make([]string, len(FB.Contents))
	copy(undoBuffer, FB.Contents)
	undoLine = Line

	// Insert lines at specified address
	before := make([]string, line+1)
	copy(before, FB.Contents[:line+1])
	after := make([]string, len(FB.Contents)-line-1)
	copy(after, FB.Contents[line+1:])
	
	FB.Contents = append(before, lines...)
	FB.Contents = append(FB.Contents, after...)

	Line = line + len(lines)
	return nil
}

// executeList implements the list command (unambiguous printing)
func executeList(rangeStr string) ([]string, error) {
	start, end, err := parseRange(rangeStr)
	if err != nil {
		return nil, err
	}

	var output []string
	for i := start; i <= end && i < len(FB.Contents); i++ {
		line := FB.Contents[i]
		// Make special characters visible
		line = strings.ReplaceAll(line, "\t", "\\t")
		line = strings.ReplaceAll(line, "\n", "\\n")
		line = strings.ReplaceAll(line, "\r", "\\r")
		line += "$"
		output = append(output, line)
	}

	return output, nil
}

// executeNumber implements the number command (print with line numbers)
func executeNumber(rangeStr string) ([]string, error) {
	start, end, err := parseRange(rangeStr)
	if err != nil {
		return nil, err
	}

	var output []string
	for i := start; i <= end && i < len(FB.Contents); i++ {
		line := fmt.Sprintf("%d\t%s", i+1, FB.Contents[i])
		output = append(output, line)
	}

	return output, nil
}

// executeEquals implements the equals command (print line number)
func executeEquals(address string) (int, error) {
	line, err := parseAddress(address)
	if err != nil {
		return 0, err
	}

	return line + 1, nil // Convert back to 1-based indexing
}

// executeShellCommand implements shell command execution
func executeShellCommand(command string) (string, error) {
	// Replace % with current filename
	if filename, ok := FB.Meta["filename"]; ok {
		command = strings.ReplaceAll(command, "%", filename)
	}

	cmd := exec.Command("sh", "-c", command)
	output, err := cmd.CombinedOutput()
	return string(output), err
}

// executeDelete implements enhanced delete command
func executeDelete(rangeStr string) error {
	start, end, err := parseRange(rangeStr)
	if err != nil {
		return err
	}

	// Save for undo
	undoBuffer = make([]string, len(FB.Contents))
	copy(undoBuffer, FB.Contents)
	undoLine = Line

	// Delete lines
	before := FB.Contents[:start]
	after := FB.Contents[end+1:]
	FB.Contents = append(before, after...)

	// Adjust current line position
	if start < len(FB.Contents) {
		Line = start
	} else if len(FB.Contents) > 0 {
		Line = len(FB.Contents) - 1
	} else {
		Line = 0
	}

	return nil
}

// executePrint implements enhanced print command
func executePrint(rangeStr string) error {
	start, end, err := parseRange(rangeStr)
	if err != nil {
		return err
	}

	for i := start; i <= end && i < len(FB.Contents); i++ {
		fmt.Println(FB.Contents[i])
		Line = i
	}

	return nil
}