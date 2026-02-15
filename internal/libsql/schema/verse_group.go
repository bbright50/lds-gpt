package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// VerseGroup is a sliding window of consecutive verses for embedding.
// These are the primary units for semantic search in RAG.
type VerseGroup struct {
	ent.Schema
}

func (VerseGroup) Fields() []ent.Field {
	return []ent.Field{
		field.Text("text").
			NotEmpty().
			Comment("Concatenated verse texts"),
		field.Bytes("embedding").
			Optional().
			Nillable().
			SchemaType(map[string]string{
				dialect.SQLite: "F32_BLOB(1024)",
			}).
			Comment("Vector embedding (1024-dim float32)"),
		field.Int("start_verse_number").
			Positive().
			Comment("First verse number in group"),
		field.Int("end_verse_number").
			Positive().
			Comment("Last verse number in group"),
	}
}

func (VerseGroup) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("chapter", Chapter.Type).
			Ref("verse_groups").
			Required().
			Unique(),
		edge.To("verses", Verse.Type),
	}
}
