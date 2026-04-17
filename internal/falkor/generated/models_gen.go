package generated

// SortDirection represents sort ordering.
type SortDirection string

const (
	SortDirectionASC  SortDirection = "ASC"
	SortDirectionDESC SortDirection = "DESC"
)

type PageInfo struct {
	HasNextPage     bool    `json:"hasNextPage"`
	HasPreviousPage bool    `json:"hasPreviousPage"`
	StartCursor     *string `json:"startCursor,omitempty"`
	EndCursor       *string `json:"endCursor,omitempty"`
}

type DeleteInfo struct {
	NodesDeleted int `json:"nodesDeleted"`
}

type BDVerseRefProps struct {
	TargetEndVerseId *string `json:"targetEndVerseId,omitempty"`
}

type BDVerseRefPropsCreateInput struct {
	TargetEndVerseId *string `json:"targetEndVerseId,omitempty"`
}

type BDVerseRefPropsUpdateInput struct {
	TargetEndVerseId *string `json:"targetEndVerseId,omitempty"`
}

type VerseCrossRefProps struct {
	Category         *string `json:"category,omitempty"`
	FootnoteMarker   *string `json:"footnoteMarker,omitempty"`
	ReferenceText    *string `json:"referenceText,omitempty"`
	TargetEndVerseId *string `json:"targetEndVerseId,omitempty"`
}

type VerseCrossRefPropsCreateInput struct {
	Category         *string `json:"category,omitempty"`
	FootnoteMarker   *string `json:"footnoteMarker,omitempty"`
	ReferenceText    *string `json:"referenceText,omitempty"`
	TargetEndVerseId *string `json:"targetEndVerseId,omitempty"`
}

type VerseCrossRefPropsUpdateInput struct {
	Category         *string `json:"category,omitempty"`
	FootnoteMarker   *string `json:"footnoteMarker,omitempty"`
	ReferenceText    *string `json:"referenceText,omitempty"`
	TargetEndVerseId *string `json:"targetEndVerseId,omitempty"`
}

type VerseTGRefProps struct {
	FootnoteMarker *string `json:"footnoteMarker,omitempty"`
	TgTopicText    *string `json:"tgTopicText,omitempty"`
	ReferenceText  *string `json:"referenceText,omitempty"`
}

type VerseTGRefPropsCreateInput struct {
	FootnoteMarker *string `json:"footnoteMarker,omitempty"`
	TgTopicText    *string `json:"tgTopicText,omitempty"`
	ReferenceText  *string `json:"referenceText,omitempty"`
}

type VerseTGRefPropsUpdateInput struct {
	FootnoteMarker *string `json:"footnoteMarker,omitempty"`
	TgTopicText    *string `json:"tgTopicText,omitempty"`
	ReferenceText  *string `json:"referenceText,omitempty"`
}

type VerseBDRefProps struct {
	FootnoteMarker *string `json:"footnoteMarker,omitempty"`
}

type VerseBDRefPropsCreateInput struct {
	FootnoteMarker *string `json:"footnoteMarker,omitempty"`
}

type VerseBDRefPropsUpdateInput struct {
	FootnoteMarker *string `json:"footnoteMarker,omitempty"`
}

type VerseJSTRefProps struct {
	FootnoteMarker *string `json:"footnoteMarker,omitempty"`
}

type VerseJSTRefPropsCreateInput struct {
	FootnoteMarker *string `json:"footnoteMarker,omitempty"`
}

type VerseJSTRefPropsUpdateInput struct {
	FootnoteMarker *string `json:"footnoteMarker,omitempty"`
}

type IDXVerseRefProps struct {
	Phrase *string `json:"phrase,omitempty"`
}

type IDXVerseRefPropsCreateInput struct {
	Phrase *string `json:"phrase,omitempty"`
}

type IDXVerseRefPropsUpdateInput struct {
	Phrase *string `json:"phrase,omitempty"`
}

type TGVerseRefProps struct {
	Phrase *string `json:"phrase,omitempty"`
}

type TGVerseRefPropsCreateInput struct {
	Phrase *string `json:"phrase,omitempty"`
}

type TGVerseRefPropsUpdateInput struct {
	Phrase *string `json:"phrase,omitempty"`
}

type Volume struct {
	Id           string  `json:"id"`
	Name         string  `json:"name"`
	Abbreviation string  `json:"abbreviation"`
	Books        []*Book `json:"books,omitempty"`
}

type VolumeCreateInput struct {
	Id           *string `json:"id,omitempty"`
	Name         string  `json:"name"`
	Abbreviation string  `json:"abbreviation"`
}

type VolumeUpdateInput struct {
	Name         *string `json:"name,omitempty"`
	Abbreviation *string `json:"abbreviation,omitempty"`
}

type VolumeWhere struct {
	Id                     *string        `json:"id,omitempty"`
	IdNot                  *string        `json:"id_NOT,omitempty"`
	IdIn                   []string       `json:"id_IN,omitempty"`
	IdNotIn                []string       `json:"id_NOT_IN,omitempty"`
	IdContains             *string        `json:"id_CONTAINS,omitempty"`
	IdStartsWith           *string        `json:"id_STARTS_WITH,omitempty"`
	IdEndsWith             *string        `json:"id_ENDS_WITH,omitempty"`
	Name                   *string        `json:"name,omitempty"`
	NameNot                *string        `json:"name_NOT,omitempty"`
	NameIn                 []string       `json:"name_IN,omitempty"`
	NameNotIn              []string       `json:"name_NOT_IN,omitempty"`
	NameGt                 *string        `json:"name_GT,omitempty"`
	NameGte                *string        `json:"name_GTE,omitempty"`
	NameLt                 *string        `json:"name_LT,omitempty"`
	NameLte                *string        `json:"name_LTE,omitempty"`
	NameContains           *string        `json:"name_CONTAINS,omitempty"`
	NameStartsWith         *string        `json:"name_STARTS_WITH,omitempty"`
	NameEndsWith           *string        `json:"name_ENDS_WITH,omitempty"`
	Abbreviation           *string        `json:"abbreviation,omitempty"`
	AbbreviationNot        *string        `json:"abbreviation_NOT,omitempty"`
	AbbreviationIn         []string       `json:"abbreviation_IN,omitempty"`
	AbbreviationNotIn      []string       `json:"abbreviation_NOT_IN,omitempty"`
	AbbreviationGt         *string        `json:"abbreviation_GT,omitempty"`
	AbbreviationGte        *string        `json:"abbreviation_GTE,omitempty"`
	AbbreviationLt         *string        `json:"abbreviation_LT,omitempty"`
	AbbreviationLte        *string        `json:"abbreviation_LTE,omitempty"`
	AbbreviationContains   *string        `json:"abbreviation_CONTAINS,omitempty"`
	AbbreviationStartsWith *string        `json:"abbreviation_STARTS_WITH,omitempty"`
	AbbreviationEndsWith   *string        `json:"abbreviation_ENDS_WITH,omitempty"`
	AND                    []*VolumeWhere `json:"AND,omitempty"`
	OR                     []*VolumeWhere `json:"OR,omitempty"`
	NOT                    *VolumeWhere   `json:"NOT,omitempty"`
}

type VolumeSort struct {
	Id           *SortDirection `json:"id,omitempty"`
	Name         *SortDirection `json:"name,omitempty"`
	Abbreviation *SortDirection `json:"abbreviation,omitempty"`
}

type VolumesConnection struct {
	Edges      []*VolumeEdge `json:"edges"`
	TotalCount int           `json:"totalCount"`
	PageInfo   PageInfo      `json:"pageInfo"`
}

type VolumeEdge struct {
	Node   *Volume `json:"node"`
	Cursor string  `json:"cursor"`
}

type CreateVolumesMutationResponse struct {
	Volumes []*Volume `json:"volumes"`
}

type UpdateVolumesMutationResponse struct {
	Volumes []*Volume `json:"volumes"`
}

type VolumeBooksConnection struct {
	Edges      []*VolumeBooksEdge `json:"edges"`
	TotalCount int                `json:"totalCount"`
	PageInfo   PageInfo           `json:"pageInfo"`
}

type VolumeBooksEdge struct {
	Node   *Book  `json:"node"`
	Cursor string `json:"cursor"`
}

type VolumeBooksFieldInput struct {
	Create  []*VolumeBooksCreateFieldInput  `json:"create,omitempty"`
	Connect []*VolumeBooksConnectFieldInput `json:"connect,omitempty"`
}

type VolumeBooksUpdateFieldInput struct {
	Create     []*VolumeBooksCreateFieldInput      `json:"create,omitempty"`
	Connect    []*VolumeBooksConnectFieldInput     `json:"connect,omitempty"`
	Disconnect []*VolumeBooksDisconnectFieldInput  `json:"disconnect,omitempty"`
	Update     []*VolumeBooksUpdateConnectionInput `json:"update,omitempty"`
	Delete     []*VolumeBooksDeleteFieldInput      `json:"delete,omitempty"`
}

type VolumeBooksCreateFieldInput struct {
	Node BookCreateInput `json:"node"`
}

type VolumeBooksConnectFieldInput struct {
	Where BookWhere `json:"where"`
}

type VolumeBooksDisconnectFieldInput struct {
	Where BookWhere `json:"where"`
}

type VolumeBooksUpdateConnectionInput struct {
	Where BookWhere        `json:"where"`
	Node  *BookUpdateInput `json:"node,omitempty"`
}

type VolumeBooksDeleteFieldInput struct {
	Where BookWhere `json:"where"`
}

type BibleDictEntry struct {
	Id        string            `json:"id"`
	Name      string            `json:"name"`
	Text      string            `json:"text"`
	Embedding []float64         `json:"embedding"`
	SeeAlso   []*BibleDictEntry `json:"seeAlso,omitempty"`
	VerseRefs []*Verse          `json:"verseRefs,omitempty"`
}

type BibleDictEntryCreateInput struct {
	Id        *string   `json:"id,omitempty"`
	Name      string    `json:"name"`
	Text      string    `json:"text"`
	Embedding []float64 `json:"embedding"`
}

type BibleDictEntryUpdateInput struct {
	Name      *string   `json:"name,omitempty"`
	Text      *string   `json:"text,omitempty"`
	Embedding []float64 `json:"embedding,omitempty"`
}

type BibleDictEntryWhere struct {
	Id             *string                `json:"id,omitempty"`
	IdNot          *string                `json:"id_NOT,omitempty"`
	IdIn           []string               `json:"id_IN,omitempty"`
	IdNotIn        []string               `json:"id_NOT_IN,omitempty"`
	IdContains     *string                `json:"id_CONTAINS,omitempty"`
	IdStartsWith   *string                `json:"id_STARTS_WITH,omitempty"`
	IdEndsWith     *string                `json:"id_ENDS_WITH,omitempty"`
	Name           *string                `json:"name,omitempty"`
	NameNot        *string                `json:"name_NOT,omitempty"`
	NameIn         []string               `json:"name_IN,omitempty"`
	NameNotIn      []string               `json:"name_NOT_IN,omitempty"`
	NameGt         *string                `json:"name_GT,omitempty"`
	NameGte        *string                `json:"name_GTE,omitempty"`
	NameLt         *string                `json:"name_LT,omitempty"`
	NameLte        *string                `json:"name_LTE,omitempty"`
	NameContains   *string                `json:"name_CONTAINS,omitempty"`
	NameStartsWith *string                `json:"name_STARTS_WITH,omitempty"`
	NameEndsWith   *string                `json:"name_ENDS_WITH,omitempty"`
	Text           *string                `json:"text,omitempty"`
	TextNot        *string                `json:"text_NOT,omitempty"`
	TextIn         []string               `json:"text_IN,omitempty"`
	TextNotIn      []string               `json:"text_NOT_IN,omitempty"`
	TextGt         *string                `json:"text_GT,omitempty"`
	TextGte        *string                `json:"text_GTE,omitempty"`
	TextLt         *string                `json:"text_LT,omitempty"`
	TextLte        *string                `json:"text_LTE,omitempty"`
	TextContains   *string                `json:"text_CONTAINS,omitempty"`
	TextStartsWith *string                `json:"text_STARTS_WITH,omitempty"`
	TextEndsWith   *string                `json:"text_ENDS_WITH,omitempty"`
	Embedding      *[]float64             `json:"embedding,omitempty"`
	EmbeddingNot   *[]float64             `json:"embedding_NOT,omitempty"`
	EmbeddingIn    [][]float64            `json:"embedding_IN,omitempty"`
	EmbeddingNotIn [][]float64            `json:"embedding_NOT_IN,omitempty"`
	EmbeddingGt    *[]float64             `json:"embedding_GT,omitempty"`
	EmbeddingGte   *[]float64             `json:"embedding_GTE,omitempty"`
	EmbeddingLt    *[]float64             `json:"embedding_LT,omitempty"`
	EmbeddingLte   *[]float64             `json:"embedding_LTE,omitempty"`
	AND            []*BibleDictEntryWhere `json:"AND,omitempty"`
	OR             []*BibleDictEntryWhere `json:"OR,omitempty"`
	NOT            *BibleDictEntryWhere   `json:"NOT,omitempty"`
}

type BibleDictEntrySort struct {
	Id        *SortDirection `json:"id,omitempty"`
	Name      *SortDirection `json:"name,omitempty"`
	Text      *SortDirection `json:"text,omitempty"`
	Embedding *SortDirection `json:"embedding,omitempty"`
}

type BibleDictEntrysConnection struct {
	Edges      []*BibleDictEntryEdge `json:"edges"`
	TotalCount int                   `json:"totalCount"`
	PageInfo   PageInfo              `json:"pageInfo"`
}

type BibleDictEntryEdge struct {
	Node   *BibleDictEntry `json:"node"`
	Cursor string          `json:"cursor"`
}

type CreateBibleDictEntrysMutationResponse struct {
	BibleDictEntrys []*BibleDictEntry `json:"bibleDictEntries"`
}

type UpdateBibleDictEntrysMutationResponse struct {
	BibleDictEntrys []*BibleDictEntry `json:"bibleDictEntries"`
}

type BibleDictEntrySimilarResult struct {
	Score float64         `json:"score"`
	Node  *BibleDictEntry `json:"node"`
}

type BibleDictEntrySeeAlsoConnection struct {
	Edges      []*BibleDictEntrySeeAlsoEdge `json:"edges"`
	TotalCount int                          `json:"totalCount"`
	PageInfo   PageInfo                     `json:"pageInfo"`
}

type BibleDictEntrySeeAlsoEdge struct {
	Node   *BibleDictEntry `json:"node"`
	Cursor string          `json:"cursor"`
}

type BibleDictEntrySeeAlsoFieldInput struct {
	Create  []*BibleDictEntrySeeAlsoCreateFieldInput  `json:"create,omitempty"`
	Connect []*BibleDictEntrySeeAlsoConnectFieldInput `json:"connect,omitempty"`
}

type BibleDictEntrySeeAlsoUpdateFieldInput struct {
	Create     []*BibleDictEntrySeeAlsoCreateFieldInput      `json:"create,omitempty"`
	Connect    []*BibleDictEntrySeeAlsoConnectFieldInput     `json:"connect,omitempty"`
	Disconnect []*BibleDictEntrySeeAlsoDisconnectFieldInput  `json:"disconnect,omitempty"`
	Update     []*BibleDictEntrySeeAlsoUpdateConnectionInput `json:"update,omitempty"`
	Delete     []*BibleDictEntrySeeAlsoDeleteFieldInput      `json:"delete,omitempty"`
}

type BibleDictEntrySeeAlsoCreateFieldInput struct {
	Node BibleDictEntryCreateInput `json:"node"`
}

type BibleDictEntrySeeAlsoConnectFieldInput struct {
	Where BibleDictEntryWhere `json:"where"`
}

type BibleDictEntrySeeAlsoDisconnectFieldInput struct {
	Where BibleDictEntryWhere `json:"where"`
}

type BibleDictEntrySeeAlsoUpdateConnectionInput struct {
	Where BibleDictEntryWhere        `json:"where"`
	Node  *BibleDictEntryUpdateInput `json:"node,omitempty"`
}

type BibleDictEntrySeeAlsoDeleteFieldInput struct {
	Where BibleDictEntryWhere `json:"where"`
}

type BibleDictEntryVerseRefsConnection struct {
	Edges      []*BibleDictEntryVerseRefsEdge `json:"edges"`
	TotalCount int                            `json:"totalCount"`
	PageInfo   PageInfo                       `json:"pageInfo"`
}

type BibleDictEntryVerseRefsEdge struct {
	Node       *Verse           `json:"node"`
	Cursor     string           `json:"cursor"`
	Properties *BDVerseRefProps `json:"properties,omitempty"`
}

type BibleDictEntryVerseRefsFieldInput struct {
	Create  []*BibleDictEntryVerseRefsCreateFieldInput  `json:"create,omitempty"`
	Connect []*BibleDictEntryVerseRefsConnectFieldInput `json:"connect,omitempty"`
}

type BibleDictEntryVerseRefsUpdateFieldInput struct {
	Create     []*BibleDictEntryVerseRefsCreateFieldInput      `json:"create,omitempty"`
	Connect    []*BibleDictEntryVerseRefsConnectFieldInput     `json:"connect,omitempty"`
	Disconnect []*BibleDictEntryVerseRefsDisconnectFieldInput  `json:"disconnect,omitempty"`
	Update     []*BibleDictEntryVerseRefsUpdateConnectionInput `json:"update,omitempty"`
	Delete     []*BibleDictEntryVerseRefsDeleteFieldInput      `json:"delete,omitempty"`
}

type BibleDictEntryVerseRefsCreateFieldInput struct {
	Node VerseCreateInput            `json:"node"`
	Edge *BDVerseRefPropsCreateInput `json:"edge,omitempty"`
}

type BibleDictEntryVerseRefsConnectFieldInput struct {
	Where VerseWhere                  `json:"where"`
	Edge  *BDVerseRefPropsCreateInput `json:"edge,omitempty"`
}

type BibleDictEntryVerseRefsDisconnectFieldInput struct {
	Where VerseWhere `json:"where"`
}

type BibleDictEntryVerseRefsUpdateConnectionInput struct {
	Where VerseWhere                  `json:"where"`
	Node  *VerseUpdateInput           `json:"node,omitempty"`
	Edge  *BDVerseRefPropsUpdateInput `json:"edge,omitempty"`
}

type BibleDictEntryVerseRefsDeleteFieldInput struct {
	Where VerseWhere `json:"where"`
}

type Chapter struct {
	Id               string        `json:"id"`
	Number           int           `json:"number"`
	Summary          *string       `json:"summary,omitempty"`
	SummaryEmbedding []float64     `json:"summaryEmbedding"`
	Url              *string       `json:"url,omitempty"`
	Book             []*Book       `json:"book,omitempty"`
	Verses           []*Verse      `json:"verses,omitempty"`
	VerseGroups      []*VerseGroup `json:"verseGroups,omitempty"`
}

type ChapterCreateInput struct {
	Id               *string   `json:"id,omitempty"`
	Number           int       `json:"number"`
	Summary          *string   `json:"summary,omitempty"`
	SummaryEmbedding []float64 `json:"summaryEmbedding"`
	Url              *string   `json:"url,omitempty"`
}

type ChapterUpdateInput struct {
	Number           *int      `json:"number,omitempty"`
	Summary          *string   `json:"summary,omitempty"`
	SummaryEmbedding []float64 `json:"summaryEmbedding,omitempty"`
	Url              *string   `json:"url,omitempty"`
}

type ChapterWhere struct {
	Id                    *string         `json:"id,omitempty"`
	IdNot                 *string         `json:"id_NOT,omitempty"`
	IdIn                  []string        `json:"id_IN,omitempty"`
	IdNotIn               []string        `json:"id_NOT_IN,omitempty"`
	IdContains            *string         `json:"id_CONTAINS,omitempty"`
	IdStartsWith          *string         `json:"id_STARTS_WITH,omitempty"`
	IdEndsWith            *string         `json:"id_ENDS_WITH,omitempty"`
	Number                *int            `json:"number,omitempty"`
	NumberNot             *int            `json:"number_NOT,omitempty"`
	NumberIn              []int           `json:"number_IN,omitempty"`
	NumberNotIn           []int           `json:"number_NOT_IN,omitempty"`
	NumberGt              *int            `json:"number_GT,omitempty"`
	NumberGte             *int            `json:"number_GTE,omitempty"`
	NumberLt              *int            `json:"number_LT,omitempty"`
	NumberLte             *int            `json:"number_LTE,omitempty"`
	Summary               *string         `json:"summary,omitempty"`
	SummaryNot            *string         `json:"summary_NOT,omitempty"`
	SummaryIn             []string        `json:"summary_IN,omitempty"`
	SummaryNotIn          []string        `json:"summary_NOT_IN,omitempty"`
	SummaryGt             *string         `json:"summary_GT,omitempty"`
	SummaryGte            *string         `json:"summary_GTE,omitempty"`
	SummaryLt             *string         `json:"summary_LT,omitempty"`
	SummaryLte            *string         `json:"summary_LTE,omitempty"`
	SummaryContains       *string         `json:"summary_CONTAINS,omitempty"`
	SummaryStartsWith     *string         `json:"summary_STARTS_WITH,omitempty"`
	SummaryEndsWith       *string         `json:"summary_ENDS_WITH,omitempty"`
	SummaryEmbedding      *[]float64      `json:"summaryEmbedding,omitempty"`
	SummaryEmbeddingNot   *[]float64      `json:"summaryEmbedding_NOT,omitempty"`
	SummaryEmbeddingIn    [][]float64     `json:"summaryEmbedding_IN,omitempty"`
	SummaryEmbeddingNotIn [][]float64     `json:"summaryEmbedding_NOT_IN,omitempty"`
	SummaryEmbeddingGt    *[]float64      `json:"summaryEmbedding_GT,omitempty"`
	SummaryEmbeddingGte   *[]float64      `json:"summaryEmbedding_GTE,omitempty"`
	SummaryEmbeddingLt    *[]float64      `json:"summaryEmbedding_LT,omitempty"`
	SummaryEmbeddingLte   *[]float64      `json:"summaryEmbedding_LTE,omitempty"`
	Url                   *string         `json:"url,omitempty"`
	UrlNot                *string         `json:"url_NOT,omitempty"`
	UrlIn                 []string        `json:"url_IN,omitempty"`
	UrlNotIn              []string        `json:"url_NOT_IN,omitempty"`
	UrlGt                 *string         `json:"url_GT,omitempty"`
	UrlGte                *string         `json:"url_GTE,omitempty"`
	UrlLt                 *string         `json:"url_LT,omitempty"`
	UrlLte                *string         `json:"url_LTE,omitempty"`
	UrlContains           *string         `json:"url_CONTAINS,omitempty"`
	UrlStartsWith         *string         `json:"url_STARTS_WITH,omitempty"`
	UrlEndsWith           *string         `json:"url_ENDS_WITH,omitempty"`
	AND                   []*ChapterWhere `json:"AND,omitempty"`
	OR                    []*ChapterWhere `json:"OR,omitempty"`
	NOT                   *ChapterWhere   `json:"NOT,omitempty"`
}

type ChapterSort struct {
	Id               *SortDirection `json:"id,omitempty"`
	Number           *SortDirection `json:"number,omitempty"`
	Summary          *SortDirection `json:"summary,omitempty"`
	SummaryEmbedding *SortDirection `json:"summaryEmbedding,omitempty"`
	Url              *SortDirection `json:"url,omitempty"`
}

type ChaptersConnection struct {
	Edges      []*ChapterEdge `json:"edges"`
	TotalCount int            `json:"totalCount"`
	PageInfo   PageInfo       `json:"pageInfo"`
}

type ChapterEdge struct {
	Node   *Chapter `json:"node"`
	Cursor string   `json:"cursor"`
}

type CreateChaptersMutationResponse struct {
	Chapters []*Chapter `json:"chapters"`
}

type UpdateChaptersMutationResponse struct {
	Chapters []*Chapter `json:"chapters"`
}

type ChapterSimilarResult struct {
	Score float64  `json:"score"`
	Node  *Chapter `json:"node"`
}

type ChapterBookConnection struct {
	Edges      []*ChapterBookEdge `json:"edges"`
	TotalCount int                `json:"totalCount"`
	PageInfo   PageInfo           `json:"pageInfo"`
}

type ChapterBookEdge struct {
	Node   *Book  `json:"node"`
	Cursor string `json:"cursor"`
}

type ChapterBookFieldInput struct {
	Create  []*ChapterBookCreateFieldInput  `json:"create,omitempty"`
	Connect []*ChapterBookConnectFieldInput `json:"connect,omitempty"`
}

type ChapterBookUpdateFieldInput struct {
	Create     []*ChapterBookCreateFieldInput      `json:"create,omitempty"`
	Connect    []*ChapterBookConnectFieldInput     `json:"connect,omitempty"`
	Disconnect []*ChapterBookDisconnectFieldInput  `json:"disconnect,omitempty"`
	Update     []*ChapterBookUpdateConnectionInput `json:"update,omitempty"`
	Delete     []*ChapterBookDeleteFieldInput      `json:"delete,omitempty"`
}

type ChapterBookCreateFieldInput struct {
	Node BookCreateInput `json:"node"`
}

type ChapterBookConnectFieldInput struct {
	Where BookWhere `json:"where"`
}

type ChapterBookDisconnectFieldInput struct {
	Where BookWhere `json:"where"`
}

type ChapterBookUpdateConnectionInput struct {
	Where BookWhere        `json:"where"`
	Node  *BookUpdateInput `json:"node,omitempty"`
}

type ChapterBookDeleteFieldInput struct {
	Where BookWhere `json:"where"`
}

type ChapterVersesConnection struct {
	Edges      []*ChapterVersesEdge `json:"edges"`
	TotalCount int                  `json:"totalCount"`
	PageInfo   PageInfo             `json:"pageInfo"`
}

type ChapterVersesEdge struct {
	Node   *Verse `json:"node"`
	Cursor string `json:"cursor"`
}

type ChapterVersesFieldInput struct {
	Create  []*ChapterVersesCreateFieldInput  `json:"create,omitempty"`
	Connect []*ChapterVersesConnectFieldInput `json:"connect,omitempty"`
}

type ChapterVersesUpdateFieldInput struct {
	Create     []*ChapterVersesCreateFieldInput      `json:"create,omitempty"`
	Connect    []*ChapterVersesConnectFieldInput     `json:"connect,omitempty"`
	Disconnect []*ChapterVersesDisconnectFieldInput  `json:"disconnect,omitempty"`
	Update     []*ChapterVersesUpdateConnectionInput `json:"update,omitempty"`
	Delete     []*ChapterVersesDeleteFieldInput      `json:"delete,omitempty"`
}

type ChapterVersesCreateFieldInput struct {
	Node VerseCreateInput `json:"node"`
}

type ChapterVersesConnectFieldInput struct {
	Where VerseWhere `json:"where"`
}

type ChapterVersesDisconnectFieldInput struct {
	Where VerseWhere `json:"where"`
}

type ChapterVersesUpdateConnectionInput struct {
	Where VerseWhere        `json:"where"`
	Node  *VerseUpdateInput `json:"node,omitempty"`
}

type ChapterVersesDeleteFieldInput struct {
	Where VerseWhere `json:"where"`
}

type ChapterVerseGroupsConnection struct {
	Edges      []*ChapterVerseGroupsEdge `json:"edges"`
	TotalCount int                       `json:"totalCount"`
	PageInfo   PageInfo                  `json:"pageInfo"`
}

type ChapterVerseGroupsEdge struct {
	Node   *VerseGroup `json:"node"`
	Cursor string      `json:"cursor"`
}

type ChapterVerseGroupsFieldInput struct {
	Create  []*ChapterVerseGroupsCreateFieldInput  `json:"create,omitempty"`
	Connect []*ChapterVerseGroupsConnectFieldInput `json:"connect,omitempty"`
}

type ChapterVerseGroupsUpdateFieldInput struct {
	Create     []*ChapterVerseGroupsCreateFieldInput      `json:"create,omitempty"`
	Connect    []*ChapterVerseGroupsConnectFieldInput     `json:"connect,omitempty"`
	Disconnect []*ChapterVerseGroupsDisconnectFieldInput  `json:"disconnect,omitempty"`
	Update     []*ChapterVerseGroupsUpdateConnectionInput `json:"update,omitempty"`
	Delete     []*ChapterVerseGroupsDeleteFieldInput      `json:"delete,omitempty"`
}

type ChapterVerseGroupsCreateFieldInput struct {
	Node VerseGroupCreateInput `json:"node"`
}

type ChapterVerseGroupsConnectFieldInput struct {
	Where VerseGroupWhere `json:"where"`
}

type ChapterVerseGroupsDisconnectFieldInput struct {
	Where VerseGroupWhere `json:"where"`
}

type ChapterVerseGroupsUpdateConnectionInput struct {
	Where VerseGroupWhere        `json:"where"`
	Node  *VerseGroupUpdateInput `json:"node,omitempty"`
}

type ChapterVerseGroupsDeleteFieldInput struct {
	Where VerseGroupWhere `json:"where"`
}

type VerseGroup struct {
	Id               string     `json:"id"`
	Text             string     `json:"text"`
	StartVerseNumber int        `json:"startVerseNumber"`
	EndVerseNumber   int        `json:"endVerseNumber"`
	Embedding        []float64  `json:"embedding"`
	Chapter          []*Chapter `json:"chapter,omitempty"`
	Verses           []*Verse   `json:"verses,omitempty"`
}

type VerseGroupCreateInput struct {
	Id               *string   `json:"id,omitempty"`
	Text             string    `json:"text"`
	StartVerseNumber int       `json:"startVerseNumber"`
	EndVerseNumber   int       `json:"endVerseNumber"`
	Embedding        []float64 `json:"embedding"`
}

type VerseGroupUpdateInput struct {
	Text             *string   `json:"text,omitempty"`
	StartVerseNumber *int      `json:"startVerseNumber,omitempty"`
	EndVerseNumber   *int      `json:"endVerseNumber,omitempty"`
	Embedding        []float64 `json:"embedding,omitempty"`
}

type VerseGroupWhere struct {
	Id                    *string            `json:"id,omitempty"`
	IdNot                 *string            `json:"id_NOT,omitempty"`
	IdIn                  []string           `json:"id_IN,omitempty"`
	IdNotIn               []string           `json:"id_NOT_IN,omitempty"`
	IdContains            *string            `json:"id_CONTAINS,omitempty"`
	IdStartsWith          *string            `json:"id_STARTS_WITH,omitempty"`
	IdEndsWith            *string            `json:"id_ENDS_WITH,omitempty"`
	Text                  *string            `json:"text,omitempty"`
	TextNot               *string            `json:"text_NOT,omitempty"`
	TextIn                []string           `json:"text_IN,omitempty"`
	TextNotIn             []string           `json:"text_NOT_IN,omitempty"`
	TextGt                *string            `json:"text_GT,omitempty"`
	TextGte               *string            `json:"text_GTE,omitempty"`
	TextLt                *string            `json:"text_LT,omitempty"`
	TextLte               *string            `json:"text_LTE,omitempty"`
	TextContains          *string            `json:"text_CONTAINS,omitempty"`
	TextStartsWith        *string            `json:"text_STARTS_WITH,omitempty"`
	TextEndsWith          *string            `json:"text_ENDS_WITH,omitempty"`
	StartVerseNumber      *int               `json:"startVerseNumber,omitempty"`
	StartVerseNumberNot   *int               `json:"startVerseNumber_NOT,omitempty"`
	StartVerseNumberIn    []int              `json:"startVerseNumber_IN,omitempty"`
	StartVerseNumberNotIn []int              `json:"startVerseNumber_NOT_IN,omitempty"`
	StartVerseNumberGt    *int               `json:"startVerseNumber_GT,omitempty"`
	StartVerseNumberGte   *int               `json:"startVerseNumber_GTE,omitempty"`
	StartVerseNumberLt    *int               `json:"startVerseNumber_LT,omitempty"`
	StartVerseNumberLte   *int               `json:"startVerseNumber_LTE,omitempty"`
	EndVerseNumber        *int               `json:"endVerseNumber,omitempty"`
	EndVerseNumberNot     *int               `json:"endVerseNumber_NOT,omitempty"`
	EndVerseNumberIn      []int              `json:"endVerseNumber_IN,omitempty"`
	EndVerseNumberNotIn   []int              `json:"endVerseNumber_NOT_IN,omitempty"`
	EndVerseNumberGt      *int               `json:"endVerseNumber_GT,omitempty"`
	EndVerseNumberGte     *int               `json:"endVerseNumber_GTE,omitempty"`
	EndVerseNumberLt      *int               `json:"endVerseNumber_LT,omitempty"`
	EndVerseNumberLte     *int               `json:"endVerseNumber_LTE,omitempty"`
	Embedding             *[]float64         `json:"embedding,omitempty"`
	EmbeddingNot          *[]float64         `json:"embedding_NOT,omitempty"`
	EmbeddingIn           [][]float64        `json:"embedding_IN,omitempty"`
	EmbeddingNotIn        [][]float64        `json:"embedding_NOT_IN,omitempty"`
	EmbeddingGt           *[]float64         `json:"embedding_GT,omitempty"`
	EmbeddingGte          *[]float64         `json:"embedding_GTE,omitempty"`
	EmbeddingLt           *[]float64         `json:"embedding_LT,omitempty"`
	EmbeddingLte          *[]float64         `json:"embedding_LTE,omitempty"`
	AND                   []*VerseGroupWhere `json:"AND,omitempty"`
	OR                    []*VerseGroupWhere `json:"OR,omitempty"`
	NOT                   *VerseGroupWhere   `json:"NOT,omitempty"`
}

type VerseGroupSort struct {
	Id               *SortDirection `json:"id,omitempty"`
	Text             *SortDirection `json:"text,omitempty"`
	StartVerseNumber *SortDirection `json:"startVerseNumber,omitempty"`
	EndVerseNumber   *SortDirection `json:"endVerseNumber,omitempty"`
	Embedding        *SortDirection `json:"embedding,omitempty"`
}

type VerseGroupsConnection struct {
	Edges      []*VerseGroupEdge `json:"edges"`
	TotalCount int               `json:"totalCount"`
	PageInfo   PageInfo          `json:"pageInfo"`
}

type VerseGroupEdge struct {
	Node   *VerseGroup `json:"node"`
	Cursor string      `json:"cursor"`
}

type CreateVerseGroupsMutationResponse struct {
	VerseGroups []*VerseGroup `json:"verseGroups"`
}

type UpdateVerseGroupsMutationResponse struct {
	VerseGroups []*VerseGroup `json:"verseGroups"`
}

type VerseGroupSimilarResult struct {
	Score float64     `json:"score"`
	Node  *VerseGroup `json:"node"`
}

type VerseGroupChapterConnection struct {
	Edges      []*VerseGroupChapterEdge `json:"edges"`
	TotalCount int                      `json:"totalCount"`
	PageInfo   PageInfo                 `json:"pageInfo"`
}

type VerseGroupChapterEdge struct {
	Node   *Chapter `json:"node"`
	Cursor string   `json:"cursor"`
}

type VerseGroupChapterFieldInput struct {
	Create  []*VerseGroupChapterCreateFieldInput  `json:"create,omitempty"`
	Connect []*VerseGroupChapterConnectFieldInput `json:"connect,omitempty"`
}

type VerseGroupChapterUpdateFieldInput struct {
	Create     []*VerseGroupChapterCreateFieldInput      `json:"create,omitempty"`
	Connect    []*VerseGroupChapterConnectFieldInput     `json:"connect,omitempty"`
	Disconnect []*VerseGroupChapterDisconnectFieldInput  `json:"disconnect,omitempty"`
	Update     []*VerseGroupChapterUpdateConnectionInput `json:"update,omitempty"`
	Delete     []*VerseGroupChapterDeleteFieldInput      `json:"delete,omitempty"`
}

type VerseGroupChapterCreateFieldInput struct {
	Node ChapterCreateInput `json:"node"`
}

type VerseGroupChapterConnectFieldInput struct {
	Where ChapterWhere `json:"where"`
}

type VerseGroupChapterDisconnectFieldInput struct {
	Where ChapterWhere `json:"where"`
}

type VerseGroupChapterUpdateConnectionInput struct {
	Where ChapterWhere        `json:"where"`
	Node  *ChapterUpdateInput `json:"node,omitempty"`
}

type VerseGroupChapterDeleteFieldInput struct {
	Where ChapterWhere `json:"where"`
}

type VerseGroupVersesConnection struct {
	Edges      []*VerseGroupVersesEdge `json:"edges"`
	TotalCount int                     `json:"totalCount"`
	PageInfo   PageInfo                `json:"pageInfo"`
}

type VerseGroupVersesEdge struct {
	Node   *Verse `json:"node"`
	Cursor string `json:"cursor"`
}

type VerseGroupVersesFieldInput struct {
	Create  []*VerseGroupVersesCreateFieldInput  `json:"create,omitempty"`
	Connect []*VerseGroupVersesConnectFieldInput `json:"connect,omitempty"`
}

type VerseGroupVersesUpdateFieldInput struct {
	Create     []*VerseGroupVersesCreateFieldInput      `json:"create,omitempty"`
	Connect    []*VerseGroupVersesConnectFieldInput     `json:"connect,omitempty"`
	Disconnect []*VerseGroupVersesDisconnectFieldInput  `json:"disconnect,omitempty"`
	Update     []*VerseGroupVersesUpdateConnectionInput `json:"update,omitempty"`
	Delete     []*VerseGroupVersesDeleteFieldInput      `json:"delete,omitempty"`
}

type VerseGroupVersesCreateFieldInput struct {
	Node VerseCreateInput `json:"node"`
}

type VerseGroupVersesConnectFieldInput struct {
	Where VerseWhere `json:"where"`
}

type VerseGroupVersesDisconnectFieldInput struct {
	Where VerseWhere `json:"where"`
}

type VerseGroupVersesUpdateConnectionInput struct {
	Where VerseWhere        `json:"where"`
	Node  *VerseUpdateInput `json:"node,omitempty"`
}

type VerseGroupVersesDeleteFieldInput struct {
	Where VerseWhere `json:"where"`
}

type Book struct {
	Id       string     `json:"id"`
	Name     string     `json:"name"`
	Slug     string     `json:"slug"`
	UrlPath  string     `json:"urlPath"`
	Volume   []*Volume  `json:"volume,omitempty"`
	Chapters []*Chapter `json:"chapters,omitempty"`
}

type BookCreateInput struct {
	Id      *string `json:"id,omitempty"`
	Name    string  `json:"name"`
	Slug    string  `json:"slug"`
	UrlPath string  `json:"urlPath"`
}

type BookUpdateInput struct {
	Name    *string `json:"name,omitempty"`
	Slug    *string `json:"slug,omitempty"`
	UrlPath *string `json:"urlPath,omitempty"`
}

type BookWhere struct {
	Id                *string      `json:"id,omitempty"`
	IdNot             *string      `json:"id_NOT,omitempty"`
	IdIn              []string     `json:"id_IN,omitempty"`
	IdNotIn           []string     `json:"id_NOT_IN,omitempty"`
	IdContains        *string      `json:"id_CONTAINS,omitempty"`
	IdStartsWith      *string      `json:"id_STARTS_WITH,omitempty"`
	IdEndsWith        *string      `json:"id_ENDS_WITH,omitempty"`
	Name              *string      `json:"name,omitempty"`
	NameNot           *string      `json:"name_NOT,omitempty"`
	NameIn            []string     `json:"name_IN,omitempty"`
	NameNotIn         []string     `json:"name_NOT_IN,omitempty"`
	NameGt            *string      `json:"name_GT,omitempty"`
	NameGte           *string      `json:"name_GTE,omitempty"`
	NameLt            *string      `json:"name_LT,omitempty"`
	NameLte           *string      `json:"name_LTE,omitempty"`
	NameContains      *string      `json:"name_CONTAINS,omitempty"`
	NameStartsWith    *string      `json:"name_STARTS_WITH,omitempty"`
	NameEndsWith      *string      `json:"name_ENDS_WITH,omitempty"`
	Slug              *string      `json:"slug,omitempty"`
	SlugNot           *string      `json:"slug_NOT,omitempty"`
	SlugIn            []string     `json:"slug_IN,omitempty"`
	SlugNotIn         []string     `json:"slug_NOT_IN,omitempty"`
	SlugGt            *string      `json:"slug_GT,omitempty"`
	SlugGte           *string      `json:"slug_GTE,omitempty"`
	SlugLt            *string      `json:"slug_LT,omitempty"`
	SlugLte           *string      `json:"slug_LTE,omitempty"`
	SlugContains      *string      `json:"slug_CONTAINS,omitempty"`
	SlugStartsWith    *string      `json:"slug_STARTS_WITH,omitempty"`
	SlugEndsWith      *string      `json:"slug_ENDS_WITH,omitempty"`
	UrlPath           *string      `json:"urlPath,omitempty"`
	UrlPathNot        *string      `json:"urlPath_NOT,omitempty"`
	UrlPathIn         []string     `json:"urlPath_IN,omitempty"`
	UrlPathNotIn      []string     `json:"urlPath_NOT_IN,omitempty"`
	UrlPathGt         *string      `json:"urlPath_GT,omitempty"`
	UrlPathGte        *string      `json:"urlPath_GTE,omitempty"`
	UrlPathLt         *string      `json:"urlPath_LT,omitempty"`
	UrlPathLte        *string      `json:"urlPath_LTE,omitempty"`
	UrlPathContains   *string      `json:"urlPath_CONTAINS,omitempty"`
	UrlPathStartsWith *string      `json:"urlPath_STARTS_WITH,omitempty"`
	UrlPathEndsWith   *string      `json:"urlPath_ENDS_WITH,omitempty"`
	AND               []*BookWhere `json:"AND,omitempty"`
	OR                []*BookWhere `json:"OR,omitempty"`
	NOT               *BookWhere   `json:"NOT,omitempty"`
}

type BookSort struct {
	Id      *SortDirection `json:"id,omitempty"`
	Name    *SortDirection `json:"name,omitempty"`
	Slug    *SortDirection `json:"slug,omitempty"`
	UrlPath *SortDirection `json:"urlPath,omitempty"`
}

type BooksConnection struct {
	Edges      []*BookEdge `json:"edges"`
	TotalCount int         `json:"totalCount"`
	PageInfo   PageInfo    `json:"pageInfo"`
}

type BookEdge struct {
	Node   *Book  `json:"node"`
	Cursor string `json:"cursor"`
}

type CreateBooksMutationResponse struct {
	Books []*Book `json:"books"`
}

type UpdateBooksMutationResponse struct {
	Books []*Book `json:"books"`
}

type BookVolumeConnection struct {
	Edges      []*BookVolumeEdge `json:"edges"`
	TotalCount int               `json:"totalCount"`
	PageInfo   PageInfo          `json:"pageInfo"`
}

type BookVolumeEdge struct {
	Node   *Volume `json:"node"`
	Cursor string  `json:"cursor"`
}

type BookVolumeFieldInput struct {
	Create  []*BookVolumeCreateFieldInput  `json:"create,omitempty"`
	Connect []*BookVolumeConnectFieldInput `json:"connect,omitempty"`
}

type BookVolumeUpdateFieldInput struct {
	Create     []*BookVolumeCreateFieldInput      `json:"create,omitempty"`
	Connect    []*BookVolumeConnectFieldInput     `json:"connect,omitempty"`
	Disconnect []*BookVolumeDisconnectFieldInput  `json:"disconnect,omitempty"`
	Update     []*BookVolumeUpdateConnectionInput `json:"update,omitempty"`
	Delete     []*BookVolumeDeleteFieldInput      `json:"delete,omitempty"`
}

type BookVolumeCreateFieldInput struct {
	Node VolumeCreateInput `json:"node"`
}

type BookVolumeConnectFieldInput struct {
	Where VolumeWhere `json:"where"`
}

type BookVolumeDisconnectFieldInput struct {
	Where VolumeWhere `json:"where"`
}

type BookVolumeUpdateConnectionInput struct {
	Where VolumeWhere        `json:"where"`
	Node  *VolumeUpdateInput `json:"node,omitempty"`
}

type BookVolumeDeleteFieldInput struct {
	Where VolumeWhere `json:"where"`
}

type BookChaptersConnection struct {
	Edges      []*BookChaptersEdge `json:"edges"`
	TotalCount int                 `json:"totalCount"`
	PageInfo   PageInfo            `json:"pageInfo"`
}

type BookChaptersEdge struct {
	Node   *Chapter `json:"node"`
	Cursor string   `json:"cursor"`
}

type BookChaptersFieldInput struct {
	Create  []*BookChaptersCreateFieldInput  `json:"create,omitempty"`
	Connect []*BookChaptersConnectFieldInput `json:"connect,omitempty"`
}

type BookChaptersUpdateFieldInput struct {
	Create     []*BookChaptersCreateFieldInput      `json:"create,omitempty"`
	Connect    []*BookChaptersConnectFieldInput     `json:"connect,omitempty"`
	Disconnect []*BookChaptersDisconnectFieldInput  `json:"disconnect,omitempty"`
	Update     []*BookChaptersUpdateConnectionInput `json:"update,omitempty"`
	Delete     []*BookChaptersDeleteFieldInput      `json:"delete,omitempty"`
}

type BookChaptersCreateFieldInput struct {
	Node ChapterCreateInput `json:"node"`
}

type BookChaptersConnectFieldInput struct {
	Where ChapterWhere `json:"where"`
}

type BookChaptersDisconnectFieldInput struct {
	Where ChapterWhere `json:"where"`
}

type BookChaptersUpdateConnectionInput struct {
	Where ChapterWhere        `json:"where"`
	Node  *ChapterUpdateInput `json:"node,omitempty"`
}

type BookChaptersDeleteFieldInput struct {
	Where ChapterWhere `json:"where"`
}

type Verse struct {
	Id                string               `json:"id"`
	Number            int                  `json:"number"`
	Reference         string               `json:"reference"`
	Text              string               `json:"text"`
	TranslationNotes  *string              `json:"translationNotes,omitempty"`
	AlternateReadings *string              `json:"alternateReadings,omitempty"`
	ExplanatoryNotes  *string              `json:"explanatoryNotes,omitempty"`
	Chapter           []*Chapter           `json:"chapter,omitempty"`
	CrossRefsOut      []*Verse             `json:"crossRefsOut,omitempty"`
	CrossRefsIn       []*Verse             `json:"crossRefsIn,omitempty"`
	TgFootnotes       []*TopicalGuideEntry `json:"tgFootnotes,omitempty"`
	BdFootnotes       []*BibleDictEntry    `json:"bdFootnotes,omitempty"`
	JstFootnotes      []*JSTPassage        `json:"jstFootnotes,omitempty"`
}

type VerseCreateInput struct {
	Id                *string `json:"id,omitempty"`
	Number            int     `json:"number"`
	Reference         string  `json:"reference"`
	Text              string  `json:"text"`
	TranslationNotes  *string `json:"translationNotes,omitempty"`
	AlternateReadings *string `json:"alternateReadings,omitempty"`
	ExplanatoryNotes  *string `json:"explanatoryNotes,omitempty"`
}

type VerseUpdateInput struct {
	Number            *int    `json:"number,omitempty"`
	Reference         *string `json:"reference,omitempty"`
	Text              *string `json:"text,omitempty"`
	TranslationNotes  *string `json:"translationNotes,omitempty"`
	AlternateReadings *string `json:"alternateReadings,omitempty"`
	ExplanatoryNotes  *string `json:"explanatoryNotes,omitempty"`
}

type VerseWhere struct {
	Id                          *string       `json:"id,omitempty"`
	IdNot                       *string       `json:"id_NOT,omitempty"`
	IdIn                        []string      `json:"id_IN,omitempty"`
	IdNotIn                     []string      `json:"id_NOT_IN,omitempty"`
	IdContains                  *string       `json:"id_CONTAINS,omitempty"`
	IdStartsWith                *string       `json:"id_STARTS_WITH,omitempty"`
	IdEndsWith                  *string       `json:"id_ENDS_WITH,omitempty"`
	Number                      *int          `json:"number,omitempty"`
	NumberNot                   *int          `json:"number_NOT,omitempty"`
	NumberIn                    []int         `json:"number_IN,omitempty"`
	NumberNotIn                 []int         `json:"number_NOT_IN,omitempty"`
	NumberGt                    *int          `json:"number_GT,omitempty"`
	NumberGte                   *int          `json:"number_GTE,omitempty"`
	NumberLt                    *int          `json:"number_LT,omitempty"`
	NumberLte                   *int          `json:"number_LTE,omitempty"`
	Reference                   *string       `json:"reference,omitempty"`
	ReferenceNot                *string       `json:"reference_NOT,omitempty"`
	ReferenceIn                 []string      `json:"reference_IN,omitempty"`
	ReferenceNotIn              []string      `json:"reference_NOT_IN,omitempty"`
	ReferenceGt                 *string       `json:"reference_GT,omitempty"`
	ReferenceGte                *string       `json:"reference_GTE,omitempty"`
	ReferenceLt                 *string       `json:"reference_LT,omitempty"`
	ReferenceLte                *string       `json:"reference_LTE,omitempty"`
	ReferenceContains           *string       `json:"reference_CONTAINS,omitempty"`
	ReferenceStartsWith         *string       `json:"reference_STARTS_WITH,omitempty"`
	ReferenceEndsWith           *string       `json:"reference_ENDS_WITH,omitempty"`
	Text                        *string       `json:"text,omitempty"`
	TextNot                     *string       `json:"text_NOT,omitempty"`
	TextIn                      []string      `json:"text_IN,omitempty"`
	TextNotIn                   []string      `json:"text_NOT_IN,omitempty"`
	TextGt                      *string       `json:"text_GT,omitempty"`
	TextGte                     *string       `json:"text_GTE,omitempty"`
	TextLt                      *string       `json:"text_LT,omitempty"`
	TextLte                     *string       `json:"text_LTE,omitempty"`
	TextContains                *string       `json:"text_CONTAINS,omitempty"`
	TextStartsWith              *string       `json:"text_STARTS_WITH,omitempty"`
	TextEndsWith                *string       `json:"text_ENDS_WITH,omitempty"`
	TranslationNotes            *string       `json:"translationNotes,omitempty"`
	TranslationNotesNot         *string       `json:"translationNotes_NOT,omitempty"`
	TranslationNotesIn          []string      `json:"translationNotes_IN,omitempty"`
	TranslationNotesNotIn       []string      `json:"translationNotes_NOT_IN,omitempty"`
	TranslationNotesGt          *string       `json:"translationNotes_GT,omitempty"`
	TranslationNotesGte         *string       `json:"translationNotes_GTE,omitempty"`
	TranslationNotesLt          *string       `json:"translationNotes_LT,omitempty"`
	TranslationNotesLte         *string       `json:"translationNotes_LTE,omitempty"`
	TranslationNotesContains    *string       `json:"translationNotes_CONTAINS,omitempty"`
	TranslationNotesStartsWith  *string       `json:"translationNotes_STARTS_WITH,omitempty"`
	TranslationNotesEndsWith    *string       `json:"translationNotes_ENDS_WITH,omitempty"`
	AlternateReadings           *string       `json:"alternateReadings,omitempty"`
	AlternateReadingsNot        *string       `json:"alternateReadings_NOT,omitempty"`
	AlternateReadingsIn         []string      `json:"alternateReadings_IN,omitempty"`
	AlternateReadingsNotIn      []string      `json:"alternateReadings_NOT_IN,omitempty"`
	AlternateReadingsGt         *string       `json:"alternateReadings_GT,omitempty"`
	AlternateReadingsGte        *string       `json:"alternateReadings_GTE,omitempty"`
	AlternateReadingsLt         *string       `json:"alternateReadings_LT,omitempty"`
	AlternateReadingsLte        *string       `json:"alternateReadings_LTE,omitempty"`
	AlternateReadingsContains   *string       `json:"alternateReadings_CONTAINS,omitempty"`
	AlternateReadingsStartsWith *string       `json:"alternateReadings_STARTS_WITH,omitempty"`
	AlternateReadingsEndsWith   *string       `json:"alternateReadings_ENDS_WITH,omitempty"`
	ExplanatoryNotes            *string       `json:"explanatoryNotes,omitempty"`
	ExplanatoryNotesNot         *string       `json:"explanatoryNotes_NOT,omitempty"`
	ExplanatoryNotesIn          []string      `json:"explanatoryNotes_IN,omitempty"`
	ExplanatoryNotesNotIn       []string      `json:"explanatoryNotes_NOT_IN,omitempty"`
	ExplanatoryNotesGt          *string       `json:"explanatoryNotes_GT,omitempty"`
	ExplanatoryNotesGte         *string       `json:"explanatoryNotes_GTE,omitempty"`
	ExplanatoryNotesLt          *string       `json:"explanatoryNotes_LT,omitempty"`
	ExplanatoryNotesLte         *string       `json:"explanatoryNotes_LTE,omitempty"`
	ExplanatoryNotesContains    *string       `json:"explanatoryNotes_CONTAINS,omitempty"`
	ExplanatoryNotesStartsWith  *string       `json:"explanatoryNotes_STARTS_WITH,omitempty"`
	ExplanatoryNotesEndsWith    *string       `json:"explanatoryNotes_ENDS_WITH,omitempty"`
	AND                         []*VerseWhere `json:"AND,omitempty"`
	OR                          []*VerseWhere `json:"OR,omitempty"`
	NOT                         *VerseWhere   `json:"NOT,omitempty"`
}

type VerseSort struct {
	Id                *SortDirection `json:"id,omitempty"`
	Number            *SortDirection `json:"number,omitempty"`
	Reference         *SortDirection `json:"reference,omitempty"`
	Text              *SortDirection `json:"text,omitempty"`
	TranslationNotes  *SortDirection `json:"translationNotes,omitempty"`
	AlternateReadings *SortDirection `json:"alternateReadings,omitempty"`
	ExplanatoryNotes  *SortDirection `json:"explanatoryNotes,omitempty"`
}

type VersesConnection struct {
	Edges      []*VerseEdge `json:"edges"`
	TotalCount int          `json:"totalCount"`
	PageInfo   PageInfo     `json:"pageInfo"`
}

type VerseEdge struct {
	Node   *Verse `json:"node"`
	Cursor string `json:"cursor"`
}

type CreateVersesMutationResponse struct {
	Verses []*Verse `json:"verses"`
}

type UpdateVersesMutationResponse struct {
	Verses []*Verse `json:"verses"`
}

type VerseChapterConnection struct {
	Edges      []*VerseChapterEdge `json:"edges"`
	TotalCount int                 `json:"totalCount"`
	PageInfo   PageInfo            `json:"pageInfo"`
}

type VerseChapterEdge struct {
	Node   *Chapter `json:"node"`
	Cursor string   `json:"cursor"`
}

type VerseChapterFieldInput struct {
	Create  []*VerseChapterCreateFieldInput  `json:"create,omitempty"`
	Connect []*VerseChapterConnectFieldInput `json:"connect,omitempty"`
}

type VerseChapterUpdateFieldInput struct {
	Create     []*VerseChapterCreateFieldInput      `json:"create,omitempty"`
	Connect    []*VerseChapterConnectFieldInput     `json:"connect,omitempty"`
	Disconnect []*VerseChapterDisconnectFieldInput  `json:"disconnect,omitempty"`
	Update     []*VerseChapterUpdateConnectionInput `json:"update,omitempty"`
	Delete     []*VerseChapterDeleteFieldInput      `json:"delete,omitempty"`
}

type VerseChapterCreateFieldInput struct {
	Node ChapterCreateInput `json:"node"`
}

type VerseChapterConnectFieldInput struct {
	Where ChapterWhere `json:"where"`
}

type VerseChapterDisconnectFieldInput struct {
	Where ChapterWhere `json:"where"`
}

type VerseChapterUpdateConnectionInput struct {
	Where ChapterWhere        `json:"where"`
	Node  *ChapterUpdateInput `json:"node,omitempty"`
}

type VerseChapterDeleteFieldInput struct {
	Where ChapterWhere `json:"where"`
}

type VerseCrossRefsOutConnection struct {
	Edges      []*VerseCrossRefsOutEdge `json:"edges"`
	TotalCount int                      `json:"totalCount"`
	PageInfo   PageInfo                 `json:"pageInfo"`
}

type VerseCrossRefsOutEdge struct {
	Node       *Verse              `json:"node"`
	Cursor     string              `json:"cursor"`
	Properties *VerseCrossRefProps `json:"properties,omitempty"`
}

type VerseCrossRefsOutFieldInput struct {
	Create  []*VerseCrossRefsOutCreateFieldInput  `json:"create,omitempty"`
	Connect []*VerseCrossRefsOutConnectFieldInput `json:"connect,omitempty"`
}

type VerseCrossRefsOutUpdateFieldInput struct {
	Create     []*VerseCrossRefsOutCreateFieldInput      `json:"create,omitempty"`
	Connect    []*VerseCrossRefsOutConnectFieldInput     `json:"connect,omitempty"`
	Disconnect []*VerseCrossRefsOutDisconnectFieldInput  `json:"disconnect,omitempty"`
	Update     []*VerseCrossRefsOutUpdateConnectionInput `json:"update,omitempty"`
	Delete     []*VerseCrossRefsOutDeleteFieldInput      `json:"delete,omitempty"`
}

type VerseCrossRefsOutCreateFieldInput struct {
	Node VerseCreateInput               `json:"node"`
	Edge *VerseCrossRefPropsCreateInput `json:"edge,omitempty"`
}

type VerseCrossRefsOutConnectFieldInput struct {
	Where VerseWhere                     `json:"where"`
	Edge  *VerseCrossRefPropsCreateInput `json:"edge,omitempty"`
}

type VerseCrossRefsOutDisconnectFieldInput struct {
	Where VerseWhere `json:"where"`
}

type VerseCrossRefsOutUpdateConnectionInput struct {
	Where VerseWhere                     `json:"where"`
	Node  *VerseUpdateInput              `json:"node,omitempty"`
	Edge  *VerseCrossRefPropsUpdateInput `json:"edge,omitempty"`
}

type VerseCrossRefsOutDeleteFieldInput struct {
	Where VerseWhere `json:"where"`
}

type VerseCrossRefsInConnection struct {
	Edges      []*VerseCrossRefsInEdge `json:"edges"`
	TotalCount int                     `json:"totalCount"`
	PageInfo   PageInfo                `json:"pageInfo"`
}

type VerseCrossRefsInEdge struct {
	Node       *Verse              `json:"node"`
	Cursor     string              `json:"cursor"`
	Properties *VerseCrossRefProps `json:"properties,omitempty"`
}

type VerseCrossRefsInFieldInput struct {
	Create  []*VerseCrossRefsInCreateFieldInput  `json:"create,omitempty"`
	Connect []*VerseCrossRefsInConnectFieldInput `json:"connect,omitempty"`
}

type VerseCrossRefsInUpdateFieldInput struct {
	Create     []*VerseCrossRefsInCreateFieldInput      `json:"create,omitempty"`
	Connect    []*VerseCrossRefsInConnectFieldInput     `json:"connect,omitempty"`
	Disconnect []*VerseCrossRefsInDisconnectFieldInput  `json:"disconnect,omitempty"`
	Update     []*VerseCrossRefsInUpdateConnectionInput `json:"update,omitempty"`
	Delete     []*VerseCrossRefsInDeleteFieldInput      `json:"delete,omitempty"`
}

type VerseCrossRefsInCreateFieldInput struct {
	Node VerseCreateInput               `json:"node"`
	Edge *VerseCrossRefPropsCreateInput `json:"edge,omitempty"`
}

type VerseCrossRefsInConnectFieldInput struct {
	Where VerseWhere                     `json:"where"`
	Edge  *VerseCrossRefPropsCreateInput `json:"edge,omitempty"`
}

type VerseCrossRefsInDisconnectFieldInput struct {
	Where VerseWhere `json:"where"`
}

type VerseCrossRefsInUpdateConnectionInput struct {
	Where VerseWhere                     `json:"where"`
	Node  *VerseUpdateInput              `json:"node,omitempty"`
	Edge  *VerseCrossRefPropsUpdateInput `json:"edge,omitempty"`
}

type VerseCrossRefsInDeleteFieldInput struct {
	Where VerseWhere `json:"where"`
}

type VerseTgFootnotesConnection struct {
	Edges      []*VerseTgFootnotesEdge `json:"edges"`
	TotalCount int                     `json:"totalCount"`
	PageInfo   PageInfo                `json:"pageInfo"`
}

type VerseTgFootnotesEdge struct {
	Node       *TopicalGuideEntry `json:"node"`
	Cursor     string             `json:"cursor"`
	Properties *VerseTGRefProps   `json:"properties,omitempty"`
}

type VerseTgFootnotesFieldInput struct {
	Create  []*VerseTgFootnotesCreateFieldInput  `json:"create,omitempty"`
	Connect []*VerseTgFootnotesConnectFieldInput `json:"connect,omitempty"`
}

type VerseTgFootnotesUpdateFieldInput struct {
	Create     []*VerseTgFootnotesCreateFieldInput      `json:"create,omitempty"`
	Connect    []*VerseTgFootnotesConnectFieldInput     `json:"connect,omitempty"`
	Disconnect []*VerseTgFootnotesDisconnectFieldInput  `json:"disconnect,omitempty"`
	Update     []*VerseTgFootnotesUpdateConnectionInput `json:"update,omitempty"`
	Delete     []*VerseTgFootnotesDeleteFieldInput      `json:"delete,omitempty"`
}

type VerseTgFootnotesCreateFieldInput struct {
	Node TopicalGuideEntryCreateInput `json:"node"`
	Edge *VerseTGRefPropsCreateInput  `json:"edge,omitempty"`
}

type VerseTgFootnotesConnectFieldInput struct {
	Where TopicalGuideEntryWhere      `json:"where"`
	Edge  *VerseTGRefPropsCreateInput `json:"edge,omitempty"`
}

type VerseTgFootnotesDisconnectFieldInput struct {
	Where TopicalGuideEntryWhere `json:"where"`
}

type VerseTgFootnotesUpdateConnectionInput struct {
	Where TopicalGuideEntryWhere        `json:"where"`
	Node  *TopicalGuideEntryUpdateInput `json:"node,omitempty"`
	Edge  *VerseTGRefPropsUpdateInput   `json:"edge,omitempty"`
}

type VerseTgFootnotesDeleteFieldInput struct {
	Where TopicalGuideEntryWhere `json:"where"`
}

type VerseBdFootnotesConnection struct {
	Edges      []*VerseBdFootnotesEdge `json:"edges"`
	TotalCount int                     `json:"totalCount"`
	PageInfo   PageInfo                `json:"pageInfo"`
}

type VerseBdFootnotesEdge struct {
	Node       *BibleDictEntry  `json:"node"`
	Cursor     string           `json:"cursor"`
	Properties *VerseBDRefProps `json:"properties,omitempty"`
}

type VerseBdFootnotesFieldInput struct {
	Create  []*VerseBdFootnotesCreateFieldInput  `json:"create,omitempty"`
	Connect []*VerseBdFootnotesConnectFieldInput `json:"connect,omitempty"`
}

type VerseBdFootnotesUpdateFieldInput struct {
	Create     []*VerseBdFootnotesCreateFieldInput      `json:"create,omitempty"`
	Connect    []*VerseBdFootnotesConnectFieldInput     `json:"connect,omitempty"`
	Disconnect []*VerseBdFootnotesDisconnectFieldInput  `json:"disconnect,omitempty"`
	Update     []*VerseBdFootnotesUpdateConnectionInput `json:"update,omitempty"`
	Delete     []*VerseBdFootnotesDeleteFieldInput      `json:"delete,omitempty"`
}

type VerseBdFootnotesCreateFieldInput struct {
	Node BibleDictEntryCreateInput   `json:"node"`
	Edge *VerseBDRefPropsCreateInput `json:"edge,omitempty"`
}

type VerseBdFootnotesConnectFieldInput struct {
	Where BibleDictEntryWhere         `json:"where"`
	Edge  *VerseBDRefPropsCreateInput `json:"edge,omitempty"`
}

type VerseBdFootnotesDisconnectFieldInput struct {
	Where BibleDictEntryWhere `json:"where"`
}

type VerseBdFootnotesUpdateConnectionInput struct {
	Where BibleDictEntryWhere         `json:"where"`
	Node  *BibleDictEntryUpdateInput  `json:"node,omitempty"`
	Edge  *VerseBDRefPropsUpdateInput `json:"edge,omitempty"`
}

type VerseBdFootnotesDeleteFieldInput struct {
	Where BibleDictEntryWhere `json:"where"`
}

type VerseJstFootnotesConnection struct {
	Edges      []*VerseJstFootnotesEdge `json:"edges"`
	TotalCount int                      `json:"totalCount"`
	PageInfo   PageInfo                 `json:"pageInfo"`
}

type VerseJstFootnotesEdge struct {
	Node       *JSTPassage       `json:"node"`
	Cursor     string            `json:"cursor"`
	Properties *VerseJSTRefProps `json:"properties,omitempty"`
}

type VerseJstFootnotesFieldInput struct {
	Create  []*VerseJstFootnotesCreateFieldInput  `json:"create,omitempty"`
	Connect []*VerseJstFootnotesConnectFieldInput `json:"connect,omitempty"`
}

type VerseJstFootnotesUpdateFieldInput struct {
	Create     []*VerseJstFootnotesCreateFieldInput      `json:"create,omitempty"`
	Connect    []*VerseJstFootnotesConnectFieldInput     `json:"connect,omitempty"`
	Disconnect []*VerseJstFootnotesDisconnectFieldInput  `json:"disconnect,omitempty"`
	Update     []*VerseJstFootnotesUpdateConnectionInput `json:"update,omitempty"`
	Delete     []*VerseJstFootnotesDeleteFieldInput      `json:"delete,omitempty"`
}

type VerseJstFootnotesCreateFieldInput struct {
	Node JSTPassageCreateInput        `json:"node"`
	Edge *VerseJSTRefPropsCreateInput `json:"edge,omitempty"`
}

type VerseJstFootnotesConnectFieldInput struct {
	Where JSTPassageWhere              `json:"where"`
	Edge  *VerseJSTRefPropsCreateInput `json:"edge,omitempty"`
}

type VerseJstFootnotesDisconnectFieldInput struct {
	Where JSTPassageWhere `json:"where"`
}

type VerseJstFootnotesUpdateConnectionInput struct {
	Where JSTPassageWhere              `json:"where"`
	Node  *JSTPassageUpdateInput       `json:"node,omitempty"`
	Edge  *VerseJSTRefPropsUpdateInput `json:"edge,omitempty"`
}

type VerseJstFootnotesDeleteFieldInput struct {
	Where JSTPassageWhere `json:"where"`
}

type IndexEntry struct {
	Id        string               `json:"id"`
	Name      string               `json:"name"`
	Embedding []float64            `json:"embedding"`
	SeeAlso   []*IndexEntry        `json:"seeAlso,omitempty"`
	TgRefs    []*TopicalGuideEntry `json:"tgRefs,omitempty"`
	BdRefs    []*BibleDictEntry    `json:"bdRefs,omitempty"`
	VerseRefs []*Verse             `json:"verseRefs,omitempty"`
}

type IndexEntryCreateInput struct {
	Id        *string   `json:"id,omitempty"`
	Name      string    `json:"name"`
	Embedding []float64 `json:"embedding"`
}

type IndexEntryUpdateInput struct {
	Name      *string   `json:"name,omitempty"`
	Embedding []float64 `json:"embedding,omitempty"`
}

type IndexEntryWhere struct {
	Id             *string            `json:"id,omitempty"`
	IdNot          *string            `json:"id_NOT,omitempty"`
	IdIn           []string           `json:"id_IN,omitempty"`
	IdNotIn        []string           `json:"id_NOT_IN,omitempty"`
	IdContains     *string            `json:"id_CONTAINS,omitempty"`
	IdStartsWith   *string            `json:"id_STARTS_WITH,omitempty"`
	IdEndsWith     *string            `json:"id_ENDS_WITH,omitempty"`
	Name           *string            `json:"name,omitempty"`
	NameNot        *string            `json:"name_NOT,omitempty"`
	NameIn         []string           `json:"name_IN,omitempty"`
	NameNotIn      []string           `json:"name_NOT_IN,omitempty"`
	NameGt         *string            `json:"name_GT,omitempty"`
	NameGte        *string            `json:"name_GTE,omitempty"`
	NameLt         *string            `json:"name_LT,omitempty"`
	NameLte        *string            `json:"name_LTE,omitempty"`
	NameContains   *string            `json:"name_CONTAINS,omitempty"`
	NameStartsWith *string            `json:"name_STARTS_WITH,omitempty"`
	NameEndsWith   *string            `json:"name_ENDS_WITH,omitempty"`
	Embedding      *[]float64         `json:"embedding,omitempty"`
	EmbeddingNot   *[]float64         `json:"embedding_NOT,omitempty"`
	EmbeddingIn    [][]float64        `json:"embedding_IN,omitempty"`
	EmbeddingNotIn [][]float64        `json:"embedding_NOT_IN,omitempty"`
	EmbeddingGt    *[]float64         `json:"embedding_GT,omitempty"`
	EmbeddingGte   *[]float64         `json:"embedding_GTE,omitempty"`
	EmbeddingLt    *[]float64         `json:"embedding_LT,omitempty"`
	EmbeddingLte   *[]float64         `json:"embedding_LTE,omitempty"`
	AND            []*IndexEntryWhere `json:"AND,omitempty"`
	OR             []*IndexEntryWhere `json:"OR,omitempty"`
	NOT            *IndexEntryWhere   `json:"NOT,omitempty"`
}

type IndexEntrySort struct {
	Id        *SortDirection `json:"id,omitempty"`
	Name      *SortDirection `json:"name,omitempty"`
	Embedding *SortDirection `json:"embedding,omitempty"`
}

type IndexEntrysConnection struct {
	Edges      []*IndexEntryEdge `json:"edges"`
	TotalCount int               `json:"totalCount"`
	PageInfo   PageInfo          `json:"pageInfo"`
}

type IndexEntryEdge struct {
	Node   *IndexEntry `json:"node"`
	Cursor string      `json:"cursor"`
}

type CreateIndexEntrysMutationResponse struct {
	IndexEntrys []*IndexEntry `json:"indexEntries"`
}

type UpdateIndexEntrysMutationResponse struct {
	IndexEntrys []*IndexEntry `json:"indexEntries"`
}

type IndexEntrySimilarResult struct {
	Score float64     `json:"score"`
	Node  *IndexEntry `json:"node"`
}

type IndexEntrySeeAlsoConnection struct {
	Edges      []*IndexEntrySeeAlsoEdge `json:"edges"`
	TotalCount int                      `json:"totalCount"`
	PageInfo   PageInfo                 `json:"pageInfo"`
}

type IndexEntrySeeAlsoEdge struct {
	Node   *IndexEntry `json:"node"`
	Cursor string      `json:"cursor"`
}

type IndexEntrySeeAlsoFieldInput struct {
	Create  []*IndexEntrySeeAlsoCreateFieldInput  `json:"create,omitempty"`
	Connect []*IndexEntrySeeAlsoConnectFieldInput `json:"connect,omitempty"`
}

type IndexEntrySeeAlsoUpdateFieldInput struct {
	Create     []*IndexEntrySeeAlsoCreateFieldInput      `json:"create,omitempty"`
	Connect    []*IndexEntrySeeAlsoConnectFieldInput     `json:"connect,omitempty"`
	Disconnect []*IndexEntrySeeAlsoDisconnectFieldInput  `json:"disconnect,omitempty"`
	Update     []*IndexEntrySeeAlsoUpdateConnectionInput `json:"update,omitempty"`
	Delete     []*IndexEntrySeeAlsoDeleteFieldInput      `json:"delete,omitempty"`
}

type IndexEntrySeeAlsoCreateFieldInput struct {
	Node IndexEntryCreateInput `json:"node"`
}

type IndexEntrySeeAlsoConnectFieldInput struct {
	Where IndexEntryWhere `json:"where"`
}

type IndexEntrySeeAlsoDisconnectFieldInput struct {
	Where IndexEntryWhere `json:"where"`
}

type IndexEntrySeeAlsoUpdateConnectionInput struct {
	Where IndexEntryWhere        `json:"where"`
	Node  *IndexEntryUpdateInput `json:"node,omitempty"`
}

type IndexEntrySeeAlsoDeleteFieldInput struct {
	Where IndexEntryWhere `json:"where"`
}

type IndexEntryTgRefsConnection struct {
	Edges      []*IndexEntryTgRefsEdge `json:"edges"`
	TotalCount int                     `json:"totalCount"`
	PageInfo   PageInfo                `json:"pageInfo"`
}

type IndexEntryTgRefsEdge struct {
	Node   *TopicalGuideEntry `json:"node"`
	Cursor string             `json:"cursor"`
}

type IndexEntryTgRefsFieldInput struct {
	Create  []*IndexEntryTgRefsCreateFieldInput  `json:"create,omitempty"`
	Connect []*IndexEntryTgRefsConnectFieldInput `json:"connect,omitempty"`
}

type IndexEntryTgRefsUpdateFieldInput struct {
	Create     []*IndexEntryTgRefsCreateFieldInput      `json:"create,omitempty"`
	Connect    []*IndexEntryTgRefsConnectFieldInput     `json:"connect,omitempty"`
	Disconnect []*IndexEntryTgRefsDisconnectFieldInput  `json:"disconnect,omitempty"`
	Update     []*IndexEntryTgRefsUpdateConnectionInput `json:"update,omitempty"`
	Delete     []*IndexEntryTgRefsDeleteFieldInput      `json:"delete,omitempty"`
}

type IndexEntryTgRefsCreateFieldInput struct {
	Node TopicalGuideEntryCreateInput `json:"node"`
}

type IndexEntryTgRefsConnectFieldInput struct {
	Where TopicalGuideEntryWhere `json:"where"`
}

type IndexEntryTgRefsDisconnectFieldInput struct {
	Where TopicalGuideEntryWhere `json:"where"`
}

type IndexEntryTgRefsUpdateConnectionInput struct {
	Where TopicalGuideEntryWhere        `json:"where"`
	Node  *TopicalGuideEntryUpdateInput `json:"node,omitempty"`
}

type IndexEntryTgRefsDeleteFieldInput struct {
	Where TopicalGuideEntryWhere `json:"where"`
}

type IndexEntryBdRefsConnection struct {
	Edges      []*IndexEntryBdRefsEdge `json:"edges"`
	TotalCount int                     `json:"totalCount"`
	PageInfo   PageInfo                `json:"pageInfo"`
}

type IndexEntryBdRefsEdge struct {
	Node   *BibleDictEntry `json:"node"`
	Cursor string          `json:"cursor"`
}

type IndexEntryBdRefsFieldInput struct {
	Create  []*IndexEntryBdRefsCreateFieldInput  `json:"create,omitempty"`
	Connect []*IndexEntryBdRefsConnectFieldInput `json:"connect,omitempty"`
}

type IndexEntryBdRefsUpdateFieldInput struct {
	Create     []*IndexEntryBdRefsCreateFieldInput      `json:"create,omitempty"`
	Connect    []*IndexEntryBdRefsConnectFieldInput     `json:"connect,omitempty"`
	Disconnect []*IndexEntryBdRefsDisconnectFieldInput  `json:"disconnect,omitempty"`
	Update     []*IndexEntryBdRefsUpdateConnectionInput `json:"update,omitempty"`
	Delete     []*IndexEntryBdRefsDeleteFieldInput      `json:"delete,omitempty"`
}

type IndexEntryBdRefsCreateFieldInput struct {
	Node BibleDictEntryCreateInput `json:"node"`
}

type IndexEntryBdRefsConnectFieldInput struct {
	Where BibleDictEntryWhere `json:"where"`
}

type IndexEntryBdRefsDisconnectFieldInput struct {
	Where BibleDictEntryWhere `json:"where"`
}

type IndexEntryBdRefsUpdateConnectionInput struct {
	Where BibleDictEntryWhere        `json:"where"`
	Node  *BibleDictEntryUpdateInput `json:"node,omitempty"`
}

type IndexEntryBdRefsDeleteFieldInput struct {
	Where BibleDictEntryWhere `json:"where"`
}

type IndexEntryVerseRefsConnection struct {
	Edges      []*IndexEntryVerseRefsEdge `json:"edges"`
	TotalCount int                        `json:"totalCount"`
	PageInfo   PageInfo                   `json:"pageInfo"`
}

type IndexEntryVerseRefsEdge struct {
	Node       *Verse            `json:"node"`
	Cursor     string            `json:"cursor"`
	Properties *IDXVerseRefProps `json:"properties,omitempty"`
}

type IndexEntryVerseRefsFieldInput struct {
	Create  []*IndexEntryVerseRefsCreateFieldInput  `json:"create,omitempty"`
	Connect []*IndexEntryVerseRefsConnectFieldInput `json:"connect,omitempty"`
}

type IndexEntryVerseRefsUpdateFieldInput struct {
	Create     []*IndexEntryVerseRefsCreateFieldInput      `json:"create,omitempty"`
	Connect    []*IndexEntryVerseRefsConnectFieldInput     `json:"connect,omitempty"`
	Disconnect []*IndexEntryVerseRefsDisconnectFieldInput  `json:"disconnect,omitempty"`
	Update     []*IndexEntryVerseRefsUpdateConnectionInput `json:"update,omitempty"`
	Delete     []*IndexEntryVerseRefsDeleteFieldInput      `json:"delete,omitempty"`
}

type IndexEntryVerseRefsCreateFieldInput struct {
	Node VerseCreateInput             `json:"node"`
	Edge *IDXVerseRefPropsCreateInput `json:"edge,omitempty"`
}

type IndexEntryVerseRefsConnectFieldInput struct {
	Where VerseWhere                   `json:"where"`
	Edge  *IDXVerseRefPropsCreateInput `json:"edge,omitempty"`
}

type IndexEntryVerseRefsDisconnectFieldInput struct {
	Where VerseWhere `json:"where"`
}

type IndexEntryVerseRefsUpdateConnectionInput struct {
	Where VerseWhere                   `json:"where"`
	Node  *VerseUpdateInput            `json:"node,omitempty"`
	Edge  *IDXVerseRefPropsUpdateInput `json:"edge,omitempty"`
}

type IndexEntryVerseRefsDeleteFieldInput struct {
	Where VerseWhere `json:"where"`
}

type JSTPassage struct {
	Id            string    `json:"id"`
	Book          string    `json:"book"`
	Chapter       string    `json:"chapter"`
	Comprises     string    `json:"comprises"`
	CompareRef    *string   `json:"compareRef,omitempty"`
	Summary       *string   `json:"summary,omitempty"`
	Text          string    `json:"text"`
	Embedding     []float64 `json:"embedding"`
	CompareVerses []*Verse  `json:"compareVerses,omitempty"`
}

type JSTPassageCreateInput struct {
	Id         *string   `json:"id,omitempty"`
	Book       string    `json:"book"`
	Chapter    string    `json:"chapter"`
	Comprises  string    `json:"comprises"`
	CompareRef *string   `json:"compareRef,omitempty"`
	Summary    *string   `json:"summary,omitempty"`
	Text       string    `json:"text"`
	Embedding  []float64 `json:"embedding"`
}

type JSTPassageUpdateInput struct {
	Book       *string   `json:"book,omitempty"`
	Chapter    *string   `json:"chapter,omitempty"`
	Comprises  *string   `json:"comprises,omitempty"`
	CompareRef *string   `json:"compareRef,omitempty"`
	Summary    *string   `json:"summary,omitempty"`
	Text       *string   `json:"text,omitempty"`
	Embedding  []float64 `json:"embedding,omitempty"`
}

type JSTPassageWhere struct {
	Id                   *string            `json:"id,omitempty"`
	IdNot                *string            `json:"id_NOT,omitempty"`
	IdIn                 []string           `json:"id_IN,omitempty"`
	IdNotIn              []string           `json:"id_NOT_IN,omitempty"`
	IdContains           *string            `json:"id_CONTAINS,omitempty"`
	IdStartsWith         *string            `json:"id_STARTS_WITH,omitempty"`
	IdEndsWith           *string            `json:"id_ENDS_WITH,omitempty"`
	Book                 *string            `json:"book,omitempty"`
	BookNot              *string            `json:"book_NOT,omitempty"`
	BookIn               []string           `json:"book_IN,omitempty"`
	BookNotIn            []string           `json:"book_NOT_IN,omitempty"`
	BookGt               *string            `json:"book_GT,omitempty"`
	BookGte              *string            `json:"book_GTE,omitempty"`
	BookLt               *string            `json:"book_LT,omitempty"`
	BookLte              *string            `json:"book_LTE,omitempty"`
	BookContains         *string            `json:"book_CONTAINS,omitempty"`
	BookStartsWith       *string            `json:"book_STARTS_WITH,omitempty"`
	BookEndsWith         *string            `json:"book_ENDS_WITH,omitempty"`
	Chapter              *string            `json:"chapter,omitempty"`
	ChapterNot           *string            `json:"chapter_NOT,omitempty"`
	ChapterIn            []string           `json:"chapter_IN,omitempty"`
	ChapterNotIn         []string           `json:"chapter_NOT_IN,omitempty"`
	ChapterGt            *string            `json:"chapter_GT,omitempty"`
	ChapterGte           *string            `json:"chapter_GTE,omitempty"`
	ChapterLt            *string            `json:"chapter_LT,omitempty"`
	ChapterLte           *string            `json:"chapter_LTE,omitempty"`
	ChapterContains      *string            `json:"chapter_CONTAINS,omitempty"`
	ChapterStartsWith    *string            `json:"chapter_STARTS_WITH,omitempty"`
	ChapterEndsWith      *string            `json:"chapter_ENDS_WITH,omitempty"`
	Comprises            *string            `json:"comprises,omitempty"`
	ComprisesNot         *string            `json:"comprises_NOT,omitempty"`
	ComprisesIn          []string           `json:"comprises_IN,omitempty"`
	ComprisesNotIn       []string           `json:"comprises_NOT_IN,omitempty"`
	ComprisesGt          *string            `json:"comprises_GT,omitempty"`
	ComprisesGte         *string            `json:"comprises_GTE,omitempty"`
	ComprisesLt          *string            `json:"comprises_LT,omitempty"`
	ComprisesLte         *string            `json:"comprises_LTE,omitempty"`
	ComprisesContains    *string            `json:"comprises_CONTAINS,omitempty"`
	ComprisesStartsWith  *string            `json:"comprises_STARTS_WITH,omitempty"`
	ComprisesEndsWith    *string            `json:"comprises_ENDS_WITH,omitempty"`
	CompareRef           *string            `json:"compareRef,omitempty"`
	CompareRefNot        *string            `json:"compareRef_NOT,omitempty"`
	CompareRefIn         []string           `json:"compareRef_IN,omitempty"`
	CompareRefNotIn      []string           `json:"compareRef_NOT_IN,omitempty"`
	CompareRefGt         *string            `json:"compareRef_GT,omitempty"`
	CompareRefGte        *string            `json:"compareRef_GTE,omitempty"`
	CompareRefLt         *string            `json:"compareRef_LT,omitempty"`
	CompareRefLte        *string            `json:"compareRef_LTE,omitempty"`
	CompareRefContains   *string            `json:"compareRef_CONTAINS,omitempty"`
	CompareRefStartsWith *string            `json:"compareRef_STARTS_WITH,omitempty"`
	CompareRefEndsWith   *string            `json:"compareRef_ENDS_WITH,omitempty"`
	Summary              *string            `json:"summary,omitempty"`
	SummaryNot           *string            `json:"summary_NOT,omitempty"`
	SummaryIn            []string           `json:"summary_IN,omitempty"`
	SummaryNotIn         []string           `json:"summary_NOT_IN,omitempty"`
	SummaryGt            *string            `json:"summary_GT,omitempty"`
	SummaryGte           *string            `json:"summary_GTE,omitempty"`
	SummaryLt            *string            `json:"summary_LT,omitempty"`
	SummaryLte           *string            `json:"summary_LTE,omitempty"`
	SummaryContains      *string            `json:"summary_CONTAINS,omitempty"`
	SummaryStartsWith    *string            `json:"summary_STARTS_WITH,omitempty"`
	SummaryEndsWith      *string            `json:"summary_ENDS_WITH,omitempty"`
	Text                 *string            `json:"text,omitempty"`
	TextNot              *string            `json:"text_NOT,omitempty"`
	TextIn               []string           `json:"text_IN,omitempty"`
	TextNotIn            []string           `json:"text_NOT_IN,omitempty"`
	TextGt               *string            `json:"text_GT,omitempty"`
	TextGte              *string            `json:"text_GTE,omitempty"`
	TextLt               *string            `json:"text_LT,omitempty"`
	TextLte              *string            `json:"text_LTE,omitempty"`
	TextContains         *string            `json:"text_CONTAINS,omitempty"`
	TextStartsWith       *string            `json:"text_STARTS_WITH,omitempty"`
	TextEndsWith         *string            `json:"text_ENDS_WITH,omitempty"`
	Embedding            *[]float64         `json:"embedding,omitempty"`
	EmbeddingNot         *[]float64         `json:"embedding_NOT,omitempty"`
	EmbeddingIn          [][]float64        `json:"embedding_IN,omitempty"`
	EmbeddingNotIn       [][]float64        `json:"embedding_NOT_IN,omitempty"`
	EmbeddingGt          *[]float64         `json:"embedding_GT,omitempty"`
	EmbeddingGte         *[]float64         `json:"embedding_GTE,omitempty"`
	EmbeddingLt          *[]float64         `json:"embedding_LT,omitempty"`
	EmbeddingLte         *[]float64         `json:"embedding_LTE,omitempty"`
	AND                  []*JSTPassageWhere `json:"AND,omitempty"`
	OR                   []*JSTPassageWhere `json:"OR,omitempty"`
	NOT                  *JSTPassageWhere   `json:"NOT,omitempty"`
}

type JSTPassageSort struct {
	Id         *SortDirection `json:"id,omitempty"`
	Book       *SortDirection `json:"book,omitempty"`
	Chapter    *SortDirection `json:"chapter,omitempty"`
	Comprises  *SortDirection `json:"comprises,omitempty"`
	CompareRef *SortDirection `json:"compareRef,omitempty"`
	Summary    *SortDirection `json:"summary,omitempty"`
	Text       *SortDirection `json:"text,omitempty"`
	Embedding  *SortDirection `json:"embedding,omitempty"`
}

type JSTPassagesConnection struct {
	Edges      []*JSTPassageEdge `json:"edges"`
	TotalCount int               `json:"totalCount"`
	PageInfo   PageInfo          `json:"pageInfo"`
}

type JSTPassageEdge struct {
	Node   *JSTPassage `json:"node"`
	Cursor string      `json:"cursor"`
}

type CreateJSTPassagesMutationResponse struct {
	JSTPassages []*JSTPassage `json:"jSTPassages"`
}

type UpdateJSTPassagesMutationResponse struct {
	JSTPassages []*JSTPassage `json:"jSTPassages"`
}

type JSTPassageSimilarResult struct {
	Score float64     `json:"score"`
	Node  *JSTPassage `json:"node"`
}

type JSTPassageCompareVersesConnection struct {
	Edges      []*JSTPassageCompareVersesEdge `json:"edges"`
	TotalCount int                            `json:"totalCount"`
	PageInfo   PageInfo                       `json:"pageInfo"`
}

type JSTPassageCompareVersesEdge struct {
	Node   *Verse `json:"node"`
	Cursor string `json:"cursor"`
}

type JSTPassageCompareVersesFieldInput struct {
	Create  []*JSTPassageCompareVersesCreateFieldInput  `json:"create,omitempty"`
	Connect []*JSTPassageCompareVersesConnectFieldInput `json:"connect,omitempty"`
}

type JSTPassageCompareVersesUpdateFieldInput struct {
	Create     []*JSTPassageCompareVersesCreateFieldInput      `json:"create,omitempty"`
	Connect    []*JSTPassageCompareVersesConnectFieldInput     `json:"connect,omitempty"`
	Disconnect []*JSTPassageCompareVersesDisconnectFieldInput  `json:"disconnect,omitempty"`
	Update     []*JSTPassageCompareVersesUpdateConnectionInput `json:"update,omitempty"`
	Delete     []*JSTPassageCompareVersesDeleteFieldInput      `json:"delete,omitempty"`
}

type JSTPassageCompareVersesCreateFieldInput struct {
	Node VerseCreateInput `json:"node"`
}

type JSTPassageCompareVersesConnectFieldInput struct {
	Where VerseWhere `json:"where"`
}

type JSTPassageCompareVersesDisconnectFieldInput struct {
	Where VerseWhere `json:"where"`
}

type JSTPassageCompareVersesUpdateConnectionInput struct {
	Where VerseWhere        `json:"where"`
	Node  *VerseUpdateInput `json:"node,omitempty"`
}

type JSTPassageCompareVersesDeleteFieldInput struct {
	Where VerseWhere `json:"where"`
}

type TopicalGuideEntry struct {
	Id        string               `json:"id"`
	Name      string               `json:"name"`
	Embedding []float64            `json:"embedding"`
	SeeAlso   []*TopicalGuideEntry `json:"seeAlso,omitempty"`
	BdRefs    []*BibleDictEntry    `json:"bdRefs,omitempty"`
	VerseRefs []*Verse             `json:"verseRefs,omitempty"`
}

type TopicalGuideEntryCreateInput struct {
	Id        *string   `json:"id,omitempty"`
	Name      string    `json:"name"`
	Embedding []float64 `json:"embedding"`
}

type TopicalGuideEntryUpdateInput struct {
	Name      *string   `json:"name,omitempty"`
	Embedding []float64 `json:"embedding,omitempty"`
}

type TopicalGuideEntryWhere struct {
	Id             *string                   `json:"id,omitempty"`
	IdNot          *string                   `json:"id_NOT,omitempty"`
	IdIn           []string                  `json:"id_IN,omitempty"`
	IdNotIn        []string                  `json:"id_NOT_IN,omitempty"`
	IdContains     *string                   `json:"id_CONTAINS,omitempty"`
	IdStartsWith   *string                   `json:"id_STARTS_WITH,omitempty"`
	IdEndsWith     *string                   `json:"id_ENDS_WITH,omitempty"`
	Name           *string                   `json:"name,omitempty"`
	NameNot        *string                   `json:"name_NOT,omitempty"`
	NameIn         []string                  `json:"name_IN,omitempty"`
	NameNotIn      []string                  `json:"name_NOT_IN,omitempty"`
	NameGt         *string                   `json:"name_GT,omitempty"`
	NameGte        *string                   `json:"name_GTE,omitempty"`
	NameLt         *string                   `json:"name_LT,omitempty"`
	NameLte        *string                   `json:"name_LTE,omitempty"`
	NameContains   *string                   `json:"name_CONTAINS,omitempty"`
	NameStartsWith *string                   `json:"name_STARTS_WITH,omitempty"`
	NameEndsWith   *string                   `json:"name_ENDS_WITH,omitempty"`
	Embedding      *[]float64                `json:"embedding,omitempty"`
	EmbeddingNot   *[]float64                `json:"embedding_NOT,omitempty"`
	EmbeddingIn    [][]float64               `json:"embedding_IN,omitempty"`
	EmbeddingNotIn [][]float64               `json:"embedding_NOT_IN,omitempty"`
	EmbeddingGt    *[]float64                `json:"embedding_GT,omitempty"`
	EmbeddingGte   *[]float64                `json:"embedding_GTE,omitempty"`
	EmbeddingLt    *[]float64                `json:"embedding_LT,omitempty"`
	EmbeddingLte   *[]float64                `json:"embedding_LTE,omitempty"`
	AND            []*TopicalGuideEntryWhere `json:"AND,omitempty"`
	OR             []*TopicalGuideEntryWhere `json:"OR,omitempty"`
	NOT            *TopicalGuideEntryWhere   `json:"NOT,omitempty"`
}

type TopicalGuideEntrySort struct {
	Id        *SortDirection `json:"id,omitempty"`
	Name      *SortDirection `json:"name,omitempty"`
	Embedding *SortDirection `json:"embedding,omitempty"`
}

type TopicalGuideEntrysConnection struct {
	Edges      []*TopicalGuideEntryEdge `json:"edges"`
	TotalCount int                      `json:"totalCount"`
	PageInfo   PageInfo                 `json:"pageInfo"`
}

type TopicalGuideEntryEdge struct {
	Node   *TopicalGuideEntry `json:"node"`
	Cursor string             `json:"cursor"`
}

type CreateTopicalGuideEntrysMutationResponse struct {
	TopicalGuideEntrys []*TopicalGuideEntry `json:"topicalGuideEntries"`
}

type UpdateTopicalGuideEntrysMutationResponse struct {
	TopicalGuideEntrys []*TopicalGuideEntry `json:"topicalGuideEntries"`
}

type TopicalGuideEntrySimilarResult struct {
	Score float64            `json:"score"`
	Node  *TopicalGuideEntry `json:"node"`
}

type TopicalGuideEntrySeeAlsoConnection struct {
	Edges      []*TopicalGuideEntrySeeAlsoEdge `json:"edges"`
	TotalCount int                             `json:"totalCount"`
	PageInfo   PageInfo                        `json:"pageInfo"`
}

type TopicalGuideEntrySeeAlsoEdge struct {
	Node   *TopicalGuideEntry `json:"node"`
	Cursor string             `json:"cursor"`
}

type TopicalGuideEntrySeeAlsoFieldInput struct {
	Create  []*TopicalGuideEntrySeeAlsoCreateFieldInput  `json:"create,omitempty"`
	Connect []*TopicalGuideEntrySeeAlsoConnectFieldInput `json:"connect,omitempty"`
}

type TopicalGuideEntrySeeAlsoUpdateFieldInput struct {
	Create     []*TopicalGuideEntrySeeAlsoCreateFieldInput      `json:"create,omitempty"`
	Connect    []*TopicalGuideEntrySeeAlsoConnectFieldInput     `json:"connect,omitempty"`
	Disconnect []*TopicalGuideEntrySeeAlsoDisconnectFieldInput  `json:"disconnect,omitempty"`
	Update     []*TopicalGuideEntrySeeAlsoUpdateConnectionInput `json:"update,omitempty"`
	Delete     []*TopicalGuideEntrySeeAlsoDeleteFieldInput      `json:"delete,omitempty"`
}

type TopicalGuideEntrySeeAlsoCreateFieldInput struct {
	Node TopicalGuideEntryCreateInput `json:"node"`
}

type TopicalGuideEntrySeeAlsoConnectFieldInput struct {
	Where TopicalGuideEntryWhere `json:"where"`
}

type TopicalGuideEntrySeeAlsoDisconnectFieldInput struct {
	Where TopicalGuideEntryWhere `json:"where"`
}

type TopicalGuideEntrySeeAlsoUpdateConnectionInput struct {
	Where TopicalGuideEntryWhere        `json:"where"`
	Node  *TopicalGuideEntryUpdateInput `json:"node,omitempty"`
}

type TopicalGuideEntrySeeAlsoDeleteFieldInput struct {
	Where TopicalGuideEntryWhere `json:"where"`
}

type TopicalGuideEntryBdRefsConnection struct {
	Edges      []*TopicalGuideEntryBdRefsEdge `json:"edges"`
	TotalCount int                            `json:"totalCount"`
	PageInfo   PageInfo                       `json:"pageInfo"`
}

type TopicalGuideEntryBdRefsEdge struct {
	Node   *BibleDictEntry `json:"node"`
	Cursor string          `json:"cursor"`
}

type TopicalGuideEntryBdRefsFieldInput struct {
	Create  []*TopicalGuideEntryBdRefsCreateFieldInput  `json:"create,omitempty"`
	Connect []*TopicalGuideEntryBdRefsConnectFieldInput `json:"connect,omitempty"`
}

type TopicalGuideEntryBdRefsUpdateFieldInput struct {
	Create     []*TopicalGuideEntryBdRefsCreateFieldInput      `json:"create,omitempty"`
	Connect    []*TopicalGuideEntryBdRefsConnectFieldInput     `json:"connect,omitempty"`
	Disconnect []*TopicalGuideEntryBdRefsDisconnectFieldInput  `json:"disconnect,omitempty"`
	Update     []*TopicalGuideEntryBdRefsUpdateConnectionInput `json:"update,omitempty"`
	Delete     []*TopicalGuideEntryBdRefsDeleteFieldInput      `json:"delete,omitempty"`
}

type TopicalGuideEntryBdRefsCreateFieldInput struct {
	Node BibleDictEntryCreateInput `json:"node"`
}

type TopicalGuideEntryBdRefsConnectFieldInput struct {
	Where BibleDictEntryWhere `json:"where"`
}

type TopicalGuideEntryBdRefsDisconnectFieldInput struct {
	Where BibleDictEntryWhere `json:"where"`
}

type TopicalGuideEntryBdRefsUpdateConnectionInput struct {
	Where BibleDictEntryWhere        `json:"where"`
	Node  *BibleDictEntryUpdateInput `json:"node,omitempty"`
}

type TopicalGuideEntryBdRefsDeleteFieldInput struct {
	Where BibleDictEntryWhere `json:"where"`
}

type TopicalGuideEntryVerseRefsConnection struct {
	Edges      []*TopicalGuideEntryVerseRefsEdge `json:"edges"`
	TotalCount int                               `json:"totalCount"`
	PageInfo   PageInfo                          `json:"pageInfo"`
}

type TopicalGuideEntryVerseRefsEdge struct {
	Node       *Verse           `json:"node"`
	Cursor     string           `json:"cursor"`
	Properties *TGVerseRefProps `json:"properties,omitempty"`
}

type TopicalGuideEntryVerseRefsFieldInput struct {
	Create  []*TopicalGuideEntryVerseRefsCreateFieldInput  `json:"create,omitempty"`
	Connect []*TopicalGuideEntryVerseRefsConnectFieldInput `json:"connect,omitempty"`
}

type TopicalGuideEntryVerseRefsUpdateFieldInput struct {
	Create     []*TopicalGuideEntryVerseRefsCreateFieldInput      `json:"create,omitempty"`
	Connect    []*TopicalGuideEntryVerseRefsConnectFieldInput     `json:"connect,omitempty"`
	Disconnect []*TopicalGuideEntryVerseRefsDisconnectFieldInput  `json:"disconnect,omitempty"`
	Update     []*TopicalGuideEntryVerseRefsUpdateConnectionInput `json:"update,omitempty"`
	Delete     []*TopicalGuideEntryVerseRefsDeleteFieldInput      `json:"delete,omitempty"`
}

type TopicalGuideEntryVerseRefsCreateFieldInput struct {
	Node VerseCreateInput            `json:"node"`
	Edge *TGVerseRefPropsCreateInput `json:"edge,omitempty"`
}

type TopicalGuideEntryVerseRefsConnectFieldInput struct {
	Where VerseWhere                  `json:"where"`
	Edge  *TGVerseRefPropsCreateInput `json:"edge,omitempty"`
}

type TopicalGuideEntryVerseRefsDisconnectFieldInput struct {
	Where VerseWhere `json:"where"`
}

type TopicalGuideEntryVerseRefsUpdateConnectionInput struct {
	Where VerseWhere                  `json:"where"`
	Node  *VerseUpdateInput           `json:"node,omitempty"`
	Edge  *TGVerseRefPropsUpdateInput `json:"edge,omitempty"`
}

type TopicalGuideEntryVerseRefsDeleteFieldInput struct {
	Where VerseWhere `json:"where"`
}

type VolumeMatchInput struct {
	Name         *string `json:"name,omitempty"`
	Abbreviation *string `json:"abbreviation,omitempty"`
}

type VolumeMergeInput struct {
	Match    *VolumeMatchInput  `json:"match"`
	OnCreate *VolumeCreateInput `json:"onCreate,omitempty"`
	OnMatch  *VolumeUpdateInput `json:"onMatch,omitempty"`
}

type MergeVolumesMutationResponse struct {
	Volumes []*Volume `json:"volumes"`
}

type BibleDictEntryMatchInput struct {
	Name *string `json:"name,omitempty"`
	Text *string `json:"text,omitempty"`
}

type BibleDictEntryMergeInput struct {
	Match    *BibleDictEntryMatchInput  `json:"match"`
	OnCreate *BibleDictEntryCreateInput `json:"onCreate,omitempty"`
	OnMatch  *BibleDictEntryUpdateInput `json:"onMatch,omitempty"`
}

type MergeBibleDictEntrysMutationResponse struct {
	BibleDictEntrys []*BibleDictEntry `json:"bibleDictEntries"`
}

type ChapterMatchInput struct {
	Number  *int    `json:"number,omitempty"`
	Summary *string `json:"summary,omitempty"`
	Url     *string `json:"url,omitempty"`
}

type ChapterMergeInput struct {
	Match    *ChapterMatchInput  `json:"match"`
	OnCreate *ChapterCreateInput `json:"onCreate,omitempty"`
	OnMatch  *ChapterUpdateInput `json:"onMatch,omitempty"`
}

type MergeChaptersMutationResponse struct {
	Chapters []*Chapter `json:"chapters"`
}

type VerseGroupMatchInput struct {
	Text             *string `json:"text,omitempty"`
	StartVerseNumber *int    `json:"startVerseNumber,omitempty"`
	EndVerseNumber   *int    `json:"endVerseNumber,omitempty"`
}

type VerseGroupMergeInput struct {
	Match    *VerseGroupMatchInput  `json:"match"`
	OnCreate *VerseGroupCreateInput `json:"onCreate,omitempty"`
	OnMatch  *VerseGroupUpdateInput `json:"onMatch,omitempty"`
}

type MergeVerseGroupsMutationResponse struct {
	VerseGroups []*VerseGroup `json:"verseGroups"`
}

type BookMatchInput struct {
	Name    *string `json:"name,omitempty"`
	Slug    *string `json:"slug,omitempty"`
	UrlPath *string `json:"urlPath,omitempty"`
}

type BookMergeInput struct {
	Match    *BookMatchInput  `json:"match"`
	OnCreate *BookCreateInput `json:"onCreate,omitempty"`
	OnMatch  *BookUpdateInput `json:"onMatch,omitempty"`
}

type MergeBooksMutationResponse struct {
	Books []*Book `json:"books"`
}

type VerseMatchInput struct {
	Number            *int    `json:"number,omitempty"`
	Reference         *string `json:"reference,omitempty"`
	Text              *string `json:"text,omitempty"`
	TranslationNotes  *string `json:"translationNotes,omitempty"`
	AlternateReadings *string `json:"alternateReadings,omitempty"`
	ExplanatoryNotes  *string `json:"explanatoryNotes,omitempty"`
}

type VerseMergeInput struct {
	Match    *VerseMatchInput  `json:"match"`
	OnCreate *VerseCreateInput `json:"onCreate,omitempty"`
	OnMatch  *VerseUpdateInput `json:"onMatch,omitempty"`
}

type MergeVersesMutationResponse struct {
	Verses []*Verse `json:"verses"`
}

type IndexEntryMatchInput struct {
	Name *string `json:"name,omitempty"`
}

type IndexEntryMergeInput struct {
	Match    *IndexEntryMatchInput  `json:"match"`
	OnCreate *IndexEntryCreateInput `json:"onCreate,omitempty"`
	OnMatch  *IndexEntryUpdateInput `json:"onMatch,omitempty"`
}

type MergeIndexEntrysMutationResponse struct {
	IndexEntrys []*IndexEntry `json:"indexEntries"`
}

type JSTPassageMatchInput struct {
	Book       *string `json:"book,omitempty"`
	Chapter    *string `json:"chapter,omitempty"`
	Comprises  *string `json:"comprises,omitempty"`
	CompareRef *string `json:"compareRef,omitempty"`
	Summary    *string `json:"summary,omitempty"`
	Text       *string `json:"text,omitempty"`
}

type JSTPassageMergeInput struct {
	Match    *JSTPassageMatchInput  `json:"match"`
	OnCreate *JSTPassageCreateInput `json:"onCreate,omitempty"`
	OnMatch  *JSTPassageUpdateInput `json:"onMatch,omitempty"`
}

type MergeJSTPassagesMutationResponse struct {
	JSTPassages []*JSTPassage `json:"jSTPassages"`
}

type TopicalGuideEntryMatchInput struct {
	Name *string `json:"name,omitempty"`
}

type TopicalGuideEntryMergeInput struct {
	Match    *TopicalGuideEntryMatchInput  `json:"match"`
	OnCreate *TopicalGuideEntryCreateInput `json:"onCreate,omitempty"`
	OnMatch  *TopicalGuideEntryUpdateInput `json:"onMatch,omitempty"`
}

type MergeTopicalGuideEntrysMutationResponse struct {
	TopicalGuideEntrys []*TopicalGuideEntry `json:"topicalGuideEntries"`
}

type ConnectVolumeBooksInput struct {
	From *VolumeWhere `json:"from"`
	To   *BookWhere   `json:"to"`
}

type ConnectInfo struct {
	RelationshipsCreated int `json:"relationshipsCreated"`
}

type ConnectBibleDictEntrySeeAlsoInput struct {
	From *BibleDictEntryWhere `json:"from"`
	To   *BibleDictEntryWhere `json:"to"`
}

type ConnectBibleDictEntryVerseRefsInput struct {
	From *BibleDictEntryWhere        `json:"from"`
	To   *VerseWhere                 `json:"to"`
	Edge *BDVerseRefPropsCreateInput `json:"edge,omitempty"`
}

type ConnectChapterBookInput struct {
	From *ChapterWhere `json:"from"`
	To   *BookWhere    `json:"to"`
}

type ConnectChapterVersesInput struct {
	From *ChapterWhere `json:"from"`
	To   *VerseWhere   `json:"to"`
}

type ConnectChapterVerseGroupsInput struct {
	From *ChapterWhere    `json:"from"`
	To   *VerseGroupWhere `json:"to"`
}

type ConnectVerseGroupChapterInput struct {
	From *VerseGroupWhere `json:"from"`
	To   *ChapterWhere    `json:"to"`
}

type ConnectVerseGroupVersesInput struct {
	From *VerseGroupWhere `json:"from"`
	To   *VerseWhere      `json:"to"`
}

type ConnectBookVolumeInput struct {
	From *BookWhere   `json:"from"`
	To   *VolumeWhere `json:"to"`
}

type ConnectBookChaptersInput struct {
	From *BookWhere    `json:"from"`
	To   *ChapterWhere `json:"to"`
}

type ConnectVerseChapterInput struct {
	From *VerseWhere   `json:"from"`
	To   *ChapterWhere `json:"to"`
}

type ConnectVerseCrossRefsOutInput struct {
	From *VerseWhere                    `json:"from"`
	To   *VerseWhere                    `json:"to"`
	Edge *VerseCrossRefPropsCreateInput `json:"edge,omitempty"`
}

type ConnectVerseCrossRefsInInput struct {
	From *VerseWhere                    `json:"from"`
	To   *VerseWhere                    `json:"to"`
	Edge *VerseCrossRefPropsCreateInput `json:"edge,omitempty"`
}

type ConnectVerseTgFootnotesInput struct {
	From *VerseWhere                 `json:"from"`
	To   *TopicalGuideEntryWhere     `json:"to"`
	Edge *VerseTGRefPropsCreateInput `json:"edge,omitempty"`
}

type ConnectVerseBdFootnotesInput struct {
	From *VerseWhere                 `json:"from"`
	To   *BibleDictEntryWhere        `json:"to"`
	Edge *VerseBDRefPropsCreateInput `json:"edge,omitempty"`
}

type ConnectVerseJstFootnotesInput struct {
	From *VerseWhere                  `json:"from"`
	To   *JSTPassageWhere             `json:"to"`
	Edge *VerseJSTRefPropsCreateInput `json:"edge,omitempty"`
}

type ConnectIndexEntrySeeAlsoInput struct {
	From *IndexEntryWhere `json:"from"`
	To   *IndexEntryWhere `json:"to"`
}

type ConnectIndexEntryTgRefsInput struct {
	From *IndexEntryWhere        `json:"from"`
	To   *TopicalGuideEntryWhere `json:"to"`
}

type ConnectIndexEntryBdRefsInput struct {
	From *IndexEntryWhere     `json:"from"`
	To   *BibleDictEntryWhere `json:"to"`
}

type ConnectIndexEntryVerseRefsInput struct {
	From *IndexEntryWhere             `json:"from"`
	To   *VerseWhere                  `json:"to"`
	Edge *IDXVerseRefPropsCreateInput `json:"edge,omitempty"`
}

type ConnectJSTPassageCompareVersesInput struct {
	From *JSTPassageWhere `json:"from"`
	To   *VerseWhere      `json:"to"`
}

type ConnectTopicalGuideEntrySeeAlsoInput struct {
	From *TopicalGuideEntryWhere `json:"from"`
	To   *TopicalGuideEntryWhere `json:"to"`
}

type ConnectTopicalGuideEntryBdRefsInput struct {
	From *TopicalGuideEntryWhere `json:"from"`
	To   *BibleDictEntryWhere    `json:"to"`
}

type ConnectTopicalGuideEntryVerseRefsInput struct {
	From *TopicalGuideEntryWhere     `json:"from"`
	To   *VerseWhere                 `json:"to"`
	Edge *TGVerseRefPropsCreateInput `json:"edge,omitempty"`
}
