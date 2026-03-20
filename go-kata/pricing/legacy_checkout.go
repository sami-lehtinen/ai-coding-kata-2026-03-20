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
)

type Country string

const (
	CountryIT Country = "IT"
	CountryDE Country = "DE"
	CountryUS Country = "US"
)

const (
	couponSave10   = "SAVE10"
	couponVipOnly  = "VIPONLY"
	couponBulk     = "BULK"
	couponFreeShip = "FREESHIP"
	couponTaxFree  = "TAXFREE"
)

const (
	maxDiscountPercent          = 40
	vipBaseDiscountPercent      = 15
	premiumHighBaseDiscountPct  = 10
	premiumLowBaseDiscountPct   = 5
	employeeBaseDiscountPercent = 30
	blackFridayExtraDiscountPct = 5
	save10DiscountPercent       = 10
	vipOnlyDiscountPercent      = 5
	bulkDiscountPercent         = 7
	save10MinSubtotalCents      = 5000
	bulkMinSubtotalCents        = 20000
	premiumBaseTierMinSubtotal  = 10000
)

const (
	defaultShippingCents              = 2500
	shippingITCents                   = 700
	shippingDECents                   = 900
	shippingUSCents                   = 1500
	blackFridayUSShippingSurcharge    = 300
	freeShipMinDiscountedSubtotal     = 8000
	vipFreeShippingMinDiscountedTotal = 15000
	premiumFreeShippingMinSubtotal    = 20000
	employeeNonITShippingSurcharge    = 500
)

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

// CalculateTotalCents calculates the total price for an order including discounts, shipping, and taxes.
// Quirk: Unknown customer types and countries fall back gracefully (0% base discount, default shipping, 0% tax).
// Quirk: Employee surcharge (non-IT countries) overrides free shipping conditions entirely—it is applied
// after free shipping checks, ensuring employees always pay at least the surcharge in eligible countries.
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

func calculateDiscountPercent(subtotal int, customerType CustomerType, coupon string, blackFriday bool) int {
	// Discounts accumulate additively and are capped at 40%. Order matters:
	// 1. Customer type discount (0-30%)
	// 2. Coupon discount (0-10%)
	// 3. Black Friday bonus (0-5%), except employees are excluded
	// This can lead to counter-intuitive behavior: a discount closer to 40% may ignore later additions.
	discountPercent := 0

	switch customerType {
	case CustomerTypeVip:
		discountPercent += vipBaseDiscountPercent
	case CustomerTypePremium:
		if subtotal >= premiumBaseTierMinSubtotal {
			discountPercent += premiumHighBaseDiscountPct
		} else {
			discountPercent += premiumLowBaseDiscountPct
		}
	case CustomerTypeEmployee:
		discountPercent += employeeBaseDiscountPercent
	}

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
	}

	if blackFriday {
		if customerType != CustomerTypeEmployee {
			discountPercent += blackFridayExtraDiscountPct
		}
	}

	// Hard cap: no discount exceeds 40%, even if rules would combine to more.
	if discountPercent > maxDiscountPercent {
		discountPercent = maxDiscountPercent
	}

	return discountPercent
}

func calculateShippingCents(discountedSubtotal int, customerType CustomerType, country Country, coupon string, blackFriday bool) int {
	shippingCents := defaultShippingCents
	switch country {
	case CountryIT:
		shippingCents = shippingITCents
	case CountryDE:
		shippingCents = shippingDECents
	case CountryUS:
		shippingCents = shippingUSCents
	}

	if blackFriday && country == CountryUS {
		shippingCents += blackFridayUSShippingSurcharge
	}

	if coupon == couponFreeShip && discountedSubtotal >= freeShipMinDiscountedSubtotal {
		shippingCents = 0
	}

	if customerType == CustomerTypeVip && discountedSubtotal >= vipFreeShippingMinDiscountedTotal {
		shippingCents = 0
	}

	if customerType == CustomerTypePremium && discountedSubtotal >= premiumFreeShippingMinSubtotal {
		shippingCents = 0
	}

	// Quirk: Employee surcharge is applied AFTER free shipping checks, overriding them entirely.
	// This means employees in DE/US pay the surcharge even if they qualified for free shipping.
	if customerType == CustomerTypeEmployee && country != CountryIT {
		shippingCents += employeeNonITShippingSurcharge
	}

	return shippingCents
}

func calculateTaxPercent(customerType CustomerType, country Country, coupon string) int {
	taxPercent := 0
	switch country {
	case CountryIT:
		taxPercent = taxITPercent
	case CountryDE:
		taxPercent = taxDEPercent
	case CountryUS:
		taxPercent = taxUSPercent
	}

	// Quirk: VIP customers in Italy pay reduced tax (20% instead of the default 22%).
	// This override happens after country-based tax is set.
	if customerType == CustomerTypeVip && country == CountryIT {
		taxPercent = vipTaxInITPercent
	}

	// Quirk: TAXFREE coupon only applies to non-Italy countries. Italy residents cannot use it.
	if coupon == couponTaxFree && country != CountryIT {
		taxPercent = 0
	}

	return taxPercent
}

// parseCustomerType converts user input to a CustomerType.
// Quirk: Unknown customer types do not error; they fall back to 0% base discount and standard rules.
func parseCustomerType(value string) CustomerType {
	return CustomerType(safe(value))
}

// parseCountry converts user input to a Country.
// Quirk: Unknown countries do not error; they fall back to default shipping (2500 cents) and 0% tax.
func parseCountry(value string) Country {
	return Country(safe(value))
}

// safe trims leading/trailing whitespace from input strings for comparison.
func safe(value string) string {
	return strings.TrimSpace(value)
}
