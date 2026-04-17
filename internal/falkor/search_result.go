package falkor

import "sort"

// EntityType identifies which node label a search result came from.
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
	defaultHopPenalty  = 0.05
	defaultVerseBonus  = 0.05
	defaultGraphLimit  = 5
	defaultSearchLimit = 10
	defaultKNN         = 20
)

// SearchResult is a single hit from any of the 6 embeddable node labels
// (Stage 1) or a 1-hop neighbour reached via graph traversal (Stage 2).
//
// ID is a FalkorDB string identifier (e.g. "v/ot/gen/1/1", "vg/nt/matt/5/1-5")
// — unlike the prior LibSQL implementation which used int primary keys.
type SearchResult struct {
	EntityType EntityType
	ID         string
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
	ChapterID        string

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

// SortByDistance returns a new slice sorted by ascending cosine distance.
// The input is not modified.
func SortByDistance(results []SearchResult) []SearchResult {
	sorted := make([]SearchResult, len(results))
	copy(sorted, results)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Distance < sorted[j].Distance
	})
	return sorted
}
