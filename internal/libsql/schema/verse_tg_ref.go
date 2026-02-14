package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// VerseTGRef is the junction entity for footnotes linking verses
// to Topical Guide topics. Approximately 18,266 edges.
type VerseTGRef struct {
	ent.Schema
}

func (VerseTGRef) Fields() []ent.Field {
	return []ent.Field{
		field.String("footnote_marker").
			NotEmpty().
			Comment("Footnote marker, e.g. '1a'"),
		field.String("reference_text").
			Optional().
			Comment("The word/phrase annotated in the verse"),
		field.String("tg_topic_text").
			Optional().
			Comment("Raw TG reference text, e.g. 'TG Birthright.'"),
		field.Int("verse_id").
			Comment("FK to verse"),
		field.Int("tg_entry_id").
			Comment("FK to topical guide entry"),
	}
}

func (VerseTGRef) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("verse", Verse.Type).
			Required().
			Unique().
			Field("verse_id"),
		edge.To("tg_entry", TopicalGuideEntry.Type).
			Required().
			Unique().
			Field("tg_entry_id"),
	}
}
