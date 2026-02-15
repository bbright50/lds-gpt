package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// JSTPassage represents a Joseph Smith Translation passage.
type JSTPassage struct {
	ent.Schema
}

func (JSTPassage) Fields() []ent.Field {
	return []ent.Field{
		field.String("book").
			NotEmpty().
			Comment("Source book name, e.g. '1 Samuel'"),
		field.String("chapter").
			NotEmpty().
			Comment("Source chapter, e.g. '16'"),
		field.String("comprises").
			NotEmpty().
			Comment("Verse range string, e.g. '14-16, 23'"),
		field.String("compare_ref").
			Optional().
			Comment("Original verse reference to compare against"),
		field.Text("summary").
			Optional().
			Comment("Summary of the JST change"),
		field.Text("text").
			NotEmpty().
			Comment("Concatenated JST verse text"),
		field.Bytes("embedding").
			Optional().
			Nillable().
			SchemaType(map[string]string{
				dialect.SQLite: "F32_BLOB(1024)",
			}).
			Comment("Vector embedding (1024-dim float32)"),
	}
}

func (JSTPassage) Edges() []ent.Edge {
	return []ent.Edge{
		// JST -> Verse: original verses this passage modifies (from compare field)
		edge.To("compare_verses", Verse.Type),

		// Back-ref: verses whose footnotes reference this JST passage
		edge.From("footnote_verses", Verse.Type).
			Ref("footnote_jst_passages").
			Through("verse_jst_refs", VerseJSTRef.Type),
	}
}
