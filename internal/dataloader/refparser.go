package dataloader

import (
	"regexp"
	"strconv"
	"strings"
)

// ParsedRef represents a single resolved scripture reference.
type ParsedRef struct {
	Volume   string // e.g. "ot"
	Slug     string // e.g. "gen"
	Chapter  int
	Verse    int
	EndVerse int // 0 if single verse, >0 if range
}

// ParseResult contains all parsed references and any errors encountered.
type ParseResult struct {
	Refs   []ParsedRef
	Errors []string
}

// RefParser parses scripture reference strings into structured references.
type RefParser struct {
	abbrevs map[string]BookInfo
}

// NewRefParser creates a new reference parser with the given abbreviation map.
func NewRefParser(abbrevs map[string]BookInfo) RefParser {
	return RefParser{abbrevs: abbrevs}
}

// Parse parses a scripture reference string that may contain multiple references
// separated by semicolons, with optional parenthetical cross-references.
//
// Supported formats:
//   - "Gen. 1:1"
//   - "Gen. 1:1-3"
//   - "Gen. 1:1; 2:3"              (book inheritance)
//   - "Mosiah 4:2; Morm. 9:11"     (compound)
//   - "D&C 76:71 (70-71)"          (parenthetical context - ignored)
//   - "Matt. 23:12 (Luke 14:11)"   (parenthetical cross-refs)
//   - "JS—M 1:12"                  (special books)
func (rp RefParser) Parse(s string) ParseResult {
	s = strings.TrimSpace(s)
	s = strings.TrimRight(s, ".")
	if s == "" {
		return ParseResult{}
	}

	return rp.parseCompound(s)
}

// parseCompound parses a compound reference string with semicolons.
// e.g. "Gen. 1:1; 2:3; Ex. 4:5" -> 3 refs (second inherits Gen.)
func (rp RefParser) parseCompound(s string) ParseResult {
	s = strings.TrimSpace(s)
	if s == "" {
		return ParseResult{}
	}

	var result ParseResult
	var lastBook string

	parts := splitSemicolons(s)
	for _, part := range parts {
		part = strings.TrimSpace(part)
		part = strings.TrimRight(part, ".")
		if part == "" {
			continue
		}

		refs := rp.parseSegment(part, lastBook)
		result.Refs = append(result.Refs, refs.Refs...)
		result.Errors = append(result.Errors, refs.Errors...)

		// Track last book for inheritance
		if len(refs.Refs) > 0 {
			last := refs.Refs[len(refs.Refs)-1]
			abbrev := rp.findAbbrev(last.Volume, last.Slug)
			if abbrev != "" {
				lastBook = abbrev
			}
		}
	}

	return result
}

// splitSemicolons splits on semicolons but not inside parentheses.
func splitSemicolons(s string) []string {
	var parts []string
	depth := 0
	start := 0
	for i, ch := range s {
		switch ch {
		case '(':
			depth++
		case ')':
			depth--
		case ';':
			if depth == 0 {
				parts = append(parts, s[start:i])
				start = i + 1
			}
		}
	}
	parts = append(parts, s[start:])
	return parts
}

// parseSegment parses a single semicolon-separated segment.
// It handles parenthetical content within the segment.
// e.g. "Matt. 23:12 (Luke 14:11; D&C 101:42)" or "D&C 76:71 (70-71)"
func (rp RefParser) parseSegment(s string, inheritBook string) ParseResult {
	s = strings.TrimSpace(s)
	if s == "" {
		return ParseResult{}
	}

	// Check for parenthetical content
	main, paren := splitParenthetical(s)
	if main == "" {
		main = s
		paren = ""
	}

	// Parse the main reference
	mainResult := rp.parseSingle(main, inheritBook)

	// Handle parenthetical content if present
	if paren != "" {
		parenResult := rp.handleParenthetical(paren, mainResult)
		mainResult.Refs = append(mainResult.Refs, parenResult.Refs...)
		mainResult.Errors = append(mainResult.Errors, parenResult.Errors...)
	}

	return mainResult
}

// splitParenthetical splits "main (paren)" into main and paren parts.
func splitParenthetical(s string) (string, string) {
	idx := strings.Index(s, "(")
	if idx < 0 {
		return s, ""
	}
	main := strings.TrimSpace(s[:idx])
	rest := s[idx+1:]
	// Find matching closing paren
	closeIdx := strings.LastIndex(rest, ")")
	if closeIdx < 0 {
		return main, strings.TrimSpace(rest)
	}
	return main, strings.TrimSpace(rest[:closeIdx])
}

// handleParenthetical processes parenthetical content.
// If it looks like a verse range (just numbers), it's contextual and ignored.
// Otherwise, it parses as independent cross-references.
func (rp RefParser) handleParenthetical(paren string, mainRefs ParseResult) ParseResult {
	// Check if this is just a verse range like "70-71" or "70–71" or "Appendix"
	if isVerseRangeOnly(paren) {
		return ParseResult{}
	}

	// Check for non-reference keywords
	lower := strings.ToLower(strings.TrimSpace(paren))
	if lower == "appendix" || strings.HasPrefix(lower, "see ") {
		return ParseResult{}
	}

	// Otherwise parse as independent references
	// These may contain semicolons (e.g. "Luke 14:11; D&C 101:42")
	return rp.parseCompound(paren)
}

// findAbbrev finds the abbreviation for a volume/slug pair.
func (rp RefParser) findAbbrev(volume, slug string) string {
	for abbrev, info := range rp.abbrevs {
		if info.Volume == volume && info.Slug == slug {
			return abbrev
		}
	}
	return ""
}

// isVerseRangeOnly checks if a string is just a verse range like "70-71" or "1, 3-5".
var verseRangeOnlyRe = regexp.MustCompile(`^[\d\s,\-–]+$`)

func isVerseRangeOnly(s string) bool {
	return verseRangeOnlyRe.MatchString(strings.TrimSpace(s))
}

// parseSingle parses a single reference (no semicolons, no parens).
// inheritBook is the abbreviation from a previous segment for book inheritance.
func (rp RefParser) parseSingle(s string, inheritBook string) ParseResult {
	s = strings.TrimSpace(s)
	if s == "" {
		return ParseResult{}
	}

	// Strip trailing period
	s = strings.TrimRight(s, ".")

	// Try to extract book abbreviation and chapter:verse
	book, chapterVerse, found := rp.extractBook(s)
	if !found {
		// Maybe it's just chapter:verse inheriting the previous book
		if inheritBook != "" {
			book = inheritBook
			chapterVerse = s
		} else {
			return ParseResult{Errors: []string{"cannot parse reference: " + s}}
		}
	}

	// Look up book info
	info, ok := rp.abbrevs[book]
	if !ok {
		return ParseResult{Errors: []string{"unknown book abbreviation: " + book}}
	}

	// Parse chapter:verse(s) - may contain commas for multiple verse ranges
	return rp.parseChapterVerse(chapterVerse, info)
}

// extractBook tries to find a known book abbreviation at the start of the string.
// Returns (abbreviation, remainder, found).
func (rp RefParser) extractBook(s string) (string, string, bool) {
	bestMatch := ""
	bestRemainder := ""

	for abbrev := range rp.abbrevs {
		if len(abbrev) <= len(bestMatch) {
			continue
		}
		if strings.HasPrefix(s, abbrev) {
			remainder := strings.TrimSpace(s[len(abbrev):])
			// Ensure the match isn't part of a longer word
			// (e.g. "Job" shouldn't match "John")
			if len(s) > len(abbrev) {
				next := s[len(abbrev)]
				if next != ' ' && next != '\t' {
					continue
				}
			}
			bestMatch = abbrev
			bestRemainder = remainder
		}
	}

	if bestMatch != "" {
		return bestMatch, bestRemainder, true
	}
	return "", "", false
}

// parseChapterVerse parses "chapter:verse" or "chapter:verse-endverse"
// or "chapter:v1, v2-v3" format.
func (rp RefParser) parseChapterVerse(s string, info BookInfo) ParseResult {
	s = strings.TrimSpace(s)
	if s == "" {
		return ParseResult{Errors: []string{"empty chapter:verse for " + info.Volume + "/" + info.Slug}}
	}

	// Split chapter from verse on ':'
	colonIdx := strings.Index(s, ":")
	if colonIdx < 0 {
		// Maybe just a chapter number (no verse specified)
		chapter, err := strconv.Atoi(strings.TrimSpace(s))
		if err != nil {
			return ParseResult{Errors: []string{"cannot parse chapter: " + s}}
		}
		// Reference to entire chapter - return chapter verse 1
		return ParseResult{Refs: []ParsedRef{{
			Volume:  info.Volume,
			Slug:    info.Slug,
			Chapter: chapter,
			Verse:   1,
		}}}
	}

	chapterStr := strings.TrimSpace(s[:colonIdx])
	verseStr := strings.TrimSpace(s[colonIdx+1:])

	chapter, err := strconv.Atoi(chapterStr)
	if err != nil {
		return ParseResult{Errors: []string{"cannot parse chapter number: " + chapterStr}}
	}

	// Parse verse part - may contain commas and ranges
	return rp.parseVerseList(verseStr, info, chapter)
}

// parseVerseList handles "1", "1-3", "1, 3-5, 7" verse specifications.
func (rp RefParser) parseVerseList(s string, info BookInfo, chapter int) ParseResult {
	var result ParseResult

	// Split on commas
	parts := strings.Split(s, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		ref, parseErr := parseVerseRange(part, info, chapter)
		if parseErr != "" {
			result.Errors = append(result.Errors, parseErr)
			continue
		}
		result.Refs = append(result.Refs, ref)
	}

	return result
}

// parseVerseRange parses "1" or "1-3" or "1–3" into a ParsedRef.
func parseVerseRange(s string, info BookInfo, chapter int) (ParsedRef, string) {
	s = strings.TrimSpace(s)

	// Normalize en-dash and em-dash to hyphen
	s = strings.ReplaceAll(s, "–", "-")
	s = strings.ReplaceAll(s, "—", "-")

	// Split on hyphen for range
	dashIdx := strings.Index(s, "-")
	if dashIdx < 0 {
		// Single verse
		verse, parseErr := strconv.Atoi(strings.TrimSpace(s))
		if parseErr != nil {
			return ParsedRef{}, "cannot parse verse number: " + s
		}
		return ParsedRef{
			Volume:  info.Volume,
			Slug:    info.Slug,
			Chapter: chapter,
			Verse:   verse,
		}, ""
	}

	// Range
	startStr := strings.TrimSpace(s[:dashIdx])
	endStr := strings.TrimSpace(s[dashIdx+1:])

	start, err1 := strconv.Atoi(startStr)
	end, err2 := strconv.Atoi(endStr)
	if err1 != nil || err2 != nil {
		return ParsedRef{}, "cannot parse verse range: " + s
	}

	return ParsedRef{
		Volume:   info.Volume,
		Slug:     info.Slug,
		Chapter:  chapter,
		Verse:    start,
		EndVerse: end,
	}, ""
}
