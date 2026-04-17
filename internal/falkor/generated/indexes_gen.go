package generated

import (
	"context"
	"fmt"
	"strings"

	"github.com/tab58/go-ormql/pkg/cypher"
	"github.com/tab58/go-ormql/pkg/driver"
)

// VectorIndexes maps logical index names to their label/property for FalkorDB vector query rewrite.
// FalkorDB does not support named vector indexes; these keys are logical identifiers
// used by the driver for query rewriting, not actual database index names.
var VectorIndexes = map[string]driver.VectorIndex{
	"verse_group_embedding":     {Label: "VerseGroup", Property: "embedding"},
	"jst_embedding":             {Label: "JSTPassage", Property: "embedding"},
	"chapter_summary_embedding": {Label: "Chapter", Property: "summaryEmbedding"},
	"bd_embedding":              {Label: "BibleDictEntry", Property: "embedding"},
	"idx_embedding":             {Label: "IndexEntry", Property: "embedding"},
	"tg_embedding":              {Label: "TopicalGuideEntry", Property: "embedding"},
}

// isAlreadyIndexed returns true if the error indicates the attribute is already indexed.
func isAlreadyIndexed(err error) bool {
	return err != nil && strings.Contains(err.Error(), "already indexed")
}

// CreateIndexes creates vector indexes for nodes with @vector directives.
func CreateIndexes(ctx context.Context, drv driver.Driver) error {
	if _, err := drv.ExecuteWrite(ctx, cypher.Statement{Query: "CREATE VECTOR INDEX FOR (n:VerseGroup) ON (n.embedding) OPTIONS {dimension: 1024, similarityFunction: 'cosine'}"}); err != nil && !isAlreadyIndexed(err) {
		return fmt.Errorf("failed to create vector index verse_group_embedding: %w", err)
	}
	if _, err := drv.ExecuteWrite(ctx, cypher.Statement{Query: "CREATE VECTOR INDEX FOR (n:JSTPassage) ON (n.embedding) OPTIONS {dimension: 1024, similarityFunction: 'cosine'}"}); err != nil && !isAlreadyIndexed(err) {
		return fmt.Errorf("failed to create vector index jst_embedding: %w", err)
	}
	if _, err := drv.ExecuteWrite(ctx, cypher.Statement{Query: "CREATE VECTOR INDEX FOR (n:Chapter) ON (n.summaryEmbedding) OPTIONS {dimension: 1024, similarityFunction: 'cosine'}"}); err != nil && !isAlreadyIndexed(err) {
		return fmt.Errorf("failed to create vector index chapter_summary_embedding: %w", err)
	}
	if _, err := drv.ExecuteWrite(ctx, cypher.Statement{Query: "CREATE VECTOR INDEX FOR (n:BibleDictEntry) ON (n.embedding) OPTIONS {dimension: 1024, similarityFunction: 'cosine'}"}); err != nil && !isAlreadyIndexed(err) {
		return fmt.Errorf("failed to create vector index bd_embedding: %w", err)
	}
	if _, err := drv.ExecuteWrite(ctx, cypher.Statement{Query: "CREATE VECTOR INDEX FOR (n:IndexEntry) ON (n.embedding) OPTIONS {dimension: 1024, similarityFunction: 'cosine'}"}); err != nil && !isAlreadyIndexed(err) {
		return fmt.Errorf("failed to create vector index idx_embedding: %w", err)
	}
	if _, err := drv.ExecuteWrite(ctx, cypher.Statement{Query: "CREATE VECTOR INDEX FOR (n:TopicalGuideEntry) ON (n.embedding) OPTIONS {dimension: 1024, similarityFunction: 'cosine'}"}); err != nil && !isAlreadyIndexed(err) {
		return fmt.Errorf("failed to create vector index tg_embedding: %w", err)
	}
	return nil
}
