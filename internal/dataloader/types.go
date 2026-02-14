package dataloader

import "sync"

// ChapterJSON represents the JSON structure of a scraped chapter file.
type ChapterJSON struct {
	URL       string            `json:"url"`
	Book      string            `json:"book"`
	Chapter   int               `json:"chapter"`
	Summary   string            `json:"summary"`
	Verses    []VerseJSON       `json:"verses"`
	Footnotes map[string]FootnoteJSON `json:"footnotes"`
}

// VerseJSON represents a verse within a chapter JSON file.
type VerseJSON struct {
	Number          int      `json:"number"`
	Text            string   `json:"text"`
	FootnoteMarkers []string `json:"footnote_markers"`
}

// FootnoteJSON represents a footnote entry in a chapter JSON file.
type FootnoteJSON struct {
	Category      string `json:"category"`
	ReferenceText string `json:"reference_text"`
	Text          string `json:"text"`
}

// TGEntryJSON represents a single entry in a topical guide topic's array.
type TGEntryJSON struct {
	Phrase    string `json:"phrase,omitempty"`
	Reference string `json:"reference"`
	Key       string `json:"key,omitempty"`
}

// BDEntryJSON represents a Bible Dictionary entry.
type BDEntryJSON struct {
	Text       string   `json:"text"`
	References []string `json:"references"`
}

// JSTChapterJSON represents a chapter-level entry in jst.json.
type JSTChapterJSON struct {
	Reference string         `json:"reference"`
	Book      string         `json:"book"`
	Chapter   string         `json:"chapter"`
	Entries   []JSTEntryJSON `json:"entries"`
}

// JSTEntryJSON represents a single JST passage entry.
type JSTEntryJSON struct {
	Comprises string      `json:"comprises"`
	Compare   string      `json:"compare"`
	Summary   string      `json:"summary"`
	Verses    []JSTVerseJSON `json:"verses"`
}

// JSTVerseJSON represents a verse within a JST passage.
type JSTVerseJSON struct {
	Number int    `json:"number"`
	Text   string `json:"text"`
}

// IDXEntryJSON represents a single entry in an index topic's array.
// Same structure as TGEntryJSON but reference type uses "IDX" instead of "TG".
type IDXEntryJSON struct {
	Phrase    string `json:"phrase,omitempty"`
	Reference string `json:"reference"`
	Key       string `json:"key,omitempty"`
}

// VerseIndex provides O(1) lookup from scripture path to database verse ID.
// Path format: "{volume}/{book-slug}/{chapter}/{verse}" e.g. "ot/gen/1/1".
type VerseIndex struct {
	byPath map[string]int
}

// NewVerseIndex creates a new empty VerseIndex.
func NewVerseIndex() VerseIndex {
	return VerseIndex{byPath: make(map[string]int, 45000)}
}

// Put adds a verse path -> ID mapping.
func (vi VerseIndex) Put(volume, slug string, chapter, verse int, dbID int) {
	key := versePath(volume, slug, chapter, verse)
	vi.byPath[key] = dbID
}

// Get returns the database ID for a verse path, or 0 and false if not found.
func (vi VerseIndex) Get(volume, slug string, chapter, verse int) (int, bool) {
	key := versePath(volume, slug, chapter, verse)
	id, ok := vi.byPath[key]
	return id, ok
}

// Len returns the number of indexed verses.
func (vi VerseIndex) Len() int {
	return len(vi.byPath)
}

func versePath(volume, slug string, chapter, verse int) string {
	return volume + "/" + slug + "/" + itoa(chapter) + "/" + itoa(verse)
}

// itoa converts an int to string without importing strconv in the hot path.
func itoa(n int) string {
	if n < 10 {
		return string(rune('0' + n))
	}
	// For larger numbers, use a simple approach
	if n < 0 {
		return "-" + itoa(-n)
	}
	result := ""
	for n > 0 {
		result = string(rune('0'+n%10)) + result
		n /= 10
	}
	return result
}

// JSTIndex provides O(1) lookup from a normalized JST reference to DB ID.
// Key format: "{book}/{chapter}" e.g. "genesis/10" or "1-samuel/16".
type JSTIndex struct {
	byRef map[string][]jstEntry
}

type jstEntry struct {
	dbID      int
	comprises string
}

// NewJSTIndex creates a new empty JSTIndex.
func NewJSTIndex() JSTIndex {
	return JSTIndex{byRef: make(map[string][]jstEntry, 120)}
}

// Put adds a JST passage to the index.
func (ji JSTIndex) Put(bookSlug, chapter, comprises string, dbID int) {
	key := bookSlug + "/" + chapter
	ji.byRef[key] = append(ji.byRef[key], jstEntry{dbID: dbID, comprises: comprises})
}

// Get returns the database ID for a JST passage matching the given book/chapter/comprises.
func (ji JSTIndex) Get(bookSlug, chapter string) []jstEntry {
	key := bookSlug + "/" + chapter
	return ji.byRef[key]
}

// LoadStats tracks counts and warnings during the loading process.
type LoadStats struct {
	Volumes        int
	Books          int
	Chapters       int
	Verses         int
	TGEntries      int
	BDEntries      int
	IDXEntries     int
	JSTPassages    int
	CrossRefs      int
	VerseTGRefs    int
	VerseBDRefs    int
	VerseJSTRefs   int
	TGVerseRefs    int
	BDVerseRefs    int
	IDXVerseRefs   int
	TGSeeAlso      int
	BDSeeAlso      int
	IDXSeeAlso     int
	IDXTGRefs      int
	IDXBDRefs      int
	TGBDRefs       int
	JSTCompares    int
	VerseGroups    int

	// Phase 6: embedding counts
	EmbVerseGroups int
	EmbChapters    int
	EmbTGEntries   int
	EmbBDEntries   int
	EmbIDXEntries  int
	EmbJSTPassages int

	mu       sync.Mutex
	Warnings []string
}

// Warn adds a warning message. Safe for concurrent use.
func (s *LoadStats) Warn(msg string) {
	s.mu.Lock()
	s.Warnings = append(s.Warnings, msg)
	s.mu.Unlock()
}
