package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// TGVerseRef is the junction entity for scripture references listed
// under a Topical Guide topic. Thousands of edges.
type TGVerseRef struct {
	ent.Schema
}

func (TGVerseRef) Fields() []ent.Field {
	return []ent.Field{
		field.String("phrase").
			Optional().
			Comment("Scripture phrase quoted, e.g. 'exalt himself shall be abased'"),
		field.Int("tg_entry_id").
			Comment("FK to topical guide entry"),
		field.Int("verse_id").
			Comment("FK to verse"),
		field.Int("target_end_verse_id").
			Optional().
			Nillable().
			Comment("Last verse ID if reference is a range"),
	}
}

func (TGVerseRef) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("tg_entry", TopicalGuideEntry.Type).
			Required().
			Unique().
			Field("tg_entry_id"),
		edge.To("verse", Verse.Type).
			Required().
			Unique().
			Field("verse_id"),
	}
}
