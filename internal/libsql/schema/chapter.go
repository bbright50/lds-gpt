package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// Chapter represents a chapter within a book.
type Chapter struct {
	ent.Schema
}

func (Chapter) Fields() []ent.Field {
	return []ent.Field{
		field.Int("number").
			Positive().
			Comment("Chapter number within the book"),
		field.Text("summary").
			Optional().
			Comment("Chapter heading/summary text"),
		field.Bytes("summary_embedding").
			Optional().
			Nillable().
			Comment("Vector embedding of summary (packed float32 blob)"),
		field.String("url").
			Optional().
			Comment("Source URL for this chapter"),
	}
}

func (Chapter) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("book", Book.Type).
			Ref("chapters").
			Required().
			Unique(),
		edge.To("verses", Verse.Type),
		edge.To("verse_groups", VerseGroup.Type),
	}
}
