package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// Volume represents a canonical scripture collection
// (e.g., Old Testament, Book of Mormon, Doctrine and Covenants).
type Volume struct {
	ent.Schema
}

func (Volume) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").
			NotEmpty().
			Unique().
			Comment("Display name, e.g. 'Old Testament'"),
		field.String("abbreviation").
			NotEmpty().
			Unique().
			Comment("URL slug, e.g. 'ot', 'bofm', 'dc-testament'"),
	}
}

func (Volume) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("books", Book.Type),
	}
}
