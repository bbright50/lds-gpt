package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// Book represents a book within a volume (e.g., Genesis, 1 Nephi, D&C).
type Book struct {
	ent.Schema
}

func (Book) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").
			NotEmpty().
			Comment("Full display name, e.g. 'The First Book of Nephi'"),
		field.String("slug").
			NotEmpty().
			Comment("URL/directory slug, e.g. '1-ne', 'gen'"),
		field.String("url_path").
			NotEmpty().
			Comment("Volume-relative path for URL construction"),
	}
}

func (Book) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("volume", Volume.Type).
			Ref("books").
			Required().
			Unique(),
		edge.To("chapters", Chapter.Type),
	}
}
