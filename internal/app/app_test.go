package app

import (
	"context"
	"testing"

	"go.uber.org/mock/gomock"

	"lds-gpt/internal/bedrockembedding/mocks"
	"lds-gpt/internal/falkor"
)

// TestDoContextualSearchForwardsOptions confirms App.DoContextualSearch
// propagates WithKNN down to falkor.DoContextualSearch. We seed 5 VerseGroup
// nodes sharing the query's embedding so Stage 1 returns all 5, then cap
// the result with WithKNN(2) — the response must be ≤ 2.
func TestDoContextualSearchForwardsOptions(t *testing.T) {
	ctrl := gomock.NewController(t)

	// Fixed fractional-valued embedding — needs decimals for
	// falkordb-go's param stringification (see Phase B writeup).
	embed := make([]float64, 1024)
	for i := range embed {
		embed[i] = 0.001
	}
	embed[0] = 0.999

	mockEmbed := mocks.NewMockClient(ctrl)
	mockEmbed.EXPECT().EmbedText(gomock.Any(), "test query").Return(embed, nil)

	client := falkor.StartFalkorContainer(t)
	ctx := context.Background()
	if err := client.Migrate(ctx); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	// Seed 5 VerseGroups with the same embedding as the query so Stage 1
	// returns up to 5. The kNN cap should trim to 2.
	vec := make([]interface{}, len(embed))
	for i, x := range embed {
		vec[i] = x
	}
	for i := 0; i < 5; i++ {
		if _, err := client.Raw().Query(
			`CREATE (:VerseGroup {
			   id: $id, text: 'group',
			   startVerseNumber: $start, endVerseNumber: $end,
			   embedding: vecf32($vec)
			 })`,
			map[string]interface{}{
				"id":    "vg/" + string(rune('a'+i)),
				"start": i*2 + 1,
				"end":   i*2 + 2,
				"vec":   vec,
			}, nil,
		); err != nil {
			t.Fatalf("seed VerseGroup %d: %v", i, err)
		}
	}

	a := NewApp(client, mockEmbed)
	results, err := a.DoContextualSearch(ctx, "test query", falkor.WithKNN(2))
	if err != nil {
		t.Fatalf("DoContextualSearch: %v", err)
	}
	if len(results) > 2 {
		t.Errorf("expected ≤2 results with WithKNN(2), got %d", len(results))
	}
	if len(results) == 0 {
		t.Errorf("expected ≥1 result, got 0 — Stage 1 kNN did not return the seeded groups")
	}
}
