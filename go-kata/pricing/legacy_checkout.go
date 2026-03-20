package pricing

import (
	"strings"
)

type CustomerType string

const (
	CustomerTypeRegular  CustomerType = "regular"
	CustomerTypeNew      CustomerType = "new"
	CustomerTypeVip      CustomerType = "vip"
	CustomerTypePremium  CustomerType = "premium"
	CustomerTypeEmployee CustomerType = "employee"
	CustomerTypePartner  CustomerType = "partner"
)

type Country string

const (
	CountryIT Country = "IT"
	CountryDE Country = "DE"
	CountryUS Country = "US"
)

// Coupon codes and names used in discount and tax calculations.
const (
	couponSave10   = "SAVE10"
	couponVipOnly  = "VIPONLY"
	couponBulk     = "BULK"
	couponFreeShip = "FREESHIP"
	couponTaxFree  = "TAXFREE"
	couponPartner5 = "PARTNER5"
)

// Discount rules: base percentages by customer type and coupon bonuses.
// Discounts accumulate and are capped at maxDiscountPercent.
const (
	maxDiscountPercent            = 40
	vipBaseDiscountPercent        = 15
	premiumHighBaseDiscountPct    = 10
	premiumLowBaseDiscountPct     = 5
	employeeBaseDiscountPercent   = 30
	partnerBaseDiscountPercent    = 12
	blackFridayExtraDiscountPct   = 5
	partnerBlackFridayDiscountPct = 3
	save10DiscountPercent         = 10
	vipOnlyDiscountPercent        = 5
	bulkDiscountPercent           = 7
	partner5DiscountPercent       = 5
	save10MinSubtotalCents        = 5000
	bulkMinSubtotalCents          = 20000
	premiumBaseTierMinSubtotal    = 10000
	partner5MinSubtotalCents      = 12000
)

// Shipping base rates by country and promotion thresholds.
// Employee surcharge in non-IT countries overrides free shipping.
// Partner customers get free shipping above a specific discounted subtotal threshold.
const (
	defaultShippingCents              = 2500
	shippingITCents                   = 700
	shippingDECents                   = 900
	shippingUSCents                   = 1500
	blackFridayUSShippingSurcharge    = 300
	freeShipMinDiscountedSubtotal     = 8000
	vipFreeShippingMinDiscountedTotal = 15000
	premiumFreeShippingMinSubtotal    = 20000
	partnerFreeShippingMinSubtotal    = 15000
	employeeNonITShippingSurcharge    = 500
)

// Tax rates by country with VIP override in specific regions.
const (
	taxITPercent      = 22
	taxDEPercent      = 19
	taxUSPercent      = 7
	vipTaxInITPercent = 20
)

type Order struct {
	CustomerType  string
	SubtotalCents int
	Country       string
	CouponCode    string
	BlackFriday   bool
}

// CalculateTotalCents computes the total order price.
//
// This function orchestrates three calculation steps:
//  1. Discounts are applied to the subtotal (capped at 40%)
//  2. Shipping is computed based on country, customer type, and discounted subtotal
//  3. Taxes are applied to the discounted subtotal, then added with shipping
//
// Return value clamps to 0 for negative totals (edge case with negative input subtotals).
//
// Quirk: Unknown customer types and countries fall back gracefully:
//   - Unknown customer type → 0% base discount
//   - Unknown country → default shipping (2500 cents) + 0% tax
//
// Quirk: Employee surcharge (non-IT countries) overrides free shipping entirely.
// Employees always pay at least the surcharge, even if other rules would waive shipping.
func CalculateTotalCents(order Order) int {
	subtotal := order.SubtotalCents
	customerType := parseCustomerType(order.CustomerType)
	country := parseCountry(order.Country)
	coupon := safe(order.CouponCode)

	discountPercent := calculateDiscountPercent(subtotal, customerType, coupon, order.BlackFriday)
	discountedSubtotal := subtotal * (100 - discountPercent) / 100
	shippingCents := calculateShippingCents(discountedSubtotal, customerType, country, coupon, order.BlackFriday)
	taxPercent := calculateTaxPercent(customerType, country, coupon)
	taxCents := discountedSubtotal * taxPercent / 100
	total := discountedSubtotal + shippingCents + taxCents

	if total < 0 {
		return 0
	}

	return total
}

// calculateDiscountPercent computes the total discount percentage for an order.
//
// Parameters:
//   - subtotal: Order subtotal in cents (used for threshold checks)
//   - customerType: Identifies the customer's tier (vip, premium, employee, partner, regular, new)
//   - coupon: Coupon code applied (may provide additional discount if thresholds are met)
//   - blackFriday: True if Black Friday promotion is active
//
// Returns: Discount percentage (0-40), where higher values represent greater discounts.
//
// Rules (applied in order, discounts accumulate until capped):
//  1. Customer type discount: employee(30%) > vip(15%) > partner(12%) > premium(5-10% tiered) > others(0%)
//  2. Coupon bonuses: SAVE10(+10% if subtotal >= 5000), VIPONLY(+5% vip only),
//     BULK(+7% if subtotal >= 20000), PARTNER5(+5% partner only if subtotal >= 12000)
//  3. Black Friday: partner(+3%) or others(+5%), employees excluded
//  4. Hard cap: 40% maximum discount percentage
func calculateDiscountPercent(subtotal int, customerType CustomerType, coupon string, blackFriday bool) int {
	discountPercent := 0

	// Step 1: Apply customer type base discount.
	switch customerType {
	case CustomerTypeVip:
		discountPercent += vipBaseDiscountPercent
	case CustomerTypePremium:
		// Premium tier: 10% for large orders (>= 10000), 5% otherwise.
		if subtotal >= premiumBaseTierMinSubtotal {
			discountPercent += premiumHighBaseDiscountPct
		} else {
			discountPercent += premiumLowBaseDiscountPct
		}
	case CustomerTypeEmployee:
		discountPercent += employeeBaseDiscountPercent
	case CustomerTypePartner:
		discountPercent += partnerBaseDiscountPercent
	}

	// Step 2: Apply coupon-based bonus discounts.
	switch coupon {
	case couponSave10:
		if subtotal >= save10MinSubtotalCents {
			discountPercent += save10DiscountPercent
		}
	case couponVipOnly:
		if customerType == CustomerTypeVip {
			discountPercent += vipOnlyDiscountPercent
		}
	case couponBulk:
		if subtotal >= bulkMinSubtotalCents {
			discountPercent += bulkDiscountPercent
		}
	case couponPartner5:
		// PARTNER5 coupon only applies to partner customers and requires minimum subtotal.
		if customerType == CustomerTypePartner && subtotal >= partner5MinSubtotalCents {
			discountPercent += partner5DiscountPercent
		}
	}

	// Step 3: Apply Black Friday bonus (except employees are excluded).
	// Note: Partner customers get different bonus (+3%) compared to others (+5%).
	if blackFriday {
		if customerType == CustomerTypeEmployee {
			// Employees get no Black Friday discount
		} else if customerType == CustomerTypePartner {
			discountPercent += partnerBlackFridayDiscountPct
		} else {
			discountPercent += blackFridayExtraDiscountPct
		}
	}

	// Step 4: Cap at maximum discount percentage.
	// Note: This can cause later discounts to be silently ignored if cap is reached.
	if discountPercent > maxDiscountPercent {
		discountPercent = maxDiscountPercent
	}

	return discountPercent
}

// shouldApplyFreeShipping determines if free shipping applies based on customer type and order subtotal.
// Free shipping is offered to high-value customers or when promotion coupons are redeemed.
//
// Returns true if any free shipping condition is met:
//   - FREESHIP coupon with discounted subtotal >= 8000 cents
//   - VIP customer with discounted subtotal >= 15000 cents
//   - Premium customer with discounted subtotal >= 20000 cents
//   - Partner customer with discounted subtotal >= 15000 cents
func shouldApplyFreeShipping(discountedSubtotal int, customerType CustomerType, coupon string) bool {
	if coupon == couponFreeShip && discountedSubtotal >= freeShipMinDiscountedSubtotal {
		return true
	}
	if customerType == CustomerTypeVip && discountedSubtotal >= vipFreeShippingMinDiscountedTotal {
		return true
	}
	if customerType == CustomerTypePremium && discountedSubtotal >= premiumFreeShippingMinSubtotal {
		return true
	}
	if customerType == CustomerTypePartner && discountedSubtotal >= partnerFreeShippingMinSubtotal {
		return true
	}
	return false
}

// calculateShippingCents computes shipping charges for an order.
//
// Parameters:
//   - discountedSubtotal: Subtotal after discounts have been applied (in cents)
//   - customerType: Customer tier determining shipping eligibility
//   - country: Destination country affecting base rates and surcharges
//   - coupon: Coupon code that may waive shipping
//   - blackFriday: True if Black Friday promotion applies (US surcharge)
//
// Returns: Shipping cost in cents.
//
// Calculation steps:
//  1. Determine base shipping by country (IT: 700, DE: 900, US: 1500, others: 2500)
//  2. Add Black Friday US surcharge (+300 cents if applicable)
//  3. Apply free shipping if customer qualifies (overrides base shipping)
//  4. Add employee surcharge in non-IT countries (+500 cents, overrides free shipping)
//
// Quirk: Employee surcharge overrides free shipping entirely. Employees always pay minimum
// shipping (base + surcharge) in DE/US, even if they qualified for free shipping.
func calculateShippingCents(discountedSubtotal int, customerType CustomerType, country Country, coupon string, blackFriday bool) int {
	// Step 1: Set base shipping rate by destination country.
	shippingCents := defaultShippingCents
	switch country {
	case CountryIT:
		shippingCents = shippingITCents
	case CountryDE:
		shippingCents = shippingDECents
	case CountryUS:
		shippingCents = shippingUSCents
	}

	// Step 2: Apply Black Friday promotion surcharge for US orders.
	if blackFriday && country == CountryUS {
		shippingCents += blackFridayUSShippingSurcharge
	}

	// Step 3: Apply free shipping if customer qualifies.
	if shouldApplyFreeShipping(discountedSubtotal, customerType, coupon) {
		shippingCents = 0
	}

	// Step 4: Apply employee surcharge in non-IT countries (overrides free shipping).
	// This ensures employees always contribute to shipping costs in those regions.
	if customerType == CustomerTypeEmployee && country != CountryIT {
		shippingCents += employeeNonITShippingSurcharge
	}

	return shippingCents
}

// calculateTaxPercent computes the applicable tax percentage for an order.
//
// Parameters:
//   - customerType: Customer tier (VIP may get tax reductions in specific countries)
//   - country: Destination country (determines base tax rate)
//   - coupon: Coupon code that may waive taxes
//
// Returns: Tax percentage (0-22).
//
// Rules (applied in order):
//  1. Base tax by country: IT(22%), DE(19%), US(7%), others(0%)
//  2. VIP override in Italy: VIP customers pay 20% instead of 22%
//  3. TAXFREE coupon: Applies only outside Italy, setting tax to 0%
//
// Quirk: TAXFREE coupon does not apply to Italy residents, even if they have VIP status.
// This preserves Italy's minimum tax collection for VIP customers.
func calculateTaxPercent(customerType CustomerType, country Country, coupon string) int {
	// Step 1: Set base tax rate by destination country.
	taxPercent := 0
	switch country {
	case CountryIT:
		taxPercent = taxITPercent
	case CountryDE:
		taxPercent = taxDEPercent
	case CountryUS:
		taxPercent = taxUSPercent
	}

	// Step 2: Apply VIP tax reduction in Italy only.
	// Quirk: This override happens after country-based tax is set, reducing IT tax for VIP.
	if customerType == CustomerTypeVip && country == CountryIT {
		taxPercent = vipTaxInITPercent
	}

	// Step 3: Apply TAXFREE coupon (non-Italy countries only).
	// Quirk: Italy residents cannot use this coupon, ensuring base tax collection.
	if coupon == couponTaxFree && country != CountryIT {
		taxPercent = 0
	}

	return taxPercent
}

// parseCustomerType normalizes and converts user input to a strongly-typed CustomerType.
// Whitespace is trimmed automatically.
//
// Quirk: Unknown customer types do not error; they fall back to 0% base discount and standard rules.
// Example: parseCustomerType("  acme  ") returns CustomerType("acme"), which matches no case
// and results in 0% customer discount.
func parseCustomerType(value string) CustomerType {
	return CustomerType(safe(value))
}

// parseCountry normalizes and converts user input to a strongly-typed Country.
// Whitespace is trimmed automatically.
//
// Quirk: Unknown countries do not error; they fall back to default shipping (2500 cents) and 0% tax.
// Example: parseCountry("  FR  ") returns Country("FR"), which matches no case and uses defaults.
func parseCountry(value string) Country {
	return Country(safe(value))
}

// safe trims leading and trailing whitespace from input strings to enable consistent comparisons.
func safe(value string) string {
	return strings.TrimSpace(value)
}
