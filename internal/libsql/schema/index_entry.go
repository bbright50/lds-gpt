package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// IndexEntry represents an entry in the Triple Combination Index.
type IndexEntry struct {
	ent.Schema
}

func (IndexEntry) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").
			NotEmpty().
			Unique().
			Comment("Entry name, e.g. 'Aaron1--brother of Moses'"),
		field.Bytes("embedding").
			Optional().
			Nillable().
			Comment("Vector embedding of name + phrase snippets (packed float32 blob)"),
	}
}

func (IndexEntry) Edges() []ent.Edge {
	return []ent.Edge{
		// Self-referencing: cross-references within the index
		edge.To("see_also", IndexEntry.Type),

		// IDX -> TG references
		edge.To("tg_refs", TopicalGuideEntry.Type),

		// IDX -> BD references
		edge.To("bd_refs", BibleDictEntry.Type),

		// IDX -> Verse references (with phrase metadata)
		edge.To("verse_refs", Verse.Type).
			Through("idx_verse_refs", IDXVerseRef.Type),
	}
}
