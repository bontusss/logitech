package service

import "strings"

// StandardItemCategories is the canonical catalogue shown to providers and
// customers. Stored delivery categories use these values.
var StandardItemCategories = []string{
	"cloth", "electronics", "document", "food", "jewelry", "other",
}

// NormalizeItemCategory maps common spellings to a canonical category. Unknown
// values are returned trimmed and lower-cased so they can still be rejected
// consistently.
func NormalizeItemCategory(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "clothing", "clothes", "fashion", "cloth":
		return "cloth"
	case "document", "documents", "paperwork":
		return "document"
	case "electronic", "electronics", "device", "devices":
		return "electronics"
	case "food", "foods", "groceries":
		return "food"
	case "jewelry", "jewellery", "jewelries":
		return "jewelry"
	case "parcel", "parcels", "package", "packages", "fragile", "bulk", "other", "":
		return "other"
	default:
		return strings.ToLower(strings.TrimSpace(value))
	}
}

// ItemCategoryAllowed applies a provider's acceptable and prohibited lists.
// Prohibited always wins. An empty acceptable list means the provider has not
// restricted categories, so everything not prohibited is allowed.
func ItemCategoryAllowed(category string, accepted, prohibited []string) bool {
	normalized := NormalizeItemCategory(category)
	if normalized == "" {
		return false
	}
	for _, value := range prohibited {
		if NormalizeItemCategory(value) == normalized {
			return false
		}
	}
	if len(accepted) == 0 {
		return true
	}
	for _, value := range accepted {
		if NormalizeItemCategory(value) == normalized {
			return true
		}
	}
	return false
}

// NormalizeItemCategoryList validates and de-duplicates a provider policy.
func NormalizeItemCategoryList(values []string) ([]string, bool) {
	seen := make(map[string]bool, len(values))
	result := make([]string, 0, len(values))
	for _, value := range values {
		category := NormalizeItemCategory(value)
		if !IsStandardItemCategory(category) {
			return nil, false
		}
		if seen[category] {
			continue
		}
		seen[category] = true
		result = append(result, category)
	}
	return result, true
}

func IsStandardItemCategory(value string) bool {
	for _, category := range StandardItemCategories {
		if category == value {
			return true
		}
	}
	return false
}
