package game

import (
	"strings"
	"testing"
)

func TestGenerateRoomCode_Length(t *testing.T) {
	code := GenerateRoomCode()
	if len(code) != 6 {
		t.Errorf("expected code length 6, got %d", len(code))
	}
}

func TestGenerateRoomCode_ValidCharset(t *testing.T) {
	// Characters that should NOT appear (ambiguous characters)
	ambiguous := "0O1IL"
	// Valid characters
	validChars := "ABCDEFGHJKMNPQRSTUVWXYZ23456789"

	// Generate many codes and verify all characters are valid
	for i := 0; i < 100; i++ {
		code := GenerateRoomCode()

		// Check for ambiguous characters
		for _, c := range ambiguous {
			if strings.ContainsRune(code, c) {
				t.Errorf("code %q contains ambiguous character %q", code, c)
			}
		}

		// Check all characters are in valid set
		for _, c := range code {
			if !strings.ContainsRune(validChars, c) {
				t.Errorf("code %q contains invalid character %q", code, c)
			}
		}
	}
}

func TestGenerateRoomCode_Uniqueness(t *testing.T) {
	seen := make(map[string]bool)
	numCodes := 1000

	for i := 0; i < numCodes; i++ {
		code := GenerateRoomCode()
		if seen[code] {
			t.Errorf("duplicate code generated: %q", code)
		}
		seen[code] = true
	}

	// With 30 possible characters and 6 positions: 30^6 = 729,000,000 possibilities
	// Probability of collision in 1000 samples is very low
	t.Logf("generated %d unique codes", len(seen))
}

func TestGenerateRoomCode_AllUppercase(t *testing.T) {
	for i := 0; i < 100; i++ {
		code := GenerateRoomCode()
		if code != strings.ToUpper(code) {
			t.Errorf("code %q contains lowercase characters", code)
		}
	}
}

func TestGenerateRoomCode_Distribution(t *testing.T) {
	// Generate many codes and check rough distribution
	charCount := make(map[rune]int)
	numCodes := 10000

	for i := 0; i < numCodes; i++ {
		code := GenerateRoomCode()
		for _, c := range code {
			charCount[c]++
		}
	}

	// With 30 chars and 60000 total chars (10000 codes * 6 chars),
	// each char should appear roughly 2000 times
	totalChars := numCodes * 6
	expectedPerChar := float64(totalChars) / 30.0

	// Allow 50% deviation (very loose check, mainly for detecting obvious bias)
	minExpected := int(expectedPerChar * 0.5)
	maxExpected := int(expectedPerChar * 1.5)

	for char, count := range charCount {
		if count < minExpected || count > maxExpected {
			t.Errorf("character %q appeared %d times (expected between %d and %d, avg ~%.0f)", char, count, minExpected, maxExpected, expectedPerChar)
		}
	}
}

// Benchmark to ensure code generation is fast
func BenchmarkGenerateRoomCode(b *testing.B) {
	for i := 0; i < b.N; i++ {
		GenerateRoomCode()
	}
}
