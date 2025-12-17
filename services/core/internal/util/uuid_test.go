package util

import "testing"

func TestGenerateUUID_Success(t *testing.T) {
	id1, err := GenerateUUID()
	if err != nil {
		t.Fatalf("GenerateUUID returned error: %v", err)
	}
	if id1 == "" {
		t.Fatal("GenerateUUID returned empty string")
	}

	id2, err := GenerateUUID()
	if err != nil {
		t.Fatalf("GenerateUUID returned error on second call: %v", err)
	}
	if id1 == id2 {
		t.Fatalf("expected different UUIDs, got the same: %s", id1)
	}
}
