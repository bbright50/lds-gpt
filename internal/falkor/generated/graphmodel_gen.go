package generated

import "github.com/tab58/go-ormql/pkg/schema"

var GraphModel = schema.GraphModel{
	Nodes: []schema.NodeDefinition{
		{
			Name:   "VerseGroup",
			Labels: []string{"VerseGroup"},
			Fields: []schema.FieldDefinition{
				{Name: "id", GraphQLType: "ID!", GoType: "string", CypherType: "STRING", IsID: true},
				{Name: "text", GraphQLType: "String!", GoType: "string", CypherType: "STRING"},
				{Name: "startVerseNumber", GraphQLType: "Int!", GoType: "int", CypherType: "INTEGER"},
				{Name: "endVerseNumber", GraphQLType: "Int!", GoType: "int", CypherType: "INTEGER"},
				{Name: "embedding", GraphQLType: "[Float!]!", GoType: "[]float64", CypherType: "LIST<FLOAT>", IsList: true},
			},
			VectorField: &schema.VectorFieldDefinition{
				Name:       "embedding",
				IndexName:  "verse_group_embedding",
				Dimensions: 1024,
				Similarity: "cosine",
			},
		},
		{
			Name:   "JSTPassage",
			Labels: []string{"JSTPassage"},
			Fields: []schema.FieldDefinition{
				{Name: "id", GraphQLType: "ID!", GoType: "string", CypherType: "STRING", IsID: true},
				{Name: "book", GraphQLType: "String!", GoType: "string", CypherType: "STRING"},
				{Name: "chapter", GraphQLType: "String!", GoType: "string", CypherType: "STRING"},
				{Name: "comprises", GraphQLType: "String!", GoType: "string", CypherType: "STRING"},
				{Name: "compareRef", GraphQLType: "String", GoType: "*string", CypherType: "STRING", Nullable: true},
				{Name: "summary", GraphQLType: "String", GoType: "*string", CypherType: "STRING", Nullable: true},
				{Name: "text", GraphQLType: "String!", GoType: "string", CypherType: "STRING"},
				{Name: "embedding", GraphQLType: "[Float!]!", GoType: "[]float64", CypherType: "LIST<FLOAT>", IsList: true},
			},
			VectorField: &schema.VectorFieldDefinition{
				Name:       "embedding",
				IndexName:  "jst_embedding",
				Dimensions: 1024,
				Similarity: "cosine",
			},
		},
		{
			Name:   "Chapter",
			Labels: []string{"Chapter"},
			Fields: []schema.FieldDefinition{
				{Name: "id", GraphQLType: "ID!", GoType: "string", CypherType: "STRING", IsID: true},
				{Name: "number", GraphQLType: "Int!", GoType: "int", CypherType: "INTEGER"},
				{Name: "summary", GraphQLType: "String", GoType: "*string", CypherType: "STRING", Nullable: true},
				{Name: "summaryEmbedding", GraphQLType: "[Float!]!", GoType: "[]float64", CypherType: "LIST<FLOAT>", IsList: true},
				{Name: "url", GraphQLType: "String", GoType: "*string", CypherType: "STRING", Nullable: true},
			},
			VectorField: &schema.VectorFieldDefinition{
				Name:       "summaryEmbedding",
				IndexName:  "chapter_summary_embedding",
				Dimensions: 1024,
				Similarity: "cosine",
			},
		},
		{
			Name:   "Verse",
			Labels: []string{"Verse"},
			Fields: []schema.FieldDefinition{
				{Name: "id", GraphQLType: "ID!", GoType: "string", CypherType: "STRING", IsID: true},
				{Name: "number", GraphQLType: "Int!", GoType: "int", CypherType: "INTEGER"},
				{Name: "reference", GraphQLType: "String!", GoType: "string", CypherType: "STRING"},
				{Name: "text", GraphQLType: "String!", GoType: "string", CypherType: "STRING"},
				{Name: "translationNotes", GraphQLType: "String", GoType: "*string", CypherType: "STRING", Nullable: true},
				{Name: "alternateReadings", GraphQLType: "String", GoType: "*string", CypherType: "STRING", Nullable: true},
				{Name: "explanatoryNotes", GraphQLType: "String", GoType: "*string", CypherType: "STRING", Nullable: true},
			},
		},
		{
			Name:   "BibleDictEntry",
			Labels: []string{"BibleDictEntry"},
			Fields: []schema.FieldDefinition{
				{Name: "id", GraphQLType: "ID!", GoType: "string", CypherType: "STRING", IsID: true},
				{Name: "name", GraphQLType: "String!", GoType: "string", CypherType: "STRING"},
				{Name: "text", GraphQLType: "String!", GoType: "string", CypherType: "STRING"},
				{Name: "embedding", GraphQLType: "[Float!]!", GoType: "[]float64", CypherType: "LIST<FLOAT>", IsList: true},
			},
			VectorField: &schema.VectorFieldDefinition{
				Name:       "embedding",
				IndexName:  "bd_embedding",
				Dimensions: 1024,
				Similarity: "cosine",
			},
		},
		{
			Name:   "IndexEntry",
			Labels: []string{"IndexEntry"},
			Fields: []schema.FieldDefinition{
				{Name: "id", GraphQLType: "ID!", GoType: "string", CypherType: "STRING", IsID: true},
				{Name: "name", GraphQLType: "String!", GoType: "string", CypherType: "STRING"},
				{Name: "embedding", GraphQLType: "[Float!]!", GoType: "[]float64", CypherType: "LIST<FLOAT>", IsList: true},
			},
			VectorField: &schema.VectorFieldDefinition{
				Name:       "embedding",
				IndexName:  "idx_embedding",
				Dimensions: 1024,
				Similarity: "cosine",
			},
		},
		{
			Name:   "TopicalGuideEntry",
			Labels: []string{"TopicalGuideEntry"},
			Fields: []schema.FieldDefinition{
				{Name: "id", GraphQLType: "ID!", GoType: "string", CypherType: "STRING", IsID: true},
				{Name: "name", GraphQLType: "String!", GoType: "string", CypherType: "STRING"},
				{Name: "embedding", GraphQLType: "[Float!]!", GoType: "[]float64", CypherType: "LIST<FLOAT>", IsList: true},
			},
			VectorField: &schema.VectorFieldDefinition{
				Name:       "embedding",
				IndexName:  "tg_embedding",
				Dimensions: 1024,
				Similarity: "cosine",
			},
		},
		{
			Name:   "Volume",
			Labels: []string{"Volume"},
			Fields: []schema.FieldDefinition{
				{Name: "id", GraphQLType: "ID!", GoType: "string", CypherType: "STRING", IsID: true},
				{Name: "name", GraphQLType: "String!", GoType: "string", CypherType: "STRING"},
				{Name: "abbreviation", GraphQLType: "String!", GoType: "string", CypherType: "STRING"},
			},
		},
		{
			Name:   "Book",
			Labels: []string{"Book"},
			Fields: []schema.FieldDefinition{
				{Name: "id", GraphQLType: "ID!", GoType: "string", CypherType: "STRING", IsID: true},
				{Name: "name", GraphQLType: "String!", GoType: "string", CypherType: "STRING"},
				{Name: "slug", GraphQLType: "String!", GoType: "string", CypherType: "STRING"},
				{Name: "urlPath", GraphQLType: "String!", GoType: "string", CypherType: "STRING"},
			},
		},
	},
	Relationships: []schema.RelationshipDefinition{
		{
			FieldName: "chapter",
			RelType:   "HAS_GROUP",
			Direction: schema.DirectionIN,
			FromNode:  "VerseGroup",
			ToNode:    "Chapter", IsList: false,
		},
		{
			FieldName: "verses",
			RelType:   "INCLUDES",
			Direction: schema.DirectionOUT,
			FromNode:  "VerseGroup",
			ToNode:    "Verse", IsList: true,
		},
		{
			FieldName: "compareVerses",
			RelType:   "COMPARES",
			Direction: schema.DirectionOUT,
			FromNode:  "JSTPassage",
			ToNode:    "Verse", IsList: true,
		},
		{
			FieldName: "book",
			RelType:   "CONTAINS",
			Direction: schema.DirectionIN,
			FromNode:  "Chapter",
			ToNode:    "Book", IsList: false,
		},
		{
			FieldName: "verses",
			RelType:   "HAS_VERSE",
			Direction: schema.DirectionOUT,
			FromNode:  "Chapter",
			ToNode:    "Verse", IsList: true,
		},
		{
			FieldName: "verseGroups",
			RelType:   "HAS_GROUP",
			Direction: schema.DirectionOUT,
			FromNode:  "Chapter",
			ToNode:    "VerseGroup", IsList: true,
		},
		{
			FieldName: "chapter",
			RelType:   "HAS_VERSE",
			Direction: schema.DirectionIN,
			FromNode:  "Verse",
			ToNode:    "Chapter", IsList: false,
		},
		{
			FieldName: "crossRefsOut",
			RelType:   "CROSS_REF",
			Direction: schema.DirectionOUT,
			FromNode:  "Verse",
			ToNode:    "Verse", IsList: true,
			Properties: &schema.PropertiesDefinition{
				TypeName: "VerseCrossRefProps",
				Fields: []schema.FieldDefinition{
					{Name: "category", GraphQLType: "String", GoType: "*string", CypherType: "STRING", Nullable: true},
					{Name: "footnoteMarker", GraphQLType: "String", GoType: "*string", CypherType: "STRING", Nullable: true},
					{Name: "referenceText", GraphQLType: "String", GoType: "*string", CypherType: "STRING", Nullable: true},
				},
			},
		},
		{
			FieldName: "crossRefsIn",
			RelType:   "CROSS_REF",
			Direction: schema.DirectionIN,
			FromNode:  "Verse",
			ToNode:    "Verse", IsList: true,
			Properties: &schema.PropertiesDefinition{
				TypeName: "VerseCrossRefProps",
				Fields: []schema.FieldDefinition{
					{Name: "category", GraphQLType: "String", GoType: "*string", CypherType: "STRING", Nullable: true},
					{Name: "footnoteMarker", GraphQLType: "String", GoType: "*string", CypherType: "STRING", Nullable: true},
					{Name: "referenceText", GraphQLType: "String", GoType: "*string", CypherType: "STRING", Nullable: true},
				},
			},
		},
		{
			FieldName: "tgFootnotes",
			RelType:   "TG_FOOTNOTE",
			Direction: schema.DirectionOUT,
			FromNode:  "Verse",
			ToNode:    "TopicalGuideEntry", IsList: true,
			Properties: &schema.PropertiesDefinition{
				TypeName: "VerseTGRefProps",
				Fields: []schema.FieldDefinition{
					{Name: "footnoteMarker", GraphQLType: "String", GoType: "*string", CypherType: "STRING", Nullable: true},
				},
			},
		},
		{
			FieldName: "bdFootnotes",
			RelType:   "BD_FOOTNOTE",
			Direction: schema.DirectionOUT,
			FromNode:  "Verse",
			ToNode:    "BibleDictEntry", IsList: true,
			Properties: &schema.PropertiesDefinition{
				TypeName: "VerseBDRefProps",
				Fields: []schema.FieldDefinition{
					{Name: "footnoteMarker", GraphQLType: "String", GoType: "*string", CypherType: "STRING", Nullable: true},
				},
			},
		},
		{
			FieldName: "jstFootnotes",
			RelType:   "JST_FOOTNOTE",
			Direction: schema.DirectionOUT,
			FromNode:  "Verse",
			ToNode:    "JSTPassage", IsList: true,
			Properties: &schema.PropertiesDefinition{
				TypeName: "VerseJSTRefProps",
				Fields: []schema.FieldDefinition{
					{Name: "footnoteMarker", GraphQLType: "String", GoType: "*string", CypherType: "STRING", Nullable: true},
				},
			},
		},
		{
			FieldName: "seeAlso",
			RelType:   "BD_SEE_ALSO",
			Direction: schema.DirectionOUT,
			FromNode:  "BibleDictEntry",
			ToNode:    "BibleDictEntry", IsList: true,
		},
		{
			FieldName: "verseRefs",
			RelType:   "BD_VERSE_REF",
			Direction: schema.DirectionOUT,
			FromNode:  "BibleDictEntry",
			ToNode:    "Verse", IsList: true,
			Properties: &schema.PropertiesDefinition{
				TypeName: "BDVerseRefProps",
				Fields: []schema.FieldDefinition{
					{Name: "targetEndVerseId", GraphQLType: "String", GoType: "*string", CypherType: "STRING", Nullable: true},
				},
			},
		},
		{
			FieldName: "seeAlso",
			RelType:   "IDX_SEE_ALSO",
			Direction: schema.DirectionOUT,
			FromNode:  "IndexEntry",
			ToNode:    "IndexEntry", IsList: true,
		},
		{
			FieldName: "tgRefs",
			RelType:   "IDX_TG_REF",
			Direction: schema.DirectionOUT,
			FromNode:  "IndexEntry",
			ToNode:    "TopicalGuideEntry", IsList: true,
		},
		{
			FieldName: "bdRefs",
			RelType:   "IDX_BD_REF",
			Direction: schema.DirectionOUT,
			FromNode:  "IndexEntry",
			ToNode:    "BibleDictEntry", IsList: true,
		},
		{
			FieldName: "verseRefs",
			RelType:   "IDX_VERSE_REF",
			Direction: schema.DirectionOUT,
			FromNode:  "IndexEntry",
			ToNode:    "Verse", IsList: true,
			Properties: &schema.PropertiesDefinition{
				TypeName: "IDXVerseRefProps",
				Fields: []schema.FieldDefinition{
					{Name: "phrase", GraphQLType: "String", GoType: "*string", CypherType: "STRING", Nullable: true},
				},
			},
		},
		{
			FieldName: "seeAlso",
			RelType:   "TG_SEE_ALSO",
			Direction: schema.DirectionOUT,
			FromNode:  "TopicalGuideEntry",
			ToNode:    "TopicalGuideEntry", IsList: true,
		},
		{
			FieldName: "bdRefs",
			RelType:   "TG_BD_REF",
			Direction: schema.DirectionOUT,
			FromNode:  "TopicalGuideEntry",
			ToNode:    "BibleDictEntry", IsList: true,
		},
		{
			FieldName: "verseRefs",
			RelType:   "TG_VERSE_REF",
			Direction: schema.DirectionOUT,
			FromNode:  "TopicalGuideEntry",
			ToNode:    "Verse", IsList: true,
			Properties: &schema.PropertiesDefinition{
				TypeName: "TGVerseRefProps",
				Fields: []schema.FieldDefinition{
					{Name: "phrase", GraphQLType: "String", GoType: "*string", CypherType: "STRING", Nullable: true},
				},
			},
		},
		{
			FieldName: "books",
			RelType:   "CONTAINS",
			Direction: schema.DirectionOUT,
			FromNode:  "Volume",
			ToNode:    "Book", IsList: true,
		},
		{
			FieldName: "volume",
			RelType:   "CONTAINS",
			Direction: schema.DirectionIN,
			FromNode:  "Book",
			ToNode:    "Volume", IsList: false,
		},
		{
			FieldName: "chapters",
			RelType:   "CONTAINS",
			Direction: schema.DirectionOUT,
			FromNode:  "Book",
			ToNode:    "Chapter", IsList: true,
		},
	},
}

var AugmentedSchemaSDL = `type VerseGroup {
  id: ID!
  text: String!
  startVerseNumber: Int!
  endVerseNumber: Int!
  embedding: [Float!]!
  chapterConnection(first: Int, after: String, where: ChapterWhere, sort: [ChapterSort!]): VerseGroupChapterConnection!
  versesConnection(first: Int, after: String, where: VerseWhere, sort: [VerseSort!]): VerseGroupVersesConnection!
}

type JSTPassage {
  id: ID!
  book: String!
  chapter: String!
  comprises: String!
  compareRef: String
  summary: String
  text: String!
  embedding: [Float!]!
  compareVersesConnection(first: Int, after: String, where: VerseWhere, sort: [VerseSort!]): JSTPassageCompareVersesConnection!
}

type Chapter {
  id: ID!
  number: Int!
  summary: String
  summaryEmbedding: [Float!]!
  url: String
  bookConnection(first: Int, after: String, where: BookWhere, sort: [BookSort!]): ChapterBookConnection!
  versesConnection(first: Int, after: String, where: VerseWhere, sort: [VerseSort!]): ChapterVersesConnection!
  verseGroupsConnection(first: Int, after: String, where: VerseGroupWhere, sort: [VerseGroupSort!]): ChapterVerseGroupsConnection!
}

type Verse {
  id: ID!
  number: Int!
  reference: String!
  text: String!
  translationNotes: String
  alternateReadings: String
  explanatoryNotes: String
  chapterConnection(first: Int, after: String, where: ChapterWhere, sort: [ChapterSort!]): VerseChapterConnection!
  crossRefsOutConnection(first: Int, after: String, where: VerseWhere, sort: [VerseSort!]): VerseCrossRefsOutConnection!
  crossRefsInConnection(first: Int, after: String, where: VerseWhere, sort: [VerseSort!]): VerseCrossRefsInConnection!
  tgFootnotesConnection(first: Int, after: String, where: TopicalGuideEntryWhere, sort: [TopicalGuideEntrySort!]): VerseTgFootnotesConnection!
  bdFootnotesConnection(first: Int, after: String, where: BibleDictEntryWhere, sort: [BibleDictEntrySort!]): VerseBdFootnotesConnection!
  jstFootnotesConnection(first: Int, after: String, where: JSTPassageWhere, sort: [JSTPassageSort!]): VerseJstFootnotesConnection!
}

type BibleDictEntry {
  id: ID!
  name: String!
  text: String!
  embedding: [Float!]!
  seeAlsoConnection(first: Int, after: String, where: BibleDictEntryWhere, sort: [BibleDictEntrySort!]): BibleDictEntrySeeAlsoConnection!
  verseRefsConnection(first: Int, after: String, where: VerseWhere, sort: [VerseSort!]): BibleDictEntryVerseRefsConnection!
}

type IndexEntry {
  id: ID!
  name: String!
  embedding: [Float!]!
  seeAlsoConnection(first: Int, after: String, where: IndexEntryWhere, sort: [IndexEntrySort!]): IndexEntrySeeAlsoConnection!
  tgRefsConnection(first: Int, after: String, where: TopicalGuideEntryWhere, sort: [TopicalGuideEntrySort!]): IndexEntryTgRefsConnection!
  bdRefsConnection(first: Int, after: String, where: BibleDictEntryWhere, sort: [BibleDictEntrySort!]): IndexEntryBdRefsConnection!
  verseRefsConnection(first: Int, after: String, where: VerseWhere, sort: [VerseSort!]): IndexEntryVerseRefsConnection!
}

type TopicalGuideEntry {
  id: ID!
  name: String!
  embedding: [Float!]!
  seeAlsoConnection(first: Int, after: String, where: TopicalGuideEntryWhere, sort: [TopicalGuideEntrySort!]): TopicalGuideEntrySeeAlsoConnection!
  bdRefsConnection(first: Int, after: String, where: BibleDictEntryWhere, sort: [BibleDictEntrySort!]): TopicalGuideEntryBdRefsConnection!
  verseRefsConnection(first: Int, after: String, where: VerseWhere, sort: [VerseSort!]): TopicalGuideEntryVerseRefsConnection!
}

type Volume {
  id: ID!
  name: String!
  abbreviation: String!
  booksConnection(first: Int, after: String, where: BookWhere, sort: [BookSort!]): VolumeBooksConnection!
}

type Book {
  id: ID!
  name: String!
  slug: String!
  urlPath: String!
  volumeConnection(first: Int, after: String, where: VolumeWhere, sort: [VolumeSort!]): BookVolumeConnection!
  chaptersConnection(first: Int, after: String, where: ChapterWhere, sort: [ChapterSort!]): BookChaptersConnection!
}

input VerseGroupWhere {
  id: ID
  id_gt: ID
  id_gte: ID
  id_lt: ID
  id_lte: ID
  id_contains: ID
  id_startsWith: ID
  id_endsWith: ID
  id_regex: ID
  id_in: [ID!]
  id_nin: [ID!]
  id_not: ID
  id_isNull: Boolean
  text: String
  text_gt: String
  text_gte: String
  text_lt: String
  text_lte: String
  text_contains: String
  text_startsWith: String
  text_endsWith: String
  text_regex: String
  text_in: [String!]
  text_nin: [String!]
  text_not: String
  text_isNull: Boolean
  startVerseNumber: Int
  startVerseNumber_gt: Int
  startVerseNumber_gte: Int
  startVerseNumber_lt: Int
  startVerseNumber_lte: Int
  startVerseNumber_in: [Int!]
  startVerseNumber_nin: [Int!]
  startVerseNumber_not: Int
  startVerseNumber_isNull: Boolean
  endVerseNumber: Int
  endVerseNumber_gt: Int
  endVerseNumber_gte: Int
  endVerseNumber_lt: Int
  endVerseNumber_lte: Int
  endVerseNumber_in: [Int!]
  endVerseNumber_nin: [Int!]
  endVerseNumber_not: Int
  endVerseNumber_isNull: Boolean
  chapter: ChapterWhere
  verses_some: VerseWhere
  AND: [VerseGroupWhere!]
  OR: [VerseGroupWhere!]
  NOT: VerseGroupWhere
}

input VerseGroupSort {
  id: SortDirection
  text: SortDirection
  startVerseNumber: SortDirection
  endVerseNumber: SortDirection
}

input VerseGroupCreateInput {
  text: String!
  startVerseNumber: Int!
  endVerseNumber: Int!
  embedding: [Float!]!
  chapter: VerseGroupChapterFieldInput
  verses: VerseGroupVersesFieldInput
}

input VerseGroupUpdateInput {
  text: String
  startVerseNumber: Int
  endVerseNumber: Int
  embedding: [Float!]
  chapter: VerseGroupChapterUpdateFieldInput
  verses: VerseGroupVersesUpdateFieldInput
}

type CreateVerseGroupsMutationResponse {
  verseGroups: [VerseGroup!]!
}

type UpdateVerseGroupsMutationResponse {
  verseGroups: [VerseGroup!]!
}

input VerseGroupMatchInput {
  text: String
  startVerseNumber: Int
  endVerseNumber: Int
}

input VerseGroupMergeInput {
  match: VerseGroupMatchInput!
  onCreate: VerseGroupCreateInput
  onMatch: VerseGroupUpdateInput
}

type MergeVerseGroupsMutationResponse {
  verseGroups: [VerseGroup!]!
}

type VerseGroupsConnection {
  edges: [VerseGroupEdge!]!
  pageInfo: PageInfo!
  totalCount: Int!
}

type VerseGroupEdge {
  node: VerseGroup!
  cursor: String!
}

type VerseGroupSimilarResult {
  score: Float!
  node: VerseGroup!
}

input JSTPassageWhere {
  id: ID
  id_gt: ID
  id_gte: ID
  id_lt: ID
  id_lte: ID
  id_contains: ID
  id_startsWith: ID
  id_endsWith: ID
  id_regex: ID
  id_in: [ID!]
  id_nin: [ID!]
  id_not: ID
  id_isNull: Boolean
  book: String
  book_gt: String
  book_gte: String
  book_lt: String
  book_lte: String
  book_contains: String
  book_startsWith: String
  book_endsWith: String
  book_regex: String
  book_in: [String!]
  book_nin: [String!]
  book_not: String
  book_isNull: Boolean
  chapter: String
  chapter_gt: String
  chapter_gte: String
  chapter_lt: String
  chapter_lte: String
  chapter_contains: String
  chapter_startsWith: String
  chapter_endsWith: String
  chapter_regex: String
  chapter_in: [String!]
  chapter_nin: [String!]
  chapter_not: String
  chapter_isNull: Boolean
  comprises: String
  comprises_gt: String
  comprises_gte: String
  comprises_lt: String
  comprises_lte: String
  comprises_contains: String
  comprises_startsWith: String
  comprises_endsWith: String
  comprises_regex: String
  comprises_in: [String!]
  comprises_nin: [String!]
  comprises_not: String
  comprises_isNull: Boolean
  compareRef: String
  compareRef_gt: String
  compareRef_gte: String
  compareRef_lt: String
  compareRef_lte: String
  compareRef_contains: String
  compareRef_startsWith: String
  compareRef_endsWith: String
  compareRef_regex: String
  compareRef_in: [String!]
  compareRef_nin: [String!]
  compareRef_not: String
  compareRef_isNull: Boolean
  summary: String
  summary_gt: String
  summary_gte: String
  summary_lt: String
  summary_lte: String
  summary_contains: String
  summary_startsWith: String
  summary_endsWith: String
  summary_regex: String
  summary_in: [String!]
  summary_nin: [String!]
  summary_not: String
  summary_isNull: Boolean
  text: String
  text_gt: String
  text_gte: String
  text_lt: String
  text_lte: String
  text_contains: String
  text_startsWith: String
  text_endsWith: String
  text_regex: String
  text_in: [String!]
  text_nin: [String!]
  text_not: String
  text_isNull: Boolean
  compareVerses_some: VerseWhere
  AND: [JSTPassageWhere!]
  OR: [JSTPassageWhere!]
  NOT: JSTPassageWhere
}

input JSTPassageSort {
  id: SortDirection
  book: SortDirection
  chapter: SortDirection
  comprises: SortDirection
  compareRef: SortDirection
  summary: SortDirection
  text: SortDirection
}

input JSTPassageCreateInput {
  book: String!
  chapter: String!
  comprises: String!
  compareRef: String
  summary: String
  text: String!
  embedding: [Float!]!
  compareVerses: JSTPassageCompareVersesFieldInput
}

input JSTPassageUpdateInput {
  book: String
  chapter: String
  comprises: String
  compareRef: String
  summary: String
  text: String
  embedding: [Float!]
  compareVerses: JSTPassageCompareVersesUpdateFieldInput
}

type CreateJSTPassagesMutationResponse {
  jSTPassages: [JSTPassage!]!
}

type UpdateJSTPassagesMutationResponse {
  jSTPassages: [JSTPassage!]!
}

input JSTPassageMatchInput {
  book: String
  chapter: String
  comprises: String
  compareRef: String
  summary: String
  text: String
}

input JSTPassageMergeInput {
  match: JSTPassageMatchInput!
  onCreate: JSTPassageCreateInput
  onMatch: JSTPassageUpdateInput
}

type MergeJSTPassagesMutationResponse {
  jSTPassages: [JSTPassage!]!
}

type JSTPassagesConnection {
  edges: [JSTPassageEdge!]!
  pageInfo: PageInfo!
  totalCount: Int!
}

type JSTPassageEdge {
  node: JSTPassage!
  cursor: String!
}

type JSTPassageSimilarResult {
  score: Float!
  node: JSTPassage!
}

input ChapterWhere {
  id: ID
  id_gt: ID
  id_gte: ID
  id_lt: ID
  id_lte: ID
  id_contains: ID
  id_startsWith: ID
  id_endsWith: ID
  id_regex: ID
  id_in: [ID!]
  id_nin: [ID!]
  id_not: ID
  id_isNull: Boolean
  number: Int
  number_gt: Int
  number_gte: Int
  number_lt: Int
  number_lte: Int
  number_in: [Int!]
  number_nin: [Int!]
  number_not: Int
  number_isNull: Boolean
  summary: String
  summary_gt: String
  summary_gte: String
  summary_lt: String
  summary_lte: String
  summary_contains: String
  summary_startsWith: String
  summary_endsWith: String
  summary_regex: String
  summary_in: [String!]
  summary_nin: [String!]
  summary_not: String
  summary_isNull: Boolean
  url: String
  url_gt: String
  url_gte: String
  url_lt: String
  url_lte: String
  url_contains: String
  url_startsWith: String
  url_endsWith: String
  url_regex: String
  url_in: [String!]
  url_nin: [String!]
  url_not: String
  url_isNull: Boolean
  book: BookWhere
  verses_some: VerseWhere
  verseGroups_some: VerseGroupWhere
  AND: [ChapterWhere!]
  OR: [ChapterWhere!]
  NOT: ChapterWhere
}

input ChapterSort {
  id: SortDirection
  number: SortDirection
  summary: SortDirection
  url: SortDirection
}

input ChapterCreateInput {
  number: Int!
  summary: String
  summaryEmbedding: [Float!]!
  url: String
  book: ChapterBookFieldInput
  verses: ChapterVersesFieldInput
  verseGroups: ChapterVerseGroupsFieldInput
}

input ChapterUpdateInput {
  number: Int
  summary: String
  summaryEmbedding: [Float!]
  url: String
  book: ChapterBookUpdateFieldInput
  verses: ChapterVersesUpdateFieldInput
  verseGroups: ChapterVerseGroupsUpdateFieldInput
}

type CreateChaptersMutationResponse {
  chapters: [Chapter!]!
}

type UpdateChaptersMutationResponse {
  chapters: [Chapter!]!
}

input ChapterMatchInput {
  number: Int
  summary: String
  url: String
}

input ChapterMergeInput {
  match: ChapterMatchInput!
  onCreate: ChapterCreateInput
  onMatch: ChapterUpdateInput
}

type MergeChaptersMutationResponse {
  chapters: [Chapter!]!
}

type ChaptersConnection {
  edges: [ChapterEdge!]!
  pageInfo: PageInfo!
  totalCount: Int!
}

type ChapterEdge {
  node: Chapter!
  cursor: String!
}

type ChapterSimilarResult {
  score: Float!
  node: Chapter!
}

input VerseWhere {
  id: ID
  id_gt: ID
  id_gte: ID
  id_lt: ID
  id_lte: ID
  id_contains: ID
  id_startsWith: ID
  id_endsWith: ID
  id_regex: ID
  id_in: [ID!]
  id_nin: [ID!]
  id_not: ID
  id_isNull: Boolean
  number: Int
  number_gt: Int
  number_gte: Int
  number_lt: Int
  number_lte: Int
  number_in: [Int!]
  number_nin: [Int!]
  number_not: Int
  number_isNull: Boolean
  reference: String
  reference_gt: String
  reference_gte: String
  reference_lt: String
  reference_lte: String
  reference_contains: String
  reference_startsWith: String
  reference_endsWith: String
  reference_regex: String
  reference_in: [String!]
  reference_nin: [String!]
  reference_not: String
  reference_isNull: Boolean
  text: String
  text_gt: String
  text_gte: String
  text_lt: String
  text_lte: String
  text_contains: String
  text_startsWith: String
  text_endsWith: String
  text_regex: String
  text_in: [String!]
  text_nin: [String!]
  text_not: String
  text_isNull: Boolean
  translationNotes: String
  translationNotes_gt: String
  translationNotes_gte: String
  translationNotes_lt: String
  translationNotes_lte: String
  translationNotes_contains: String
  translationNotes_startsWith: String
  translationNotes_endsWith: String
  translationNotes_regex: String
  translationNotes_in: [String!]
  translationNotes_nin: [String!]
  translationNotes_not: String
  translationNotes_isNull: Boolean
  alternateReadings: String
  alternateReadings_gt: String
  alternateReadings_gte: String
  alternateReadings_lt: String
  alternateReadings_lte: String
  alternateReadings_contains: String
  alternateReadings_startsWith: String
  alternateReadings_endsWith: String
  alternateReadings_regex: String
  alternateReadings_in: [String!]
  alternateReadings_nin: [String!]
  alternateReadings_not: String
  alternateReadings_isNull: Boolean
  explanatoryNotes: String
  explanatoryNotes_gt: String
  explanatoryNotes_gte: String
  explanatoryNotes_lt: String
  explanatoryNotes_lte: String
  explanatoryNotes_contains: String
  explanatoryNotes_startsWith: String
  explanatoryNotes_endsWith: String
  explanatoryNotes_regex: String
  explanatoryNotes_in: [String!]
  explanatoryNotes_nin: [String!]
  explanatoryNotes_not: String
  explanatoryNotes_isNull: Boolean
  chapter: ChapterWhere
  crossRefsOut_some: VerseWhere
  crossRefsIn_some: VerseWhere
  tgFootnotes_some: TopicalGuideEntryWhere
  bdFootnotes_some: BibleDictEntryWhere
  jstFootnotes_some: JSTPassageWhere
  AND: [VerseWhere!]
  OR: [VerseWhere!]
  NOT: VerseWhere
}

input VerseSort {
  id: SortDirection
  number: SortDirection
  reference: SortDirection
  text: SortDirection
  translationNotes: SortDirection
  alternateReadings: SortDirection
  explanatoryNotes: SortDirection
}

input VerseCreateInput {
  number: Int!
  reference: String!
  text: String!
  translationNotes: String
  alternateReadings: String
  explanatoryNotes: String
  chapter: VerseChapterFieldInput
  crossRefsOut: VerseCrossRefsOutFieldInput
  crossRefsIn: VerseCrossRefsInFieldInput
  tgFootnotes: VerseTgFootnotesFieldInput
  bdFootnotes: VerseBdFootnotesFieldInput
  jstFootnotes: VerseJstFootnotesFieldInput
}

input VerseUpdateInput {
  number: Int
  reference: String
  text: String
  translationNotes: String
  alternateReadings: String
  explanatoryNotes: String
  chapter: VerseChapterUpdateFieldInput
  crossRefsOut: VerseCrossRefsOutUpdateFieldInput
  crossRefsIn: VerseCrossRefsInUpdateFieldInput
  tgFootnotes: VerseTgFootnotesUpdateFieldInput
  bdFootnotes: VerseBdFootnotesUpdateFieldInput
  jstFootnotes: VerseJstFootnotesUpdateFieldInput
}

type CreateVersesMutationResponse {
  verses: [Verse!]!
}

type UpdateVersesMutationResponse {
  verses: [Verse!]!
}

input VerseMatchInput {
  number: Int
  reference: String
  text: String
  translationNotes: String
  alternateReadings: String
  explanatoryNotes: String
}

input VerseMergeInput {
  match: VerseMatchInput!
  onCreate: VerseCreateInput
  onMatch: VerseUpdateInput
}

type MergeVersesMutationResponse {
  verses: [Verse!]!
}

type VersesConnection {
  edges: [VerseEdge!]!
  pageInfo: PageInfo!
  totalCount: Int!
}

type VerseEdge {
  node: Verse!
  cursor: String!
}

input BibleDictEntryWhere {
  id: ID
  id_gt: ID
  id_gte: ID
  id_lt: ID
  id_lte: ID
  id_contains: ID
  id_startsWith: ID
  id_endsWith: ID
  id_regex: ID
  id_in: [ID!]
  id_nin: [ID!]
  id_not: ID
  id_isNull: Boolean
  name: String
  name_gt: String
  name_gte: String
  name_lt: String
  name_lte: String
  name_contains: String
  name_startsWith: String
  name_endsWith: String
  name_regex: String
  name_in: [String!]
  name_nin: [String!]
  name_not: String
  name_isNull: Boolean
  text: String
  text_gt: String
  text_gte: String
  text_lt: String
  text_lte: String
  text_contains: String
  text_startsWith: String
  text_endsWith: String
  text_regex: String
  text_in: [String!]
  text_nin: [String!]
  text_not: String
  text_isNull: Boolean
  seeAlso_some: BibleDictEntryWhere
  verseRefs_some: VerseWhere
  AND: [BibleDictEntryWhere!]
  OR: [BibleDictEntryWhere!]
  NOT: BibleDictEntryWhere
}

input BibleDictEntrySort {
  id: SortDirection
  name: SortDirection
  text: SortDirection
}

input BibleDictEntryCreateInput {
  name: String!
  text: String!
  embedding: [Float!]!
  seeAlso: BibleDictEntrySeeAlsoFieldInput
  verseRefs: BibleDictEntryVerseRefsFieldInput
}

input BibleDictEntryUpdateInput {
  name: String
  text: String
  embedding: [Float!]
  seeAlso: BibleDictEntrySeeAlsoUpdateFieldInput
  verseRefs: BibleDictEntryVerseRefsUpdateFieldInput
}

type CreateBibleDictEntriesMutationResponse {
  bibleDictEntries: [BibleDictEntry!]!
}

type UpdateBibleDictEntriesMutationResponse {
  bibleDictEntries: [BibleDictEntry!]!
}

input BibleDictEntryMatchInput {
  name: String
  text: String
}

input BibleDictEntryMergeInput {
  match: BibleDictEntryMatchInput!
  onCreate: BibleDictEntryCreateInput
  onMatch: BibleDictEntryUpdateInput
}

type MergeBibleDictEntriesMutationResponse {
  bibleDictEntries: [BibleDictEntry!]!
}

type BibleDictEntriesConnection {
  edges: [BibleDictEntryEdge!]!
  pageInfo: PageInfo!
  totalCount: Int!
}

type BibleDictEntryEdge {
  node: BibleDictEntry!
  cursor: String!
}

type BibleDictEntrySimilarResult {
  score: Float!
  node: BibleDictEntry!
}

input IndexEntryWhere {
  id: ID
  id_gt: ID
  id_gte: ID
  id_lt: ID
  id_lte: ID
  id_contains: ID
  id_startsWith: ID
  id_endsWith: ID
  id_regex: ID
  id_in: [ID!]
  id_nin: [ID!]
  id_not: ID
  id_isNull: Boolean
  name: String
  name_gt: String
  name_gte: String
  name_lt: String
  name_lte: String
  name_contains: String
  name_startsWith: String
  name_endsWith: String
  name_regex: String
  name_in: [String!]
  name_nin: [String!]
  name_not: String
  name_isNull: Boolean
  seeAlso_some: IndexEntryWhere
  tgRefs_some: TopicalGuideEntryWhere
  bdRefs_some: BibleDictEntryWhere
  verseRefs_some: VerseWhere
  AND: [IndexEntryWhere!]
  OR: [IndexEntryWhere!]
  NOT: IndexEntryWhere
}

input IndexEntrySort {
  id: SortDirection
  name: SortDirection
}

input IndexEntryCreateInput {
  name: String!
  embedding: [Float!]!
  seeAlso: IndexEntrySeeAlsoFieldInput
  tgRefs: IndexEntryTgRefsFieldInput
  bdRefs: IndexEntryBdRefsFieldInput
  verseRefs: IndexEntryVerseRefsFieldInput
}

input IndexEntryUpdateInput {
  name: String
  embedding: [Float!]
  seeAlso: IndexEntrySeeAlsoUpdateFieldInput
  tgRefs: IndexEntryTgRefsUpdateFieldInput
  bdRefs: IndexEntryBdRefsUpdateFieldInput
  verseRefs: IndexEntryVerseRefsUpdateFieldInput
}

type CreateIndexEntriesMutationResponse {
  indexEntries: [IndexEntry!]!
}

type UpdateIndexEntriesMutationResponse {
  indexEntries: [IndexEntry!]!
}

input IndexEntryMatchInput {
  name: String
}

input IndexEntryMergeInput {
  match: IndexEntryMatchInput!
  onCreate: IndexEntryCreateInput
  onMatch: IndexEntryUpdateInput
}

type MergeIndexEntriesMutationResponse {
  indexEntries: [IndexEntry!]!
}

type IndexEntriesConnection {
  edges: [IndexEntryEdge!]!
  pageInfo: PageInfo!
  totalCount: Int!
}

type IndexEntryEdge {
  node: IndexEntry!
  cursor: String!
}

type IndexEntrySimilarResult {
  score: Float!
  node: IndexEntry!
}

input TopicalGuideEntryWhere {
  id: ID
  id_gt: ID
  id_gte: ID
  id_lt: ID
  id_lte: ID
  id_contains: ID
  id_startsWith: ID
  id_endsWith: ID
  id_regex: ID
  id_in: [ID!]
  id_nin: [ID!]
  id_not: ID
  id_isNull: Boolean
  name: String
  name_gt: String
  name_gte: String
  name_lt: String
  name_lte: String
  name_contains: String
  name_startsWith: String
  name_endsWith: String
  name_regex: String
  name_in: [String!]
  name_nin: [String!]
  name_not: String
  name_isNull: Boolean
  seeAlso_some: TopicalGuideEntryWhere
  bdRefs_some: BibleDictEntryWhere
  verseRefs_some: VerseWhere
  AND: [TopicalGuideEntryWhere!]
  OR: [TopicalGuideEntryWhere!]
  NOT: TopicalGuideEntryWhere
}

input TopicalGuideEntrySort {
  id: SortDirection
  name: SortDirection
}

input TopicalGuideEntryCreateInput {
  name: String!
  embedding: [Float!]!
  seeAlso: TopicalGuideEntrySeeAlsoFieldInput
  bdRefs: TopicalGuideEntryBdRefsFieldInput
  verseRefs: TopicalGuideEntryVerseRefsFieldInput
}

input TopicalGuideEntryUpdateInput {
  name: String
  embedding: [Float!]
  seeAlso: TopicalGuideEntrySeeAlsoUpdateFieldInput
  bdRefs: TopicalGuideEntryBdRefsUpdateFieldInput
  verseRefs: TopicalGuideEntryVerseRefsUpdateFieldInput
}

type CreateTopicalGuideEntriesMutationResponse {
  topicalGuideEntries: [TopicalGuideEntry!]!
}

type UpdateTopicalGuideEntriesMutationResponse {
  topicalGuideEntries: [TopicalGuideEntry!]!
}

input TopicalGuideEntryMatchInput {
  name: String
}

input TopicalGuideEntryMergeInput {
  match: TopicalGuideEntryMatchInput!
  onCreate: TopicalGuideEntryCreateInput
  onMatch: TopicalGuideEntryUpdateInput
}

type MergeTopicalGuideEntriesMutationResponse {
  topicalGuideEntries: [TopicalGuideEntry!]!
}

type TopicalGuideEntriesConnection {
  edges: [TopicalGuideEntryEdge!]!
  pageInfo: PageInfo!
  totalCount: Int!
}

type TopicalGuideEntryEdge {
  node: TopicalGuideEntry!
  cursor: String!
}

type TopicalGuideEntrySimilarResult {
  score: Float!
  node: TopicalGuideEntry!
}

input VolumeWhere {
  id: ID
  id_gt: ID
  id_gte: ID
  id_lt: ID
  id_lte: ID
  id_contains: ID
  id_startsWith: ID
  id_endsWith: ID
  id_regex: ID
  id_in: [ID!]
  id_nin: [ID!]
  id_not: ID
  id_isNull: Boolean
  name: String
  name_gt: String
  name_gte: String
  name_lt: String
  name_lte: String
  name_contains: String
  name_startsWith: String
  name_endsWith: String
  name_regex: String
  name_in: [String!]
  name_nin: [String!]
  name_not: String
  name_isNull: Boolean
  abbreviation: String
  abbreviation_gt: String
  abbreviation_gte: String
  abbreviation_lt: String
  abbreviation_lte: String
  abbreviation_contains: String
  abbreviation_startsWith: String
  abbreviation_endsWith: String
  abbreviation_regex: String
  abbreviation_in: [String!]
  abbreviation_nin: [String!]
  abbreviation_not: String
  abbreviation_isNull: Boolean
  books_some: BookWhere
  AND: [VolumeWhere!]
  OR: [VolumeWhere!]
  NOT: VolumeWhere
}

input VolumeSort {
  id: SortDirection
  name: SortDirection
  abbreviation: SortDirection
}

input VolumeCreateInput {
  name: String!
  abbreviation: String!
  books: VolumeBooksFieldInput
}

input VolumeUpdateInput {
  name: String
  abbreviation: String
  books: VolumeBooksUpdateFieldInput
}

type CreateVolumesMutationResponse {
  volumes: [Volume!]!
}

type UpdateVolumesMutationResponse {
  volumes: [Volume!]!
}

input VolumeMatchInput {
  name: String
  abbreviation: String
}

input VolumeMergeInput {
  match: VolumeMatchInput!
  onCreate: VolumeCreateInput
  onMatch: VolumeUpdateInput
}

type MergeVolumesMutationResponse {
  volumes: [Volume!]!
}

type VolumesConnection {
  edges: [VolumeEdge!]!
  pageInfo: PageInfo!
  totalCount: Int!
}

type VolumeEdge {
  node: Volume!
  cursor: String!
}

input BookWhere {
  id: ID
  id_gt: ID
  id_gte: ID
  id_lt: ID
  id_lte: ID
  id_contains: ID
  id_startsWith: ID
  id_endsWith: ID
  id_regex: ID
  id_in: [ID!]
  id_nin: [ID!]
  id_not: ID
  id_isNull: Boolean
  name: String
  name_gt: String
  name_gte: String
  name_lt: String
  name_lte: String
  name_contains: String
  name_startsWith: String
  name_endsWith: String
  name_regex: String
  name_in: [String!]
  name_nin: [String!]
  name_not: String
  name_isNull: Boolean
  slug: String
  slug_gt: String
  slug_gte: String
  slug_lt: String
  slug_lte: String
  slug_contains: String
  slug_startsWith: String
  slug_endsWith: String
  slug_regex: String
  slug_in: [String!]
  slug_nin: [String!]
  slug_not: String
  slug_isNull: Boolean
  urlPath: String
  urlPath_gt: String
  urlPath_gte: String
  urlPath_lt: String
  urlPath_lte: String
  urlPath_contains: String
  urlPath_startsWith: String
  urlPath_endsWith: String
  urlPath_regex: String
  urlPath_in: [String!]
  urlPath_nin: [String!]
  urlPath_not: String
  urlPath_isNull: Boolean
  volume: VolumeWhere
  chapters_some: ChapterWhere
  AND: [BookWhere!]
  OR: [BookWhere!]
  NOT: BookWhere
}

input BookSort {
  id: SortDirection
  name: SortDirection
  slug: SortDirection
  urlPath: SortDirection
}

input BookCreateInput {
  name: String!
  slug: String!
  urlPath: String!
  volume: BookVolumeFieldInput
  chapters: BookChaptersFieldInput
}

input BookUpdateInput {
  name: String
  slug: String
  urlPath: String
  volume: BookVolumeUpdateFieldInput
  chapters: BookChaptersUpdateFieldInput
}

type CreateBooksMutationResponse {
  books: [Book!]!
}

type UpdateBooksMutationResponse {
  books: [Book!]!
}

input BookMatchInput {
  name: String
  slug: String
  urlPath: String
}

input BookMergeInput {
  match: BookMatchInput!
  onCreate: BookCreateInput
  onMatch: BookUpdateInput
}

type MergeBooksMutationResponse {
  books: [Book!]!
}

type BooksConnection {
  edges: [BookEdge!]!
  pageInfo: PageInfo!
  totalCount: Int!
}

type BookEdge {
  node: Book!
  cursor: String!
}

input VerseGroupChapterFieldInput {
  create: [VerseGroupChapterCreateFieldInput!]
  connect: [VerseGroupChapterConnectFieldInput!]
}

input VerseGroupChapterCreateFieldInput {
  node: ChapterCreateInput!
}

input VerseGroupChapterConnectFieldInput {
  where: ChapterWhere
}

input VerseGroupChapterUpdateFieldInput {
  create: [VerseGroupChapterCreateFieldInput!]
  connect: [VerseGroupChapterConnectFieldInput!]
  disconnect: [VerseGroupChapterDisconnectFieldInput!]
  update: VerseGroupChapterUpdateConnectionInput
  delete: [VerseGroupChapterDeleteFieldInput!]
}

input VerseGroupChapterDisconnectFieldInput {
  where: ChapterWhere
}

input VerseGroupChapterDeleteFieldInput {
  where: ChapterWhere
}

input VerseGroupChapterUpdateConnectionInput {
  where: ChapterWhere
  node: ChapterUpdateInput
}

type VerseGroupChapterConnection {
  edges: [VerseGroupChapterEdge!]!
  pageInfo: PageInfo!
  totalCount: Int!
}

type VerseGroupChapterEdge {
  node: Chapter!
  cursor: String!
}

input VerseGroupVersesFieldInput {
  create: [VerseGroupVersesCreateFieldInput!]
  connect: [VerseGroupVersesConnectFieldInput!]
}

input VerseGroupVersesCreateFieldInput {
  node: VerseCreateInput!
}

input VerseGroupVersesConnectFieldInput {
  where: VerseWhere
}

input VerseGroupVersesUpdateFieldInput {
  create: [VerseGroupVersesCreateFieldInput!]
  connect: [VerseGroupVersesConnectFieldInput!]
  disconnect: [VerseGroupVersesDisconnectFieldInput!]
  update: VerseGroupVersesUpdateConnectionInput
  delete: [VerseGroupVersesDeleteFieldInput!]
}

input VerseGroupVersesDisconnectFieldInput {
  where: VerseWhere
}

input VerseGroupVersesDeleteFieldInput {
  where: VerseWhere
}

input VerseGroupVersesUpdateConnectionInput {
  where: VerseWhere
  node: VerseUpdateInput
}

type VerseGroupVersesConnection {
  edges: [VerseGroupVersesEdge!]!
  pageInfo: PageInfo!
  totalCount: Int!
}

type VerseGroupVersesEdge {
  node: Verse!
  cursor: String!
}

input JSTPassageCompareVersesFieldInput {
  create: [JSTPassageCompareVersesCreateFieldInput!]
  connect: [JSTPassageCompareVersesConnectFieldInput!]
}

input JSTPassageCompareVersesCreateFieldInput {
  node: VerseCreateInput!
}

input JSTPassageCompareVersesConnectFieldInput {
  where: VerseWhere
}

input JSTPassageCompareVersesUpdateFieldInput {
  create: [JSTPassageCompareVersesCreateFieldInput!]
  connect: [JSTPassageCompareVersesConnectFieldInput!]
  disconnect: [JSTPassageCompareVersesDisconnectFieldInput!]
  update: JSTPassageCompareVersesUpdateConnectionInput
  delete: [JSTPassageCompareVersesDeleteFieldInput!]
}

input JSTPassageCompareVersesDisconnectFieldInput {
  where: VerseWhere
}

input JSTPassageCompareVersesDeleteFieldInput {
  where: VerseWhere
}

input JSTPassageCompareVersesUpdateConnectionInput {
  where: VerseWhere
  node: VerseUpdateInput
}

type JSTPassageCompareVersesConnection {
  edges: [JSTPassageCompareVersesEdge!]!
  pageInfo: PageInfo!
  totalCount: Int!
}

type JSTPassageCompareVersesEdge {
  node: Verse!
  cursor: String!
}

input ChapterBookFieldInput {
  create: [ChapterBookCreateFieldInput!]
  connect: [ChapterBookConnectFieldInput!]
}

input ChapterBookCreateFieldInput {
  node: BookCreateInput!
}

input ChapterBookConnectFieldInput {
  where: BookWhere
}

input ChapterBookUpdateFieldInput {
  create: [ChapterBookCreateFieldInput!]
  connect: [ChapterBookConnectFieldInput!]
  disconnect: [ChapterBookDisconnectFieldInput!]
  update: ChapterBookUpdateConnectionInput
  delete: [ChapterBookDeleteFieldInput!]
}

input ChapterBookDisconnectFieldInput {
  where: BookWhere
}

input ChapterBookDeleteFieldInput {
  where: BookWhere
}

input ChapterBookUpdateConnectionInput {
  where: BookWhere
  node: BookUpdateInput
}

type ChapterBookConnection {
  edges: [ChapterBookEdge!]!
  pageInfo: PageInfo!
  totalCount: Int!
}

type ChapterBookEdge {
  node: Book!
  cursor: String!
}

input ChapterVersesFieldInput {
  create: [ChapterVersesCreateFieldInput!]
  connect: [ChapterVersesConnectFieldInput!]
}

input ChapterVersesCreateFieldInput {
  node: VerseCreateInput!
}

input ChapterVersesConnectFieldInput {
  where: VerseWhere
}

input ChapterVersesUpdateFieldInput {
  create: [ChapterVersesCreateFieldInput!]
  connect: [ChapterVersesConnectFieldInput!]
  disconnect: [ChapterVersesDisconnectFieldInput!]
  update: ChapterVersesUpdateConnectionInput
  delete: [ChapterVersesDeleteFieldInput!]
}

input ChapterVersesDisconnectFieldInput {
  where: VerseWhere
}

input ChapterVersesDeleteFieldInput {
  where: VerseWhere
}

input ChapterVersesUpdateConnectionInput {
  where: VerseWhere
  node: VerseUpdateInput
}

type ChapterVersesConnection {
  edges: [ChapterVersesEdge!]!
  pageInfo: PageInfo!
  totalCount: Int!
}

type ChapterVersesEdge {
  node: Verse!
  cursor: String!
}

input ChapterVerseGroupsFieldInput {
  create: [ChapterVerseGroupsCreateFieldInput!]
  connect: [ChapterVerseGroupsConnectFieldInput!]
}

input ChapterVerseGroupsCreateFieldInput {
  node: VerseGroupCreateInput!
}

input ChapterVerseGroupsConnectFieldInput {
  where: VerseGroupWhere
}

input ChapterVerseGroupsUpdateFieldInput {
  create: [ChapterVerseGroupsCreateFieldInput!]
  connect: [ChapterVerseGroupsConnectFieldInput!]
  disconnect: [ChapterVerseGroupsDisconnectFieldInput!]
  update: ChapterVerseGroupsUpdateConnectionInput
  delete: [ChapterVerseGroupsDeleteFieldInput!]
}

input ChapterVerseGroupsDisconnectFieldInput {
  where: VerseGroupWhere
}

input ChapterVerseGroupsDeleteFieldInput {
  where: VerseGroupWhere
}

input ChapterVerseGroupsUpdateConnectionInput {
  where: VerseGroupWhere
  node: VerseGroupUpdateInput
}

type ChapterVerseGroupsConnection {
  edges: [ChapterVerseGroupsEdge!]!
  pageInfo: PageInfo!
  totalCount: Int!
}

type ChapterVerseGroupsEdge {
  node: VerseGroup!
  cursor: String!
}

input VerseChapterFieldInput {
  create: [VerseChapterCreateFieldInput!]
  connect: [VerseChapterConnectFieldInput!]
}

input VerseChapterCreateFieldInput {
  node: ChapterCreateInput!
}

input VerseChapterConnectFieldInput {
  where: ChapterWhere
}

input VerseChapterUpdateFieldInput {
  create: [VerseChapterCreateFieldInput!]
  connect: [VerseChapterConnectFieldInput!]
  disconnect: [VerseChapterDisconnectFieldInput!]
  update: VerseChapterUpdateConnectionInput
  delete: [VerseChapterDeleteFieldInput!]
}

input VerseChapterDisconnectFieldInput {
  where: ChapterWhere
}

input VerseChapterDeleteFieldInput {
  where: ChapterWhere
}

input VerseChapterUpdateConnectionInput {
  where: ChapterWhere
  node: ChapterUpdateInput
}

type VerseChapterConnection {
  edges: [VerseChapterEdge!]!
  pageInfo: PageInfo!
  totalCount: Int!
}

type VerseChapterEdge {
  node: Chapter!
  cursor: String!
}

input VerseCrossRefsOutFieldInput {
  create: [VerseCrossRefsOutCreateFieldInput!]
  connect: [VerseCrossRefsOutConnectFieldInput!]
}

input VerseCrossRefsOutCreateFieldInput {
  node: VerseCreateInput!
  edge: VerseCrossRefPropsCreateInput
}

input VerseCrossRefsOutConnectFieldInput {
  where: VerseWhere
  edge: VerseCrossRefPropsCreateInput
}

input VerseCrossRefsOutUpdateFieldInput {
  create: [VerseCrossRefsOutCreateFieldInput!]
  connect: [VerseCrossRefsOutConnectFieldInput!]
  disconnect: [VerseCrossRefsOutDisconnectFieldInput!]
  update: VerseCrossRefsOutUpdateConnectionInput
  delete: [VerseCrossRefsOutDeleteFieldInput!]
}

input VerseCrossRefsOutDisconnectFieldInput {
  where: VerseWhere
}

input VerseCrossRefsOutDeleteFieldInput {
  where: VerseWhere
}

input VerseCrossRefsOutUpdateConnectionInput {
  where: VerseWhere
  node: VerseUpdateInput
  edge: VerseCrossRefPropsUpdateInput
}

type VerseCrossRefsOutConnection {
  edges: [VerseCrossRefsOutEdge!]!
  pageInfo: PageInfo!
  totalCount: Int!
}

type VerseCrossRefsOutEdge {
  node: Verse!
  cursor: String!
  properties: VerseCrossRefProps
}

input VerseCrossRefPropsCreateInput {
  category: String
  footnoteMarker: String
  referenceText: String
}

input VerseCrossRefPropsUpdateInput {
  category: String
  footnoteMarker: String
  referenceText: String
}

type VerseCrossRefProps {
  category: String
  footnoteMarker: String
  referenceText: String
}

input VerseCrossRefsInFieldInput {
  create: [VerseCrossRefsInCreateFieldInput!]
  connect: [VerseCrossRefsInConnectFieldInput!]
}

input VerseCrossRefsInCreateFieldInput {
  node: VerseCreateInput!
  edge: VerseCrossRefPropsCreateInput
}

input VerseCrossRefsInConnectFieldInput {
  where: VerseWhere
  edge: VerseCrossRefPropsCreateInput
}

input VerseCrossRefsInUpdateFieldInput {
  create: [VerseCrossRefsInCreateFieldInput!]
  connect: [VerseCrossRefsInConnectFieldInput!]
  disconnect: [VerseCrossRefsInDisconnectFieldInput!]
  update: VerseCrossRefsInUpdateConnectionInput
  delete: [VerseCrossRefsInDeleteFieldInput!]
}

input VerseCrossRefsInDisconnectFieldInput {
  where: VerseWhere
}

input VerseCrossRefsInDeleteFieldInput {
  where: VerseWhere
}

input VerseCrossRefsInUpdateConnectionInput {
  where: VerseWhere
  node: VerseUpdateInput
  edge: VerseCrossRefPropsUpdateInput
}

type VerseCrossRefsInConnection {
  edges: [VerseCrossRefsInEdge!]!
  pageInfo: PageInfo!
  totalCount: Int!
}

type VerseCrossRefsInEdge {
  node: Verse!
  cursor: String!
  properties: VerseCrossRefProps
}

input VerseTgFootnotesFieldInput {
  create: [VerseTgFootnotesCreateFieldInput!]
  connect: [VerseTgFootnotesConnectFieldInput!]
}

input VerseTgFootnotesCreateFieldInput {
  node: TopicalGuideEntryCreateInput!
  edge: VerseTGRefPropsCreateInput
}

input VerseTgFootnotesConnectFieldInput {
  where: TopicalGuideEntryWhere
  edge: VerseTGRefPropsCreateInput
}

input VerseTgFootnotesUpdateFieldInput {
  create: [VerseTgFootnotesCreateFieldInput!]
  connect: [VerseTgFootnotesConnectFieldInput!]
  disconnect: [VerseTgFootnotesDisconnectFieldInput!]
  update: VerseTgFootnotesUpdateConnectionInput
  delete: [VerseTgFootnotesDeleteFieldInput!]
}

input VerseTgFootnotesDisconnectFieldInput {
  where: TopicalGuideEntryWhere
}

input VerseTgFootnotesDeleteFieldInput {
  where: TopicalGuideEntryWhere
}

input VerseTgFootnotesUpdateConnectionInput {
  where: TopicalGuideEntryWhere
  node: TopicalGuideEntryUpdateInput
  edge: VerseTGRefPropsUpdateInput
}

type VerseTgFootnotesConnection {
  edges: [VerseTgFootnotesEdge!]!
  pageInfo: PageInfo!
  totalCount: Int!
}

type VerseTgFootnotesEdge {
  node: TopicalGuideEntry!
  cursor: String!
  properties: VerseTGRefProps
}

input VerseTGRefPropsCreateInput {
  footnoteMarker: String
}

input VerseTGRefPropsUpdateInput {
  footnoteMarker: String
}

type VerseTGRefProps {
  footnoteMarker: String
}

input VerseBdFootnotesFieldInput {
  create: [VerseBdFootnotesCreateFieldInput!]
  connect: [VerseBdFootnotesConnectFieldInput!]
}

input VerseBdFootnotesCreateFieldInput {
  node: BibleDictEntryCreateInput!
  edge: VerseBDRefPropsCreateInput
}

input VerseBdFootnotesConnectFieldInput {
  where: BibleDictEntryWhere
  edge: VerseBDRefPropsCreateInput
}

input VerseBdFootnotesUpdateFieldInput {
  create: [VerseBdFootnotesCreateFieldInput!]
  connect: [VerseBdFootnotesConnectFieldInput!]
  disconnect: [VerseBdFootnotesDisconnectFieldInput!]
  update: VerseBdFootnotesUpdateConnectionInput
  delete: [VerseBdFootnotesDeleteFieldInput!]
}

input VerseBdFootnotesDisconnectFieldInput {
  where: BibleDictEntryWhere
}

input VerseBdFootnotesDeleteFieldInput {
  where: BibleDictEntryWhere
}

input VerseBdFootnotesUpdateConnectionInput {
  where: BibleDictEntryWhere
  node: BibleDictEntryUpdateInput
  edge: VerseBDRefPropsUpdateInput
}

type VerseBdFootnotesConnection {
  edges: [VerseBdFootnotesEdge!]!
  pageInfo: PageInfo!
  totalCount: Int!
}

type VerseBdFootnotesEdge {
  node: BibleDictEntry!
  cursor: String!
  properties: VerseBDRefProps
}

input VerseBDRefPropsCreateInput {
  footnoteMarker: String
}

input VerseBDRefPropsUpdateInput {
  footnoteMarker: String
}

type VerseBDRefProps {
  footnoteMarker: String
}

input VerseJstFootnotesFieldInput {
  create: [VerseJstFootnotesCreateFieldInput!]
  connect: [VerseJstFootnotesConnectFieldInput!]
}

input VerseJstFootnotesCreateFieldInput {
  node: JSTPassageCreateInput!
  edge: VerseJSTRefPropsCreateInput
}

input VerseJstFootnotesConnectFieldInput {
  where: JSTPassageWhere
  edge: VerseJSTRefPropsCreateInput
}

input VerseJstFootnotesUpdateFieldInput {
  create: [VerseJstFootnotesCreateFieldInput!]
  connect: [VerseJstFootnotesConnectFieldInput!]
  disconnect: [VerseJstFootnotesDisconnectFieldInput!]
  update: VerseJstFootnotesUpdateConnectionInput
  delete: [VerseJstFootnotesDeleteFieldInput!]
}

input VerseJstFootnotesDisconnectFieldInput {
  where: JSTPassageWhere
}

input VerseJstFootnotesDeleteFieldInput {
  where: JSTPassageWhere
}

input VerseJstFootnotesUpdateConnectionInput {
  where: JSTPassageWhere
  node: JSTPassageUpdateInput
  edge: VerseJSTRefPropsUpdateInput
}

type VerseJstFootnotesConnection {
  edges: [VerseJstFootnotesEdge!]!
  pageInfo: PageInfo!
  totalCount: Int!
}

type VerseJstFootnotesEdge {
  node: JSTPassage!
  cursor: String!
  properties: VerseJSTRefProps
}

input VerseJSTRefPropsCreateInput {
  footnoteMarker: String
}

input VerseJSTRefPropsUpdateInput {
  footnoteMarker: String
}

type VerseJSTRefProps {
  footnoteMarker: String
}

input BibleDictEntrySeeAlsoFieldInput {
  create: [BibleDictEntrySeeAlsoCreateFieldInput!]
  connect: [BibleDictEntrySeeAlsoConnectFieldInput!]
}

input BibleDictEntrySeeAlsoCreateFieldInput {
  node: BibleDictEntryCreateInput!
}

input BibleDictEntrySeeAlsoConnectFieldInput {
  where: BibleDictEntryWhere
}

input BibleDictEntrySeeAlsoUpdateFieldInput {
  create: [BibleDictEntrySeeAlsoCreateFieldInput!]
  connect: [BibleDictEntrySeeAlsoConnectFieldInput!]
  disconnect: [BibleDictEntrySeeAlsoDisconnectFieldInput!]
  update: BibleDictEntrySeeAlsoUpdateConnectionInput
  delete: [BibleDictEntrySeeAlsoDeleteFieldInput!]
}

input BibleDictEntrySeeAlsoDisconnectFieldInput {
  where: BibleDictEntryWhere
}

input BibleDictEntrySeeAlsoDeleteFieldInput {
  where: BibleDictEntryWhere
}

input BibleDictEntrySeeAlsoUpdateConnectionInput {
  where: BibleDictEntryWhere
  node: BibleDictEntryUpdateInput
}

type BibleDictEntrySeeAlsoConnection {
  edges: [BibleDictEntrySeeAlsoEdge!]!
  pageInfo: PageInfo!
  totalCount: Int!
}

type BibleDictEntrySeeAlsoEdge {
  node: BibleDictEntry!
  cursor: String!
}

input BibleDictEntryVerseRefsFieldInput {
  create: [BibleDictEntryVerseRefsCreateFieldInput!]
  connect: [BibleDictEntryVerseRefsConnectFieldInput!]
}

input BibleDictEntryVerseRefsCreateFieldInput {
  node: VerseCreateInput!
  edge: BDVerseRefPropsCreateInput
}

input BibleDictEntryVerseRefsConnectFieldInput {
  where: VerseWhere
  edge: BDVerseRefPropsCreateInput
}

input BibleDictEntryVerseRefsUpdateFieldInput {
  create: [BibleDictEntryVerseRefsCreateFieldInput!]
  connect: [BibleDictEntryVerseRefsConnectFieldInput!]
  disconnect: [BibleDictEntryVerseRefsDisconnectFieldInput!]
  update: BibleDictEntryVerseRefsUpdateConnectionInput
  delete: [BibleDictEntryVerseRefsDeleteFieldInput!]
}

input BibleDictEntryVerseRefsDisconnectFieldInput {
  where: VerseWhere
}

input BibleDictEntryVerseRefsDeleteFieldInput {
  where: VerseWhere
}

input BibleDictEntryVerseRefsUpdateConnectionInput {
  where: VerseWhere
  node: VerseUpdateInput
  edge: BDVerseRefPropsUpdateInput
}

type BibleDictEntryVerseRefsConnection {
  edges: [BibleDictEntryVerseRefsEdge!]!
  pageInfo: PageInfo!
  totalCount: Int!
}

type BibleDictEntryVerseRefsEdge {
  node: Verse!
  cursor: String!
  properties: BDVerseRefProps
}

input BDVerseRefPropsCreateInput {
  targetEndVerseId: String
}

input BDVerseRefPropsUpdateInput {
  targetEndVerseId: String
}

type BDVerseRefProps {
  targetEndVerseId: String
}

input IndexEntrySeeAlsoFieldInput {
  create: [IndexEntrySeeAlsoCreateFieldInput!]
  connect: [IndexEntrySeeAlsoConnectFieldInput!]
}

input IndexEntrySeeAlsoCreateFieldInput {
  node: IndexEntryCreateInput!
}

input IndexEntrySeeAlsoConnectFieldInput {
  where: IndexEntryWhere
}

input IndexEntrySeeAlsoUpdateFieldInput {
  create: [IndexEntrySeeAlsoCreateFieldInput!]
  connect: [IndexEntrySeeAlsoConnectFieldInput!]
  disconnect: [IndexEntrySeeAlsoDisconnectFieldInput!]
  update: IndexEntrySeeAlsoUpdateConnectionInput
  delete: [IndexEntrySeeAlsoDeleteFieldInput!]
}

input IndexEntrySeeAlsoDisconnectFieldInput {
  where: IndexEntryWhere
}

input IndexEntrySeeAlsoDeleteFieldInput {
  where: IndexEntryWhere
}

input IndexEntrySeeAlsoUpdateConnectionInput {
  where: IndexEntryWhere
  node: IndexEntryUpdateInput
}

type IndexEntrySeeAlsoConnection {
  edges: [IndexEntrySeeAlsoEdge!]!
  pageInfo: PageInfo!
  totalCount: Int!
}

type IndexEntrySeeAlsoEdge {
  node: IndexEntry!
  cursor: String!
}

input IndexEntryTgRefsFieldInput {
  create: [IndexEntryTgRefsCreateFieldInput!]
  connect: [IndexEntryTgRefsConnectFieldInput!]
}

input IndexEntryTgRefsCreateFieldInput {
  node: TopicalGuideEntryCreateInput!
}

input IndexEntryTgRefsConnectFieldInput {
  where: TopicalGuideEntryWhere
}

input IndexEntryTgRefsUpdateFieldInput {
  create: [IndexEntryTgRefsCreateFieldInput!]
  connect: [IndexEntryTgRefsConnectFieldInput!]
  disconnect: [IndexEntryTgRefsDisconnectFieldInput!]
  update: IndexEntryTgRefsUpdateConnectionInput
  delete: [IndexEntryTgRefsDeleteFieldInput!]
}

input IndexEntryTgRefsDisconnectFieldInput {
  where: TopicalGuideEntryWhere
}

input IndexEntryTgRefsDeleteFieldInput {
  where: TopicalGuideEntryWhere
}

input IndexEntryTgRefsUpdateConnectionInput {
  where: TopicalGuideEntryWhere
  node: TopicalGuideEntryUpdateInput
}

type IndexEntryTgRefsConnection {
  edges: [IndexEntryTgRefsEdge!]!
  pageInfo: PageInfo!
  totalCount: Int!
}

type IndexEntryTgRefsEdge {
  node: TopicalGuideEntry!
  cursor: String!
}

input IndexEntryBdRefsFieldInput {
  create: [IndexEntryBdRefsCreateFieldInput!]
  connect: [IndexEntryBdRefsConnectFieldInput!]
}

input IndexEntryBdRefsCreateFieldInput {
  node: BibleDictEntryCreateInput!
}

input IndexEntryBdRefsConnectFieldInput {
  where: BibleDictEntryWhere
}

input IndexEntryBdRefsUpdateFieldInput {
  create: [IndexEntryBdRefsCreateFieldInput!]
  connect: [IndexEntryBdRefsConnectFieldInput!]
  disconnect: [IndexEntryBdRefsDisconnectFieldInput!]
  update: IndexEntryBdRefsUpdateConnectionInput
  delete: [IndexEntryBdRefsDeleteFieldInput!]
}

input IndexEntryBdRefsDisconnectFieldInput {
  where: BibleDictEntryWhere
}

input IndexEntryBdRefsDeleteFieldInput {
  where: BibleDictEntryWhere
}

input IndexEntryBdRefsUpdateConnectionInput {
  where: BibleDictEntryWhere
  node: BibleDictEntryUpdateInput
}

type IndexEntryBdRefsConnection {
  edges: [IndexEntryBdRefsEdge!]!
  pageInfo: PageInfo!
  totalCount: Int!
}

type IndexEntryBdRefsEdge {
  node: BibleDictEntry!
  cursor: String!
}

input IndexEntryVerseRefsFieldInput {
  create: [IndexEntryVerseRefsCreateFieldInput!]
  connect: [IndexEntryVerseRefsConnectFieldInput!]
}

input IndexEntryVerseRefsCreateFieldInput {
  node: VerseCreateInput!
  edge: IDXVerseRefPropsCreateInput
}

input IndexEntryVerseRefsConnectFieldInput {
  where: VerseWhere
  edge: IDXVerseRefPropsCreateInput
}

input IndexEntryVerseRefsUpdateFieldInput {
  create: [IndexEntryVerseRefsCreateFieldInput!]
  connect: [IndexEntryVerseRefsConnectFieldInput!]
  disconnect: [IndexEntryVerseRefsDisconnectFieldInput!]
  update: IndexEntryVerseRefsUpdateConnectionInput
  delete: [IndexEntryVerseRefsDeleteFieldInput!]
}

input IndexEntryVerseRefsDisconnectFieldInput {
  where: VerseWhere
}

input IndexEntryVerseRefsDeleteFieldInput {
  where: VerseWhere
}

input IndexEntryVerseRefsUpdateConnectionInput {
  where: VerseWhere
  node: VerseUpdateInput
  edge: IDXVerseRefPropsUpdateInput
}

type IndexEntryVerseRefsConnection {
  edges: [IndexEntryVerseRefsEdge!]!
  pageInfo: PageInfo!
  totalCount: Int!
}

type IndexEntryVerseRefsEdge {
  node: Verse!
  cursor: String!
  properties: IDXVerseRefProps
}

input IDXVerseRefPropsCreateInput {
  phrase: String
}

input IDXVerseRefPropsUpdateInput {
  phrase: String
}

type IDXVerseRefProps {
  phrase: String
}

input TopicalGuideEntrySeeAlsoFieldInput {
  create: [TopicalGuideEntrySeeAlsoCreateFieldInput!]
  connect: [TopicalGuideEntrySeeAlsoConnectFieldInput!]
}

input TopicalGuideEntrySeeAlsoCreateFieldInput {
  node: TopicalGuideEntryCreateInput!
}

input TopicalGuideEntrySeeAlsoConnectFieldInput {
  where: TopicalGuideEntryWhere
}

input TopicalGuideEntrySeeAlsoUpdateFieldInput {
  create: [TopicalGuideEntrySeeAlsoCreateFieldInput!]
  connect: [TopicalGuideEntrySeeAlsoConnectFieldInput!]
  disconnect: [TopicalGuideEntrySeeAlsoDisconnectFieldInput!]
  update: TopicalGuideEntrySeeAlsoUpdateConnectionInput
  delete: [TopicalGuideEntrySeeAlsoDeleteFieldInput!]
}

input TopicalGuideEntrySeeAlsoDisconnectFieldInput {
  where: TopicalGuideEntryWhere
}

input TopicalGuideEntrySeeAlsoDeleteFieldInput {
  where: TopicalGuideEntryWhere
}

input TopicalGuideEntrySeeAlsoUpdateConnectionInput {
  where: TopicalGuideEntryWhere
  node: TopicalGuideEntryUpdateInput
}

type TopicalGuideEntrySeeAlsoConnection {
  edges: [TopicalGuideEntrySeeAlsoEdge!]!
  pageInfo: PageInfo!
  totalCount: Int!
}

type TopicalGuideEntrySeeAlsoEdge {
  node: TopicalGuideEntry!
  cursor: String!
}

input TopicalGuideEntryBdRefsFieldInput {
  create: [TopicalGuideEntryBdRefsCreateFieldInput!]
  connect: [TopicalGuideEntryBdRefsConnectFieldInput!]
}

input TopicalGuideEntryBdRefsCreateFieldInput {
  node: BibleDictEntryCreateInput!
}

input TopicalGuideEntryBdRefsConnectFieldInput {
  where: BibleDictEntryWhere
}

input TopicalGuideEntryBdRefsUpdateFieldInput {
  create: [TopicalGuideEntryBdRefsCreateFieldInput!]
  connect: [TopicalGuideEntryBdRefsConnectFieldInput!]
  disconnect: [TopicalGuideEntryBdRefsDisconnectFieldInput!]
  update: TopicalGuideEntryBdRefsUpdateConnectionInput
  delete: [TopicalGuideEntryBdRefsDeleteFieldInput!]
}

input TopicalGuideEntryBdRefsDisconnectFieldInput {
  where: BibleDictEntryWhere
}

input TopicalGuideEntryBdRefsDeleteFieldInput {
  where: BibleDictEntryWhere
}

input TopicalGuideEntryBdRefsUpdateConnectionInput {
  where: BibleDictEntryWhere
  node: BibleDictEntryUpdateInput
}

type TopicalGuideEntryBdRefsConnection {
  edges: [TopicalGuideEntryBdRefsEdge!]!
  pageInfo: PageInfo!
  totalCount: Int!
}

type TopicalGuideEntryBdRefsEdge {
  node: BibleDictEntry!
  cursor: String!
}

input TopicalGuideEntryVerseRefsFieldInput {
  create: [TopicalGuideEntryVerseRefsCreateFieldInput!]
  connect: [TopicalGuideEntryVerseRefsConnectFieldInput!]
}

input TopicalGuideEntryVerseRefsCreateFieldInput {
  node: VerseCreateInput!
  edge: TGVerseRefPropsCreateInput
}

input TopicalGuideEntryVerseRefsConnectFieldInput {
  where: VerseWhere
  edge: TGVerseRefPropsCreateInput
}

input TopicalGuideEntryVerseRefsUpdateFieldInput {
  create: [TopicalGuideEntryVerseRefsCreateFieldInput!]
  connect: [TopicalGuideEntryVerseRefsConnectFieldInput!]
  disconnect: [TopicalGuideEntryVerseRefsDisconnectFieldInput!]
  update: TopicalGuideEntryVerseRefsUpdateConnectionInput
  delete: [TopicalGuideEntryVerseRefsDeleteFieldInput!]
}

input TopicalGuideEntryVerseRefsDisconnectFieldInput {
  where: VerseWhere
}

input TopicalGuideEntryVerseRefsDeleteFieldInput {
  where: VerseWhere
}

input TopicalGuideEntryVerseRefsUpdateConnectionInput {
  where: VerseWhere
  node: VerseUpdateInput
  edge: TGVerseRefPropsUpdateInput
}

type TopicalGuideEntryVerseRefsConnection {
  edges: [TopicalGuideEntryVerseRefsEdge!]!
  pageInfo: PageInfo!
  totalCount: Int!
}

type TopicalGuideEntryVerseRefsEdge {
  node: Verse!
  cursor: String!
  properties: TGVerseRefProps
}

input TGVerseRefPropsCreateInput {
  phrase: String
}

input TGVerseRefPropsUpdateInput {
  phrase: String
}

type TGVerseRefProps {
  phrase: String
}

input VolumeBooksFieldInput {
  create: [VolumeBooksCreateFieldInput!]
  connect: [VolumeBooksConnectFieldInput!]
}

input VolumeBooksCreateFieldInput {
  node: BookCreateInput!
}

input VolumeBooksConnectFieldInput {
  where: BookWhere
}

input VolumeBooksUpdateFieldInput {
  create: [VolumeBooksCreateFieldInput!]
  connect: [VolumeBooksConnectFieldInput!]
  disconnect: [VolumeBooksDisconnectFieldInput!]
  update: VolumeBooksUpdateConnectionInput
  delete: [VolumeBooksDeleteFieldInput!]
}

input VolumeBooksDisconnectFieldInput {
  where: BookWhere
}

input VolumeBooksDeleteFieldInput {
  where: BookWhere
}

input VolumeBooksUpdateConnectionInput {
  where: BookWhere
  node: BookUpdateInput
}

type VolumeBooksConnection {
  edges: [VolumeBooksEdge!]!
  pageInfo: PageInfo!
  totalCount: Int!
}

type VolumeBooksEdge {
  node: Book!
  cursor: String!
}

input BookVolumeFieldInput {
  create: [BookVolumeCreateFieldInput!]
  connect: [BookVolumeConnectFieldInput!]
}

input BookVolumeCreateFieldInput {
  node: VolumeCreateInput!
}

input BookVolumeConnectFieldInput {
  where: VolumeWhere
}

input BookVolumeUpdateFieldInput {
  create: [BookVolumeCreateFieldInput!]
  connect: [BookVolumeConnectFieldInput!]
  disconnect: [BookVolumeDisconnectFieldInput!]
  update: BookVolumeUpdateConnectionInput
  delete: [BookVolumeDeleteFieldInput!]
}

input BookVolumeDisconnectFieldInput {
  where: VolumeWhere
}

input BookVolumeDeleteFieldInput {
  where: VolumeWhere
}

input BookVolumeUpdateConnectionInput {
  where: VolumeWhere
  node: VolumeUpdateInput
}

type BookVolumeConnection {
  edges: [BookVolumeEdge!]!
  pageInfo: PageInfo!
  totalCount: Int!
}

type BookVolumeEdge {
  node: Volume!
  cursor: String!
}

input BookChaptersFieldInput {
  create: [BookChaptersCreateFieldInput!]
  connect: [BookChaptersConnectFieldInput!]
}

input BookChaptersCreateFieldInput {
  node: ChapterCreateInput!
}

input BookChaptersConnectFieldInput {
  where: ChapterWhere
}

input BookChaptersUpdateFieldInput {
  create: [BookChaptersCreateFieldInput!]
  connect: [BookChaptersConnectFieldInput!]
  disconnect: [BookChaptersDisconnectFieldInput!]
  update: BookChaptersUpdateConnectionInput
  delete: [BookChaptersDeleteFieldInput!]
}

input BookChaptersDisconnectFieldInput {
  where: ChapterWhere
}

input BookChaptersDeleteFieldInput {
  where: ChapterWhere
}

input BookChaptersUpdateConnectionInput {
  where: ChapterWhere
  node: ChapterUpdateInput
}

type BookChaptersConnection {
  edges: [BookChaptersEdge!]!
  pageInfo: PageInfo!
  totalCount: Int!
}

type BookChaptersEdge {
  node: Chapter!
  cursor: String!
}

input ConnectVerseGroupChapterInput {
  from: VerseGroupWhere!
  to: ChapterWhere!
}

input ConnectVerseGroupVersesInput {
  from: VerseGroupWhere!
  to: VerseWhere!
}

input ConnectJSTPassageCompareVersesInput {
  from: JSTPassageWhere!
  to: VerseWhere!
}

input ConnectChapterBookInput {
  from: ChapterWhere!
  to: BookWhere!
}

input ConnectChapterVersesInput {
  from: ChapterWhere!
  to: VerseWhere!
}

input ConnectChapterVerseGroupsInput {
  from: ChapterWhere!
  to: VerseGroupWhere!
}

input ConnectVerseChapterInput {
  from: VerseWhere!
  to: ChapterWhere!
}

input ConnectVerseCrossRefsOutInput {
  from: VerseWhere!
  to: VerseWhere!
  edge: VerseCrossRefPropsCreateInput
}

input ConnectVerseCrossRefsInInput {
  from: VerseWhere!
  to: VerseWhere!
  edge: VerseCrossRefPropsCreateInput
}

input ConnectVerseTgFootnotesInput {
  from: VerseWhere!
  to: TopicalGuideEntryWhere!
  edge: VerseTGRefPropsCreateInput
}

input ConnectVerseBdFootnotesInput {
  from: VerseWhere!
  to: BibleDictEntryWhere!
  edge: VerseBDRefPropsCreateInput
}

input ConnectVerseJstFootnotesInput {
  from: VerseWhere!
  to: JSTPassageWhere!
  edge: VerseJSTRefPropsCreateInput
}

input ConnectBibleDictEntrySeeAlsoInput {
  from: BibleDictEntryWhere!
  to: BibleDictEntryWhere!
}

input ConnectBibleDictEntryVerseRefsInput {
  from: BibleDictEntryWhere!
  to: VerseWhere!
  edge: BDVerseRefPropsCreateInput
}

input ConnectIndexEntrySeeAlsoInput {
  from: IndexEntryWhere!
  to: IndexEntryWhere!
}

input ConnectIndexEntryTgRefsInput {
  from: IndexEntryWhere!
  to: TopicalGuideEntryWhere!
}

input ConnectIndexEntryBdRefsInput {
  from: IndexEntryWhere!
  to: BibleDictEntryWhere!
}

input ConnectIndexEntryVerseRefsInput {
  from: IndexEntryWhere!
  to: VerseWhere!
  edge: IDXVerseRefPropsCreateInput
}

input ConnectTopicalGuideEntrySeeAlsoInput {
  from: TopicalGuideEntryWhere!
  to: TopicalGuideEntryWhere!
}

input ConnectTopicalGuideEntryBdRefsInput {
  from: TopicalGuideEntryWhere!
  to: BibleDictEntryWhere!
}

input ConnectTopicalGuideEntryVerseRefsInput {
  from: TopicalGuideEntryWhere!
  to: VerseWhere!
  edge: TGVerseRefPropsCreateInput
}

input ConnectVolumeBooksInput {
  from: VolumeWhere!
  to: BookWhere!
}

input ConnectBookVolumeInput {
  from: BookWhere!
  to: VolumeWhere!
}

input ConnectBookChaptersInput {
  from: BookWhere!
  to: ChapterWhere!
}

type DeleteInfo {
  nodesDeleted: Int!
  relationshipsDeleted: Int!
}

type PageInfo {
  hasNextPage: Boolean!
  hasPreviousPage: Boolean!
  startCursor: String
  endCursor: String
}

type ConnectInfo {
  relationshipsCreated: Int!
}

enum SortDirection {
  ASC
  DESC
}

type Query {
  verseGroups(where: VerseGroupWhere, sort: [VerseGroupSort!]): [VerseGroup!]!
  verseGroupsConnection(first: Int, after: String, where: VerseGroupWhere, sort: [VerseGroupSort!]): VerseGroupsConnection!
  verseGroupsSimilar(vector: [Float!]!, first: Int = 10): [VerseGroupSimilarResult!]!
  jSTPassages(where: JSTPassageWhere, sort: [JSTPassageSort!]): [JSTPassage!]!
  jSTPassagesConnection(first: Int, after: String, where: JSTPassageWhere, sort: [JSTPassageSort!]): JSTPassagesConnection!
  jSTPassagesSimilar(vector: [Float!]!, first: Int = 10): [JSTPassageSimilarResult!]!
  chapters(where: ChapterWhere, sort: [ChapterSort!]): [Chapter!]!
  chaptersConnection(first: Int, after: String, where: ChapterWhere, sort: [ChapterSort!]): ChaptersConnection!
  chaptersSimilar(vector: [Float!]!, first: Int = 10): [ChapterSimilarResult!]!
  verses(where: VerseWhere, sort: [VerseSort!]): [Verse!]!
  versesConnection(first: Int, after: String, where: VerseWhere, sort: [VerseSort!]): VersesConnection!
  bibleDictEntries(where: BibleDictEntryWhere, sort: [BibleDictEntrySort!]): [BibleDictEntry!]!
  bibleDictEntriesConnection(first: Int, after: String, where: BibleDictEntryWhere, sort: [BibleDictEntrySort!]): BibleDictEntriesConnection!
  bibleDictEntriesSimilar(vector: [Float!]!, first: Int = 10): [BibleDictEntrySimilarResult!]!
  indexEntries(where: IndexEntryWhere, sort: [IndexEntrySort!]): [IndexEntry!]!
  indexEntriesConnection(first: Int, after: String, where: IndexEntryWhere, sort: [IndexEntrySort!]): IndexEntriesConnection!
  indexEntriesSimilar(vector: [Float!]!, first: Int = 10): [IndexEntrySimilarResult!]!
  topicalGuideEntries(where: TopicalGuideEntryWhere, sort: [TopicalGuideEntrySort!]): [TopicalGuideEntry!]!
  topicalGuideEntriesConnection(first: Int, after: String, where: TopicalGuideEntryWhere, sort: [TopicalGuideEntrySort!]): TopicalGuideEntriesConnection!
  topicalGuideEntriesSimilar(vector: [Float!]!, first: Int = 10): [TopicalGuideEntrySimilarResult!]!
  volumes(where: VolumeWhere, sort: [VolumeSort!]): [Volume!]!
  volumesConnection(first: Int, after: String, where: VolumeWhere, sort: [VolumeSort!]): VolumesConnection!
  books(where: BookWhere, sort: [BookSort!]): [Book!]!
  booksConnection(first: Int, after: String, where: BookWhere, sort: [BookSort!]): BooksConnection!
}

type Mutation {
  createVerseGroups(input: [VerseGroupCreateInput!]!): CreateVerseGroupsMutationResponse!
  updateVerseGroups(where: VerseGroupWhere, update: VerseGroupUpdateInput): UpdateVerseGroupsMutationResponse!
  deleteVerseGroups(where: VerseGroupWhere): DeleteInfo!
  mergeVerseGroups(input: [VerseGroupMergeInput!]!): MergeVerseGroupsMutationResponse!
  createJSTPassages(input: [JSTPassageCreateInput!]!): CreateJSTPassagesMutationResponse!
  updateJSTPassages(where: JSTPassageWhere, update: JSTPassageUpdateInput): UpdateJSTPassagesMutationResponse!
  deleteJSTPassages(where: JSTPassageWhere): DeleteInfo!
  mergeJSTPassages(input: [JSTPassageMergeInput!]!): MergeJSTPassagesMutationResponse!
  createChapters(input: [ChapterCreateInput!]!): CreateChaptersMutationResponse!
  updateChapters(where: ChapterWhere, update: ChapterUpdateInput): UpdateChaptersMutationResponse!
  deleteChapters(where: ChapterWhere): DeleteInfo!
  mergeChapters(input: [ChapterMergeInput!]!): MergeChaptersMutationResponse!
  createVerses(input: [VerseCreateInput!]!): CreateVersesMutationResponse!
  updateVerses(where: VerseWhere, update: VerseUpdateInput): UpdateVersesMutationResponse!
  deleteVerses(where: VerseWhere): DeleteInfo!
  mergeVerses(input: [VerseMergeInput!]!): MergeVersesMutationResponse!
  createBibleDictEntries(input: [BibleDictEntryCreateInput!]!): CreateBibleDictEntriesMutationResponse!
  updateBibleDictEntries(where: BibleDictEntryWhere, update: BibleDictEntryUpdateInput): UpdateBibleDictEntriesMutationResponse!
  deleteBibleDictEntries(where: BibleDictEntryWhere): DeleteInfo!
  mergeBibleDictEntries(input: [BibleDictEntryMergeInput!]!): MergeBibleDictEntriesMutationResponse!
  createIndexEntries(input: [IndexEntryCreateInput!]!): CreateIndexEntriesMutationResponse!
  updateIndexEntries(where: IndexEntryWhere, update: IndexEntryUpdateInput): UpdateIndexEntriesMutationResponse!
  deleteIndexEntries(where: IndexEntryWhere): DeleteInfo!
  mergeIndexEntries(input: [IndexEntryMergeInput!]!): MergeIndexEntriesMutationResponse!
  createTopicalGuideEntries(input: [TopicalGuideEntryCreateInput!]!): CreateTopicalGuideEntriesMutationResponse!
  updateTopicalGuideEntries(where: TopicalGuideEntryWhere, update: TopicalGuideEntryUpdateInput): UpdateTopicalGuideEntriesMutationResponse!
  deleteTopicalGuideEntries(where: TopicalGuideEntryWhere): DeleteInfo!
  mergeTopicalGuideEntries(input: [TopicalGuideEntryMergeInput!]!): MergeTopicalGuideEntriesMutationResponse!
  createVolumes(input: [VolumeCreateInput!]!): CreateVolumesMutationResponse!
  updateVolumes(where: VolumeWhere, update: VolumeUpdateInput): UpdateVolumesMutationResponse!
  deleteVolumes(where: VolumeWhere): DeleteInfo!
  mergeVolumes(input: [VolumeMergeInput!]!): MergeVolumesMutationResponse!
  createBooks(input: [BookCreateInput!]!): CreateBooksMutationResponse!
  updateBooks(where: BookWhere, update: BookUpdateInput): UpdateBooksMutationResponse!
  deleteBooks(where: BookWhere): DeleteInfo!
  mergeBooks(input: [BookMergeInput!]!): MergeBooksMutationResponse!
  connectVerseGroupChapter(input: [ConnectVerseGroupChapterInput!]!): ConnectInfo!
  connectVerseGroupVerses(input: [ConnectVerseGroupVersesInput!]!): ConnectInfo!
  connectJSTPassageCompareVerses(input: [ConnectJSTPassageCompareVersesInput!]!): ConnectInfo!
  connectChapterBook(input: [ConnectChapterBookInput!]!): ConnectInfo!
  connectChapterVerses(input: [ConnectChapterVersesInput!]!): ConnectInfo!
  connectChapterVerseGroups(input: [ConnectChapterVerseGroupsInput!]!): ConnectInfo!
  connectVerseChapter(input: [ConnectVerseChapterInput!]!): ConnectInfo!
  connectVerseCrossRefsOut(input: [ConnectVerseCrossRefsOutInput!]!): ConnectInfo!
  connectVerseCrossRefsIn(input: [ConnectVerseCrossRefsInInput!]!): ConnectInfo!
  connectVerseTgFootnotes(input: [ConnectVerseTgFootnotesInput!]!): ConnectInfo!
  connectVerseBdFootnotes(input: [ConnectVerseBdFootnotesInput!]!): ConnectInfo!
  connectVerseJstFootnotes(input: [ConnectVerseJstFootnotesInput!]!): ConnectInfo!
  connectBibleDictEntrySeeAlso(input: [ConnectBibleDictEntrySeeAlsoInput!]!): ConnectInfo!
  connectBibleDictEntryVerseRefs(input: [ConnectBibleDictEntryVerseRefsInput!]!): ConnectInfo!
  connectIndexEntrySeeAlso(input: [ConnectIndexEntrySeeAlsoInput!]!): ConnectInfo!
  connectIndexEntryTgRefs(input: [ConnectIndexEntryTgRefsInput!]!): ConnectInfo!
  connectIndexEntryBdRefs(input: [ConnectIndexEntryBdRefsInput!]!): ConnectInfo!
  connectIndexEntryVerseRefs(input: [ConnectIndexEntryVerseRefsInput!]!): ConnectInfo!
  connectTopicalGuideEntrySeeAlso(input: [ConnectTopicalGuideEntrySeeAlsoInput!]!): ConnectInfo!
  connectTopicalGuideEntryBdRefs(input: [ConnectTopicalGuideEntryBdRefsInput!]!): ConnectInfo!
  connectTopicalGuideEntryVerseRefs(input: [ConnectTopicalGuideEntryVerseRefsInput!]!): ConnectInfo!
  connectVolumeBooks(input: [ConnectVolumeBooksInput!]!): ConnectInfo!
  connectBookVolume(input: [ConnectBookVolumeInput!]!): ConnectInfo!
  connectBookChapters(input: [ConnectBookChaptersInput!]!): ConnectInfo!
}

`
