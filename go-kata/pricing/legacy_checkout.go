package pricing

import "strings"

type Order struct {
	CustomerType  string
	SubtotalCents int
	Country       string
	CouponCode    string
	BlackFriday   bool
}

func CalculateTotalCents(order Order) int {
	subtotal := order.SubtotalCents
	customerType := safe(order.CustomerType)
	country := safe(order.Country)
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

func calculateDiscountPercent(subtotal int, customerType, coupon string, blackFriday bool) int {
	discountPercent := 0

	if customerType == "vip" {
		discountPercent += 15
	} else if customerType == "premium" {
		if subtotal >= 10000 {
			discountPercent += 10
		} else {
			discountPercent += 5
		}
	} else if customerType == "employee" {
		discountPercent += 30
	}

	if coupon == "SAVE10" {
		if subtotal >= 5000 {
			discountPercent += 10
		}
	} else if coupon == "VIPONLY" {
		if customerType == "vip" {
			discountPercent += 5
		}
	} else if coupon == "BULK" {
		if subtotal >= 20000 {
			discountPercent += 7
		}
	}

	if blackFriday {
		if customerType != "employee" {
			discountPercent += 5
		}
	}

	if discountPercent > 40 {
		discountPercent = 40
	}

	return discountPercent
}

func calculateShippingCents(discountedSubtotal int, customerType, country, coupon string, blackFriday bool) int {
	shippingCents := 2500
	if country == "IT" {
		shippingCents = 700
	} else if country == "DE" {
		shippingCents = 900
	} else if country == "US" {
		shippingCents = 1500
	}

	if blackFriday && country == "US" {
		shippingCents += 300
	}

	if coupon == "FREESHIP" && discountedSubtotal >= 8000 {
		shippingCents = 0
	}

	if customerType == "vip" && discountedSubtotal >= 15000 {
		shippingCents = 0
	}

	if customerType == "premium" && discountedSubtotal >= 20000 {
		shippingCents = 0
	}

	if customerType == "employee" && country != "IT" {
		shippingCents += 500
	}

	return shippingCents
}

func calculateTaxPercent(customerType, country, coupon string) int {
	taxPercent := 0
	if country == "IT" {
		taxPercent = 22
	} else if country == "DE" {
		taxPercent = 19
	} else if country == "US" {
		taxPercent = 7
	}

	if customerType == "vip" && country == "IT" {
		taxPercent = 20
	}

	if coupon == "TAXFREE" && country != "IT" {
		taxPercent = 0
	}

	return taxPercent
}

func safe(value string) string {
	return strings.TrimSpace(value)
}
