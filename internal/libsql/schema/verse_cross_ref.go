package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// VerseCrossRef is the junction entity for scripture cross-references
// between verses (from footnotes). Approximately 23,616 edges.
type VerseCrossRef struct {
	ent.Schema
}

func (VerseCrossRef) Fields() []ent.Field {
	return []ent.Field{
		field.String("footnote_marker").
			NotEmpty().
			Comment("Footnote marker, e.g. '1a', '13c'"),
		field.String("reference_text").
			Optional().
			Comment("The word/phrase in the verse being annotated"),
		field.Int("verse_id").
			Comment("FK to source verse (owner side)"),
		field.Int("cross_ref_target_id").
			Comment("FK to target verse"),
		field.Int("target_end_verse_id").
			Optional().
			Nillable().
			Comment("Last verse ID if reference is a range"),
	}
}

func (VerseCrossRef) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("verse", Verse.Type).
			Required().
			Unique().
			Field("verse_id"),
		edge.To("cross_ref_target", Verse.Type).
			Required().
			Unique().
			Field("cross_ref_target_id"),
	}
}
