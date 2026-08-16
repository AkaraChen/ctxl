package store

import (
	"regexp"
	"strings"
	"unicode"
)

var (
	headingLine = regexp.MustCompile(`^(#{1,6})\s+(.+?)\s*$`)
	fenceLine   = regexp.MustCompile(`^(\s{0,3})(` + "`" + `{3,}|~{3,})(.*)$`)
)

type mdSection struct {
	Heading string
	Level   int
	Body    string
	Start   int
	End     int
}

func parseSections(content string) []mdSection {
	lines := strings.Split(content, "\n")
	starts := lineStarts(content)
	type head struct {
		heading       string
		level         int
		start         int
		headingLength int
	}
	var heads []head
	var fence string
	for i, line := range lines {
		line = strings.TrimRightFunc(line, func(r rune) bool { return r == '\r' })
		if m := fenceLine.FindStringSubmatch(line); m != nil {
			tok := m[2]
			ch := string(tok[0])
			if fence == "" {
				fence = ch
			} else if strings.HasPrefix(tok, fence) && len(tok) >= 3 {
				fence = ""
			}
			continue
		}
		if fence != "" {
			continue
		}
		if m := headingLine.FindStringSubmatch(line); m != nil {
			heads = append(heads, head{
				heading:       strings.TrimSpace(m[2]),
				level:         len(m[1]),
				start:         starts[i],
				headingLength: len(m[0]),
			})
		}
	}
	out := make([]mdSection, 0, len(heads))
	for i, h := range heads {
		end := len(content)
		if i+1 < len(heads) {
			end = heads[i+1].start
		}
		body := content[h.start+h.headingLength : end]
		body = strings.TrimLeft(body, "\r\n")
		body = strings.TrimRight(body, " \t\r\n")
		out = append(out, mdSection{Heading: h.heading, Level: h.level, Body: body, Start: h.start, End: end})
	}
	return out
}

func lineStarts(content string) []int {
	starts := []int{0}
	for i := 0; i < len(content); i++ {
		if content[i] == '\n' {
			starts = append(starts, i+1)
		}
	}
	return starts
}

func findSection(sections []mdSection, heading string) (mdSection, bool) {
	want := normalizeHeading(heading)
	for _, s := range sections {
		if normalizeHeading(s.Heading) == want {
			return s, true
		}
	}
	return mdSection{}, false
}

func normalizeHeading(h string) string {
	return strings.ToLower(strings.TrimSpace(h))
}

func renderSection(heading, body string, level int) string {
	if level < 1 {
		level = 2
	}
	prefix := strings.Repeat("#", level)
	trimmed := strings.TrimRightFunc(strings.TrimLeft(body, "\r\n"), unicode.IsSpace)
	return prefix + " " + heading + "\n\n" + trimmed + "\n"
}

func upsertSection(content, heading, body string) string {
	block := renderSection(heading, body, 2)
	if strings.TrimSpace(content) == "" {
		return block
	}
	secs := parseSections(content)
	target, ok := findSection(secs, heading)
	if !ok {
		trimmed := strings.TrimRight(content, "\r\n")
		return trimmed + "\n\n" + block
	}
	before := strings.TrimRight(content[:target.Start], "\r\n")
	after := strings.TrimLeft(content[target.End:], "\r\n")
	middle := block
	if before != "" {
		middle = before + "\n\n" + block
	}
	if after != "" {
		return middle + "\n" + after
	}
	return middle
}
