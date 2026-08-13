/**
 * Mapping des noms de pays Discogs vers codes ISO-3166-1 alpha-2
 * 
 * Ce fichier contient les correspondances entre les noms de pays
 * retournés par l'API Discogs et leurs codes ISO à 2 lettres.
 */

export const countryMapping = {
  // Pays européens
  'Germany': 'DE',
  'France': 'FR',
  'UK': 'GB',
  'United Kingdom': 'GB',
  'Italy': 'IT',
  'Spain': 'ES',
  'Netherlands': 'NL',
  'Belgium': 'BE',
  'Switzerland': 'CH',
  'Austria': 'AT',
  'Sweden': 'SE',
  'Norway': 'NO',
  'Denmark': 'DK',
  'Finland': 'FI',
  'Poland': 'PL',
  'Czech Republic': 'CZ',
  'Hungary': 'HU',
  'Portugal': 'PT',
  'Greece': 'GR',
  'Ireland': 'IE',
  'Luxembourg': 'LU',
  'Iceland': 'IS',
  'Croatia': 'HR',
  'Slovenia': 'SI',
  'Slovakia': 'SK',
  'Romania': 'RO',
  'Bulgaria': 'BG',
  'Serbia': 'RS',
  'Ukraine': 'UA',
  'Russia': 'RU',

  // Amérique du Nord
  'US': 'US',
  'USA': 'US',
  'United States': 'US',
  'United States of America': 'US',
  'Canada': 'CA',
  'Mexico': 'MX',

  // Amérique du Sud
  'Brazil': 'BR',
  'Argentina': 'AR',
  'Chile': 'CL',
  'Colombia': 'CO',
  'Peru': 'PE',
  'Venezuela': 'VE',

  // Asie
  'Japan': 'JP',
  'China': 'CN',
  'South Korea': 'KR',
  'Korea': 'KR',
  'Taiwan': 'TW',
  'Hong Kong': 'HK',
  'Singapore': 'SG',
  'Thailand': 'TH',
  'Malaysia': 'MY',
  'Indonesia': 'ID',
  'Philippines': 'PH',
  'India': 'IN',
  'Israel': 'IL',
  'Turkey': 'TR',

  // Océanie
  'Australia': 'AU',
  'New Zealand': 'NZ',

  // Afrique
  'South Africa': 'ZA',
  'Egypt': 'EG',
  'Morocco': 'MA',

  // ✅ Zones géographiques (à conserver car présentes dans votre base)
  'Europe': 'EU',       // ID 17 → pressages multi-pays
  'Worldwide': 'WW',    // ID 20 → distribution globale
  'European Union': 'EU',

  // ✅ Cas Discogs fréquents avec variantes
  'EU': 'EU',
  'WW': 'WW',
  'Czech': 'CZ',
  'Czechoslovakia': 'CZ',
  'U.S.A.': 'US',
  'U.S.': 'US',
  'Great Britain': 'GB',
  'Holland': 'NL',
  'Deutschland': 'DE',
  'Deutsch': 'DE',
  'Allemagne': 'FR',
  'Espagne': 'ES',
  'Italie': 'IT',
  'Belgique': 'BE',
  'Suisse': 'CH',
  'Österreich': 'AT',
  'Suède': 'SE',
  'Norvège': 'NO',
  'Danemark': 'DK',
  'Finlande': 'FI',
  'Pologne': 'PL',
  'Hongrie': 'HU',
  'Portugal': 'PT',
  'Grèce': 'GR',
  'Irlande': 'IE',
  'Luxembourg': 'LU',
  'Islande': 'IS',
  'Croatie': 'HR',
  'Slovénie': 'SI',
  'Slovaquie': 'SK',
  'Roumanie': 'RO',
  'Bulgarie': 'BG',
  'Serbie': 'RS',
  'Ukraine': 'UA',
  'Russie': 'RU',
  'Canada (Québec)': 'CA',
  'Mexique': 'MX',
  'Brésil': 'BR',
  'Argentine': 'AR',
  'Chili': 'CL',
  'Colombie': 'CO',
  'Pérou': 'PE',
  'Vénézuela': 'VE',
  'Japon': 'JP',
  'Chine': 'CN',
  'Corée du Sud': 'KR',
  'Taïwan': 'TW',
  'Singapour': 'SG',
  'Thaïlande': 'TH',
  'Malaisie': 'MY',
  'Indonésie': 'ID',
  'Philippines': 'PH',
  'Inde': 'IN',
  'Israël': 'IL',
  'Turquie': 'TR',
  'Afrique du Sud': 'ZA',
  'Égypte': 'EG',
  'Maroc': 'MA',
  'Australie': 'AU',
  'Nouvelle-Zélande': 'NZ',

  // ✅ Valeurs spéciales
  'Unknown': 'XX',
  'Unknown Country': 'XX',
  'N/A': 'XX',
  '': 'XX',
  '—': 'XX',
  '-': 'XX'
};

/**
 * Nettoie et normalise un nom de pays pour la comparaison
 * @param {string} str
 * @returns {string}
 */
const normalizeForMatch = (str) => {
  if (!str) return '';
  return str
    .trim()
    .toLowerCase()
    .replace(/\s+/g, ' ')           // espaces multiples → 1
    .replace(/[.,\-()]/g, '')       // supprime . , - ( )
    .replace(/’/g, "'")              // guillemet courbe → apostrophe droite
    ;

};

/**
 * Convertit un nom de pays Discogs en code ISO
 * @param {string} countryName - Nom du pays retourné par Discogs
 * @param {boolean} [debug=false] - Active les logs de debug
 * @returns {string|null} Code ISO à 2 lettres ou null si non trouvé
 */
export const getCountryCode = (countryName, debug = false) => {
  if (!countryName || typeof countryName !== 'string') {
    debug && console.warn(`[countryMapping] Entrée invalide:`, countryName);
    return 'XX';
  }

  const raw = countryName;
  const normalized = normalizeForMatch(raw);

  if (debug) console.log(`[countryMapping] Input: "${raw}" → normalized: "${normalized}"`);

  // 1. Recherche exacte (normalisée)
  const exactMatch = Object.keys(countryMapping).find(
    key => normalizeForMatch(key) === normalized
  );

  if (exactMatch) {
    debug && console.log(`[countryMapping] ✅ Exact match: "${exactMatch}" → "${countryMapping[exactMatch]}"`);
    return countryMapping[exactMatch];
  }

  // 2. Recherche partielle (substring) — priorité aux pays réels (pas EU/WW/XX)
  const prioritizedKeys = Object.keys(countryMapping).sort((a, b) => {
    // Mettre EU/WW/XX à la fin
    const aIsSpecial = ['EU', 'WW', 'XX'].includes(countryMapping[a]);
    const bIsSpecial = ['EU', 'WW', 'XX'].includes(countryMapping[b]);
    if (aIsSpecial && !bIsSpecial) return 1;
    if (!aIsSpecial && bIsSpecial) return -1;
    return 0;
  });

  const partialMatch = prioritizedKeys.find(
    key => normalized.includes(normalizeForMatch(key))
  );

  if (partialMatch) {
    const code = countryMapping[partialMatch];
    debug && console.log(`[countryMapping] ✅ Partial match: "${raw}" contains "${partialMatch}" → "${code}"`);
    return code;
  }

  // 3. Fallback : essayer de matcher via code ISO direct (ex: 'FR' → 'FR')
  if (/^[A-Z]{2}$/.test(normalized.toUpperCase()) && isValidCountryCode(normalized)) {
    debug && console.log(`[countryMapping] ✅ Code ISO détecté: "${normalized.toUpperCase()}"`);
    return normalized.toUpperCase();
  }

  // 4. Non trouvé → XX
  debug && console.warn(`[countryMapping] ❌ Aucun match pour: "${raw}"`);
  return 'XX';
};

/**
 * Vérifie si un code pays est valide
 * @param {string} code - Code ISO à vérifier
 * @returns {boolean}
 */
export const isValidCountryCode = (code) => {
  if (!code || typeof code !== 'string') return false;
  const upper = code.toUpperCase();
  return Object.values(countryMapping).includes(upper);
};

// ✅ Utilitaire pour vérifier la couverture de votre base
export const debugMissingCountries = (countryNames = []) => {
  const missing = [];
  const found = [];
  countryNames.forEach(name => {
    const code = getCountryCode(name);
    if (code === 'XX') {
      missing.push(name);
    } else {
      found.push({ name, code });
    }
  });
  console.table([
    { 'Total': countryNames.length },
    { 'Trouvés': found.length },
    { 'Manquants': missing.length }
  ]);
  if (missing.length > 0) {
    console.warn('[countryMapping] Pays non reconnus:', missing);
  }
  return { found, missing };
};