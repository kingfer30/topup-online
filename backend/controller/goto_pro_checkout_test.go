package controller

import "testing"

func TestMonthlyCheckoutTier(t *testing.T) {
	cases := map[string]string{
		"pro":      "pro",
		"pro+":     "pro+",
		"pro_plus": "pro+",
		"ProPlus":  "pro+",
		"ultra":    "ultra",
	}
	for input, want := range cases {
		got, err := monthlyCheckoutTier(input)
		if err != nil {
			t.Fatalf("%s: %v", input, err)
		}
		if got != want {
			t.Fatalf("%s: got %s want %s", input, got, want)
		}
	}
	if _, err := monthlyCheckoutTier("plus"); err == nil {
		t.Fatal("plus should be rejected")
	}
}

func TestNormalizeCursorCookieToken(t *testing.T) {
	got, err := normalizeCursorCookieToken("user_abc::eyJ.test")
	if err != nil {
		t.Fatal(err)
	}
	if got != "user_abc%3A%3AeyJ.test" {
		t.Fatalf("got %s", got)
	}
	encoded := "user_abc%3A%3AeyJ.test"
	got, err = normalizeCursorCookieToken("WorkosCursorSessionToken=" + encoded + "; other=1")
	if err != nil {
		t.Fatal(err)
	}
	if got != encoded {
		t.Fatalf("got %s", got)
	}
	if _, err := normalizeCursorCookieToken("not-a-token"); err == nil {
		t.Fatal("invalid token should fail")
	}
}

func TestParseCheckoutURL(t *testing.T) {
	if got := parseCheckoutURL([]byte(`"https://checkout.stripe.com/c/pay/cs_test"`)); got != "https://checkout.stripe.com/c/pay/cs_test" {
		t.Fatalf("string: %s", got)
	}
	if got := parseCheckoutURL([]byte(`{"url":"https://checkout.stripe.com/c/pay/cs_obj"}`)); got != "https://checkout.stripe.com/c/pay/cs_obj" {
		t.Fatalf("object: %s", got)
	}
}

func TestCheckoutSessionID(t *testing.T) {
	got, err := checkoutSessionID("https://checkout.stripe.com/c/pay/cs_live_123#fid")
	if err != nil {
		t.Fatal(err)
	}
	if got != "cs_live_123" {
		t.Fatalf("got %s", got)
	}
	if _, err := checkoutSessionID("https://cursor.com/dashboard"); err == nil {
		t.Fatal("dashboard url should fail")
	}
}
