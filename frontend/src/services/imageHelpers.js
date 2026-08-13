// frontend/src/services/imageHelpers.js

/**
 * Service d'aide à la manipulation d'images
 */
export const imageHelpers = {
  /**
   * Convertit un File en Base64
   * @param {File} file - Fichier à convertir
   * @returns {Promise<string>} - String Base64
   */
  async fileToBase64(file) {
    return new Promise((resolve, reject) => {
      const reader = new FileReader()
      
      reader.onload = () => {
        resolve(reader.result)
      }
      
      reader.onerror = (error) => {
        reject(new Error('Erreur lors de la lecture du fichier'))
      }
      
      reader.readAsDataURL(file)
    })
  },

  /**
   * Redimensionne une image en conservant le ratio
   * @param {File} file - Fichier image à redimensionner
   * @param {number} maxWidth - Largeur maximale
   * @param {number} maxHeight - Hauteur maximale
   * @param {number} quality - Qualité JPEG (0-1)
   * @returns {Promise<File>} - Nouveau fichier redimensionné
   */
  async resizeImage(file, maxWidth = 800, maxHeight = 800, quality = 0.85) {
    return new Promise((resolve, reject) => {
      const reader = new FileReader()
      
      reader.onload = (e) => {
        const img = new Image()
        
        img.onload = () => {
          // Calculer les nouvelles dimensions en conservant le ratio
          let width = img.width
          let height = img.height
          
          if (width > height) {
            if (width > maxWidth) {
              height = Math.round((height * maxWidth) / width)
              width = maxWidth
            }
          } else {
            if (height > maxHeight) {
              width = Math.round((width * maxHeight) / height)
              height = maxHeight
            }
          }
          
          // Créer un canvas pour redimensionner
          const canvas = document.createElement('canvas')
          canvas.width = width
          canvas.height = height
          
          const ctx = canvas.getContext('2d')
          ctx.drawImage(img, 0, 0, width, height)
          
          // Convertir en Blob puis en File
          canvas.toBlob(
            (blob) => {
              if (!blob) {
                reject(new Error('Erreur lors de la création du blob'))
                return
              }
              
              const resizedFile = new File([blob], file.name, {
                type: file.type,
                lastModified: Date.now()
              })
              
              resolve(resizedFile)
            },
            file.type,
            quality
          )
        }
        
        img.onerror = () => {
          reject(new Error('Erreur lors du chargement de l\'image'))
        }
        
        img.src = e.target.result
      }
      
      reader.onerror = () => {
        reject(new Error('Erreur lors de la lecture du fichier'))
      }
      
      reader.readAsDataURL(file)
    })
  },

  /**
   * Vérifie si le fichier est une image valide
   * @param {File} file - Fichier à vérifier
   * @returns {boolean}
   */
  isValidImageType(file) {
    const validTypes = [
      'image/jpeg',
      'image/jpg',
      'image/png',
      'image/webp',
      'image/gif'
    ]
    return validTypes.includes(file.type)
  },

  /**
   * Vérifie la taille du fichier (en MB)
   * @param {File} file - Fichier à vérifier
   * @param {number} maxSizeMB - Taille maximale en MB
   * @returns {boolean}
   */
  isValidFileSize(file, maxSizeMB = 10) {
    const maxSizeBytes = maxSizeMB * 1024 * 1024
    return file.size <= maxSizeBytes
  },

  /**
   * Valide un fichier image (type + taille)
   * @param {File} file - Fichier à valider
   * @param {number} maxSizeMB - Taille max en MB
   * @returns {Object} { valid: boolean, error?: string }
   */
  validateImageFile(file, maxSizeMB = 10) {
    if (!this.isValidImageType(file)) {
      return {
        valid: false,
        error: 'Format de fichier non supporté. Utilisez JPG, PNG, WEBP ou GIF.'
      }
    }
    
    if (!this.isValidFileSize(file, maxSizeMB)) {
      return {
        valid: false,
        error: `Le fichier est trop volumineux. Taille maximale: ${maxSizeMB}MB`
      }
    }
    
    return { valid: true }
  },

  /**
   * Obtient l'URL optimisée d'une image Discogs
   * @param {string} url - URL originale Discogs
   * @param {string} size - Taille souhaitée ('150', '500', 'full')
   * @returns {string} - URL optimisée
   */
  getOptimizedDiscogsUrl(url, size = '500') {
    if (!url) return null
    
    // Discogs utilise des URLs avec des tailles dans le path
    // Ex: /150/ pour thumbnail, /500/ pour moyenne qualité
    
    // Remplacer la taille existante par la nouvelle
    let optimizedUrl = url.replace(/\/\d+\//, `/${size}/`)
    
    // Si aucune taille n'est présente, essayer d'en ajouter une
    if (optimizedUrl === url && size !== 'full') {
      optimizedUrl = url.replace(/\/([^\/]+\.(jpg|jpeg|png|gif))$/i, `/${size}/$1`)
    }
    
    return optimizedUrl
  },

  /**
   * Crée une URL temporaire pour prévisualiser un fichier
   * @param {File} file - Fichier image
   * @returns {string} - URL temporaire (blob)
   */
  createPreviewUrl(file) {
    return URL.createObjectURL(file)
  },

  /**
   * Libère une URL temporaire créée avec createPreviewUrl
   * @param {string} url - URL à libérer
   */
  revokePreviewUrl(url) {
    if (url && url.startsWith('blob:')) {
      URL.revokeObjectURL(url)
    }
  },

  /**
   * Extrait les dimensions d'une image
   * @param {File} file - Fichier image
   * @returns {Promise<Object>} - { width, height }
   */
  async getImageDimensions(file) {
    return new Promise((resolve, reject) => {
      const reader = new FileReader()
      
      reader.onload = (e) => {
        const img = new Image()
        
        img.onload = () => {
          resolve({
            width: img.width,
            height: img.height
          })
        }
        
        img.onerror = () => {
          reject(new Error('Impossible de lire les dimensions'))
        }
        
        img.src = e.target.result
      }
      
      reader.onerror = () => {
        reject(new Error('Erreur lors de la lecture du fichier'))
      }
      
      reader.readAsDataURL(file)
    })
  },

  /**
   * Crée une miniature carrée (crop centré)
   * @param {File} file - Fichier image
   * @param {number} size - Taille du carré
   * @returns {Promise<File>} - Miniature
   */
  async createThumbnail(file, size = 200) {
    return new Promise((resolve, reject) => {
      const reader = new FileReader()
      
      reader.onload = (e) => {
        const img = new Image()
        
        img.onload = () => {
          const canvas = document.createElement('canvas')
          canvas.width = size
          canvas.height = size
          
          const ctx = canvas.getContext('2d')
          
          // Calculer le crop centré
          const minDim = Math.min(img.width, img.height)
          const sx = (img.width - minDim) / 2
          const sy = (img.height - minDim) / 2
          
          ctx.drawImage(
            img,
            sx, sy, minDim, minDim,  // source
            0, 0, size, size          // destination
          )
          
          canvas.toBlob(
            (blob) => {
              if (!blob) {
                reject(new Error('Erreur création thumbnail'))
                return
              }
              
              const thumbnailFile = new File(
                [blob],
                `thumb_${file.name}`,
                { type: file.type }
              )
              
              resolve(thumbnailFile)
            },
            file.type,
            0.85
          )
        }
        
        img.onerror = () => {
          reject(new Error('Erreur chargement image'))
        }
        
        img.src = e.target.result
      }
      
      reader.onerror = () => {
        reject(new Error('Erreur lecture fichier'))
      }
      
      reader.readAsDataURL(file)
    })
  },

  /**
   * Formatte la taille d'un fichier en lecture humaine
   * @param {number} bytes - Taille en bytes
   * @returns {string} - Ex: "2.5 MB"
   */
  formatFileSize(bytes) {
    if (bytes === 0) return '0 Bytes'
    
    const k = 1024
    const sizes = ['Bytes', 'KB', 'MB', 'GB']
    const i = Math.floor(Math.log(bytes) / Math.log(k))
    
    return Math.round((bytes / Math.pow(k, i)) * 100) / 100 + ' ' + sizes[i]
  }
}

export default imageHelpers
