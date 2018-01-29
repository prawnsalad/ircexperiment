package worker

import "sync"

type ISupportBuilder struct {
	sync.Mutex
	entries map[string]string
}

func NewISupportBuilder() *ISupportBuilder {
	return &ISupportBuilder{
		entries: make(map[string]string),
	}
}

func (builder *ISupportBuilder) Set(name string, val string) {
	builder.Lock()
	builder.entries[name] = val
	builder.Unlock()
}

func (builder *ISupportBuilder) Empty() {
	builder.Lock()
	builder.entries = make(map[string]string)
	builder.Unlock()
}

func (builder *ISupportBuilder) AsString() string {
	builder.Lock()
	line := ""
	for name, val := range builder.entries {
		line += name
		if val != "" {
			line += "=" + val
		}
		line += " "
	}
	builder.Unlock()

	return line
}

// Join an array of strings to a single line. If the resulting line exceeds maxLen then start
// a new line.
func joinStringsWithMaxLength(lines []string, maxLen int, seperator string) []string {
	out := []string{}
	currentLine := ""

	for _, line := range lines {
		if len(currentLine+seperator+line) > maxLen {
			out = append(out, currentLine)
			currentLine = ""
		} else {
			currentLine += seperator + line
		}
	}

	if currentLine != "" {
		out = append(out, currentLine)
	}

	return out
}
