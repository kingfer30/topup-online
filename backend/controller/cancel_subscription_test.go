package controller

import "testing"

func TestParseGrokSubscriptionsPremium(t *testing.T) {
	state := append(protoVarint(1, grokStatusActive), protoVarint(6, grokPlatformStripe)...)
	state = append(state, protoString(9, "Premium")...)
	sub := append(protoString(1, "csub_a4fRbHP9ZA6GFwO8l4qS"), protoBytes(2, state)...)
	raw := grpcFrame(protoBytes(1, sub))

	got := parseGrokSubscriptions(raw)
	if len(got) != 1 {
		t.Fatalf("got %d subscriptions", len(got))
	}
	if got[0].ID != "csub_a4fRbHP9ZA6GFwO8l4qS" || got[0].Status != grokStatusActive || got[0].Platform != grokPlatformStripe || got[0].Plan != "Premium" {
		t.Fatalf("parsed %+v", got[0])
	}
}

func protoVarint(field, value int) []byte {
	out := []byte{byte(field << 3)}
	return appendVarint(out, value)
}

func protoBytes(field int, value []byte) []byte {
	out := []byte{byte(field<<3 | 2)}
	out = appendVarint(out, len(value))
	return append(out, value...)
}
