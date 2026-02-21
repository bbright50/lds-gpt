package libsql

import (
	"context"
	"fmt"
	"testing"
)

// --- Task 3: Verse-producing edges ---

func TestTraverseVerseGroupToVerses(t *testing.T) {
	client := TestClient(t)
	ctx := context.Background()
	ec := client.Ent()

	vol := createTestVolume(t, ctx, ec)
	book := createTestBook(t, ctx, ec, vol)
	ch := createTestChapter(t, ctx, ec, book)

	v1 := createTestVerse(t, ctx, ec, ch, 1, "First verse", "Test 1:1")
	v2 := createTestVerse(t, ctx, ec, ch, 2, "Second verse", "Test 1:2")
	v3 := createTestVerse(t, ctx, ec, ch, 3, "Third verse", "Test 1:3")

	vg, err := ec.VerseGroup.Create().
		SetText("First verse Second verse Third verse").
		SetStartVerseNumber(1).
		SetEndVerseNumber(3).
		SetChapter(ch).
		SetEmbedding(makeUnitEmbedding(0)).
		AddVerses(v1, v2, v3).
		Save(ctx)
	if err != nil {
		t.Fatalf("creating verse group: %v", err)
	}

	seed := SearchResult{
		EntityType: EntityVerseGroup,
		ID:         vg.ID,
		Distance:   0.10,
	}

	results, err := traverseEdges(ctx, client.Ent(), seed, defaultGraphLimit)
	if err != nil {
		t.Fatalf("traverseEdges: %v", err)
	}

	// Should find 3 verses from the verse group.
	verseResults := filterByType(results, EntityVerse)
	if len(verseResults) != 3 {
		t.Errorf("expected 3 verse results from verse_group, got %d", len(verseResults))
	}

	// Verify EntityVerse type and metadata populated.
	for _, r := range verseResults {
		if r.EntityType != EntityVerse {
			t.Errorf("expected EntityVerse, got %q", r.EntityType)
		}
		if r.Metadata.VerseNumber == 0 {
			t.Error("expected VerseNumber to be populated")
		}
		if r.Metadata.Reference == "" {
			t.Error("expected Reference to be populated")
		}
	}
}

func TestTraverseChapterToVerses(t *testing.T) {
	client := TestClient(t)
	ctx := context.Background()
	ec := client.Ent()

	vol := createTestVolume(t, ctx, ec)
	book := createTestBook(t, ctx, ec, vol)
	ch := createTestChapter(t, ctx, ec, book)

	createTestVerse(t, ctx, ec, ch, 1, "Chapter verse one", "Test 1:1")
	createTestVerse(t, ctx, ec, ch, 2, "Chapter verse two", "Test 1:2")

	seed := SearchResult{
		EntityType: EntityChapter,
		ID:         ch.ID,
		Distance:   0.15,
	}

	results, err := traverseEdges(ctx, client.Ent(), seed, defaultGraphLimit)
	if err != nil {
		t.Fatalf("traverseEdges: %v", err)
	}

	verseResults := filterByType(results, EntityVerse)
	if len(verseResults) != 2 {
		t.Errorf("expected 2 verse results from chapter, got %d", len(verseResults))
	}
}

func TestTraverseTGToVerseRefs(t *testing.T) {
	client := TestClient(t)
	ctx := context.Background()
	ec := client.Ent()

	vol := createTestVolume(t, ctx, ec)
	book := createTestBook(t, ctx, ec, vol)
	ch := createTestChapter(t, ctx, ec, book)
	v := createTestVerse(t, ctx, ec, ch, 1, "Faith verse", "Heb. 11:1")

	tg, err := ec.TopicalGuideEntry.Create().
		SetName("Faith").
		SetEmbedding(makeUnitEmbedding(0)).
		Save(ctx)
	if err != nil {
		t.Fatalf("creating TG entry: %v", err)
	}

	_, err = ec.TGVerseRef.Create().
		SetPhrase("substance of things hoped for").
		SetTgEntry(tg).
		SetVerse(v).
		Save(ctx)
	if err != nil {
		t.Fatalf("creating TG->verse ref: %v", err)
	}

	seed := SearchResult{
		EntityType: EntityTopicalGuide,
		ID:         tg.ID,
		Distance:   0.12,
	}

	results, err := traverseEdges(ctx, client.Ent(), seed, defaultGraphLimit)
	if err != nil {
		t.Fatalf("traverseEdges: %v", err)
	}

	verseResults := filterByType(results, EntityVerse)
	if len(verseResults) != 1 {
		t.Errorf("expected 1 verse from TG verse_refs, got %d", len(verseResults))
	}
	if len(verseResults) > 0 && verseResults[0].Metadata.Reference != "Heb. 11:1" {
		t.Errorf("expected reference 'Heb. 11:1', got %q", verseResults[0].Metadata.Reference)
	}
}

func TestTraverseBDToVerseRefs(t *testing.T) {
	client := TestClient(t)
	ctx := context.Background()
	ec := client.Ent()

	vol := createTestVolume(t, ctx, ec)
	book := createTestBook(t, ctx, ec, vol)
	ch := createTestChapter(t, ctx, ec, book)
	v := createTestVerse(t, ctx, ec, ch, 1, "BD verse", "Gen. 1:1")

	bd, err := ec.BibleDictEntry.Create().
		SetName("Aaron").
		SetText("Son of Amram").
		SetEmbedding(makeUnitEmbedding(0)).
		Save(ctx)
	if err != nil {
		t.Fatalf("creating BD entry: %v", err)
	}

	_, err = ec.BDVerseRef.Create().
		SetBdEntry(bd).
		SetVerse(v).
		Save(ctx)
	if err != nil {
		t.Fatalf("creating BD->verse ref: %v", err)
	}

	seed := SearchResult{
		EntityType: EntityBibleDict,
		ID:         bd.ID,
		Distance:   0.20,
	}

	results, err := traverseEdges(ctx, client.Ent(), seed, defaultGraphLimit)
	if err != nil {
		t.Fatalf("traverseEdges: %v", err)
	}

	verseResults := filterByType(results, EntityVerse)
	if len(verseResults) != 1 {
		t.Errorf("expected 1 verse from BD verse_refs, got %d", len(verseResults))
	}
}

func TestTraverseIDXToVerseRefs(t *testing.T) {
	client := TestClient(t)
	ctx := context.Background()
	ec := client.Ent()

	vol := createTestVolume(t, ctx, ec)
	book := createTestBook(t, ctx, ec, vol)
	ch := createTestChapter(t, ctx, ec, book)
	v := createTestVerse(t, ctx, ec, ch, 1, "IDX verse", "Alma 32:21")

	idx, err := ec.IndexEntry.Create().
		SetName("Faith").
		SetEmbedding(makeUnitEmbedding(0)).
		Save(ctx)
	if err != nil {
		t.Fatalf("creating IDX entry: %v", err)
	}

	_, err = ec.IDXVerseRef.Create().
		SetPhrase("faith is not to have a perfect knowledge").
		SetIndexEntry(idx).
		SetVerse(v).
		Save(ctx)
	if err != nil {
		t.Fatalf("creating IDX->verse ref: %v", err)
	}

	seed := SearchResult{
		EntityType: EntityIndex,
		ID:         idx.ID,
		Distance:   0.18,
	}

	results, err := traverseEdges(ctx, client.Ent(), seed, defaultGraphLimit)
	if err != nil {
		t.Fatalf("traverseEdges: %v", err)
	}

	verseResults := filterByType(results, EntityVerse)
	if len(verseResults) != 1 {
		t.Errorf("expected 1 verse from IDX verse_refs, got %d", len(verseResults))
	}
}

func TestTraverseJSTToCompareVerses(t *testing.T) {
	client := TestClient(t)
	ctx := context.Background()
	ec := client.Ent()

	vol := createTestVolume(t, ctx, ec)
	book := createTestBook(t, ctx, ec, vol)
	ch := createTestChapter(t, ctx, ec, book)
	v := createTestVerse(t, ctx, ec, ch, 1, "Original verse", "1 Sam. 16:14")

	jst, err := ec.JSTPassage.Create().
		SetBook("1 Samuel").
		SetChapter("16").
		SetComprises("14").
		SetText("JST text").
		SetEmbedding(makeUnitEmbedding(0)).
		Save(ctx)
	if err != nil {
		t.Fatalf("creating JST passage: %v", err)
	}

	err = jst.Update().AddCompareVerses(v).Exec(ctx)
	if err != nil {
		t.Fatalf("adding compare verse: %v", err)
	}

	seed := SearchResult{
		EntityType: EntityJSTPassage,
		ID:         jst.ID,
		Distance:   0.25,
	}

	results, err := traverseEdges(ctx, client.Ent(), seed, defaultGraphLimit)
	if err != nil {
		t.Fatalf("traverseEdges: %v", err)
	}

	verseResults := filterByType(results, EntityVerse)
	if len(verseResults) != 1 {
		t.Errorf("expected 1 verse from JST compare_verses, got %d", len(verseResults))
	}
}

func TestTraverseGraphLimitRespected(t *testing.T) {
	client := TestClient(t)
	ctx := context.Background()
	ec := client.Ent()

	vol := createTestVolume(t, ctx, ec)
	book := createTestBook(t, ctx, ec, vol)
	ch := createTestChapter(t, ctx, ec, book)

	// Create 10 verses in the chapter (more than defaultGraphLimit=5).
	for i := 1; i <= 10; i++ {
		createTestVerse(t, ctx, ec, ch, i, "Verse text", fmt.Sprintf("Test 1:%d", i))
	}

	seed := SearchResult{
		EntityType: EntityChapter,
		ID:         ch.ID,
		Distance:   0.10,
	}

	results, err := traverseEdges(ctx, client.Ent(), seed, defaultGraphLimit)
	if err != nil {
		t.Fatalf("traverseEdges: %v", err)
	}

	verseResults := filterByType(results, EntityVerse)
	if len(verseResults) > defaultGraphLimit {
		t.Errorf("expected at most %d verse results (graph limit), got %d", defaultGraphLimit, len(verseResults))
	}
}

// --- Task 4: Study-help-producing edges ---

func TestTraverseTGToSeeAlso(t *testing.T) {
	client := TestClient(t)
	ctx := context.Background()
	ec := client.Ent()

	tg1, err := ec.TopicalGuideEntry.Create().
		SetName("Faith").
		SetEmbedding(makeUnitEmbedding(0)).
		Save(ctx)
	if err != nil {
		t.Fatalf("creating TG entry: %v", err)
	}

	tg2, err := ec.TopicalGuideEntry.Create().
		SetName("Trust").
		Save(ctx)
	if err != nil {
		t.Fatalf("creating TG entry: %v", err)
	}

	err = tg1.Update().AddSeeAlso(tg2).Exec(ctx)
	if err != nil {
		t.Fatalf("adding TG see_also: %v", err)
	}

	seed := SearchResult{
		EntityType: EntityTopicalGuide,
		ID:         tg1.ID,
		Distance:   0.10,
	}

	results, err := traverseEdges(ctx, client.Ent(), seed, defaultGraphLimit)
	if err != nil {
		t.Fatalf("traverseEdges: %v", err)
	}

	tgResults := filterByType(results, EntityTopicalGuide)
	if len(tgResults) < 1 {
		t.Errorf("expected at least 1 TG see_also result, got %d", len(tgResults))
	}
}

func TestTraverseTGToBDRefs(t *testing.T) {
	client := TestClient(t)
	ctx := context.Background()
	ec := client.Ent()

	tg, err := ec.TopicalGuideEntry.Create().
		SetName("Atonement").
		SetEmbedding(makeUnitEmbedding(0)).
		Save(ctx)
	if err != nil {
		t.Fatalf("creating TG entry: %v", err)
	}

	bd, err := ec.BibleDictEntry.Create().
		SetName("Atonement").
		SetText("The reconciliation of man with God").
		Save(ctx)
	if err != nil {
		t.Fatalf("creating BD entry: %v", err)
	}

	err = tg.Update().AddBdRefs(bd).Exec(ctx)
	if err != nil {
		t.Fatalf("adding TG->BD ref: %v", err)
	}

	seed := SearchResult{
		EntityType: EntityTopicalGuide,
		ID:         tg.ID,
		Distance:   0.10,
	}

	results, err := traverseEdges(ctx, client.Ent(), seed, defaultGraphLimit)
	if err != nil {
		t.Fatalf("traverseEdges: %v", err)
	}

	bdResults := filterByType(results, EntityBibleDict)
	if len(bdResults) < 1 {
		t.Errorf("expected at least 1 BD result from TG bd_refs, got %d", len(bdResults))
	}
}

func TestTraverseBDToSeeAlso(t *testing.T) {
	client := TestClient(t)
	ctx := context.Background()
	ec := client.Ent()

	bd1, err := ec.BibleDictEntry.Create().
		SetName("Aaron").
		SetText("text").
		SetEmbedding(makeUnitEmbedding(0)).
		Save(ctx)
	if err != nil {
		t.Fatalf("creating BD entry: %v", err)
	}

	bd2, err := ec.BibleDictEntry.Create().
		SetName("Aaronic Priesthood").
		SetText("text").
		Save(ctx)
	if err != nil {
		t.Fatalf("creating BD entry: %v", err)
	}

	err = bd1.Update().AddSeeAlso(bd2).Exec(ctx)
	if err != nil {
		t.Fatalf("adding BD see_also: %v", err)
	}

	seed := SearchResult{
		EntityType: EntityBibleDict,
		ID:         bd1.ID,
		Distance:   0.10,
	}

	results, err := traverseEdges(ctx, client.Ent(), seed, defaultGraphLimit)
	if err != nil {
		t.Fatalf("traverseEdges: %v", err)
	}

	bdResults := filterByType(results, EntityBibleDict)
	if len(bdResults) < 1 {
		t.Errorf("expected at least 1 BD see_also result, got %d", len(bdResults))
	}
}

func TestTraverseIDXToSeeAlso(t *testing.T) {
	client := TestClient(t)
	ctx := context.Background()
	ec := client.Ent()

	idx1, err := ec.IndexEntry.Create().
		SetName("Aaron").
		SetEmbedding(makeUnitEmbedding(0)).
		Save(ctx)
	if err != nil {
		t.Fatalf("creating IDX entry: %v", err)
	}

	idx2, err := ec.IndexEntry.Create().
		SetName("Priesthood, Aaronic").
		Save(ctx)
	if err != nil {
		t.Fatalf("creating IDX entry: %v", err)
	}

	err = idx1.Update().AddSeeAlso(idx2).Exec(ctx)
	if err != nil {
		t.Fatalf("adding IDX see_also: %v", err)
	}

	seed := SearchResult{
		EntityType: EntityIndex,
		ID:         idx1.ID,
		Distance:   0.10,
	}

	results, err := traverseEdges(ctx, client.Ent(), seed, defaultGraphLimit)
	if err != nil {
		t.Fatalf("traverseEdges: %v", err)
	}

	idxResults := filterByType(results, EntityIndex)
	if len(idxResults) < 1 {
		t.Errorf("expected at least 1 IDX see_also result, got %d", len(idxResults))
	}
}

func TestTraverseIDXToTGRefs(t *testing.T) {
	client := TestClient(t)
	ctx := context.Background()
	ec := client.Ent()

	tg, err := ec.TopicalGuideEntry.Create().
		SetName("Atonement").
		Save(ctx)
	if err != nil {
		t.Fatalf("creating TG entry: %v", err)
	}

	idx, err := ec.IndexEntry.Create().
		SetName("Atonement").
		SetEmbedding(makeUnitEmbedding(0)).
		Save(ctx)
	if err != nil {
		t.Fatalf("creating IDX entry: %v", err)
	}

	err = idx.Update().AddTgRefs(tg).Exec(ctx)
	if err != nil {
		t.Fatalf("adding IDX->TG ref: %v", err)
	}

	seed := SearchResult{
		EntityType: EntityIndex,
		ID:         idx.ID,
		Distance:   0.10,
	}

	results, err := traverseEdges(ctx, client.Ent(), seed, defaultGraphLimit)
	if err != nil {
		t.Fatalf("traverseEdges: %v", err)
	}

	tgResults := filterByType(results, EntityTopicalGuide)
	if len(tgResults) < 1 {
		t.Errorf("expected at least 1 TG result from IDX tg_refs, got %d", len(tgResults))
	}
}

func TestTraverseIDXToBDRefs(t *testing.T) {
	client := TestClient(t)
	ctx := context.Background()
	ec := client.Ent()

	bd, err := ec.BibleDictEntry.Create().
		SetName("Atonement").
		SetText("text").
		Save(ctx)
	if err != nil {
		t.Fatalf("creating BD entry: %v", err)
	}

	idx, err := ec.IndexEntry.Create().
		SetName("Atonement").
		SetEmbedding(makeUnitEmbedding(0)).
		Save(ctx)
	if err != nil {
		t.Fatalf("creating IDX entry: %v", err)
	}

	err = idx.Update().AddBdRefs(bd).Exec(ctx)
	if err != nil {
		t.Fatalf("adding IDX->BD ref: %v", err)
	}

	seed := SearchResult{
		EntityType: EntityIndex,
		ID:         idx.ID,
		Distance:   0.10,
	}

	results, err := traverseEdges(ctx, client.Ent(), seed, defaultGraphLimit)
	if err != nil {
		t.Fatalf("traverseEdges: %v", err)
	}

	bdResults := filterByType(results, EntityBibleDict)
	if len(bdResults) < 1 {
		t.Errorf("expected at least 1 BD result from IDX bd_refs, got %d", len(bdResults))
	}
}

func TestTraverseUnknownEntityTypeReturnsEmpty(t *testing.T) {
	client := TestClient(t)
	ctx := context.Background()

	seed := SearchResult{
		EntityType: EntityType("unknown"),
		ID:         999,
		Distance:   0.10,
	}

	results, err := traverseEdges(ctx, client.Ent(), seed, defaultGraphLimit)
	if err != nil {
		t.Fatalf("traverseEdges for unknown type: %v", err)
	}

	if len(results) != 0 {
		t.Errorf("expected 0 results for unknown entity type, got %d", len(results))
	}
}

// filterByType is a test helper that filters results by entity type.
func filterByType(results []SearchResult, et EntityType) []SearchResult {
	filtered := make([]SearchResult, 0)
	for _, r := range results {
		if r.EntityType == et {
			filtered = append(filtered, r)
		}
	}
	return filtered
}
