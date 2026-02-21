package app

import (
	"context"
	"testing"

	"go.uber.org/mock/gomock"

	"lds-gpt/internal/bedrockembedding/mocks"
	"lds-gpt/internal/libsql"
	"lds-gpt/internal/utils/vec"
)

func TestDoContextualSearchForwardsOptions(t *testing.T) {
	ctrl := gomock.NewController(t)

	// Set up a mock embedding client that returns a fixed embedding.
	mockEmbed := mocks.NewMockClient(ctrl)
	floats := make([]float64, 1024)
	floats[0] = 1.0
	mockEmbed.EXPECT().
		EmbedText(gomock.Any(), "test query").
		Return(floats, nil)

	// Create a real libsql client with test data.
	client := libsql.TestClient(t)
	ctx := context.Background()
	ec := client.Ent()

	vol, err := ec.Volume.Create().
		SetName("Test").
		SetAbbreviation("test").
		Save(ctx)
	if err != nil {
		t.Fatalf("creating volume: %v", err)
	}

	book, err := ec.Book.Create().
		SetName("Test Book").
		SetSlug("test").
		SetURLPath("test/test").
		SetVolume(vol).
		Save(ctx)
	if err != nil {
		t.Fatalf("creating book: %v", err)
	}

	ch, err := ec.Chapter.Create().
		SetNumber(1).
		SetBook(book).
		Save(ctx)
	if err != nil {
		t.Fatalf("creating chapter: %v", err)
	}

	// Create 5 verse groups, all with the same embedding as the query.
	queryEmbedding := vec.Float64sToFloat32Bytes(floats)
	for i := 0; i < 5; i++ {
		_, err := ec.VerseGroup.Create().
			SetText("verse group text").
			SetStartVerseNumber(i*2 + 1).
			SetEndVerseNumber(i*2 + 2).
			SetChapter(ch).
			SetEmbedding(queryEmbedding).
			Save(ctx)
		if err != nil {
			t.Fatalf("creating verse group %d: %v", i, err)
		}
	}

	a := NewApp(client, mockEmbed)

	// Search with kNN=2. If options are forwarded, we should get at most 2 results.
	results, err := a.DoContextualSearch(ctx, "test query", libsql.WithKNN(2))
	if err != nil {
		t.Fatalf("DoContextualSearch: %v", err)
	}

	if len(results) > 2 {
		t.Errorf("expected at most 2 results with WithKNN(2), got %d (options not forwarded)", len(results))
	}
}
