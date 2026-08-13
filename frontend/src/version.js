import { version } from '../package.json'

// Source unique de vérité pour la version affichée sur l'écran de connexion
// et dans le panneau "À propos" — incrémenter la version de package.json
// met à jour les deux d'un coup.
export const APP_VERSION = version

// __BUILD_TIME__ est injecté par Vite (voir vite.config.js `define`) comme
// l'horodatage ISO de la compilation de ce bundle (ou du démarrage du
// serveur de dev).
export const APP_BUILD_TIME = __BUILD_TIME__
