/** Curated preset of common proxy/VPN server locations, so admins can pick a
 * location from a dropdown instead of typing a free-text region + ISO code
 * by hand. Selecting a preset sets both the display label and country_code
 * used for the flag shown on the Nodes page. */
export interface ServerLocationPreset {
  code: string; // ISO 3166-1 alpha-2
  city: string;
  label: string; // "Frankfurt, DE"
}

export const SERVER_LOCATIONS: ServerLocationPreset[] = [
  { code: "DE", city: "Frankfurt", label: "Frankfurt, DE" },
  { code: "NL", city: "Amsterdam", label: "Amsterdam, NL" },
  { code: "GB", city: "London", label: "London, GB" },
  { code: "FR", city: "Paris", label: "Paris, FR" },
  { code: "US", city: "New York", label: "New York, US" },
  { code: "US", city: "Los Angeles", label: "Los Angeles, US" },
  { code: "CA", city: "Toronto", label: "Toronto, CA" },
  { code: "JP", city: "Tokyo", label: "Tokyo, JP" },
  { code: "KR", city: "Seoul", label: "Seoul, KR" },
  { code: "SG", city: "Singapore", label: "Singapore, SG" },
  { code: "HK", city: "Hong Kong", label: "Hong Kong, HK" },
  { code: "AU", city: "Sydney", label: "Sydney, AU" },
  { code: "IN", city: "Mumbai", label: "Mumbai, IN" },
  { code: "TR", city: "Istanbul", label: "Istanbul, TR" },
  { code: "AE", city: "Dubai", label: "Dubai, AE" },
  { code: "IR", city: "Tehran", label: "Tehran, IR" },
  { code: "RU", city: "Moscow", label: "Moscow, RU" },
  { code: "BR", city: "São Paulo", label: "São Paulo, BR" },
  { code: "ZA", city: "Johannesburg", label: "Johannesburg, ZA" },
  { code: "IT", city: "Milan", label: "Milan, IT" },
  { code: "ES", city: "Madrid", label: "Madrid, ES" },
  { code: "PL", city: "Warsaw", label: "Warsaw, PL" },
  { code: "SE", city: "Stockholm", label: "Stockholm, SE" },
  { code: "FI", city: "Helsinki", label: "Helsinki, FI" },
  { code: "CH", city: "Zurich", label: "Zurich, CH" },
  { code: "UA", city: "Kyiv", label: "Kyiv, UA" },
  { code: "PK", city: "Karachi", label: "Karachi, PK" },
  { code: "ID", city: "Jakarta", label: "Jakarta, ID" },
  { code: "VN", city: "Ho Chi Minh City", label: "Ho Chi Minh City, VN" },
  { code: "TH", city: "Bangkok", label: "Bangkok, TH" },
];

export function findLocationPreset(region?: string, countryCode?: string): ServerLocationPreset | undefined {
  if (!region && !countryCode) return undefined;
  return SERVER_LOCATIONS.find(
    (p) => p.label === region || (countryCode && p.code === countryCode.toUpperCase() && p.city === region),
  );
}
