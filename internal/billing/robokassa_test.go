package billing

import (
	"net/url"
	"strings"
	"testing"
)

// The signature formulas come from Robokassa's documentation and are the whole
// of the security here. These use the worked example from it: MerchantLogin
// "demo", OutSum "100.00", InvId 1, password "password" — MD5 of
// "demo:100.00:1:password".
func TestPayLinkSignature(t *testing.T) {
	r := &Robokassa{Login: "demo", Pass1: "password", Pass2: "password2", Test: true}
	link, err := r.PayLink(1, 10000, "Подписка", nil)
	if err != nil {
		t.Fatalf("pay link: %v", err)
	}
	u, err := url.Parse(link)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	q := u.Query()
	if got := q.Get("OutSum"); got != "100.00" {
		t.Errorf("OutSum = %q, want 100.00", got)
	}
	if got := q.Get("IsTest"); got != "1" {
		t.Error("test mode did not reach the link")
	}
	want := md5hex("demo:100.00:1:password")
	if got := q.Get("SignatureValue"); got != want {
		t.Errorf("signature = %s, want %s", got, want)
	}
}

// Money must never be assumed: a shop left unconfigured has to refuse to build
// a link rather than produce one that silently cannot be paid.
func TestPayLinkRefusesWithoutCredentials(t *testing.T) {
	for _, r := range []*Robokassa{
		{},
		{Login: "demo"},
		{Login: "demo", Pass1: "p1"},
	} {
		if _, err := r.PayLink(1, 100, "x", nil); err != ErrNotConfigured {
			t.Errorf("%+v built a link without credentials", r)
		}
	}
	ok := &Robokassa{Login: "demo", Pass1: "p1", Pass2: "p2"}
	if _, err := ok.PayLink(0, 100, "x", nil); err == nil {
		t.Error("invoice id 0 was accepted; Robokassa requires a positive one")
	}
}

// Granting a subscription rests entirely on this check.
func TestVerifyResultUsesTheSecondPassword(t *testing.T) {
	r := &Robokassa{Login: "demo", Pass1: "first", Pass2: "second"}
	good := md5hex("100.00:1:second")

	if !r.VerifyResult("100.00", 1, good, nil) {
		t.Fatal("a correct callback signature was rejected")
	}
	if !r.VerifyResult("100.00", 1, strings.ToUpper(good), nil) {
		t.Error("Robokassa may send the digest uppercase; it must still verify")
	}
	// Signed with the wrong password — this is what a forged callback looks like
	// if the first password ever leaked through the payment link.
	if r.VerifyResult("100.00", 1, md5hex("100.00:1:first"), nil) {
		t.Error("a callback signed with password #1 was accepted")
	}
	for _, bad := range []struct {
		sum string
		inv int64
	}{
		{"1000.00", 1}, // amount tampered with
		{"100.00", 2},  // someone else's invoice
	} {
		if r.VerifyResult(bad.sum, bad.inv, good, nil) {
			t.Errorf("tampered callback accepted: sum=%s inv=%d", bad.sum, bad.inv)
		}
	}
	if r.VerifyResult("100.00", 1, "", nil) {
		t.Error("an empty signature was accepted")
	}
}

// Custom parameters take part in the signature, sorted by name. Leaving one out
// or ordering them differently silently invalidates every callback.
func TestShpParametersAreSignedInOrder(t *testing.T) {
	r := &Robokassa{Login: "demo", Pass1: "p1", Pass2: "p2"}
	shp := url.Values{}
	shp.Set("Shp_user", "7")
	shp.Set("Shp_months", "1")

	want := md5hex("100.00:1:p2:Shp_months=1:Shp_user=7")
	if !r.VerifyResult("100.00", 1, want, shp) {
		t.Fatal("Shp parameters are not signed in sorted order")
	}
	// The same callback without them must not verify, or a forger could simply
	// drop the parameters that say whose subscription this is.
	if r.VerifyResult("100.00", 1, want, nil) {
		t.Error("signature verified with the Shp parameters removed")
	}
	if got := ShpValues(url.Values{"Shp_a": {"1"}, "Other": {"2"}}); len(got) != 1 || got.Get("Shp_a") != "1" {
		t.Errorf("ShpValues kept the wrong fields: %v", got)
	}
}

func TestAmountFormatting(t *testing.T) {
	for _, c := range []struct {
		minor int64
		want  string
	}{{49000, "490.00"}, {100, "1.00"}, {5, "0.05"}, {0, "0.00"}, {123456, "1234.56"}} {
		if got := Amount(c.minor); got != c.want {
			t.Errorf("Amount(%d) = %q, want %q", c.minor, got, c.want)
		}
	}
}

func md5hex(s string) string {
	r := &Robokassa{}
	return r.digest(s)
}

// Robokassa echoes the sum back as text and is not strict about the decimals,
// so every shape it can send has to come back as the same number of kopeks.
// Getting this wrong rejects real payments, which is the failure a student
// notices and we do not.
func TestParseAmount(t *testing.T) {
	ok := map[string]int64{
		"490":        49000,
		"490.0":      49000,
		"490.00":     49000,
		"490.10":     49010,
		"490,50":     49050, // some locales send a comma
		"4000.00":    400000,
		"12000":      1200000,
		" 490.00 ":   49000,
		"0.01":       1,
		"0":          0,
		"1234567.89": 123456789,
	}
	for in, want := range ok {
		got, valid := ParseAmount(in)
		if !valid || got != want {
			t.Errorf("ParseAmount(%q) = %d, %v; want %d, true", in, got, valid, want)
		}
	}

	bad := []string{"", "abc", "490.123", "-490.00", "4 90", "490.0x", "1e3"}
	for _, in := range bad {
		if got, valid := ParseAmount(in); valid {
			t.Errorf("ParseAmount(%q) = %d, true; want it refused", in, got)
		}
	}

	// Whatever Amount writes, ParseAmount must read back unchanged — the two
	// are used on opposite ends of the same payment.
	for _, minor := range []int64{0, 1, 99, 49000, 400000, 1200000} {
		if got, valid := ParseAmount(Amount(minor)); !valid || got != minor {
			t.Errorf("round trip of %d gave %d, %v", minor, got, valid)
		}
	}
}

// The plan catalogue is what checkout charges, so the parts that would cost
// money if wrong are pinned: the free tier cannot be bought, and a price
// override that is nonsense does not put a plan on sale for nothing.
func TestPlans(t *testing.T) {
	free, ok := PlanByID(PlanFree)
	if !ok || free.Purchasable() {
		t.Fatalf("free plan = %+v, ok = %v", free, ok)
	}
	month, ok := PlanByID(PlanMonth)
	if !ok || !month.Purchasable() || month.Months != 1 || month.Lifetime() {
		t.Fatalf("month plan = %+v, ok = %v", month, ok)
	}
	life, ok := PlanByID(PlanLifetime)
	if !ok || !life.Purchasable() || !life.Lifetime() {
		t.Fatalf("lifetime plan = %+v, ok = %v", life, ok)
	}
	if _, ok := PlanByID("gratis"); ok {
		t.Error("an unknown plan was accepted")
	}

	t.Setenv("PRICE_MONTH_MINOR", "150000")
	if p, _ := PlanByID(PlanMonth); p.AmountMinor != 150000 {
		t.Errorf("override ignored: %d", p.AmountMinor)
	}
	for _, junk := range []string{"", "free", "0", "-1"} {
		t.Setenv("PRICE_MONTH_MINOR", junk)
		if p, _ := PlanByID(PlanMonth); p.AmountMinor != 400000 {
			t.Errorf("PRICE_MONTH_MINOR=%q gave %d, want the default", junk, p.AmountMinor)
		}
	}
}
