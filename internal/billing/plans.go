package billing

import (
	"os"
	"strconv"
)

// Plan is one line on the pricing page.
//
// The catalogue is code rather than a table because it is three entries that
// change once a year, and a table would need an admin screen, a migration and
// a way to stop a half-edited plan from reaching checkout. Prices are the one
// part that does move, so those come from the environment.
type Plan struct {
	// ID is what the client sends to checkout. It is also written to the
	// payment row, so renaming one breaks the history.
	ID string
	// Months the purchase grants. 0 means it never expires.
	Months int
	// AmountMinor is in kopeks; 0 marks the plan as not for sale.
	AmountMinor int64
	Currency    string
}

// Purchasable reports whether checkout will take money for this plan. The free
// tier is listed on the pricing page but cannot be bought.
func (p Plan) Purchasable() bool { return p.AmountMinor > 0 }

// Lifetime reports whether the plan grants access that never expires.
func (p Plan) Lifetime() bool { return p.Months == 0 }

// Plan identifiers. These are stored in the database, so they are constants.
const (
	PlanFree     = "free"
	PlanMonth    = "month"
	PlanLifetime = "lifetime"
)

// FreeLaunchesPerWeek is how many sandboxes a student without a subscription
// may start in a rolling seven days. The limit exists because a micro-VM is the
// most expensive thing on the host — far more than any number of page views —
// so it is what the subscription actually sells.
const FreeLaunchesPerWeek = 10

// Plans returns the catalogue, cheapest first.
//
//	PRICE_MONTH_MINOR     price of a month, in kopeks (default 400000 = 4000 ₽)
//	PRICE_LIFETIME_MINOR  price of lifetime access   (default 1200000 = 12000 ₽)
func Plans() []Plan {
	return []Plan{
		{ID: PlanFree, Months: 0, AmountMinor: 0, Currency: "RUB"},
		{ID: PlanMonth, Months: 1, AmountMinor: priceEnv("PRICE_MONTH_MINOR", 400_000), Currency: "RUB"},
		{ID: PlanLifetime, Months: 0, AmountMinor: priceEnv("PRICE_LIFETIME_MINOR", 1_200_000), Currency: "RUB"},
	}
}

// PlanByID finds a plan; the second result is false for an unknown id.
func PlanByID(id string) (Plan, bool) {
	for _, p := range Plans() {
		if p.ID == id {
			return p, true
		}
	}
	return Plan{}, false
}

// priceEnv reads a price override. A price that cannot be parsed, or one that
// is zero or negative, falls back to the default: a typo in the environment
// must not quietly put a plan on sale for nothing.
func priceEnv(name string, def int64) int64 {
	n, err := strconv.ParseInt(os.Getenv(name), 10, 64)
	if err != nil || n <= 0 {
		return def
	}
	return n
}
