package util

import "testing"

func TestGenerateUUID_IsValid(t *testing.T) {
	id := GenerateUUID()
	if id == "" || !IsValidUUID(id) {
		t.Fatalf("GenerateUUID produced invalid id: %q", id)
	}
}
