package tests

import (
	"testing"

	"legacycheckoutkata/pricing"

	"github.com/stretchr/testify/require"
)

func TestCalculateTotalCents_CurrentBehavior(t *testing.T) {
	testCases := []struct {
		name     string
		order    pricing.Order
		expected int
	}{
		{
			name: "regular IT baseline from readme",
			order: pricing.Order{
				CustomerType:  "regular",
				SubtotalCents: 10000,
				Country:       "IT",
				CouponCode:    "",
				BlackFriday:   false,
			},
			expected: 12900,
		},
		{
			name: "premium DE SAVE10 from readme",
			order: pricing.Order{
				CustomerType:  "premium",
				SubtotalCents: 10000,
				Country:       "DE",
				CouponCode:    "SAVE10",
				BlackFriday:   false,
			},
			expected: 10420,
		},
		{
			name: "vip IT VIPONLY from readme",
			order: pricing.Order{
				CustomerType:  "vip",
				SubtotalCents: 18000,
				Country:       "IT",
				CouponCode:    "VIPONLY",
				BlackFriday:   false,
			},
			expected: 17980,
		},
		{
			name: "save10 below threshold does not apply",
			order: pricing.Order{
				CustomerType:  "regular",
				SubtotalCents: 4999,
				Country:       "DE",
				CouponCode:    "SAVE10",
				BlackFriday:   false,
			},
			expected: 6848,
		},
		{
			name: "employee does not get black friday discount",
			order: pricing.Order{
				CustomerType:  "employee",
				SubtotalCents: 10000,
				Country:       "US",
				CouponCode:    "",
				BlackFriday:   true,
			},
			expected: 9790,
		},
		{
			name: "taxfree sets non IT tax to zero",
			order: pricing.Order{
				CustomerType:  "vip",
				SubtotalCents: 10000,
				Country:       "US",
				CouponCode:    "TAXFREE",
				BlackFriday:   false,
			},
			expected: 10000,
		},
		{
			name: "freeship checks discounted subtotal",
			order: pricing.Order{
				CustomerType:  "premium",
				SubtotalCents: 10000,
				Country:       "IT",
				CouponCode:    "FREESHIP",
				BlackFriday:   false,
			},
			expected: 10980,
		},
		{
			name: "black friday adds US shipping surcharge",
			order: pricing.Order{
				CustomerType:  "new",
				SubtotalCents: 10000,
				Country:       "US",
				CouponCode:    "",
				BlackFriday:   true,
			},
			expected: 11965,
		},
		{
			name: "unknown country uses default shipping and no tax",
			order: pricing.Order{
				CustomerType:  "mystery",
				SubtotalCents: 10000,
				Country:       "FR",
				CouponCode:    "",
				BlackFriday:   false,
			},
			expected: 12500,
		},
		{
			name: "safe trims spaces in string fields",
			order: pricing.Order{
				CustomerType:  " vip ",
				SubtotalCents: 18000,
				Country:       " IT ",
				CouponCode:    " VIPONLY ",
				BlackFriday:   false,
			},
			expected: 17980,
		},
		{
			name: "negative totals clamp to zero",
			order: pricing.Order{
				CustomerType:  "regular",
				SubtotalCents: -1000,
				Country:       "IT",
				CouponCode:    "",
				BlackFriday:   false,
			},
			expected: 0,
		},
		{
			name: "viponly coupon ignored for non vip",
			order: pricing.Order{
				CustomerType:  "regular",
				SubtotalCents: 20000,
				Country:       "IT",
				CouponCode:    "VIPONLY",
				BlackFriday:   false,
			},
			expected: 25100,
		},
		{
			name: "discount cap at forty percent",
			order: pricing.Order{
				CustomerType:  "employee",
				SubtotalCents: 30000,
				Country:       "US",
				CouponCode:    "SAVE10",
				BlackFriday:   true,
			},
			expected: 21560,
		},
		{
			name: "bulk coupon applies at threshold",
			order: pricing.Order{
				CustomerType:  "regular",
				SubtotalCents: 20000,
				Country:       "DE",
				CouponCode:    "BULK",
				BlackFriday:   false,
			},
			expected: 23034,
		},
		{
			name: "premium gets free shipping above discounted threshold",
			order: pricing.Order{
				CustomerType:  "premium",
				SubtotalCents: 25000,
				Country:       "DE",
				CouponCode:    "",
				BlackFriday:   false,
			},
			expected: 26775,
		},
		{
			name: "freeship still allows employee surcharge",
			order: pricing.Order{
				CustomerType:  "employee",
				SubtotalCents: 12000,
				Country:       "DE",
				CouponCode:    "FREESHIP",
				BlackFriday:   false,
			},
			expected: 10496,
		},
		{
			name: "taxfree does not apply in IT",
			order: pricing.Order{
				CustomerType:  "regular",
				SubtotalCents: 10000,
				Country:       "IT",
				CouponCode:    "TAXFREE",
				BlackFriday:   false,
			},
			expected: 12900,
		},
		{
			name: "partner IT baseline",
			order: pricing.Order{
				CustomerType:  "partner",
				SubtotalCents: 10000,
				Country:       "IT",
				CouponCode:    "",
				BlackFriday:   false,
			},
			expected: 11436,
		},
		{
			name: "partner with PARTNER5 coupon at threshold",
			order: pricing.Order{
				CustomerType:  "partner",
				SubtotalCents: 15000,
				Country:       "IT",
				CouponCode:    "PARTNER5",
				BlackFriday:   false,
			},
			expected: 15889,
		},
		{
			name: "partner PARTNER5 below threshold does not apply",
			order: pricing.Order{
				CustomerType:  "partner",
				SubtotalCents: 10000,
				Country:       "DE",
				CouponCode:    "PARTNER5",
				BlackFriday:   false,
			},
			expected: 11372,
		},
		{
			name: "partner gets free shipping at threshold",
			order: pricing.Order{
				CustomerType:  "partner",
				SubtotalCents: 20000,
				Country:       "US",
				CouponCode:    "",
				BlackFriday:   false,
			},
			expected: 18832,
		},
		{
			name: "partner black friday gets 3% instead of 5%",
			order: pricing.Order{
				CustomerType:  "partner",
				SubtotalCents: 10000,
				Country:       "IT",
				CouponCode:    "",
				BlackFriday:   true,
			},
			expected: 11070,
		},
		{
			name: "partner PARTNER5 coupon ignored for non-partner",
			order: pricing.Order{
				CustomerType:  "regular",
				SubtotalCents: 15000,
				Country:       "IT",
				CouponCode:    "PARTNER5",
				BlackFriday:   false,
			},
			expected: 19000,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := pricing.CalculateTotalCents(tc.order)
			require.Equal(t, tc.expected, actual)
		})
	}
}
