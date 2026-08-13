<script setup>
// records-manager/frontend/src/App.vue

// Importe le composant clé de Vue Router
import { RouterView } from 'vue-router'
import { useThemeStore } from './stores/theme'

// Instancie le store une fois au montage pour appliquer data-theme sur <html>
useThemeStore()
</script>

<template>
  <RouterView />
</template>

<style>
/* ----------------------------------------------------
    Design tokens (alignés sur l'appli Finance)
    ---------------------------------------------------- */
:root {
  --bg: #0f1420;
  --bg-soft: #161d2e;
  --bg-elevated: #1c2438;
  --line: #2a3350;
  --line-soft: rgba(var(--tint-rgb), 0.06);

  --text: #f3f5fa;
  --text-soft: #b8c1d9;
  --text-dim: #7c88a6;

  --accent: #3b82f6;
  --accent-soft: #60a5fa;
  --accent-sand: #f2a878;
  --accent-blue: #2f6fe0;
  --accent-rose: #f97066;
  --danger: #ef4444;

  --tint-rgb: 255, 255, 255;
  --panel-bg: linear-gradient(135deg, rgba(52, 74, 122, 0.95) 0%, rgba(23, 31, 52, 0.97) 100%);
  --modal-bg: #1a2338;
  --modal-overlay: rgba(6, 9, 18, 0.72);
  --body-bg:
    radial-gradient(circle at top left, rgba(59, 130, 246, 0.12), transparent 32%),
    linear-gradient(180deg, #0b0f19 0%, #0f1420 100%);
  --sidebar-bg:
    radial-gradient(circle at top left, rgba(59, 130, 246, 0.07), transparent 28%),
    #0b0f19;

  --hero-gradient: linear-gradient(135deg, #0a1128 0%, #152852 32%, #2f5382 58%, #c98a63 84%, #f0ba8e 100%);

  --negative-text: #fca5a5;
  --positive-text: #86efac;

  --shadow: 0 18px 40px rgba(0, 0, 0, 0.38);
  --radius-xl: 28px;
  --radius-lg: 20px;
  --radius-md: 16px;
  --radius-sm: 12px;
  --radius-pill: 999px;

  /* Variables de tableau, utilisées par .data-table dans toutes les vues CRUD
     (y compris celles pas encore restylées individuellement) — adaptatives
     au thème via --tint-rgb/--line plutôt que des couleurs violettes fixes. */
  --table-header-color: var(--accent);
  --odd-row-color: rgba(var(--tint-rgb), 0.035);
  --even-row-color: var(--bg-elevated);
  --border-color: var(--line);
}

:root[data-theme='light'] {
  --bg: #f4f6fb;
  --bg-soft: #ffffff;
  --bg-elevated: #ffffff;
  --line: #dbe2f0;
  --line-soft: rgba(15, 23, 42, 0.08);

  --text: #10172a;
  --text-soft: #45516c;
  --text-dim: #697490;

  --tint-rgb: 15, 23, 42;
  --panel-bg: linear-gradient(180deg, #ffffff, #f5f8fd);
  --modal-bg: #ffffff;
  --modal-overlay: rgba(15, 23, 42, 0.35);
  --body-bg:
    radial-gradient(circle at top left, rgba(59, 130, 246, 0.08), transparent 30%),
    #eef2fa;
  --sidebar-bg:
    radial-gradient(circle at top left, rgba(59, 130, 246, 0.05), transparent 28%),
    #ffffff;

  --hero-gradient: linear-gradient(135deg, #101d3f 0%, #1d3a68 32%, #3f6796 58%, #dba077 84%, #f5cca5 100%);

  --negative-text: #b91c1c;
  --positive-text: #15803d;

  --shadow: 0 18px 40px rgba(15, 23, 42, 0.12);
}

*,
*::before,
*::after {
    box-sizing: border-box;
    margin: 0;
    position: relative;
    font-weight: normal;
}

body {
    min-height: 100vh;
    color: var(--text);
    background: var(--body-bg);
    line-height: 1.6;
    font-family: Inter, system-ui, -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, 'Fira Sans', 'Droid Sans', 'Helvetica Neue', sans-serif;
    font-size: 15px;
    text-rendering: optimizeLegibility;
    -webkit-font-smoothing: antialiased;
    -moz-osx-font-smoothing: grayscale;
}

/* Styles pour les liens de base */
a {
    text-decoration: none;
    color: var(--accent);
    transition: 0.4s;
}

/* Styles pour les boutons de base (Peut être affiné dans des composants spécifiques) */
button {
    cursor: pointer;
    border: none;
    padding: 8px 15px;
    border-radius: 4px;
    transition: background-color 0.3s, opacity 0.3s;
}

button:disabled {
    opacity: 0.6;
    cursor: not-allowed;
}

/* ----------------------------------------------------
    Classes utilitaires partagées (alignées sur Finance)
    ---------------------------------------------------- */
.primary-btn,
.ghost-btn {
  min-height: 44px;
  padding: 0 20px;
  border-radius: var(--radius-pill);
  transition: 0.2s ease;
}

.primary-btn {
  background: linear-gradient(135deg, var(--accent), var(--accent-blue));
  color: white;
  font-weight: 600;
}

.primary-btn:hover {
  filter: brightness(1.08);
  transform: translateY(-1px);
}

.ghost-btn {
  background: rgba(var(--tint-rgb), 0.04);
  color: var(--text-soft);
  border: 1px solid var(--line);
}

.ghost-btn:hover {
  background: rgba(var(--tint-rgb), 0.07);
  color: var(--text);
}

.primary-btn:disabled,
.ghost-btn:disabled {
  opacity: 0.45;
  cursor: not-allowed;
  filter: none;
  pointer-events: none;
}

.form-field {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.form-field span {
  color: var(--text-soft);
  font-size: 13px;
}

.form-field input,
.form-field select,
.form-field textarea {
  width: 100%;
  min-height: 46px;
  padding: 0 14px;
  border-radius: 14px;
  border: 1px solid var(--line);
  background: rgba(var(--tint-rgb), 0.04);
  color: var(--text);
  outline: none;
}

/* Le rendu natif d'un <select> (surtout sous Safari) ne respecte pas
   toujours min-height de la même façon qu'un <input> — on reprend donc la
   main sur son apparence pour garantir la même hauteur partout. */
.form-field select {
  cursor: pointer;
  padding-right: 36px;
  -webkit-appearance: none;
  -moz-appearance: none;
  appearance: none;
  background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 24 24' fill='none' stroke='%23888' stroke-width='2' stroke-linecap='round' stroke-linejoin='round'%3E%3Cpolyline points='6 9 12 15 18 9'%3E%3C/polyline%3E%3C/svg%3E");
  background-repeat: no-repeat;
  background-position: right 12px center;
  background-size: 16px;
}

.form-field textarea {
  min-height: 120px;
  padding: 14px;
  resize: vertical;
}

.form-field input::placeholder,
.form-field textarea::placeholder {
  color: var(--text-dim);
}

.form-field input:focus,
.form-field select:focus,
.form-field textarea:focus {
  border-color: rgba(59, 130, 246, 0.55);
  box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.15);
}

.form-field-checkbox {
  flex-direction: row;
  align-items: center;
  gap: 10px;
}

.form-field-checkbox input[type='checkbox'] {
  width: 18px;
  height: 18px;
  accent-color: var(--accent);
  cursor: pointer;
}

.form-error {
  margin: 0;
  color: var(--negative-text);
  font-size: 13px;
}

.form-success {
  margin: 0;
  padding: 0.85rem 1rem;
  border-radius: 14px;
  background: rgba(134, 239, 172, 0.14);
  color: var(--positive-text);
  border: 1px solid rgba(134, 239, 172, 0.24);
  font-size: 0.92rem;
}

.danger-btn {
  min-height: 44px;
  padding: 0 20px;
  border-radius: var(--radius-pill);
  background: rgba(220, 38, 38, 0.14);
  color: var(--negative-text);
  border: 1px solid rgba(220, 38, 38, 0.28);
  font-weight: 600;
  transition: 0.2s ease;
}

.danger-btn:hover {
  background: rgba(220, 38, 38, 0.22);
  transform: translateY(-1px);
}

.danger-btn:disabled {
  opacity: 0.45;
  cursor: not-allowed;
  transform: none;
  pointer-events: none;
}

.panel {
  background: var(--panel-bg);
  border: 1px solid var(--line-soft);
  border-radius: var(--radius-xl);
  box-shadow: var(--shadow);
}

.panel-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 14px;
  margin-bottom: 14px;
}

.panel-header h2 {
  margin: 4px 0 0;
  font-size: 20px;
}

.icon-action-btn {
  flex-shrink: 0;
  width: 36px;
  height: 36px;
  padding: 0;
  border: 1px solid rgba(var(--tint-rgb), 0.08);
  border-radius: 10px;
  display: grid;
  place-items: center;
  background: rgba(var(--tint-rgb), 0.03);
  color: rgba(var(--tint-rgb), 0.6);
  cursor: pointer;
  transition: background 140ms ease, color 140ms ease, transform 140ms ease;
}

.icon-action-btn:hover {
  background: rgba(var(--tint-rgb), 0.09);
  color: var(--text);
  transform: translateY(-1px);
}

.icon-action-btn-danger:hover {
  background: rgba(220, 38, 38, 0.14);
  color: var(--negative-text);
}

.modal-overlay {
  position: fixed;
  inset: 0;
  z-index: 50;
  display: grid;
  place-items: center;
  padding: 24px;
  background: var(--modal-overlay);
  backdrop-filter: blur(6px);
}

.modal-card {
  width: min(680px, 100%);
  max-height: min(90vh, 920px);
  overflow-y: auto;
  padding: 22px;
  border-radius: 24px;
  background: var(--modal-bg);
  border: 1px solid var(--line-soft);
  box-shadow: var(--shadow);
}

.modal-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
  margin-bottom: 18px;
}

.modal-header h2 {
  margin: 4px 0 0;
  font-size: 24px;
  line-height: 1.15;
}

.modal-actions {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  margin-top: 6px;
}

@media (max-width: 680px) {
  .modal-overlay {
    padding: 16px;
    align-items: end;
  }

  .modal-card {
    width: 100%;
    max-height: 92vh;
    padding: 18px;
    border-radius: 20px 20px 0 0;
  }

  .modal-header {
    flex-direction: column;
    align-items: stretch;
  }

  .modal-actions {
    flex-direction: column-reverse;
  }

  .modal-actions .primary-btn,
  .modal-actions .ghost-btn,
  .modal-actions .danger-btn {
    width: 100%;
  }
}

/* ----------------------------------------------------
    CORRECTIONS CRITIQUES GLOBALES POUR LES TABLEAUX
    (inchangé jusqu'au restylage des vues CRUD en passe 2)
    ---------------------------------------------------- */

/* 1. DÉBLOQUER L'AFFICHAGE DE L'EN-TÊTE ET LES BORDURES */
/* Annule les masquages agressifs des frameworks mobiles et les bordures: none */
table, thead, tbody, tr, th, td {
    border-collapse: collapse !important;
    border-spacing: 0 !important;
    /* Annuler display: none ou display: block des resets */
    display: revert !important;
    visibility: visible !important;
    border: none !important;
}

/* Forcer la couleur de fond des cellules pour les bandes alternées */
.data-table tbody tr.odd-row td {
    background-color: var(--odd-row-color) !important;
}

.data-table tbody tr.even-row td {
    background-color: var(--even-row-color) !important;
}
</style>
