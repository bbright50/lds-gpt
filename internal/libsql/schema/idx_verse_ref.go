package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// IDXVerseRef is the junction entity for scripture references in
// Triple Combination Index entries. Thousands of edges.
type IDXVerseRef struct {
	ent.Schema
}

func (IDXVerseRef) Fields() []ent.Field {
	return []ent.Field{
		field.String("phrase").
			Optional().
			Comment("Quoted phrase from the index entry"),
		field.Int("index_entry_id").
			Comment("FK to index entry"),
		field.Int("verse_id").
			Comment("FK to verse"),
		field.Int("target_end_verse_id").
			Optional().
			Nillable().
			Comment("Last verse ID if reference is a range"),
	}
}

func (IDXVerseRef) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("index_entry", IndexEntry.Type).
			Required().
			Unique().
			Field("index_entry_id"),
		edge.To("verse", Verse.Type).
			Required().
			Unique().
			Field("verse_id"),
	}
}
