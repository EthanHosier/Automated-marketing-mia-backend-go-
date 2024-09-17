package storage

import (
	"encoding/json"
	"errors"
	"testing"
)

func TestExtractAndRemoveSimilarity(t *testing.T) {
	type Example struct {
		ID          int
		Similarity  float64
		Description string
	}

	type ExampleWithoutSimilarity struct {
		ID          int
		Description string
	}

	tests := []struct {
		name           string
		input          interface{}
		expectedSim    float64
		expectedErr    error
		expectedStruct interface{}
	}{
		{
			name: "Successful extraction and removal",
			input: Example{
				ID:          1,
				Similarity:  0.85,
				Description: "A sample description",
			},
			expectedSim: 0.85,
			expectedErr: nil,
			expectedStruct: ExampleWithoutSimilarity{
				ID:          1,
				Description: "A sample description",
			},
		},
		{
			name: "Field not present",
			input: ExampleWithoutSimilarity{
				ID:          1,
				Description: "A sample description",
			},
			expectedSim: 0,
			expectedErr: errors.New("field 'Similarity' not found"),
			expectedStruct: ExampleWithoutSimilarity{
				ID:          1,
				Description: "A sample description",
			},
		},
		{
			name:           "Non-struct input",
			input:          "string input",
			expectedSim:    0,
			expectedErr:    errors.New("input must be a struct"),
			expectedStruct: "string input",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			similarity, updatedStruct, err := extractAndRemoveSimilarity(tt.input)

			if tt.expectedErr != nil {
				if err == nil || err.Error() != tt.expectedErr.Error() {
					t.Fatalf("expected error: %v, got: %v", tt.expectedErr, err)
				}
			} else {
				if err != nil {
					t.Fatalf("expected no error, got: %v", err)
				}
			}

			expectedJson, err := json.Marshal(tt.expectedStruct)
			if err != nil {
				t.Fatalf("failed to marshal expected struct: %v", err)
			}

			actualJson, err := json.Marshal(updatedStruct)
			if err != nil {
				t.Fatalf("failed to marshal actual struct: %v", err)
			}

			if !(similarity == tt.expectedSim && string(expectedJson) == string(actualJson)) {
				t.Fatalf("expected similarity: %v, got: %v; expected struct: %v, got: %v", tt.expectedSim, similarity, string(expectedJson), string(actualJson))
			}

		})
	}
}
