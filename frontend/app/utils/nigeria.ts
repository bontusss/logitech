export const NIGERIAN_LOCATIONS: Record<string, string[]> = {
  Abia: ['Aba', 'Arochukwu', 'Bende', 'Ohafia', 'Umuahia'],
  Adamawa: ['Ganye', 'Girei', 'Jimeta', 'Mubi', 'Numan', 'Yola'],
  'Akwa Ibom': ['Eket', 'Ikot Ekpene', 'Oron', 'Uyo'],
  Anambra: ['Awka', 'Ekwulobia', 'Nnewi', 'Onitsha'],
  Bauchi: ['Azare', 'Bauchi', 'Jamaare', 'Misau'],
  Bayelsa: ['Brass', 'Nembe', 'Ogbia', 'Yenagoa'],
  Benue: ['Gboko', 'Katsina-Ala', 'Makurdi', 'Otukpo'],
  Borno: ['Bama', 'Biu', 'Dikwa', 'Maiduguri'],
  'Cross River': ['Calabar', 'Ikom', 'Ogoja', 'Ugep'],
  Delta: ['Abraka', 'Agbor', 'Asaba', 'Effurun', 'Sapele', 'Warri'],
  Ebonyi: ['Abakaliki', 'Afikpo', 'Ikwo', 'Onueke'],
  Edo: ['Auchi', 'Benin City', 'Ekpoma', 'Irrua'],
  Ekiti: ['Ado-Ekiti', 'Ikere-Ekiti', 'Ijero-Ekiti', 'Oye-Ekiti'],
  Enugu: ['Enugu', 'Nsukka', 'Oji River', 'Udi'],
  FCT: ['Abuja', 'Bwari', 'Gwagwalada', 'Kuje', 'Kubwa'],
  Gombe: ['Billiri', 'Deba', 'Gombe', 'Kaltungo'],
  Imo: ['Ideato', 'Mbaise', 'Oguta', 'Orlu', 'Owerri'],
  Jigawa: ['Dutse', 'Gumel', 'Hadejia', 'Kazaure'],
  Kaduna: ['Birnin Gwari', 'Kaduna', 'Kafanchan', 'Zaria'],
  Kano: ['Bichi', 'Gaya', 'Kano', 'Rano'],
  Katsina: ['Daura', 'Funtua', 'Katsina', 'Malumfashi'],
  Kebbi: ['Argungu', 'Birnin Kebbi', 'Jega', 'Yauri'],
  Kogi: ['Anyigba', 'Idah', 'Kabba', 'Lokoja'],
  Kwara: ['Ilorin', 'Jebba', 'Offa', 'Omu-Aran'],
  Lagos: ['Badagry', 'Epe', 'Ikorodu', 'Ikeja', 'Lagos Island', 'Lekki'],
  Nasarawa: ['Akwanga', 'Keffi', 'Lafia', 'Nasarawa'],
  Niger: ['Bida', 'Kontagora', 'Minna', 'Suleja'],
  Ogun: ['Abeokuta', 'Ijebu Ode', 'Ota', 'Sagamu'],
  Ondo: ['Akure', 'Ikare-Akoko', 'Ondo', 'Owo'],
  Osun: ['Ede', 'Ife', 'Ilesa', 'Osogbo'],
  Oyo: ['Ibadan', 'Iseyin', 'Ogbomoso', 'Oyo'],
  Plateau: ['Bukuru', 'Jos', 'Pankshin', 'Shendam'],
  Rivers: ['Bonny', 'Obio-Akpor', 'Port Harcourt', 'Okrika'],
  Sokoto: ['Gwadabawa', 'Illela', 'Sokoto', 'Tambuwal'],
  Taraba: ['Bali', 'Gembu', 'Jalingo', 'Takum', 'Wukari'],
  Yobe: ['Damaturu', 'Gashua', 'Nguru', 'Potiskum'],
  Zamfara: ['Anka', 'Gusau', 'Kaura Namoda', 'Talata Mafara'],
}

export const NIGERIAN_STATES = Object.keys(NIGERIAN_LOCATIONS).sort((a, b) =>
  a.localeCompare(b),
)

export function normalizeNigerianLocalPhone(value: string) {
  let digits = value.replace(/\D/g, '')
  if (digits.startsWith('234')) digits = digits.slice(3)
  if (digits.startsWith('0')) digits = digits.slice(1)
  return digits.slice(0, 10)
}

export function isValidNigerianPhone(value: string) {
  return /^[789]\d{9}$/.test(normalizeNigerianLocalPhone(value))
}
