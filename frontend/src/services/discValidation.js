// frontend/src/services/discValidation.js

/**
 * Service de validation des données de disques
 * Contient toutes les règles métier de validation
 */
export const discValidation = {
  /**
   * Valide le nom de l'artiste
   * @param {string} artist - Nom de l'artiste
   * @returns {Object} { valid: boolean, error?: string }
   */
  validateArtist(artist) {
    if (!artist || artist.trim().length === 0) {
      return { 
        valid: false, 
        error: 'L\'artiste est requis' 
      }
    }
    
    if (artist.trim().length < 2) {
      return { 
        valid: false, 
        error: 'Le nom de l\'artiste doit contenir au moins 2 caractères' 
      }
    }
    
    if (artist.length > 255) {
      return { 
        valid: false, 
        error: 'Le nom de l\'artiste est trop long (max 255 caractères)' 
      }
    }
    
    return { valid: true }
  },

  /**
   * Valide le titre de l'album
   * @param {string} title - Titre de l'album
   * @returns {Object} { valid: boolean, error?: string }
   */
  validateTitle(title) {
    if (!title || title.trim().length === 0) {
      return { 
        valid: false, 
        error: 'Le titre est requis' 
      }
    }
    
    if (title.trim().length < 2) {
      return { 
        valid: false, 
        error: 'Le titre doit contenir au moins 2 caractères' 
      }
    }
    
    if (title.length > 255) {
      return { 
        valid: false, 
        error: 'Le titre est trop long (max 255 caractères)' 
      }
    }
    
    return { valid: true }
  },

  /**
   * Valide l'année de sortie
   * @param {string|number} year - Année de sortie
   * @returns {Object} { valid: boolean, error?: string }
   */
  validateYear(year) {
    // L'année est optionnelle
    if (!year || year === '') {
      return { valid: true }
    }
    
    const yearNum = parseInt(year)
    const currentYear = new Date().getFullYear()
    
    if (isNaN(yearNum)) {
      return { 
        valid: false, 
        error: 'L\'année doit être un nombre' 
      }
    }
    
    if (yearNum < 1900) {
      return { 
        valid: false, 
        error: 'L\'année ne peut pas être antérieure à 1900' 
      }
    }
    
    if (yearNum > currentYear + 1) {
      return { 
        valid: false, 
        error: `L'année ne peut pas dépasser ${currentYear + 1}` 
      }
    }
    
    return { valid: true }
  },

  /**
   * Valide un code-barres
   * @param {string} barcode - Code-barres à valider
   * @returns {Object} { valid: boolean, error?: string }
   */
  validateBarcode(barcode) {
    // Le code-barres est optionnel
    if (!barcode || barcode === '') {
      return { valid: true }
    }
    
    const cleanBarcode = barcode.trim()
    
    // Vérifier que c'est uniquement des chiffres
    if (!/^\d+$/.test(cleanBarcode)) {
      return { 
        valid: false, 
        error: 'Le code-barres ne doit contenir que des chiffres' 
      }
    }
    
    // Vérifier la longueur (standards: EAN-8, EAN-13, UPC-A, etc.)
    if (cleanBarcode.length < 8 || cleanBarcode.length > 14) {
      return { 
        valid: false, 
        error: 'Le code-barres doit contenir entre 8 et 14 chiffres' 
      }
    }
    
    return { valid: true }
  },

  /**
   * Valide le genre musical
   * @param {string} genre - Genre musical
   * @returns {Object} { valid: boolean, error?: string }
   */
  validateGenre(genre) {
    // Le genre est optionnel
    if (!genre || genre === '') {
      return { valid: true }
    }
    
    if (genre.length > 100) {
      return { 
        valid: false, 
        error: 'Le genre est trop long (max 100 caractères)' 
      }
    }
    
    return { valid: true }
  },

  /**
   * Valide le format (vinyle, CD, etc.)
   * @param {string} format - Format du support
   * @returns {Object} { valid: boolean, error?: string }
   */
  validateFormat(format) {
    // Le format est optionnel
    if (!format || format === '') {
      return { valid: true }
    }
    
    if (format.length > 50) {
      return { 
        valid: false, 
        error: 'Le format est trop long (max 50 caractères)' 
      }
    }
    
    return { valid: true }
  },

  /**
   * Valide le label/maison de disques
   * @param {string} label - Nom du label
   * @returns {Object} { valid: boolean, error?: string }
   */
  validateLabel(label) {
    // Le label est optionnel
    if (!label || label === '') {
      return { valid: true }
    }
    
    if (label.length > 255) {
      return { 
        valid: false, 
        error: 'Le label est trop long (max 255 caractères)' 
      }
    }
    
    return { valid: true }
  },

  /**
   * Valide le pays
   * @param {string} country - Pays de sortie
   * @returns {Object} { valid: boolean, error?: string }
   */
  validateCountry(country) {
    // Le pays est optionnel
    if (!country || country === '') {
      return { valid: true }
    }
    
    if (country.length > 100) {
      return { 
        valid: false, 
        error: 'Le pays est trop long (max 100 caractères)' 
      }
    }
    
    return { valid: true }
  },

  /**
   * Valide les termes de recherche (pour recherche manuelle)
   * @param {string} title - Titre recherché
   * @param {string} artist - Artiste recherché
   * @returns {Object} { valid: boolean, error?: string }
   */
  validateSearchTerms(title, artist) {
    const titleLength = (title || '').trim().length
    const artistLength = (artist || '').trim().length
    
    // Au moins un des deux doit faire 2 caractères minimum
    if (titleLength < 2 && artistLength < 2) {
      return {
        valid: false,
        error: 'Saisissez au moins 2 caractères dans le titre ou l\'artiste'
      }
    }
    
    return { valid: true }
  },

  /**
   * Validation complète d'un disque
   * @param {Object} disc - Objet disque à valider
   * @returns {Object} { valid: boolean, errors: Object }
   */
  validateDisc(disc) {
    const errors = {}
    
    // Validation artiste (obligatoire)
    const artistValidation = this.validateArtist(disc.artist)
    if (!artistValidation.valid) {
      errors.artist = artistValidation.error
    }
    
    // Validation titre (obligatoire)
    const titleValidation = this.validateTitle(disc.title)
    if (!titleValidation.valid) {
      errors.title = titleValidation.error
    }
    
    // Validation année (optionnel mais doit être valide si fourni)
    const yearValidation = this.validateYear(disc.year)
    if (!yearValidation.valid) {
      errors.year = yearValidation.error
    }
    
    // Validation code-barres (optionnel mais doit être valide si fourni)
    const barcodeValidation = this.validateBarcode(disc.barcode)
    if (!barcodeValidation.valid) {
      errors.barcode = barcodeValidation.error
    }
    
    // Validation genre
    const genreValidation = this.validateGenre(disc.genre)
    if (!genreValidation.valid) {
      errors.genre = genreValidation.error
    }
    
    // Validation format
    const formatValidation = this.validateFormat(disc.format)
    if (!formatValidation.valid) {
      errors.format = formatValidation.error
    }
    
    // Validation label
    const labelValidation = this.validateLabel(disc.label)
    if (!labelValidation.valid) {
      errors.label = labelValidation.error
    }
    
    // Validation pays
    const countryValidation = this.validateCountry(disc.country)
    if (!countryValidation.valid) {
      errors.country = countryValidation.error
    }
    
    return {
      valid: Object.keys(errors).length === 0,
      errors
    }
  },

  /**
   * Vérifie si au moins les champs obligatoires sont remplis
   * @param {Object} disc - Objet disque
   * @returns {boolean}
   */
  hasRequiredFields(disc) {
    return !!(
      disc.artist && 
      disc.artist.trim().length >= 2 &&
      disc.title && 
      disc.title.trim().length >= 2
    )
  }
}

export default discValidation
