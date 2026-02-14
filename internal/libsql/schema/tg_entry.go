package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// TopicalGuideEntry represents a topic in the Topical Guide.
type TopicalGuideEntry struct {
	ent.Schema
}

func (TopicalGuideEntry) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").
			NotEmpty().
			Unique().
			Comment("Topic name, e.g. 'Atonement', 'Birthright'"),
		field.Bytes("embedding").
			Optional().
			Nillable().
			Comment("Vector embedding of name + phrase snippets (packed float32 blob)"),
	}
}

func (TopicalGuideEntry) Edges() []ent.Edge {
	return []ent.Edge{
		// Self-referencing: "see also" cross-references within TG
		edge.To("see_also", TopicalGuideEntry.Type),

		// TG -> BD references
		edge.To("bd_refs", BibleDictEntry.Type),

		// TG -> Verse references (with phrase metadata)
		edge.To("verse_refs", Verse.Type).
			Through("tg_verse_refs", TGVerseRef.Type),

		// Back-ref: verses whose footnotes reference this TG entry
		edge.From("footnote_verses", Verse.Type).
			Ref("footnote_tg_entries").
			Through("verse_tg_refs", VerseTGRef.Type),

		// Back-ref: index entries referencing this TG entry
		edge.From("index_refs", IndexEntry.Type).
			Ref("tg_refs"),
	}
}
