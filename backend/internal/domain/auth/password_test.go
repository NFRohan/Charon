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

func TestRecommendedArgon2idParamsForProduction(t *testing.T) {
	t.Parallel()

	defaults := DefaultArgon2idParams()
	production := RecommendedArgon2idParams("production")

	if production.Memory <= defaults.Memory {
		t.Fatalf("expected production memory to be greater than default, got %d <= %d", production.Memory, defaults.Memory)
	}
	if production.Iterations <= defaults.Iterations {
		t.Fatalf("expected production iterations to be greater than default, got %d <= %d", production.Iterations, defaults.Iterations)
	}
	if production.Parallelism != defaults.Parallelism {
		t.Fatalf("expected production parallelism to stay stable, got %d", production.Parallelism)
	}
}
