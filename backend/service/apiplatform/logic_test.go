package apiplatform

import "testing"

func TestQuotaToUSD(t *testing.T) {
	got := QuotaToUSD(258123314, 500000)
	if got != 516.246628 {
		t.Fatalf("got %v", got)
	}
	if QuotaToUSD(500000, 0) != 1 {
		t.Fatal("default quota per unit")
	}
}

func TestParseTestStatus(t *testing.T) {
	code, drop := ParseTestStatus(200, true, "")
	if code != 200 || drop {
		t.Fatalf("success: %d %v", code, drop)
	}
	code, drop = ParseTestStatus(200, false, "bad response status code 401, message: no")
	if code != 401 || !drop {
		t.Fatalf("401: %d %v", code, drop)
	}
	code, drop = ParseTestStatus(200, false, "bad response status code 429, message: slow")
	if code != 429 || drop {
		t.Fatalf("429: %d %v", code, drop)
	}
	code, drop = ParseTestStatus(401, false, "Unauthorized, insufficient privileges")
	if code != 401 || drop {
		t.Fatalf("auth 401 should not drop model: %d %v", code, drop)
	}
}

func TestRemoveModel(t *testing.T) {
	next, removed := RemoveModel("gpt-5.6-sol, gpt-6-luna", "gpt-5.6-sol")
	if !removed || next != "gpt-6-luna" {
		t.Fatalf("got %q %v", next, removed)
	}
	next, removed = RemoveModel("gpt-6-luna", "missing")
	if removed || next != "gpt-6-luna" {
		t.Fatalf("missing: %q %v", next, removed)
	}
}

func TestUnsettledAfterPay(t *testing.T) {
	if UnsettledAfter(150, 100) != 50 {
		t.Fatal("grown usage should remain unsettled")
	}
	if UnsettledAfter(100, 100) != 0 {
		t.Fatal("exact pay")
	}
}
