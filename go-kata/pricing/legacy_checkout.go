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

	if customerType == CustomerTypeVip && country == CountryIT {
		taxPercent = vipTaxInITPercent
	}

	if coupon == couponTaxFree && country != CountryIT {
		taxPercent = 0
	}

	return taxPercent
}

func parseCustomerType(value string) CustomerType {
	return CustomerType(safe(value))
}

func parseCountry(value string) Country {
	return Country(safe(value))
}

func safe(value string) string {
	return strings.TrimSpace(value)
}
