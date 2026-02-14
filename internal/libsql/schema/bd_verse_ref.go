package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// BDVerseRef is the junction entity for scripture references in
// Bible Dictionary entries. Thousands of edges.
type BDVerseRef struct {
	ent.Schema
}

func (BDVerseRef) Fields() []ent.Field {
	return []ent.Field{
		field.Int("bd_entry_id").
			Comment("FK to bible dictionary entry"),
		field.Int("verse_id").
			Comment("FK to verse"),
		field.Int("target_end_verse_id").
			Optional().
			Nillable().
			Comment("Last verse ID if reference is a range"),
	}
}

func (BDVerseRef) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("bd_entry", BibleDictEntry.Type).
			Required().
			Unique().
			Field("bd_entry_id"),
		edge.To("verse", Verse.Type).
			Required().
			Unique().
			Field("verse_id"),
	}
}
