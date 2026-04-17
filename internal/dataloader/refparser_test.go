package dataloader

import (
	"testing"
)

func testAbbrevs() map[string]BookInfo {
	return buildAbbreviationMap()
}

func TestRefParser_Parse(t *testing.T) {
	rp := NewRefParser(testAbbrevs())

	tests := []struct {
		name     string
		input    string
		wantRefs []ParsedRef
		wantErrs int
	}{
		{
			name:  "single verse",
			input: "Gen. 1:1",
			wantRefs: []ParsedRef{
				{Volume: "ot", Slug: "gen", Chapter: 1, Verse: 1},
			},
		},
		{
			name:  "single verse with trailing period",
			input: "Gen. 1:1.",
			wantRefs: []ParsedRef{
				{Volume: "ot", Slug: "gen", Chapter: 1, Verse: 1},
			},
		},
		{
			name:  "verse range with hyphen",
			input: "Gen. 1:1-3",
			wantRefs: []ParsedRef{
				{Volume: "ot", Slug: "gen", Chapter: 1, Verse: 1, EndVerse: 3},
			},
		},
		{
			name:  "verse range with en-dash",
			input: "Ex. 6:16–20",
			wantRefs: []ParsedRef{
				{Volume: "ot", Slug: "ex", Chapter: 6, Verse: 16, EndVerse: 20},
			},
		},
		{
			name:  "compound semicolon different books",
			input: "Mosiah 4:2; Morm. 9:11",
			wantRefs: []ParsedRef{
				{Volume: "bofm", Slug: "mosiah", Chapter: 4, Verse: 2},
				{Volume: "bofm", Slug: "morm", Chapter: 9, Verse: 11},
			},
		},
		{
			name:  "book inheritance across semicolons",
			input: "D&C 14:9; 76:24",
			wantRefs: []ParsedRef{
				{Volume: "dc-testament", Slug: "dc", Chapter: 14, Verse: 9},
				{Volume: "dc-testament", Slug: "dc", Chapter: 76, Verse: 24},
			},
		},
		{
			name:  "parenthetical cross-refs",
			input: "Matt. 23:12 (Luke 14:11; D&C 101:42)",
			wantRefs: []ParsedRef{
				{Volume: "nt", Slug: "matt", Chapter: 23, Verse: 12},
				{Volume: "nt", Slug: "luke", Chapter: 14, Verse: 11},
				{Volume: "dc-testament", Slug: "dc", Chapter: 101, Verse: 42},
			},
		},
		{
			name:  "parenthetical range context",
			input: "D&C 76:71 (70-71)",
			wantRefs: []ParsedRef{
				{Volume: "dc-testament", Slug: "dc", Chapter: 76, Verse: 71},
			},
		},
		{
			name:  "JS—M special book",
			input: "JS—M 1:12",
			wantRefs: []ParsedRef{
				{Volume: "pgp", Slug: "js-m", Chapter: 1, Verse: 12},
			},
		},
		{
			name:  "A of F special book",
			input: "A of F 1:4",
			wantRefs: []ParsedRef{
				{Volume: "pgp", Slug: "a-of-f", Chapter: 1, Verse: 4},
			},
		},
		{
			name:  "W of M special book",
			input: "W of M 1:7",
			wantRefs: []ParsedRef{
				{Volume: "bofm", Slug: "w-of-m", Chapter: 1, Verse: 7},
			},
		},
		{
			name:  "verse list with commas",
			input: "1 Chr. 3:19 (17–19); Ezra 2:2 (1–2); Hag. 1:1",
			wantRefs: []ParsedRef{
				{Volume: "ot", Slug: "1-chr", Chapter: 3, Verse: 19},
				{Volume: "ot", Slug: "ezra", Chapter: 2, Verse: 2},
				{Volume: "ot", Slug: "hag", Chapter: 1, Verse: 1},
			},
		},
		{
			name:  "complex compound with ranges",
			input: "D&C 76:71 (70–71); 88:45; Moses 2:16; Abr. 4:16",
			wantRefs: []ParsedRef{
				{Volume: "dc-testament", Slug: "dc", Chapter: 76, Verse: 71},
				{Volume: "dc-testament", Slug: "dc", Chapter: 88, Verse: 45},
				{Volume: "pgp", Slug: "moses", Chapter: 2, Verse: 16},
				{Volume: "pgp", Slug: "abr", Chapter: 4, Verse: 16},
			},
		},
		{
			name:  "1 Nephi",
			input: "1 Ne. 1:1",
			wantRefs: []ParsedRef{
				{Volume: "bofm", Slug: "1-ne", Chapter: 1, Verse: 1},
			},
		},
		{
			name:  "2 Corinthians range",
			input: "2 Cor. 5:17-21",
			wantRefs: []ParsedRef{
				{Volume: "nt", Slug: "2-cor", Chapter: 5, Verse: 17, EndVerse: 21},
			},
		},
		{
			name:  "Psalms",
			input: "Ps. 119:105",
			wantRefs: []ParsedRef{
				{Volume: "ot", Slug: "ps", Chapter: 119, Verse: 105},
			},
		},
		{
			name:  "Revelation",
			input: "Rev. 21:4",
			wantRefs: []ParsedRef{
				{Volume: "nt", Slug: "rev", Chapter: 21, Verse: 4},
			},
		},
		{
			name:     "empty string",
			input:    "",
			wantRefs: nil,
		},
		{
			name:     "unknown book",
			input:    "Foobar 1:1",
			wantErrs: 1,
		},
		{
			name:  "OD (Official Declaration)",
			input: "OD 1:1",
			wantRefs: []ParsedRef{
				{Volume: "dc-testament", Slug: "od", Chapter: 1, Verse: 1},
			},
		},
		{
			name:  "multiple semicolons with inheritance",
			input: "Alma 5:14; 7:14; 27:27",
			wantRefs: []ParsedRef{
				{Volume: "bofm", Slug: "alma", Chapter: 5, Verse: 14},
				{Volume: "bofm", Slug: "alma", Chapter: 7, Verse: 14},
				{Volume: "bofm", Slug: "alma", Chapter: 27, Verse: 27},
			},
		},
		{
			name:  "3 Nephi",
			input: "3 Ne. 12:22",
			wantRefs: []ParsedRef{
				{Volume: "bofm", Slug: "3-ne", Chapter: 12, Verse: 22},
			},
		},
		{
			name:  "Moses",
			input: "Moses 1:29",
			wantRefs: []ParsedRef{
				{Volume: "pgp", Slug: "moses", Chapter: 1, Verse: 29},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := rp.Parse(tt.input)

			if tt.wantErrs > 0 {
				if len(result.Errors) < tt.wantErrs {
					t.Errorf("expected at least %d errors, got %d: %v", tt.wantErrs, len(result.Errors), result.Errors)
				}
				return
			}

			if len(result.Errors) > 0 {
				t.Errorf("unexpected errors: %v", result.Errors)
			}

			if len(result.Refs) != len(tt.wantRefs) {
				t.Fatalf("got %d refs, want %d. Refs: %+v", len(result.Refs), len(tt.wantRefs), result.Refs)
			}

			for i, want := range tt.wantRefs {
				got := result.Refs[i]
				if got.Volume != want.Volume || got.Slug != want.Slug ||
					got.Chapter != want.Chapter || got.Verse != want.Verse ||
					got.EndVerse != want.EndVerse {
					t.Errorf("ref[%d]: got %+v, want %+v", i, got, want)
				}
			}
		})
	}
}

func TestRefParser_BookInheritance(t *testing.T) {
	rp := NewRefParser(testAbbrevs())

	// When second segment has no book, it should inherit from first
	result := rp.Parse("Isa. 53:3; 55:1")
	if len(result.Errors) > 0 {
		t.Fatalf("unexpected errors: %v", result.Errors)
	}
	if len(result.Refs) != 2 {
		t.Fatalf("expected 2 refs, got %d", len(result.Refs))
	}

	// Both should be Isaiah
	for i, ref := range result.Refs {
		if ref.Volume != "ot" || ref.Slug != "isa" {
			t.Errorf("ref[%d]: expected ot/isa, got %s/%s", i, ref.Volume, ref.Slug)
		}
	}

	if result.Refs[0].Chapter != 53 || result.Refs[0].Verse != 3 {
		t.Errorf("ref[0]: expected 53:3, got %d:%d", result.Refs[0].Chapter, result.Refs[0].Verse)
	}
	if result.Refs[1].Chapter != 55 || result.Refs[1].Verse != 1 {
		t.Errorf("ref[1]: expected 55:1, got %d:%d", result.Refs[1].Chapter, result.Refs[1].Verse)
	}
}

func TestVerseIndex(t *testing.T) {
	vi := NewVerseIndex()

	vi.Put("ot", "gen", 1, 1, "v/ot/gen/1/1")
	vi.Put("bofm", "1-ne", 3, 7, "v/bofm/1-ne/3/7")

	tests := []struct {
		name    string
		volume  string
		slug    string
		chapter int
		verse   int
		wantID  string
		wantOK  bool
	}{
		{"found gen 1:1", "ot", "gen", 1, 1, "v/ot/gen/1/1", true},
		{"found 1 ne 3:7", "bofm", "1-ne", 3, 7, "v/bofm/1-ne/3/7", true},
		{"not found", "ot", "gen", 1, 2, "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, ok := vi.Get(tt.volume, tt.slug, tt.chapter, tt.verse)
			if ok != tt.wantOK || id != tt.wantID {
				t.Errorf("Get(%s/%s/%d/%d) = (%q, %v), want (%q, %v)",
					tt.volume, tt.slug, tt.chapter, tt.verse,
					id, ok, tt.wantID, tt.wantOK)
			}
		})
	}

	if vi.Len() != 2 {
		t.Errorf("Len() = %d, want 2", vi.Len())
	}
}
