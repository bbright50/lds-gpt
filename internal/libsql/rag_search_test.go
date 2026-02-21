package libsql

import (
	"context"
	"fmt"
	"testing"

	"lds-gpt/internal/utils/vec"
)

// makeUnitEmbedding creates a 1024-dim unit vector with 1.0 at the given axis index.
func makeUnitEmbedding(axis int) []byte {
	f := make([]float32, 1024)
	f[axis] = 1.0
	return vec.Float32sToBytes(f)
}

func TestSearchVerseGroups(t *testing.T) {
	client := TestClient(t)
	ctx := context.Background()
	ec := client.Ent()

	vol := createTestVolume(t, ctx, ec)
	book := createTestBook(t, ctx, ec, vol)
	ch := createTestChapter(t, ctx, ec, book)

	// Create two verse groups with different embeddings.
	_, err := ec.VerseGroup.Create().
		SetText("Faith is the substance of things hoped for").
		SetStartVerseNumber(1).
		SetEndVerseNumber(3).
		SetChapter(ch).
		SetEmbedding(makeUnitEmbedding(0)).
		Save(ctx)
	if err != nil {
		t.Fatalf("creating verse group 1: %v", err)
	}

	_, err = ec.VerseGroup.Create().
		SetText("Repentance brings forgiveness").
		SetStartVerseNumber(4).
		SetEndVerseNumber(5).
		SetChapter(ch).
		SetEmbedding(makeUnitEmbedding(1)).
		Save(ctx)
	if err != nil {
		t.Fatalf("creating verse group 2: %v", err)
	}

	results, err := searchVerseGroups(ctx, client.Sqlx(), makeUnitEmbedding(0), 10)
	if err != nil {
		t.Fatalf("searchVerseGroups: %v", err)
	}

	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}

	// First result should be the closest (axis 0 matched).
	if results[0].EntityType != EntityVerseGroup {
		t.Errorf("expected EntityVerseGroup, got %q", results[0].EntityType)
	}
	if results[0].Text != "Faith is the substance of things hoped for" {
		t.Errorf("unexpected text: %q", results[0].Text)
	}
	if results[0].Distance > 0.01 {
		t.Errorf("expected near-zero distance for match, got %f", results[0].Distance)
	}
	if results[0].Metadata.StartVerseNumber != 1 || results[0].Metadata.EndVerseNumber != 3 {
		t.Errorf("unexpected verse range: %d-%d", results[0].Metadata.StartVerseNumber, results[0].Metadata.EndVerseNumber)
	}

	// Second result should be farther away.
	if results[1].Distance <= results[0].Distance {
		t.Errorf("expected second result farther, got distances %f and %f", results[0].Distance, results[1].Distance)
	}
}

func TestSearchChapters(t *testing.T) {
	client := TestClient(t)
	ctx := context.Background()
	ec := client.Ent()

	vol := createTestVolume(t, ctx, ec)
	book := createTestBook(t, ctx, ec, vol)

	ch, err := ec.Chapter.Create().
		SetNumber(1).
		SetSummary("Nephi begins the record").
		SetURL("https://example.com/1-ne/1").
		SetBook(book).
		SetSummaryEmbedding(makeUnitEmbedding(0)).
		Save(ctx)
	if err != nil {
		t.Fatalf("creating chapter: %v", err)
	}

	results, err := searchChapters(ctx, client.Sqlx(), makeUnitEmbedding(0), 10)
	if err != nil {
		t.Fatalf("searchChapters: %v", err)
	}

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}

	r := results[0]
	if r.EntityType != EntityChapter {
		t.Errorf("expected EntityChapter, got %q", r.EntityType)
	}
	if r.ID != ch.ID {
		t.Errorf("expected ID %d, got %d", ch.ID, r.ID)
	}
	if r.Text != "Nephi begins the record" {
		t.Errorf("unexpected text: %q", r.Text)
	}
	if r.Metadata.ChapterNumber != 1 {
		t.Errorf("expected chapter number 1, got %d", r.Metadata.ChapterNumber)
	}
	if r.Metadata.URL != "https://example.com/1-ne/1" {
		t.Errorf("unexpected URL: %q", r.Metadata.URL)
	}
}

func TestSearchTopicalGuide(t *testing.T) {
	client := TestClient(t)
	ctx := context.Background()
	ec := client.Ent()

	_, err := ec.TopicalGuideEntry.Create().
		SetName("Faith").
		SetEmbedding(makeUnitEmbedding(0)).
		Save(ctx)
	if err != nil {
		t.Fatalf("creating TG entry: %v", err)
	}

	results, err := searchTopicalGuide(ctx, client.Sqlx(), makeUnitEmbedding(0), 10)
	if err != nil {
		t.Fatalf("searchTopicalGuide: %v", err)
	}

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].EntityType != EntityTopicalGuide {
		t.Errorf("expected EntityTopicalGuide, got %q", results[0].EntityType)
	}
	if results[0].Name != "Faith" {
		t.Errorf("expected name 'Faith', got %q", results[0].Name)
	}
}

func TestSearchBibleDict(t *testing.T) {
	client := TestClient(t)
	ctx := context.Background()
	ec := client.Ent()

	_, err := ec.BibleDictEntry.Create().
		SetName("Aaron").
		SetText("Son of Amram and Jochebed").
		SetEmbedding(makeUnitEmbedding(0)).
		Save(ctx)
	if err != nil {
		t.Fatalf("creating BD entry: %v", err)
	}

	results, err := searchBibleDict(ctx, client.Sqlx(), makeUnitEmbedding(0), 10)
	if err != nil {
		t.Fatalf("searchBibleDict: %v", err)
	}

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].EntityType != EntityBibleDict {
		t.Errorf("expected EntityBibleDict, got %q", results[0].EntityType)
	}
	if results[0].Name != "Aaron" {
		t.Errorf("expected name 'Aaron', got %q", results[0].Name)
	}
	if results[0].Text != "Son of Amram and Jochebed" {
		t.Errorf("unexpected text: %q", results[0].Text)
	}
}

func TestSearchIndex(t *testing.T) {
	client := TestClient(t)
	ctx := context.Background()
	ec := client.Ent()

	_, err := ec.IndexEntry.Create().
		SetName("Atonement").
		SetEmbedding(makeUnitEmbedding(0)).
		Save(ctx)
	if err != nil {
		t.Fatalf("creating index entry: %v", err)
	}

	results, err := searchIndex(ctx, client.Sqlx(), makeUnitEmbedding(0), 10)
	if err != nil {
		t.Fatalf("searchIndex: %v", err)
	}

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].EntityType != EntityIndex {
		t.Errorf("expected EntityIndex, got %q", results[0].EntityType)
	}
	if results[0].Name != "Atonement" {
		t.Errorf("expected name 'Atonement', got %q", results[0].Name)
	}
}

func TestSearchJSTPassages(t *testing.T) {
	client := TestClient(t)
	ctx := context.Background()
	ec := client.Ent()

	_, err := ec.JSTPassage.Create().
		SetBook("1 Samuel").
		SetChapter("16").
		SetComprises("14-16, 23").
		SetCompareRef("1 Samuel 16:14-16, 23").
		SetSummary("The evil spirit is not from the Lord").
		SetText("But the Spirit of the Lord departed from Saul").
		SetEmbedding(makeUnitEmbedding(0)).
		Save(ctx)
	if err != nil {
		t.Fatalf("creating JST passage: %v", err)
	}

	results, err := searchJSTPassages(ctx, client.Sqlx(), makeUnitEmbedding(0), 10)
	if err != nil {
		t.Fatalf("searchJSTPassages: %v", err)
	}

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}

	r := results[0]
	if r.EntityType != EntityJSTPassage {
		t.Errorf("expected EntityJSTPassage, got %q", r.EntityType)
	}
	if r.Text != "But the Spirit of the Lord departed from Saul" {
		t.Errorf("unexpected text: %q", r.Text)
	}
	if r.Metadata.Book != "1 Samuel" {
		t.Errorf("expected book '1 Samuel', got %q", r.Metadata.Book)
	}
	if r.Metadata.Chapter != "16" {
		t.Errorf("expected chapter '16', got %q", r.Metadata.Chapter)
	}
	if r.Metadata.CompareRef != "1 Samuel 16:14-16, 23" {
		t.Errorf("unexpected compare_ref: %q", r.Metadata.CompareRef)
	}
	if r.Metadata.Summary != "The evil spirit is not from the Lord" {
		t.Errorf("unexpected summary: %q", r.Metadata.Summary)
	}
}

func TestSearchLimitRespected(t *testing.T) {
	client := TestClient(t)
	ctx := context.Background()
	ec := client.Ent()

	// Create 5 TG entries, all with embeddings.
	names := []string{"Faith", "Hope", "Charity", "Repentance", "Baptism"}
	for i, name := range names {
		_, err := ec.TopicalGuideEntry.Create().
			SetName(name).
			SetEmbedding(makeUnitEmbedding(i)).
			Save(ctx)
		if err != nil {
			t.Fatalf("creating TG entry %q: %v", name, err)
		}
	}

	// Search with limit=2.
	results, err := searchTopicalGuide(ctx, client.Sqlx(), makeUnitEmbedding(0), 2)
	if err != nil {
		t.Fatalf("searchTopicalGuide: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("expected 2 results with limit=2, got %d", len(results))
	}
}

func TestSearchEmptyTable(t *testing.T) {
	client := TestClient(t)
	ctx := context.Background()

	// Search an empty table should return empty results, not error.
	results, err := searchTopicalGuide(ctx, client.Sqlx(), makeUnitEmbedding(0), 10)
	if err != nil {
		t.Fatalf("searchTopicalGuide on empty table: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("expected 0 results on empty table, got %d", len(results))
	}
}

func TestDoContextualSearch(t *testing.T) {
	client := TestClient(t)
	ctx := context.Background()
	ec := client.Ent()

	vol := createTestVolume(t, ctx, ec)
	book := createTestBook(t, ctx, ec, vol)
	ch := createTestChapter(t, ctx, ec, book)

	// Seed data across multiple entity types, all with the same embedding.
	queryEmbedding := makeUnitEmbedding(0)

	_, err := ec.VerseGroup.Create().
		SetText("Test verse group").
		SetStartVerseNumber(1).
		SetEndVerseNumber(2).
		SetChapter(ch).
		SetEmbedding(queryEmbedding).
		Save(ctx)
	if err != nil {
		t.Fatalf("creating verse group: %v", err)
	}

	_, err = ec.TopicalGuideEntry.Create().
		SetName("TestTopic").
		SetEmbedding(queryEmbedding).
		Save(ctx)
	if err != nil {
		t.Fatalf("creating TG entry: %v", err)
	}

	_, err = ec.BibleDictEntry.Create().
		SetName("TestEntry").
		SetText("Test definition").
		SetEmbedding(queryEmbedding).
		Save(ctx)
	if err != nil {
		t.Fatalf("creating BD entry: %v", err)
	}

	results, err := client.DoContextualSearch(ctx, queryEmbedding)
	if err != nil {
		t.Fatalf("DoContextualSearch: %v", err)
	}

	// We seeded 3 entity types, so at least 3 results expected.
	if len(results) < 3 {
		t.Errorf("expected at least 3 results, got %d", len(results))
	}

	// Verify results are sorted by rank score (distance - typeBonus).
	rankScore := func(r SearchResult) float64 {
		bonus := 0.0
		if r.EntityType == EntityVerse {
			bonus = defaultVerseBonus
		}
		return r.Distance - bonus
	}
	for i := 1; i < len(results); i++ {
		if rankScore(results[i]) < rankScore(results[i-1]) {
			t.Errorf("results not sorted by rank score: index %d score %f < index %d score %f",
				i, rankScore(results[i]), i-1, rankScore(results[i-1]))
		}
	}

	// Verify we got results from multiple entity types.
	typeSet := make(map[EntityType]bool)
	for _, r := range results {
		typeSet[r.EntityType] = true
	}
	if !typeSet[EntityVerseGroup] {
		t.Error("missing EntityVerseGroup in results")
	}
	if !typeSet[EntityTopicalGuide] {
		t.Error("missing EntityTopicalGuide in results")
	}
	if !typeSet[EntityBibleDict] {
		t.Error("missing EntityBibleDict in results")
	}
}

func TestDoContextualSearchEmptyEmbedding(t *testing.T) {
	client := TestClient(t)
	ctx := context.Background()

	_, err := client.DoContextualSearch(ctx, nil)
	if err == nil {
		t.Error("expected error for nil embedding, got nil")
	}

	_, err = client.DoContextualSearch(ctx, []byte{})
	if err == nil {
		t.Error("expected error for empty embedding, got nil")
	}
}

func TestSortByDistance(t *testing.T) {
	original := []SearchResult{
		{ID: 1, Distance: 0.5},
		{ID: 2, Distance: 0.1},
		{ID: 3, Distance: 0.9},
		{ID: 4, Distance: 0.3},
	}

	sorted := SortByDistance(original)

	// Verify sorted order.
	expectedOrder := []int{2, 4, 1, 3}
	for i, want := range expectedOrder {
		if sorted[i].ID != want {
			t.Errorf("sorted[%d].ID = %d, want %d", i, sorted[i].ID, want)
		}
	}

	// Verify original is unchanged (immutability).
	originalOrder := []int{1, 2, 3, 4}
	for i, want := range originalOrder {
		if original[i].ID != want {
			t.Errorf("original[%d].ID = %d, want %d (original was mutated)", i, original[i].ID, want)
		}
	}
}

func TestSortByDistanceEmpty(t *testing.T) {
	sorted := SortByDistance(nil)
	if len(sorted) != 0 {
		t.Errorf("expected empty slice for nil input, got %d", len(sorted))
	}

	sorted = SortByDistance([]SearchResult{})
	if len(sorted) != 0 {
		t.Errorf("expected empty slice for empty input, got %d", len(sorted))
	}
}

// --- Task 7: Integration test for 3-stage pipeline ---

func TestDoContextualSearchWithGraphTraversal(t *testing.T) {
	client := TestClient(t)
	ctx := context.Background()
	ec := client.Ent()

	vol := createTestVolume(t, ctx, ec)
	book := createTestBook(t, ctx, ec, vol)
	ch := createTestChapter(t, ctx, ec, book)

	queryEmbedding := makeUnitEmbedding(0)

	// Create verses that are NOT directly searchable by vector (no embedding).
	v1 := createTestVerse(t, ctx, ec, ch, 1, "I Nephi having been born of goodly parents", "1 Ne. 1:1")
	v2 := createTestVerse(t, ctx, ec, ch, 2, "Therefore I was taught in the learning of my father", "1 Ne. 1:2")

	// Create a verse group that includes these verses AND has a vector embedding.
	_, err := ec.VerseGroup.Create().
		SetText("I Nephi having been born of goodly parents Therefore I was taught").
		SetStartVerseNumber(1).
		SetEndVerseNumber(2).
		SetChapter(ch).
		SetEmbedding(queryEmbedding).
		AddVerses(v1, v2).
		Save(ctx)
	if err != nil {
		t.Fatalf("creating verse group: %v", err)
	}

	// Create a TG entry with embedding that links to a verse.
	tg, err := ec.TopicalGuideEntry.Create().
		SetName("Faith").
		SetEmbedding(queryEmbedding).
		Save(ctx)
	if err != nil {
		t.Fatalf("creating TG entry: %v", err)
	}

	v3 := createTestVerse(t, ctx, ec, ch, 3, "Faith is the substance of things hoped for", "Heb. 11:1")
	_, err = ec.TGVerseRef.Create().
		SetPhrase("substance of things hoped for").
		SetTgEntry(tg).
		SetVerse(v3).
		Save(ctx)
	if err != nil {
		t.Fatalf("creating TG->verse ref: %v", err)
	}

	// Run the full pipeline.
	results, err := client.DoContextualSearch(ctx, queryEmbedding)
	if err != nil {
		t.Fatalf("DoContextualSearch: %v", err)
	}

	// Stage 1 should find the verse_group and TG entry (vector hits).
	// Stage 2 should traverse edges to find individual verses (v1, v2 from verse_group, v3 from TG).
	// Stage 3 should rank verses higher than verse_groups when distances are close.

	typeSet := make(map[EntityType]bool)
	for _, r := range results {
		typeSet[r.EntityType] = true
	}

	if !typeSet[EntityVerseGroup] {
		t.Error("missing EntityVerseGroup in results")
	}
	if !typeSet[EntityTopicalGuide] {
		t.Error("missing EntityTopicalGuide in results")
	}
	if !typeSet[EntityVerse] {
		t.Error("missing EntityVerse in pipeline results — graph traversal did not surface individual verses")
	}

	// Verify verses have correct metadata.
	for _, r := range results {
		if r.EntityType == EntityVerse {
			if r.Metadata.VerseNumber == 0 {
				t.Error("EntityVerse result missing VerseNumber metadata")
			}
			if r.Metadata.Reference == "" {
				t.Error("EntityVerse result missing Reference metadata")
			}
		}
	}
}

func TestDoContextualSearchKNNTrimsAfterRanking(t *testing.T) {
	client := TestClient(t)
	ctx := context.Background()
	ec := client.Ent()

	vol := createTestVolume(t, ctx, ec)
	book := createTestBook(t, ctx, ec, vol)
	ch := createTestChapter(t, ctx, ec, book)

	queryEmbedding := makeUnitEmbedding(0)

	// Create multiple verse groups with embeddings to generate lots of candidates.
	for i := range 5 {
		vg, err := ec.VerseGroup.Create().
			SetText("verse group text").
			SetStartVerseNumber(i*3 + 1).
			SetEndVerseNumber(i*3 + 3).
			SetChapter(ch).
			SetEmbedding(queryEmbedding).
			Save(ctx)
		if err != nil {
			t.Fatalf("creating verse group %d: %v", i, err)
		}

		// Add verses to each group so graph traversal produces more results.
		for j := 1; j <= 3; j++ {
			v := createTestVerse(t, ctx, ec, ch, i*3+j, "verse text", fmt.Sprintf("Test 1:%d", i*3+j))
			err = vg.Update().AddVerses(v).Exec(ctx)
			if err != nil {
				t.Fatalf("adding verse to group: %v", err)
			}
		}
	}

	// With kNN=3, final result set should be trimmed.
	results, err := client.DoContextualSearch(ctx, queryEmbedding, WithKNN(3))
	if err != nil {
		t.Fatalf("DoContextualSearch: %v", err)
	}

	if len(results) > 3 {
		t.Errorf("expected at most 3 results with WithKNN(3), got %d", len(results))
	}
}
