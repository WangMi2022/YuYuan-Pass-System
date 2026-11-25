package captcha

import "testing"

func TestRedisStoreVerifyRejectsEmptyChallenge(t *testing.T) {
	store := NewDefaultRedisStore()
	for _, challenge := range []struct {
		id     string
		answer string
	}{
		{},
		{id: "challenge-id"},
		{answer: "1234"},
		{id: "   ", answer: "   "},
	} {
		if store.Verify(challenge.id, challenge.answer, true) {
			t.Fatalf("empty Redis captcha challenge was accepted: %#v", challenge)
		}
	}
}
