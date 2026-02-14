package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// TranslationNote represents a Hebrew translation footnote (trn category).
type TranslationNote struct {
	Marker     string `json:"marker"`
	HebrewText string `json:"hebrew_text"`
}

// AlternateReading represents an alternate reading footnote (or category).
type AlternateReading struct {
	Marker string `json:"marker"`
	Text   string `json:"text"`
}

// ExplanatoryNote represents an explanatory footnote (ie category).
type ExplanatoryNote struct {
	Marker string `json:"marker"`
	Text   string `json:"text"`
}

// Verse is the atomic unit of scripture and the primary node in the knowledge graph.
type Verse struct {
	ent.Schema
}

func (Verse) Fields() []ent.Field {
	return []ent.Field{
		field.Int("number").
			Positive().
			Comment("Verse number within chapter"),
		field.Text("text").
			NotEmpty().
			Comment("Full verse text"),
		field.String("reference").
			NotEmpty().
			Comment("Canonical display reference, e.g. '1 Ne. 1:1'"),
		field.JSON("translation_notes", []TranslationNote{}).
			Optional().
			Comment("Hebrew translation footnotes (trn)"),
		field.JSON("alternate_readings", []AlternateReading{}).
			Optional().
			Comment("Alternate reading footnotes (or)"),
		field.JSON("explanatory_notes", []ExplanatoryNote{}).
			Optional().
			Comment("Explanatory footnotes (ie)"),
	}
}

func (Verse) Edges() []ent.Edge {
	return []ent.Edge{
		// Structural: chapter this verse belongs to
		edge.From("chapter", Chapter.Type).
			Ref("verses").
			Required().
			Unique(),

		// Cross-references from this verse to other verses (through VerseCrossRef)
		edge.To("cross_ref_targets", Verse.Type).
			Through("verse_cross_refs", VerseCrossRef.Type),

		// Footnotes: this verse -> TG entries (through VerseTGRef)
		edge.To("footnote_tg_entries", TopicalGuideEntry.Type).
			Through("verse_tg_refs", VerseTGRef.Type),

		// Footnotes: this verse -> BD entries (through VerseBDRef)
		edge.To("footnote_bd_entries", BibleDictEntry.Type).
			Through("verse_bd_refs", VerseBDRef.Type),

		// Footnotes: this verse -> JST passages (through VerseJSTRef)
		edge.To("footnote_jst_passages", JSTPassage.Type).
			Through("verse_jst_refs", VerseJSTRef.Type),

		// Back-refs: TG topics that reference this verse
		edge.From("tg_refs", TopicalGuideEntry.Type).
			Ref("verse_refs").
			Through("tg_verse_refs", TGVerseRef.Type),

		// Back-refs: BD entries that reference this verse
		edge.From("bd_refs", BibleDictEntry.Type).
			Ref("verse_refs").
			Through("bd_verse_refs", BDVerseRef.Type),

		// Back-refs: Index entries that reference this verse
		edge.From("idx_refs", IndexEntry.Type).
			Ref("verse_refs").
			Through("idx_verse_refs", IDXVerseRef.Type),

		// Back-refs: JST passages that modify this verse (compare)
		edge.From("jst_compares", JSTPassage.Type).
			Ref("compare_verses"),

		// Back-refs: verse groups containing this verse
		edge.From("verse_groups", VerseGroup.Type).
			Ref("verses"),
	}
}
