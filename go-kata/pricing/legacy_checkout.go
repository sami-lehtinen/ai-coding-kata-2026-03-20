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
		discountPercent += 15
	case CustomerTypePremium:
		if subtotal >= 10000 {
			discountPercent += 10
		} else {
			discountPercent += 5
		}
	case CustomerTypeEmployee:
		discountPercent += 30
	}

	switch coupon {
	case "SAVE10":
		if subtotal >= 5000 {
			discountPercent += 10
		}
	case "VIPONLY":
		if customerType == CustomerTypeVip {
			discountPercent += 5
		}
	case "BULK":
		if subtotal >= 20000 {
			discountPercent += 7
		}
	}

	if blackFriday {
		if customerType != CustomerTypeEmployee {
			discountPercent += 5
		}
	}

	if discountPercent > 40 {
		discountPercent = 40
	}

	return discountPercent
}

func calculateShippingCents(discountedSubtotal int, customerType CustomerType, country Country, coupon string, blackFriday bool) int {
	shippingCents := 2500
	switch country {
	case CountryIT:
		shippingCents = 700
	case CountryDE:
		shippingCents = 900
	case CountryUS:
		shippingCents = 1500
	}

	if blackFriday && country == CountryUS {
		shippingCents += 300
	}

	if coupon == "FREESHIP" && discountedSubtotal >= 8000 {
		shippingCents = 0
	}

	if customerType == CustomerTypeVip && discountedSubtotal >= 15000 {
		shippingCents = 0
	}

	if customerType == CustomerTypePremium && discountedSubtotal >= 20000 {
		shippingCents = 0
	}

	if customerType == CustomerTypeEmployee && country != CountryIT {
		shippingCents += 500
	}

	return shippingCents
}

func calculateTaxPercent(customerType CustomerType, country Country, coupon string) int {
	taxPercent := 0
	switch country {
	case CountryIT:
		taxPercent = 22
	case CountryDE:
		taxPercent = 19
	case CountryUS:
		taxPercent = 7
	}

	if customerType == CustomerTypeVip && country == CountryIT {
		taxPercent = 20
	}

	if coupon == "TAXFREE" && country != CountryIT {
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
