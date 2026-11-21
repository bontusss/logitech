export interface ItemCategoryOption {
  value: string
  label: string
  emoji: string
  hint: string
}

export const ITEM_CATEGORIES: ItemCategoryOption[] = [
  { value: 'cloth', label: 'Clothing', emoji: '👕', hint: 'Fashion & fabric' },
  { value: 'electronics', label: 'Electronics', emoji: '🎧', hint: 'Devices & accessories' },
  { value: 'document', label: 'Documents', emoji: '📄', hint: 'Paperwork & files' },
  { value: 'food', label: 'Food', emoji: '🍱', hint: 'Meals & groceries' },
  { value: 'jewelry', label: 'Jewelry', emoji: '✨', hint: 'Watches & valuables' },
  { value: 'other', label: 'Other', emoji: '📦', hint: 'Something else' },
]

export function itemCategoryLabel(value: string) {
  return ITEM_CATEGORIES.find((category) => category.value === value)?.label ?? 'Other'
}
