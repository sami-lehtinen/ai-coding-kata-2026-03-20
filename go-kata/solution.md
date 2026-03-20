# Solution: Legacy Checkout Refactoring

## What Was Wrong in the Legacy Design

The original `CalculateTotalCents()` function was a monolithic 100+ line function with significant architectural problems:

1. Deeply Nested Conditionals
2. Mixed Responsibilities
3. Duplicate Business Rules
4. Magic Numbers and Strings
5. Hidden Quirks
6. Poor Extensibility

## What Changed

1. Extracted Helper Functions
   - `calculateDiscountPercent()`
   - `calculateShippingCents()`
   - `calculateTaxPercent()`
   - `shouldApplyFreeShipping()`
2. Introduced Typed Enums
3. Named Constants for All Magic Values
4. Replaced Nested If/Else with Switch Statements
5. Comprehensive Documentation
6. Successfully Added Partner Customer Type

---

## Why the New Structure Is Easier to Extend

- Business logic is easier to understand
- Enums and consts are used -> New enums can be easily added
- Clear responsibility for each function
- Code is documented
- Code is testable

## AI Suggestion That Was Rejected and Why

### Rejected Suggestion: Strict Enum Validation with Panics

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

Broke Backward Compatibility
