package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// VerseJSTRef is the junction entity for footnotes linking verses
// to JST passages. Approximately 499 edges.
type VerseJSTRef struct {
	ent.Schema
}

func (VerseJSTRef) Fields() []ent.Field {
	return []ent.Field{
		field.String("footnote_marker").
			NotEmpty().
			Comment("Footnote marker"),
		field.Int("verse_id").
			Comment("FK to verse"),
		field.Int("jst_passage_id").
			Comment("FK to JST passage"),
	}
}

func (VerseJSTRef) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("verse", Verse.Type).
			Required().
			Unique().
			Field("verse_id"),
		edge.To("jst_passage", JSTPassage.Type).
			Required().
			Unique().
			Field("jst_passage_id"),
	}
}
