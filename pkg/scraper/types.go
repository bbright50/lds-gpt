package scraper

// Chapter represents a full chapter of scripture with metadata, verses, and footnotes.
type Chapter struct {
	URL       string              `json:"url"`
	Book      string              `json:"book"`
	Chapter   int                 `json:"chapter"`
	Summary   string              `json:"summary"`
	Verses    []Verse             `json:"verses"`
	Footnotes map[string]Footnote `json:"footnotes"`
}

// Verse represents a single verse within a chapter.
type Verse struct {
	Number          int      `json:"number"`
	Text            string   `json:"text"`
	FootnoteMarkers []string `json:"footnote_markers"`
}

// Footnote represents a single footnote entry tied to a verse marker.
type Footnote struct {
	Category      string `json:"category"`
	ReferenceText string `json:"reference_text"`
	Text          string `json:"text"`
}
