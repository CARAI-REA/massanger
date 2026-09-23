package prodguard

import "testing"

func TestIsWeakSecret(t *testing.T) {
	if !IsWeakSecret("rooms-auth-secret") {
		t.Fatal("expected weak")
	}
	if IsWeakSecret("a-very-long-unique-production-secret-key-32b") {
		t.Fatal("expected strong")
	}
}

func TestForbidLocalhostURL(t *testing.T) {
	if err := ForbidLocalhostURL("url", "wss://calls.example/sfu"); err != nil {
		t.Fatal(err)
	}
	if err := ForbidLocalhostURL("url", "ws://localhost:8082/v1/ws"); err == nil {
		t.Fatal("expected error")
	}
}
