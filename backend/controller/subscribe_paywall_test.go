package controller

import (
	"strings"
	"testing"
)

func TestPickMonthlySKU(t *testing.T) {
	interval := append(protoVarintField(1, 1), protoVarintField(2, intervalUnitMonth)...)
	price := protoEmbed(3, protoEmbed(2, interval))
	monthly := append(protoString(1, "sku_plus"), protoString(9, "Plus")...)
	monthly = append(monthly, price...)

	yearlyInterval := append(protoVarintField(1, 1), protoVarintField(2, 3)...)
	yearlyPrice := protoEmbed(3, protoEmbed(2, yearlyInterval))
	yearly := append(protoString(1, "sku_plus_year"), protoString(9, "Plus")...)
	yearly = append(yearly, yearlyPrice...)

	payload := append(protoEmbed(3, yearly), protoEmbed(3, monthly)...)
	sku, err := pickMonthlySKU(payload, "Plus")
	if err != nil {
		t.Fatal(err)
	}
	if sku != "sku_plus" {
		t.Fatalf("sku = %s", sku)
	}
}

func TestAcquireRequestRoundTrip(t *testing.T) {
	body := acquireRequest("cart_test", "pm_test", "user@example.com", billingAddress{
		Line1:   billingLine1,
		City:    billingCity,
		State:   billingState,
		Postal:  billingPostal,
		Country: billingCountry,
	})
	fields := protoFields(body)
	got := map[int]string{}
	for _, field := range fields {
		if field.wt == 2 {
			got[field.num] = string(field.raw)
		}
	}
	if got[1] != "cart_test" {
		t.Fatalf("cart = %q", got[1])
	}
	if string(protoFields([]byte(got[2]))[0].raw) != "pm_test" {
		t.Fatal("payment method missing")
	}
	if !isPrintable(got[4]) || len(got[4]) < 32 {
		t.Fatal("idempotency missing")
	}
	address := protoFields([]byte(got[7]))
	if string(address[0].raw) != billingCountry || string(address[4].raw) != billingCity {
		t.Fatalf("address = %#v", address)
	}
	if got[9] != "user@example.com" {
		t.Fatalf("email = %q", got[9])
	}
}

func TestRedactSubscribeLog(t *testing.T) {
	raw := "卡 4242424242424242 会话 pps-session.v1.CiRzZXNzaW9uVG9rZW4 凭证 eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxIn0.signature pk_live_51PksddHJohyvID2cABC"
	got := redactSubscribeLog(raw)
	if strings.Contains(got, "4242424242424242") || strings.Contains(got, "pps-session.v1.Ci") || strings.Contains(got, "eyJhbGci") || strings.Contains(got, "pk_live_51") {
		t.Fatalf("secret leaked: %s", got)
	}
}

func TestAttachPaymentMethodRequest(t *testing.T) {
	body := attachPaymentMethodRequest("team_123", "pm_test", billingAddress{
		Line1: billingLine1, City: billingCity, State: billingState, Postal: billingPostal, Country: billingCountry,
	})
	fields := protoFields(body)
	got := map[int]protoField{}
	for _, field := range fields {
		got[field.num] = field
	}
	if got[1].numVal != billingPlatformStripe {
		t.Fatalf("provider = %d", got[1].numVal)
	}
	if string(got[2].raw) != "pm_test" || string(got[5].raw) != "team_123" {
		t.Fatalf("pm=%q team=%q", got[2].raw, got[5].raw)
	}
	address := protoFields(got[7].raw)
	if string(address[0].raw) != billingCountry || string(address[4].raw) != billingCity {
		t.Fatal("address mismatch")
	}
}

func TestDeclineReasonText(t *testing.T) {
	payload := protoVarintField(4, 2)
	if declineReasonText(payload, 4) != "余额不足" {
		t.Fatal(declineReasonText(payload, 4))
	}
}

func TestCartTeamID(t *testing.T) {
	payload := protoString(8, "team_abc")
	if cartTeamID(payload) != "team_abc" {
		t.Fatal(cartTeamID(payload))
	}
}

func TestNormalizeBillingDefaults(t *testing.T) {
	address, err := normalizeBilling("", "", "", "", "")
	if err != nil {
		t.Fatal(err)
	}
	if address.Line1 != billingLine1 || address.City != billingCity || address.State != billingState || address.Postal != billingPostal || address.Country != billingCountry {
		t.Fatalf("%+v", address)
	}
	custom, err := normalizeBilling(" 1 Main ", " Bend ", " or ", "97701", "us")
	if err != nil {
		t.Fatal(err)
	}
	if custom.Country != "US" || custom.State != "OR" || custom.City != "Bend" {
		t.Fatalf("%+v", custom)
	}
}

func TestNormalizeCard(t *testing.T) {
	number, month, year, cvc, err := normalizeCard("4242 4242 4242 4242", "8", "29", "123")
	if err != nil {
		t.Fatal(err)
	}
	if number != "4242424242424242" || month != "08" || year != "2029" || cvc != "123" {
		t.Fatalf("%s %s %s %s", number, month, year, cvc)
	}
	if _, _, _, _, err := normalizeCard("4242", "8", "29", "123"); err == nil {
		t.Fatal("short card accepted")
	}
}
