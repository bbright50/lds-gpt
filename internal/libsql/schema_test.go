package libsql

import (
	"context"
	"math"
	"testing"

	"lds-gpt/internal/libsql/generated"
	"lds-gpt/internal/libsql/schema"
	"lds-gpt/internal/utils/vec"
)

func TestMigrationCreatesAllTables(t *testing.T) {
	client := TestClient(t)

	expectedTables := []string{
		"volumes",
		"books",
		"chapters",
		"verses",
		"verse_groups",
		"topical_guide_entries",
		"bible_dict_entries",
		"index_entries",
		"jst_passages",
		"verse_cross_refs",
		"verse_tg_refs",
		"verse_bd_refs",
		"verse_jst_refs",
		"tg_verse_refs",
		"bd_verse_refs",
		"idx_verse_refs",
		"bible_dict_entry_see_also",
		"index_entry_see_also",
		"index_entry_tg_refs",
		"index_entry_bd_refs",
		"jst_passage_compare_verses",
		"topical_guide_entry_see_also",
		"topical_guide_entry_bd_refs",
		"verse_group_verses",
	}

	for _, table := range expectedTables {
		var count int
		err := client.Sqlx().Get(&count,
			"SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name=?", table)
		if err != nil {
			t.Errorf("querying for table %q: %v", table, err)
			continue
		}
		if count != 1 {
			t.Errorf("table %q not found in database", table)
		}
	}
}

func TestScriptureHierarchyCRUD(t *testing.T) {
	client := TestClient(t)
	ctx := context.Background()
	ec := client.Ent()

	// Create Volume -> Book -> Chapter -> Verse hierarchy.
	vol, err := ec.Volume.Create().
		SetName("Book of Mormon").
		SetAbbreviation("bofm").
		Save(ctx)
	if err != nil {
		t.Fatalf("creating volume: %v", err)
	}

	book, err := ec.Book.Create().
		SetName("The First Book of Nephi").
		SetSlug("1-ne").
		SetURLPath("bofm/1-ne").
		SetVolume(vol).
		Save(ctx)
	if err != nil {
		t.Fatalf("creating book: %v", err)
	}

	ch, err := ec.Chapter.Create().
		SetNumber(1).
		SetSummary("Nephi begins the record of his people").
		SetURL("https://example.com/bofm/1-ne/1").
		SetBook(book).
		Save(ctx)
	if err != nil {
		t.Fatalf("creating chapter: %v", err)
	}

	v1, err := ec.Verse.Create().
		SetNumber(1).
		SetText("I, Nephi, having been born of goodly parents").
		SetReference("1 Ne. 1:1").
		SetChapter(ch).
		Save(ctx)
	if err != nil {
		t.Fatalf("creating verse 1: %v", err)
	}

	v2, err := ec.Verse.Create().
		SetNumber(2).
		SetText("Yea, I make a record in the language of my father").
		SetReference("1 Ne. 1:2").
		SetChapter(ch).
		Save(ctx)
	if err != nil {
		t.Fatalf("creating verse 2: %v", err)
	}

	// Verify traversal: Volume -> Book -> Chapter -> Verses.
	books, err := vol.QueryBooks().All(ctx)
	if err != nil {
		t.Fatalf("querying books: %v", err)
	}
	if len(books) != 1 || books[0].Name != "The First Book of Nephi" {
		t.Errorf("expected 1 book 'The First Book of Nephi', got %v", books)
	}

	chapters, err := book.QueryChapters().All(ctx)
	if err != nil {
		t.Fatalf("querying chapters: %v", err)
	}
	if len(chapters) != 1 || chapters[0].Number != 1 {
		t.Errorf("expected 1 chapter with number 1, got %v", chapters)
	}

	verses, err := ch.QueryVerses().All(ctx)
	if err != nil {
		t.Fatalf("querying verses: %v", err)
	}
	if len(verses) != 2 {
		t.Errorf("expected 2 verses, got %d", len(verses))
	}

	// Reverse traversal: Verse -> Chapter -> Book -> Volume.
	foundCh, err := v1.QueryChapter().Only(ctx)
	if err != nil {
		t.Fatalf("querying chapter from verse: %v", err)
	}
	if foundCh.ID != ch.ID {
		t.Errorf("expected chapter ID %d, got %d", ch.ID, foundCh.ID)
	}

	foundBook, err := foundCh.QueryBook().Only(ctx)
	if err != nil {
		t.Fatalf("querying book from chapter: %v", err)
	}
	if foundBook.Slug != "1-ne" {
		t.Errorf("expected slug '1-ne', got %q", foundBook.Slug)
	}

	_ = v2 // used in creation above
}

func TestVerseAnnotationFields(t *testing.T) {
	client := TestClient(t)
	ctx := context.Background()
	ec := client.Ent()

	vol := createTestVolume(t, ctx, ec)
	book := createTestBook(t, ctx, ec, vol)
	ch := createTestChapter(t, ctx, ec, book)

	// Create verse with all annotation types.
	v, err := ec.Verse.Create().
		SetNumber(1).
		SetText("Test verse").
		SetReference("Test 1:1").
		SetChapter(ch).
		SetTranslationNotes([]schema.TranslationNote{
			{Marker: "1a", HebrewText: "adam"},
		}).
		SetAlternateReadings([]schema.AlternateReading{
			{Marker: "1b", Text: "or, mankind"},
		}).
		SetExplanatoryNotes([]schema.ExplanatoryNote{
			{Marker: "1c", Text: "i.e., the first man"},
		}).
		Save(ctx)
	if err != nil {
		t.Fatalf("creating verse with annotations: %v", err)
	}

	// Re-query and verify JSON fields.
	found, err := ec.Verse.Get(ctx, v.ID)
	if err != nil {
		t.Fatalf("querying verse: %v", err)
	}

	if len(found.TranslationNotes) != 1 || found.TranslationNotes[0].HebrewText != "adam" {
		t.Errorf("translation_notes mismatch: %+v", found.TranslationNotes)
	}
	if len(found.AlternateReadings) != 1 || found.AlternateReadings[0].Text != "or, mankind" {
		t.Errorf("alternate_readings mismatch: %+v", found.AlternateReadings)
	}
	if len(found.ExplanatoryNotes) != 1 || found.ExplanatoryNotes[0].Text != "i.e., the first man" {
		t.Errorf("explanatory_notes mismatch: %+v", found.ExplanatoryNotes)
	}
}

func TestVerseCrossRefEdge(t *testing.T) {
	client := TestClient(t)
	ctx := context.Background()
	ec := client.Ent()

	vol := createTestVolume(t, ctx, ec)
	book := createTestBook(t, ctx, ec, vol)
	ch := createTestChapter(t, ctx, ec, book)

	v1 := createTestVerse(t, ctx, ec, ch, 1, "Source verse", "Test 1:1")
	v2 := createTestVerse(t, ctx, ec, ch, 2, "Target verse", "Test 1:2")

	// Create cross-reference edge with metadata.
	_, err := ec.VerseCrossRef.Create().
		SetFootnoteMarker("1a").
		SetReferenceText("goodly").
		SetVerse(v1).
		SetCrossRefTarget(v2).
		Save(ctx)
	if err != nil {
		t.Fatalf("creating cross-ref: %v", err)
	}

	// Traverse through the cross-reference edge.
	targets, err := v1.QueryCrossRefTargets().All(ctx)
	if err != nil {
		t.Fatalf("querying cross-ref targets: %v", err)
	}
	if len(targets) != 1 || targets[0].ID != v2.ID {
		t.Errorf("expected cross-ref to verse %d, got %v", v2.ID, targets)
	}

	// Query the junction entity for metadata.
	refs, err := ec.VerseCrossRef.Query().All(ctx)
	if err != nil {
		t.Fatalf("querying verse_cross_refs: %v", err)
	}
	if len(refs) != 1 {
		t.Fatalf("expected 1 cross-ref, got %d", len(refs))
	}
	if refs[0].FootnoteMarker != "1a" {
		t.Errorf("expected footnote_marker '1a', got %q", refs[0].FootnoteMarker)
	}
	if refs[0].ReferenceText != "goodly" {
		t.Errorf("expected reference_text 'goodly', got %q", refs[0].ReferenceText)
	}
}

func TestStudyHelpEdges(t *testing.T) {
	client := TestClient(t)
	ctx := context.Background()
	ec := client.Ent()

	vol := createTestVolume(t, ctx, ec)
	book := createTestBook(t, ctx, ec, vol)
	ch := createTestChapter(t, ctx, ec, book)
	v := createTestVerse(t, ctx, ec, ch, 1, "Test verse", "Test 1:1")

	// Create TG entry and link from verse footnote.
	tg, err := ec.TopicalGuideEntry.Create().
		SetName("Birthright").
		Save(ctx)
	if err != nil {
		t.Fatalf("creating TG entry: %v", err)
	}

	_, err = ec.VerseTGRef.Create().
		SetFootnoteMarker("1a").
		SetReferenceText("born").
		SetTgTopicText("TG Birthright.").
		SetVerse(v).
		SetTgEntry(tg).
		Save(ctx)
	if err != nil {
		t.Fatalf("creating verse->TG ref: %v", err)
	}

	// Create BD entry and link from verse footnote.
	bd, err := ec.BibleDictEntry.Create().
		SetName("Aaron").
		SetText("Son of Amram and Jochebed").
		Save(ctx)
	if err != nil {
		t.Fatalf("creating BD entry: %v", err)
	}

	_, err = ec.VerseBDRef.Create().
		SetFootnoteMarker("2a").
		SetVerse(v).
		SetBdEntry(bd).
		Save(ctx)
	if err != nil {
		t.Fatalf("creating verse->BD ref: %v", err)
	}

	// Create TG -> Verse reference (from TG data).
	_, err = ec.TGVerseRef.Create().
		SetPhrase("exalt himself shall be abased").
		SetTgEntry(tg).
		SetVerse(v).
		Save(ctx)
	if err != nil {
		t.Fatalf("creating TG->verse ref: %v", err)
	}

	// Traverse verse -> TG entries (via footnotes).
	tgEntries, err := v.QueryFootnoteTgEntries().All(ctx)
	if err != nil {
		t.Fatalf("querying footnote TG entries: %v", err)
	}
	if len(tgEntries) != 1 || tgEntries[0].Name != "Birthright" {
		t.Errorf("expected TG entry 'Birthright', got %v", tgEntries)
	}

	// Traverse TG -> verse refs.
	tgVerses, err := tg.QueryVerseRefs().All(ctx)
	if err != nil {
		t.Fatalf("querying TG verse refs: %v", err)
	}
	if len(tgVerses) != 1 || tgVerses[0].ID != v.ID {
		t.Errorf("expected verse %d from TG, got %v", v.ID, tgVerses)
	}

	// Traverse verse -> BD entries (via footnotes).
	bdEntries, err := v.QueryFootnoteBdEntries().All(ctx)
	if err != nil {
		t.Fatalf("querying footnote BD entries: %v", err)
	}
	if len(bdEntries) != 1 || bdEntries[0].Name != "Aaron" {
		t.Errorf("expected BD entry 'Aaron', got %v", bdEntries)
	}
}

func TestSelfRefEdges(t *testing.T) {
	client := TestClient(t)
	ctx := context.Background()
	ec := client.Ent()

	// TG see-also.
	tg1, err := ec.TopicalGuideEntry.Create().SetName("Faith").Save(ctx)
	if err != nil {
		t.Fatalf("creating TG entry: %v", err)
	}
	tg2, err := ec.TopicalGuideEntry.Create().SetName("Trust").Save(ctx)
	if err != nil {
		t.Fatalf("creating TG entry: %v", err)
	}

	err = tg1.Update().AddSeeAlso(tg2).Exec(ctx)
	if err != nil {
		t.Fatalf("adding TG see_also: %v", err)
	}

	seeAlso, err := tg1.QuerySeeAlso().All(ctx)
	if err != nil {
		t.Fatalf("querying TG see_also: %v", err)
	}
	if len(seeAlso) != 1 || seeAlso[0].Name != "Trust" {
		t.Errorf("expected TG see_also 'Trust', got %v", seeAlso)
	}

	// BD see-also.
	bd1, err := ec.BibleDictEntry.Create().SetName("Aaron").SetText("text").Save(ctx)
	if err != nil {
		t.Fatalf("creating BD entry: %v", err)
	}
	bd2, err := ec.BibleDictEntry.Create().SetName("Aaronic Priesthood").SetText("text").Save(ctx)
	if err != nil {
		t.Fatalf("creating BD entry: %v", err)
	}

	err = bd1.Update().AddSeeAlso(bd2).Exec(ctx)
	if err != nil {
		t.Fatalf("adding BD see_also: %v", err)
	}

	bdSeeAlso, err := bd1.QuerySeeAlso().All(ctx)
	if err != nil {
		t.Fatalf("querying BD see_also: %v", err)
	}
	if len(bdSeeAlso) != 1 || bdSeeAlso[0].Name != "Aaronic Priesthood" {
		t.Errorf("expected BD see_also 'Aaronic Priesthood', got %v", bdSeeAlso)
	}

	// IDX see-also.
	idx1, err := ec.IndexEntry.Create().SetName("Aaron1--brother of Moses").Save(ctx)
	if err != nil {
		t.Fatalf("creating IDX entry: %v", err)
	}
	idx2, err := ec.IndexEntry.Create().SetName("Priesthood, Aaronic").Save(ctx)
	if err != nil {
		t.Fatalf("creating IDX entry: %v", err)
	}

	err = idx1.Update().AddSeeAlso(idx2).Exec(ctx)
	if err != nil {
		t.Fatalf("adding IDX see_also: %v", err)
	}

	idxSeeAlso, err := idx1.QuerySeeAlso().All(ctx)
	if err != nil {
		t.Fatalf("querying IDX see_also: %v", err)
	}
	if len(idxSeeAlso) != 1 || idxSeeAlso[0].Name != "Priesthood, Aaronic" {
		t.Errorf("expected IDX see_also 'Priesthood, Aaronic', got %v", idxSeeAlso)
	}
}

func TestCrossEntityEdges(t *testing.T) {
	client := TestClient(t)
	ctx := context.Background()
	ec := client.Ent()

	tg, err := ec.TopicalGuideEntry.Create().SetName("Atonement").Save(ctx)
	if err != nil {
		t.Fatalf("creating TG: %v", err)
	}
	bd, err := ec.BibleDictEntry.Create().SetName("Atonement").SetText("text").Save(ctx)
	if err != nil {
		t.Fatalf("creating BD: %v", err)
	}
	idx, err := ec.IndexEntry.Create().SetName("Atonement").Save(ctx)
	if err != nil {
		t.Fatalf("creating IDX: %v", err)
	}

	// TG -> BD.
	err = tg.Update().AddBdRefs(bd).Exec(ctx)
	if err != nil {
		t.Fatalf("adding TG->BD ref: %v", err)
	}

	// IDX -> TG.
	err = idx.Update().AddTgRefs(tg).Exec(ctx)
	if err != nil {
		t.Fatalf("adding IDX->TG ref: %v", err)
	}

	// IDX -> BD.
	err = idx.Update().AddBdRefs(bd).Exec(ctx)
	if err != nil {
		t.Fatalf("adding IDX->BD ref: %v", err)
	}

	// Verify TG -> BD.
	bdRefs, err := tg.QueryBdRefs().All(ctx)
	if err != nil {
		t.Fatalf("querying TG->BD: %v", err)
	}
	if len(bdRefs) != 1 || bdRefs[0].Name != "Atonement" {
		t.Errorf("expected TG->BD 'Atonement', got %v", bdRefs)
	}

	// Verify IDX -> TG.
	tgRefs, err := idx.QueryTgRefs().All(ctx)
	if err != nil {
		t.Fatalf("querying IDX->TG: %v", err)
	}
	if len(tgRefs) != 1 {
		t.Errorf("expected 1 IDX->TG ref, got %d", len(tgRefs))
	}

	// Verify reverse: BD -> TG back-refs.
	tgBackRefs, err := bd.QueryTgRefs().All(ctx)
	if err != nil {
		t.Fatalf("querying BD<-TG: %v", err)
	}
	if len(tgBackRefs) != 1 {
		t.Errorf("expected 1 BD<-TG back-ref, got %d", len(tgBackRefs))
	}
}

func TestJSTPassageEdges(t *testing.T) {
	client := TestClient(t)
	ctx := context.Background()
	ec := client.Ent()

	vol := createTestVolume(t, ctx, ec)
	book := createTestBook(t, ctx, ec, vol)
	ch := createTestChapter(t, ctx, ec, book)
	v := createTestVerse(t, ctx, ec, ch, 1, "Original verse", "Test 1:1")

	jst, err := ec.JSTPassage.Create().
		SetBook("1 Samuel").
		SetChapter("16").
		SetComprises("14-16, 23").
		SetCompareRef("1 Samuel 16:14-16, 23").
		SetSummary("The evil spirit is not from the Lord").
		SetText("But the Spirit of the Lord departed from Saul").
		Save(ctx)
	if err != nil {
		t.Fatalf("creating JST passage: %v", err)
	}

	// JST -> Verse (compare).
	err = jst.Update().AddCompareVerses(v).Exec(ctx)
	if err != nil {
		t.Fatalf("adding JST compare verse: %v", err)
	}

	// Verse -> JST footnote ref.
	_, err = ec.VerseJSTRef.Create().
		SetFootnoteMarker("1a").
		SetVerse(v).
		SetJstPassage(jst).
		Save(ctx)
	if err != nil {
		t.Fatalf("creating verse->JST ref: %v", err)
	}

	// Verify JST -> Verse compare.
	compareVerses, err := jst.QueryCompareVerses().All(ctx)
	if err != nil {
		t.Fatalf("querying JST compare verses: %v", err)
	}
	if len(compareVerses) != 1 {
		t.Errorf("expected 1 compare verse, got %d", len(compareVerses))
	}

	// Verify Verse -> JST footnote.
	jstPassages, err := v.QueryFootnoteJstPassages().All(ctx)
	if err != nil {
		t.Fatalf("querying verse JST footnotes: %v", err)
	}
	if len(jstPassages) != 1 || jstPassages[0].Book != "1 Samuel" {
		t.Errorf("expected JST passage for '1 Samuel', got %v", jstPassages)
	}
}

func TestVerseGroupEdge(t *testing.T) {
	client := TestClient(t)
	ctx := context.Background()
	ec := client.Ent()

	vol := createTestVolume(t, ctx, ec)
	book := createTestBook(t, ctx, ec, vol)
	ch := createTestChapter(t, ctx, ec, book)

	v1 := createTestVerse(t, ctx, ec, ch, 1, "Verse one", "Test 1:1")
	v2 := createTestVerse(t, ctx, ec, ch, 2, "Verse two", "Test 1:2")
	v3 := createTestVerse(t, ctx, ec, ch, 3, "Verse three", "Test 1:3")

	vg, err := ec.VerseGroup.Create().
		SetText("Verse one Verse two Verse three").
		SetStartVerseNumber(1).
		SetEndVerseNumber(3).
		SetChapter(ch).
		AddVerses(v1, v2, v3).
		Save(ctx)
	if err != nil {
		t.Fatalf("creating verse group: %v", err)
	}

	// Query verses in group.
	groupVerses, err := vg.QueryVerses().All(ctx)
	if err != nil {
		t.Fatalf("querying verse group verses: %v", err)
	}
	if len(groupVerses) != 3 {
		t.Errorf("expected 3 verses in group, got %d", len(groupVerses))
	}

	// Reverse: verse -> verse groups.
	groups, err := v1.QueryVerseGroups().All(ctx)
	if err != nil {
		t.Fatalf("querying verse->groups: %v", err)
	}
	if len(groups) != 1 || groups[0].ID != vg.ID {
		t.Errorf("expected verse to be in 1 group, got %d", len(groups))
	}
}

func TestEmbeddingRoundTrip(t *testing.T) {
	client := TestClient(t)
	ctx := context.Background()
	ec := client.Ent()

	vol := createTestVolume(t, ctx, ec)
	book := createTestBook(t, ctx, ec, vol)
	ch := createTestChapter(t, ctx, ec, book)

	// Encode a 1024-dim embedding as packed float32 bytes.
	embedding := make([]float32, 1024)
	embedding[0] = 0.1
	embedding[1] = 0.2
	embedding[2] = 0.3
	embedding[3] = 0.4
	embedding[4] = 0.5
	blobBytes := vec.Float32sToBytes(embedding)

	vg, err := ec.VerseGroup.Create().
		SetText("Test verse group text").
		SetStartVerseNumber(1).
		SetEndVerseNumber(3).
		SetChapter(ch).
		SetEmbedding(blobBytes).
		Save(ctx)
	if err != nil {
		t.Fatalf("creating verse group with embedding: %v", err)
	}

	// Re-query and verify embedding round-trip.
	found, err := ec.VerseGroup.Get(ctx, vg.ID)
	if err != nil {
		t.Fatalf("querying verse group: %v", err)
	}

	if found.Embedding == nil {
		t.Fatal("embedding is nil after round-trip")
	}

	decoded := vec.BytesToFloat32s(*found.Embedding)
	if len(decoded) != len(embedding) {
		t.Fatalf("expected %d floats, got %d", len(embedding), len(decoded))
	}
	for i, want := range embedding {
		if math.Abs(float64(decoded[i]-want)) > 1e-6 {
			t.Errorf("embedding[%d] = %f, want %f", i, decoded[i], want)
		}
	}

	// Test chapter summary embedding.
	err = ch.Update().SetSummaryEmbedding(blobBytes).Exec(ctx)
	if err != nil {
		t.Fatalf("setting chapter summary embedding: %v", err)
	}

	foundCh, err := ec.Chapter.Get(ctx, ch.ID)
	if err != nil {
		t.Fatalf("querying chapter: %v", err)
	}
	if foundCh.SummaryEmbedding == nil {
		t.Error("chapter summary_embedding is nil after round-trip")
	}
}

func TestVectorSearch(t *testing.T) {
	client := TestClient(t)
	ctx := context.Background()
	ec := client.Ent()

	vol := createTestVolume(t, ctx, ec)
	book := createTestBook(t, ctx, ec, vol)
	ch := createTestChapter(t, ctx, ec, book)

	// Create a 1024-dim embedding (unit vector along first axis).
	embedding := make([]float32, 1024)
	embedding[0] = 1.0
	blobBytes := vec.Float32sToBytes(embedding)

	_, err := ec.VerseGroup.Create().
		SetText("Vector search test").
		SetStartVerseNumber(1).
		SetEndVerseNumber(1).
		SetChapter(ch).
		SetEmbedding(blobBytes).
		Save(ctx)
	if err != nil {
		t.Fatalf("creating verse group with embedding: %v", err)
	}

	// Query using vector_distance_cos: self-distance should be ~0.
	var distance float64
	err = client.Sqlx().GetContext(ctx, &distance,
		"SELECT vector_distance_cos(embedding, ?) FROM verse_groups WHERE embedding IS NOT NULL LIMIT 1",
		blobBytes,
	)
	if err != nil {
		t.Fatalf("vector_distance_cos query: %v", err)
	}
	if distance > 1e-6 {
		t.Errorf("expected self-distance ~0, got %f", distance)
	}
}

func TestTwoHopTraversal(t *testing.T) {
	client := TestClient(t)
	ctx := context.Background()
	ec := client.Ent()

	vol := createTestVolume(t, ctx, ec)
	book := createTestBook(t, ctx, ec, vol)
	ch := createTestChapter(t, ctx, ec, book)

	// Create 3 verses forming a chain: v1 -> v2 -> v3.
	v1 := createTestVerse(t, ctx, ec, ch, 1, "First verse", "Test 1:1")
	v2 := createTestVerse(t, ctx, ec, ch, 2, "Second verse", "Test 1:2")
	v3 := createTestVerse(t, ctx, ec, ch, 3, "Third verse", "Test 1:3")

	_, err := ec.VerseCrossRef.Create().
		SetFootnoteMarker("1a").
		SetVerse(v1).
		SetCrossRefTarget(v2).
		Save(ctx)
	if err != nil {
		t.Fatalf("creating cross-ref v1->v2: %v", err)
	}

	_, err = ec.VerseCrossRef.Create().
		SetFootnoteMarker("2a").
		SetVerse(v2).
		SetCrossRefTarget(v3).
		Save(ctx)
	if err != nil {
		t.Fatalf("creating cross-ref v2->v3: %v", err)
	}

	// 1-hop from v1: should find v2.
	hop1, err := v1.QueryCrossRefTargets().All(ctx)
	if err != nil {
		t.Fatalf("1-hop query: %v", err)
	}
	if len(hop1) != 1 || hop1[0].ID != v2.ID {
		t.Errorf("1-hop: expected verse %d, got %v", v2.ID, hop1)
	}

	// 2-hop from v1: v1 -> v2 -> v3.
	hop2, err := v1.QueryCrossRefTargets().
		QueryCrossRefTargets().
		All(ctx)
	if err != nil {
		t.Fatalf("2-hop query: %v", err)
	}
	if len(hop2) != 1 || hop2[0].ID != v3.ID {
		t.Errorf("2-hop: expected verse %d, got %v", v3.ID, hop2)
	}
}

// Helper functions.

func createTestVolume(t *testing.T, ctx context.Context, ec *generated.Client) *generated.Volume {
	t.Helper()
	vol, err := ec.Volume.Create().
		SetName("Test Volume").
		SetAbbreviation("test").
		Save(ctx)
	if err != nil {
		t.Fatalf("creating test volume: %v", err)
	}
	return vol
}

func createTestBook(t *testing.T, ctx context.Context, ec *generated.Client, vol *generated.Volume) *generated.Book {
	t.Helper()
	book, err := ec.Book.Create().
		SetName("Test Book").
		SetSlug("test-book").
		SetURLPath("test/test-book").
		SetVolume(vol).
		Save(ctx)
	if err != nil {
		t.Fatalf("creating test book: %v", err)
	}
	return book
}

func createTestChapter(t *testing.T, ctx context.Context, ec *generated.Client, book *generated.Book) *generated.Chapter {
	t.Helper()
	ch, err := ec.Chapter.Create().
		SetNumber(1).
		SetBook(book).
		Save(ctx)
	if err != nil {
		t.Fatalf("creating test chapter: %v", err)
	}
	return ch
}

func createTestVerse(t *testing.T, ctx context.Context, ec *generated.Client, ch *generated.Chapter, num int, text, ref string) *generated.Verse {
	t.Helper()
	v, err := ec.Verse.Create().
		SetNumber(num).
		SetText(text).
		SetReference(ref).
		SetChapter(ch).
		Save(ctx)
	if err != nil {
		t.Fatalf("creating test verse %d: %v", num, err)
	}
	return v
}

