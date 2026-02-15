package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect"
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
			SchemaType(map[string]string{
				dialect.SQLite: "F32_BLOB(1024)",
			}).
			Comment("Vector embedding (1024-dim float32)"),
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
