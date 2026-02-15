package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// BibleDictEntry represents an entry in the Bible Dictionary.
type BibleDictEntry struct {
	ent.Schema
}

func (BibleDictEntry) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").
			NotEmpty().
			Unique().
			Comment("Entry name, e.g. 'Aaron', 'Aaronic Priesthood'"),
		field.Text("text").
			NotEmpty().
			Comment("Full definition/article text"),
		field.Bytes("embedding").
			Optional().
			Nillable().
			SchemaType(map[string]string{
				dialect.SQLite: "F32_BLOB(1024)",
			}).
			Comment("Vector embedding (1024-dim float32)"),
	}
}

func (BibleDictEntry) Edges() []ent.Edge {
	return []ent.Edge{
		// Self-referencing: "See also" cross-references within BD
		edge.To("see_also", BibleDictEntry.Type),

		// BD -> Verse references (with optional range metadata)
		edge.To("verse_refs", Verse.Type).
			Through("bd_verse_refs", BDVerseRef.Type),

		// Back-ref: TG entries referencing this BD entry
		edge.From("tg_refs", TopicalGuideEntry.Type).
			Ref("bd_refs"),

		// Back-ref: verses whose footnotes reference this BD entry
		edge.From("footnote_verses", Verse.Type).
			Ref("footnote_bd_entries").
			Through("verse_bd_refs", VerseBDRef.Type),

		// Back-ref: index entries referencing this BD entry
		edge.From("index_refs", IndexEntry.Type).
			Ref("bd_refs"),
	}
}
