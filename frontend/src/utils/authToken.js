// Lit le token d'auth depuis sessionStorage (session courante, effacé à la
// fermeture du navigateur/onglet) puis localStorage (si "Se souvenir de moi"
// a été coché lors de la connexion) — voir stores/auth.js pour l'écriture.
export function getAuthToken() {
  return sessionStorage.getItem('user_token') || localStorage.getItem('user_token') || null
}
