package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// VerseBDRef is the junction entity for footnotes linking verses
// to Bible Dictionary entries. Approximately 33 edges.
type VerseBDRef struct {
	ent.Schema
}

func (VerseBDRef) Fields() []ent.Field {
	return []ent.Field{
		field.String("footnote_marker").
			NotEmpty().
			Comment("Footnote marker"),
		field.String("reference_text").
			Optional().
			Comment("The word/phrase annotated in the verse"),
		field.Int("verse_id").
			Comment("FK to verse"),
		field.Int("bd_entry_id").
			Comment("FK to bible dictionary entry"),
	}
}

func (VerseBDRef) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("verse", Verse.Type).
			Required().
			Unique().
			Field("verse_id"),
		edge.To("bd_entry", BibleDictEntry.Type).
			Required().
			Unique().
			Field("bd_entry_id"),
	}
}
