package libsql

import "testing"

func TestEntityVerseConstant(t *testing.T) {
	// EntityVerse must exist as a valid EntityType constant.
	var et EntityType = EntityVerse
	if et != "verse" {
		t.Errorf("EntityVerse = %q, want %q", et, "verse")
	}
}

func TestResultMetaVerseFields(t *testing.T) {
	// ResultMeta must have VerseNumber and Reference fields for EntityVerse results.
	meta := ResultMeta{
		VerseNumber: 7,
		Reference:   "1 Ne. 3:7",
	}

	if meta.VerseNumber != 7 {
		t.Errorf("VerseNumber = %d, want 7", meta.VerseNumber)
	}
	if meta.Reference != "1 Ne. 3:7" {
		t.Errorf("Reference = %q, want %q", meta.Reference, "1 Ne. 3:7")
	}
}

func TestPipelineConstants(t *testing.T) {
	tests := []struct {
		name     string
		got      float64
		expected float64
	}{
		{"defaultHopPenalty", defaultHopPenalty, 0.05},
		{"defaultVerseBonus", defaultVerseBonus, 0.05},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.expected {
				t.Errorf("%s = %f, want %f", tt.name, tt.got, tt.expected)
			}
		})
	}
}

func TestDefaultGraphLimit(t *testing.T) {
	if defaultGraphLimit != 5 {
		t.Errorf("defaultGraphLimit = %d, want 5", defaultGraphLimit)
	}
}
