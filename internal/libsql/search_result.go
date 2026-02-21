package libsql

import "sort"

// EntityType identifies which database table a search result came from.
type EntityType string

const (
	EntityVerseGroup   EntityType = "verse_group"
	EntityChapter      EntityType = "chapter"
	EntityTopicalGuide EntityType = "topical_guide"
	EntityBibleDict    EntityType = "bible_dict"
	EntityIndex        EntityType = "index"
	EntityJSTPassage   EntityType = "jst_passage"
	EntityVerse        EntityType = "verse"
)

// Pipeline constants for graph traversal and heuristic re-ranking.
const (
	defaultHopPenalty = 0.05
	defaultVerseBonus = 0.05
	defaultGraphLimit = 5
)

// SearchResult represents a single vector search hit from any of the 6 entity tables.
type SearchResult struct {
	EntityType EntityType
	ID         int
	Name       string
	Text       string
	Distance   float64
	Metadata   ResultMeta
}

// ResultMeta holds entity-specific fields that vary by EntityType.
type ResultMeta struct {
	// VerseGroup fields.
	StartVerseNumber int
	EndVerseNumber   int
	ChapterID        int

	// Chapter fields.
	ChapterNumber int
	URL           string

	// JST fields.
	Book       string
	Chapter    string
	Comprises  string
	CompareRef string
	Summary    string

	// Verse fields.
	VerseNumber int
	Reference   string
}

// SortByDistance returns a new slice of results sorted by ascending cosine distance.
// The original slice is not modified.
func SortByDistance(results []SearchResult) []SearchResult {
	sorted := make([]SearchResult, len(results))
	copy(sorted, results)

	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Distance < sorted[j].Distance
	})

	return sorted
}
