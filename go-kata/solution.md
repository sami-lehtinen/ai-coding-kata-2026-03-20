# Solution: Legacy Checkout Refactoring

## What Was Wrong in the Legacy Design

The original `CalculateTotalCents()` function was a monolithic 100+ line function with significant architectural problems:

### 1. **Deeply Nested Conditionals**

- Multiple levels of if/else if chains for customer type, coupon, and country logic
- Made the code difficult to scan and reason about
- Hidden dependencies between conditions were hard to spot

### 2. **Mixed Responsibilities**

- Single function handled discounts, shipping, taxes, and edge cases all at once
- Made it hard to change one aspect without risk of breaking others
- Testing required exercising the entire calculation path

### 3. **Duplicate Business Rules**

- Conditional chains repeated logic patterns (e.g., checking `country == "IT"` multiple times)
- Threshold checks scattered throughout (5000, 8000, 10000, 15000, 20000 cents)
- Easy to introduce inconsistencies when rules evolved

### 4. **Magic Numbers and Strings**

- Hard-coded thresholds (5000, 8000, 10000, 15000, 20000) appeared multiple times
- Percentages (5, 7, 10, 15, 30, 40) were unexplained
- Coupon names ("SAVE10", "VIPONLY", etc.) appeared as literal strings in conditions

### 5. **Hidden Quirks**

- Employee surcharge overriding free shipping was implemented but not documented
- VIP tax reduction in Italy was a subtle edge case
- TAXFREE coupon not applying to Italy was counterintuitive
- No comments explaining why these behaviors existed

### 6. **Poor Extensibility**

- Adding a new customer type required editing the discount calculation, shipping calculation, and tax calculation separately
- Risk of missing one location and introducing bugs
- No clear pattern to follow when adding new rules

---

## What Changed

### 1. **Extracted Helper Functions**

Created three focused functions to separate concerns:

- **`calculateDiscountPercent()`** — handles all discount logic (customer type, coupons, Black Friday cap)
- **`calculateShippingCents()`** — handles shipping logic (country rates, surcharges, free shipping)
- **`calculateTaxPercent()`** — handles tax logic (country rates, customer overrides, coupon exemptions)
- **`shouldApplyFreeShipping()`** — centralized free shipping eligibility logic

### 2. **Introduced Typed Enums**

- `CustomerType` constant type with allowed values: regular, new, vip, premium, employee, partner
- `Country` constant type with allowed values: IT, DE, US
- Enables type safety without external dependencies
- Graceful fallback for unknown values (preserves original behavior)

### 3. **Named Constants for All Magic Values**

Organized into semantic groups:

**Discount Rules:**

- `partnerBaseDiscountPercent = 12`
- `vipBaseDiscountPercent = 15`
- `employeeBaseDiscountPercent = 30`
- `maxDiscountPercent = 40`
- Plus coupon percentages and thresholds

**Shipping Rules:**

- Country base rates: `shippingITCents = 700`, `shippingDECents = 900`, `shippingUSCents = 1500`
- Free shipping thresholds: `vipFreeShippingMinDiscountedTotal = 15000`
- Surcharges: `blackFridayUSShippingSurcharge = 300`

**Tax Rates:**

- `taxITPercent = 22`, `taxDEPercent = 19`, `taxUSPercent = 7`
- VIP override: `vipTaxInITPercent = 20`

### 4. **Replaced Nested If/Else with Switch Statements**

- Switch statements are more readable than nested if/else chains
- Easier to spot missing cases
- Clear fall-through pattern and default behavior

### 5. **Comprehensive Documentation**

- Go doc comments for all public and helper functions
- Parameter and return value documentation
- Inline comments explaining quirks and non-obvious behaviors
- Examples showing how edge cases are handled

### 6. **Successfully Added Partner Customer Type**

- Added `CustomerTypePartner` to enum
- Added partner base discount (12%) to discount calculation
- Added PARTNER5 coupon with conditional threshold
- Added partner Black Friday rule (3% instead of 5%)
- Added partner free shipping threshold (15000 cents discounted subtotal)
- All integrated without modifying `CalculateTotalCents()`
- 7 new tests verify correct behavior

---

## Why the New Structure Is Easier to Extend

### 1. **Adding a New Customer Type**

**Before:** Modify discount calculation, shipping calculation, and tax calculation separately. High risk of omissions.

**After:**

1. Add constant to `CustomerType`
2. Add one case to discount switch
3. Add one case to shipping switch (if needed)
4. Add one case to tax switch (if needed)
5. Write tests

Clear, systematic pattern. Demonstrated by successfully adding `partner` type.

### 2. **Adding a New Coupon**

**Before:** Add another `else if` branch in the coupon section, potentially duplicate threshold checks.

**After:**

1. Add constant for coupon code
2. Add one case to discount switch with threshold check
3. Optionally add to tax switch if coupon affects taxes
4. Write tests

Pattern is uniform and easy to follow.

### 3. **Changing a Threshold or Rule**

**Before:** Hunt through the code for all places using `5000` or `15000`, change carefully.

**After:** Change one named constant. All usages automatically update.

### 4. **Understanding the Business Logic**

**Before:** Read 100+ lines of nested conditionals, mentally track variable state.

**After:** Read focused helper functions with clear names and documentation. Quirks are documented inline.

### 5. **Testing New Features**

**Before:** Required understanding and potentially replicating existing calculation steps.

**After:** Can write unit-style tests for each helper function independently. Reuse existing test patterns.

---

## AI Suggestion That Was Rejected and Why

### **Rejected Suggestion: Strict Enum Validation with Panics**

**What Was Proposed:**
The AI suggested adding strict validation that would panic if an unsupported `CustomerType` or `Country` was provided:

```go
func parseCustomerType(value string) CustomerType {
    clean := CustomerType(safe(value))
    switch clean {
    case CustomerTypeRegular, CustomerTypeNew, ...:
        return clean
    default:
        panic(fmt.Sprintf("unsupported customer type: %q", value))
    }
}
```

**Why It Was Rejected:**

1. **Broke Backward Compatibility** — The original code had graceful fallback behavior:
   - Unknown customer type → 0% base discount (treated as "regular"/"new")
   - Unknown country → default shipping (2500 cents) + 0% tax
2. **Existing Test Relied on Fallback** — Test case `"unknown country uses default shipping and no tax"` with customer type `"mystery"` and country `"FR"` expected a specific total (12500), not a panic.

3. **Constraint Violation** — README explicitly stated: "Preserve existing behavior unless the new requirement explicitly changes it." Changing silent fallback to loud panic violated this.

4. **Pragmatic Trade-off** — While strict validation is good practice, the kata prioritizes preserving behavior and incremental change over adopting "best practices" that break compatibility.

**Decision:** Reverted to permissive parsing that silently casts unknown values to the enum type. This preserves the original legacy behavior while still benefiting from enum typing and switch-based clarity.

---

## Summary

The refactoring successfully transformed a unmaintainable 100+ line monolithic function into a well-organized, documented, and extensible checkout calculator. The helper functions separate concerns, named constants eliminate magic numbers, and switch statements replace nested conditionals with clear patterns.

The new partner customer type was added cleanly without modifying the core calculation logic, demonstrating the improved extensibility. The solution maintains 100% backward compatibility while setting up a foundation for future growth.
