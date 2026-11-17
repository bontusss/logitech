package auth

import "testing"

func TestHashAndVerifyPassword(t *testing.T) {
	hash, err := HashPassword("correct horse battery staple")
	if err != nil {
		t.Fatalf("hash: %v", err)
	}
	if hash == "correct horse battery staple" {
		t.Fatal("password was stored in plaintext")
	}

	ok, err := VerifyPassword(hash, "correct horse battery staple")
	if err != nil || !ok {
		t.Fatalf("expected valid password, ok=%v err=%v", ok, err)
	}

	ok, err = VerifyPassword(hash, "wrong password")
	if err != nil {
		t.Fatalf("verify wrong password: %v", err)
	}
	if ok {
		t.Fatal("wrong password was accepted")
	}
}

func TestHashPasswordUsesUniqueSalt(t *testing.T) {
	first, err := HashPassword("same-password")
	if err != nil {
		t.Fatal(err)
	}
	second, err := HashPassword("same-password")
	if err != nil {
		t.Fatal(err)
	}
	if first == second {
		t.Fatal("identical hashes for the same password; salt is not random")
	}
}

func TestVerifyPasswordRejectsMalformedHash(t *testing.T) {
	if _, err := VerifyPassword("not-a-hash", "password"); err == nil {
		t.Fatal("expected malformed hash error")
	}
}
