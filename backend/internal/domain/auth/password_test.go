package auth

import "testing"

func TestArgon2idHasher(t *testing.T) {
	t.Parallel()

	hasher := NewArgon2idHasher(DefaultArgon2idParams())
	hash, err := hasher.Hash("ChangeMe123!")
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}

	if err := hasher.Verify(hash, "ChangeMe123!"); err != nil {
		t.Fatalf("verify password: %v", err)
	}

	if err := hasher.Verify(hash, "wrong-password"); err == nil {
		t.Fatal("expected verify to fail for wrong password")
	}
}
