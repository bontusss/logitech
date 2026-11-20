export type ProviderBusinessType = 'intercity' | 'lastmile'

export const providerBusinessTypes: Array<{ value: ProviderBusinessType; label: string }> = [
  { value: 'intercity', label: 'Intercity' },
  { value: 'lastmile', label: 'Last mile' },
]

export function parseProviderBusinessTypes(value: string): ProviderBusinessType[] {
  const normalized = value.toLowerCase()
  const types = new Set<ProviderBusinessType>()

  if (normalized.includes('intercity')) types.add('intercity')
  if (normalized.includes('lastmile') || normalized.includes('last mile')) types.add('lastmile')

  return providerBusinessTypes
    .map((type) => type.value)
    .filter((type) => types.has(type))
}

export function serializeProviderBusinessTypes(types: ProviderBusinessType[]): string {
  return types.join(',')
}
