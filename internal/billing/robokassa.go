// Package billing talks to Robokassa: it builds the link a student is sent to
// and checks the signature on the callback that says they paid.
//
// Only what a subscription needs is here — no recurring charges, no refunds.
// Robokassa's own API has all of that; adding it later means adding a third
// password and more signature forms, and none of it is worth carrying unused.
//
// Two passwords, two different jobs, and mixing them up is the classic way to
// get this wrong: password #1 signs the link the student follows and the
// Success redirect they come back on; password #2 signs the Result callback
// Robokassa makes server-to-server. The Result callback is the only one that
// may be believed — a student controls their own browser and can forge a
// Success redirect, but not a request they never see.
package billing

import (
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"
)

// Robokassa holds the merchant credentials and the endpoint to send students to.
type Robokassa struct {
	Login string // MerchantLogin
	Pass1 string // signs the payment link and the Success redirect
	Pass2 string // signs the Result callback
	// Test puts the shop in Robokassa's test mode: the payment page takes test
	// cards and no money moves. Default on — real charges must be a deliberate
	// act, not something that happens because a variable was left unset.
	Test bool
	// SHA256 switches the digest. Robokassa defaults to MD5 and the shop's
	// setting must match this, or every signature is rejected as invalid.
	SHA256 bool
	// PayURL is the merchant endpoint; empty uses Robokassa's own.
	PayURL string
}

const defaultPayURL = "https://auth.robokassa.ru/Merchant/Index.aspx"

// ErrNotConfigured is returned when the merchant credentials are missing, so a
// caller can fall back instead of building a link that cannot work.
var ErrNotConfigured = errors.New("robokassa is not configured")

// Configured reports whether there is enough to build and verify a payment.
func (r *Robokassa) Configured() bool {
	return r != nil && r.Login != "" && r.Pass1 != "" && r.Pass2 != ""
}

func (r *Robokassa) digest(s string) string {
	if r.SHA256 {
		sum := sha256.Sum256([]byte(s))
		return hex.EncodeToString(sum[:])
	}
	sum := md5.Sum([]byte(s))
	return hex.EncodeToString(sum[:])
}

// Amount renders a minor-unit amount the way Robokassa wants it in both the
// link and the signature: a plain decimal with two places. The signature is
// over this exact text, so "490" and "490.00" are different signatures and one
// of them is wrong.
func Amount(minor int64) string {
	return fmt.Sprintf("%d.%02d", minor/100, minor%100)
}

// ParseAmount reads back what Amount wrote: a decimal sum in roubles, into
// kopeks. Robokassa is not strict about the number of decimals it echoes, so
// "490", "490.0" and "490.00" all have to mean the same thing.
//
// Parsing as text rather than through a float is deliberate: 490.10 has no
// exact float representation, and a payment check is the last place to want a
// rounding argument.
func ParseAmount(s string) (int64, bool) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, false
	}
	whole, frac, hasFrac := strings.Cut(s, ".")
	if !hasFrac {
		whole, frac, hasFrac = strings.Cut(s, ",")
	}
	if hasFrac {
		switch len(frac) {
		case 1:
			frac += "0"
		case 2:
		default:
			return 0, false
		}
	} else {
		frac = "00"
	}
	rub, err := strconv.ParseInt(whole, 10, 64)
	if err != nil || rub < 0 {
		return 0, false
	}
	kop, err := strconv.ParseInt(frac, 10, 64)
	if err != nil || kop < 0 {
		return 0, false
	}
	return rub*100 + kop, true
}

// shpSuffix appends the custom Shp_ parameters to a signature base. Robokassa
// requires them sorted by name, and every one of them must be included —
// leaving a single parameter out makes the signature mismatch with no
// indication of which one it was.
func shpSuffix(base string, shp url.Values) string {
	if len(shp) == 0 {
		return base
	}
	keys := make([]string, 0, len(shp))
	for k := range shp {
		if strings.HasPrefix(k, "Shp_") {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)
	for _, k := range keys {
		base += fmt.Sprintf(":%s=%s", k, shp.Get(k))
	}
	return base
}

// PayLink is the URL to send the student to.
//
// invID is the invoice number Robokassa will quote back in the callback; it
// must be a positive integer unique to the shop, which is why the payment row's
// own id is used for it.
func (r *Robokassa) PayLink(invID int64, amountMinor int64, description string, shp url.Values) (string, error) {
	if !r.Configured() {
		return "", ErrNotConfigured
	}
	if invID <= 0 {
		return "", fmt.Errorf("robokassa: invoice id must be positive, got %d", invID)
	}
	sum := Amount(amountMinor)
	sig := r.digest(shpSuffix(fmt.Sprintf("%s:%s:%d:%s", r.Login, sum, invID, r.Pass1), shp))

	q := url.Values{}
	q.Set("MerchantLogin", r.Login)
	q.Set("OutSum", sum)
	q.Set("InvId", fmt.Sprintf("%d", invID))
	q.Set("Description", description)
	q.Set("SignatureValue", sig)
	q.Set("Culture", "ru")
	q.Set("Encoding", "utf-8")
	if r.Test {
		q.Set("IsTest", "1")
	}
	for k := range shp {
		if strings.HasPrefix(k, "Shp_") {
			q.Set(k, shp.Get(k))
		}
	}

	base := r.PayURL
	if base == "" {
		base = defaultPayURL
	}
	u, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("robokassa: bad payment url %q: %w", base, err)
	}
	u.RawQuery = q.Encode()
	return u.String(), nil
}

// VerifyResult checks the signature on the server-to-server callback. The whole
// decision to grant a subscription rests on this returning true, so it is
// deliberately strict: the amount and invoice are taken from the request as
// text and signed as such, and every Shp_ parameter present is included.
func (r *Robokassa) VerifyResult(outSum string, invID int64, signature string, shp url.Values) bool {
	if !r.Configured() || signature == "" {
		return false
	}
	want := r.digest(shpSuffix(fmt.Sprintf("%s:%d:%s", outSum, invID, r.Pass2), shp))
	return strings.EqualFold(signature, want)
}

// VerifySuccess checks the signature on the redirect the student's browser
// arrives on. It proves the link was ours, not that money moved — the student
// is holding this request and can replay it. Use it to decide what page to
// show, never to grant anything.
func (r *Robokassa) VerifySuccess(outSum string, invID int64, signature string, shp url.Values) bool {
	if !r.Configured() || signature == "" {
		return false
	}
	want := r.digest(shpSuffix(fmt.Sprintf("%s:%d:%s", outSum, invID, r.Pass1), shp))
	return strings.EqualFold(signature, want)
}

// ShpValues keeps only the Shp_ parameters out of a form, which is all that
// takes part in a signature.
func ShpValues(form url.Values) url.Values {
	out := url.Values{}
	for k := range form {
		if strings.HasPrefix(k, "Shp_") {
			out.Set(k, form.Get(k))
		}
	}
	return out
}

// RobokassaFromEnv reads the merchant credentials from the environment.
//
// Returns nil when the login or either password is missing, which is the normal
// state in development: checkout then falls back to the stub provider instead of
// building a link that cannot work.
//
// The passwords are secrets and belong in the cluster's secret, never in the
// repository or an image.
//
//	ROBOKASSA_LOGIN      MerchantLogin
//	ROBOKASSA_PASSWORD1  signs the payment link and the Success redirect
//	ROBOKASSA_PASSWORD2  signs the Result callback
//	ROBOKASSA_TEST       "0"/"false" leaves test mode; anything else, including
//	                     unset, keeps it on
//	ROBOKASSA_SHA256     "1"/"true" if the shop is set to SHA-256 instead of MD5
//	ROBOKASSA_PAY_URL    override the merchant endpoint (tests, staging)
func RobokassaFromEnv() *Robokassa {
	r := &Robokassa{
		Login:  strings.TrimSpace(os.Getenv("ROBOKASSA_LOGIN")),
		Pass1:  os.Getenv("ROBOKASSA_PASSWORD1"),
		Pass2:  os.Getenv("ROBOKASSA_PASSWORD2"),
		SHA256: envTrue("ROBOKASSA_SHA256"),
		PayURL: strings.TrimSpace(os.Getenv("ROBOKASSA_PAY_URL")),
		// Leaving test mode has to be written out explicitly. Getting this
		// backwards means taking real money from students by accident.
		Test: os.Getenv("ROBOKASSA_TEST") != "0" && !strings.EqualFold(os.Getenv("ROBOKASSA_TEST"), "false"),
	}
	if !r.Configured() {
		return nil
	}
	return r
}

func envTrue(key string) bool {
	v := strings.TrimSpace(os.Getenv(key))
	return v == "1" || strings.EqualFold(v, "true")
}
