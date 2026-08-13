<!-- frontend/src/components/DiscsModal/DiscsModal.vue -->
<template>
  <Teleport to="body">
    <Transition name="modal-fade">
      <div v-if="isOpen" class="discs-modal-overlay" @mousedown="onOverlayMouseDown" @click="onOverlayClick">
        <div class="discs-modal" :class="{ 'is-saving': isSaving }">
          <!-- En-tête du modal -->
          <div class="modal-header">
            <div class="header-content">
              <h2 class="modal-title">
                <span v-if="discData">✏️ Modifier le disque</span>
                <span v-else>➕ Ajouter un disque</span>
              </h2>
              <p class="modal-subtitle">
                {{ discData ? 'Modifiez les informations du disque' : 'Ajoutez un nouveau disque à votre collection' }}
              </p>
            </div>
            <button type="button" @click="handleCloseRequest" class="close-button" :disabled="isSaving">
              ✕
            </button>
          </div>
          <!-- Corps du modal -->
          <div class="modal-body">
            <!-- Étape 1: Recherche Discogs (uniquement création) -->
            <div v-if="!discData && showSearchStep" class="search-step">
              <div class="step-header">
                <div class="step-number active">1</div>
                <h3 class="step-title">🔍 Recherche automatique</h3>
              </div>
              <!-- Recherche par code-barres avancée -->
              <div class="search-method">
                <h4 class="method-title">📊 Recherche par code-barres</h4>
                <div class="barcode-input-container">
                  <div class="barcode-input-group">
                    <input ref="barcodeInput" v-model="searchBarcode" type="text"
                      placeholder="Ex: 0196587842024 ou scannez avec une douchette" class="barcode-input"
                      :class="{ 'scanner-input': scannerDetected || isWatchingClipboard }" @focus="handleBarcodeFocus"
                      @blur="checkBarcodeExists" @keyup.enter="handleBarcodeEnter" :disabled="searchLoading" />
                    <button type="button" @click="openBarcodeScanner" class="scan-button" title="Scanner avec la caméra"
                      :disabled="searchLoading">
                      📷
                    </button>
                    <button type="button" @click="triggerFileImport" class="scan-button"
                      title="Importer depuis un fichier" :disabled="searchLoading">
                      📄
                    </button>
                    <button type="button" @click="toggleClipboardWatch"
                      :class="['scan-button', { 'active': isWatchingClipboard }]"
                      :title="isWatchingClipboard ? 'Arrêter la surveillance' : 'Mode Scan Auto (Clipboard)'"
                      :disabled="searchLoading">
                      {{ isWatchingClipboard ? '⏸️' : '📋' }}
                    </button>
                    <input ref="fileImportInput" type="file" accept=".txt" style="display: none"
                      @change="handleFileImport" />
                    <button @click="handleBarcodeEnter" :disabled="!searchBarcode || searchLoading"
                      class="search-button">
                      <span v-if="searchLoading" class="spinner small"></span>
                      <span v-else class="icon">🔍</span>
                      <span class="button-text">Rechercher</span>
                    </button>
                  </div>
                  <div class="input-help scanner-detected" v-if="isWatchingClipboard">
                    📋 Mode Scan Auto ACTIF - Scannez avec votre douchette maintenant !
                  </div>
                  <div class="input-help scanner-detected" v-else-if="scannerDetected">
                    ⌨️ Douchette détectée • Buffer: {{ scanBuffer }}
                  </div>
                  <div class="input-help" v-else>
                    💡 📋 = Scan auto | 📷 = Caméra | 📄 = Fichier
                  </div>
                  <div v-if="scannerDetected" class="reset-scan-zone">
                    <button @click="resetScannerBuffer" class="btn-reset-scan" :disabled="searchLoading">
                      🔄 Réinitialiser
                    </button>
                    <span class="reset-help">Prêt pour un nouveau scan</span>
                  </div>
                </div>
              </div>
              <div class="or-separator">
                <div class="or-line"></div>
                <div class="or-text">OU</div>
                <div class="or-line"></div>
              </div>
              <div class="search-method">
                <h4 class="method-title">📝 Recherche manuelle</h4>
                <div class="manual-search-form">
                  <div class="form-row">
                    <input v-model="manualSearch.title" type="text" placeholder="Titre de l'album" class="search-input"
                      :disabled="searchLoading" @keyup.enter="searchManual" />
                    <input v-model="manualSearch.artist" type="text" placeholder="Artiste" class="search-input"
                      :disabled="searchLoading" @keyup.enter="searchManual" />
                  </div>
                  <div class="form-row" style="margin-top: 12px;">
                    <input v-model="manualSearch.country" type="text" placeholder="Pays (ex: France, US, Germany)" class="search-input"
                      :disabled="searchLoading" @keyup.enter="searchManual" />
                    <input v-model="manualSearch.year" type="text" placeholder="Année (ex: 1985)" class="search-input"
                      :disabled="searchLoading" @keyup.enter="searchManual" maxlength="4" />
                  </div>
                  <div class="form-row centered" style="margin-top: 16px;">
                    <button @click="searchManual" class="search-button"
                      :disabled="!hasValidSearchTerms || searchLoading">
                      <span v-if="searchLoading" class="spinner small"></span>
                      <span v-else class="icon">🔍</span>
                      <span class="button-text">Rechercher</span>
                    </button>
                  </div>
                  <div class="search-status" v-if="searchResults.length === 0 && searchAttempted">
                    ❌ Aucun résultat trouvé
                  </div>
                  <div class="search-status success" v-else-if="searchResults.length > 0">
                    ✅ {{ searchResults.length }} résultat(s) trouvé(s)
                  </div>
                  <!-- ✅ Info limite Discogs -->
                  <div v-if="searchResults.length === 50" class="search-status warning">
                    <span class="icon">⚠️</span>
                    <span>Plus de 50 résultats — affinez avec pays ou année</span>
                  </div>
                </div>
              </div>
              <div v-if="searchResults.length > 0" class="search-results">
                <h4 class="results-title">
                  <span class="icon">🎵</span>
                  <span>{{ searchResults.length }} résultat(s) trouvé(s)</span>
                </h4>
                <div class="results-grid">
                  <div v-for="result in searchResults" :key="result.id" class="result-card"
                    :class="{ 'selected': selectedResultId === result.id }" @click="selectSearchResult(result)">
                    <!-- Pochette avec badge prix -->
                    <div class="result-cover">
                      <img v-if="result.cover_url" :src="getImageUrl(result.cover_url)" :alt="result.title"
                        class="cover-image" @error="handleImageError" />
                      <div v-else class="cover-placeholder">
                        <span class="placeholder-icon">💿</span>
                      </div>
                      <div v-if="getAveragePrice(result.prices)" class="price-badge">
                        {{ formatCurrency(getAveragePrice(result.prices)) }}
                      </div>
                      <div v-else-if="result.pricesLoading" class="price-badge loading">...</div>
                    </div>
                    <!-- Infos -->
                    <div class="result-info">
                      <h5 class="result-title" :title="result.title">{{ result.title || 'Titre inconnu' }}</h5>
                      <p v-if="result.artist || result.artists?.[0]?.name" class="result-artist">
                        {{ result.artist || result.artists?.[0]?.name }}
                      </p>
                      <div class="result-meta">
                        <span v-if="result.year" class="meta-item">📅 {{ result.year }}</span>
                        <span v-if="result.country" class="meta-item">🌍 {{ result.country }}</span>
                        <span v-if="result.formats && result.formats.length" class="meta-item">💿 {{ result.formats.join(', ') }}</span>
                        <span v-if="result.genres && result.genres.length" class="meta-item" :title="result.genres.join(', ')">
                          🎵 {{ truncateGenres(result.genres) }}
                        </span>
                        <span v-if="result.label" class="meta-item" :title="result.label">🏷️ {{ result.label }}</span>
                      </div>
                      <!-- Footer -->
                      <div class="result-footer">
                        <a v-if="result.id" :href="`https://www.discogs.com/release/${result.id}`" 
                           target="_blank" rel="noopener" class="discogs-link" @click.stop>
                          <span class="discogs-icon"></span>
                          r{{ result.id }}
                        </a>
                        <button @click.stop="selectSearchResult(result)" class="select-button"
                          :class="{ 'is-selected': selectedResultId === result.id }">
                          {{ selectedResultId === result.id ? '✓' : 'Choisir' }}
                        </button>
                      </div>
                    </div>
                  </div>
                </div>
              </div>
              <div v-if="coverPreview.found" class="preview-section">
                <div class="preview-header">
                  <h4>✅ Résultat sélectionné</h4>
                  <button @click="resetPreview" class="btn secondary small">
                    ❌ Changer
                  </button>
                </div>
                <div class="preview-content">
                  <img v-if="coverPreview.url" :src="getImageUrl(coverPreview.url)" alt="Pochette sélectionnée"
                    class="preview-image" @error="handleImageError" />
                  <div v-else class="preview-image preview-placeholder">
                    <span>?</span>
                  </div>
                  <div class="preview-info">
                    <h5>{{ coverPreview.title }}</h5>
                    <div class="preview-details">
                      <div v-if="coverPreview.artist" class="detail-item">
                        <strong>Artiste:</strong> {{ coverPreview.artist }}
                      </div>
                      <div v-if="coverPreview.year" class="detail-item">
                        <strong>Année:</strong> {{ coverPreview.year }}
                      </div>
                      <div v-if="coverPreview.genres && coverPreview.genres.length" class="detail-item">
                        <strong>Genres:</strong> {{ coverPreview.genres.join(', ') }}
                      </div>
                      <div v-if="coverPreview.formats && coverPreview.formats.length" class="detail-item">
                        <strong>Formats:</strong> {{ coverPreview.formats.join(', ') }}
                      </div>
                      <div v-if="coverPreview.country" class="detail-item">
                        <strong>Pays:</strong> {{ coverPreview.country }}
                      </div>
                      <div v-if="coverPreview.label" class="detail-item">
                        <strong>Label:</strong> {{ coverPreview.label }}
                      </div>
                      <div v-if="getAveragePrice(coverPreview.prices)" class="detail-item price-detail">
                        <strong>💶 Prix moyen:</strong> {{ formatCurrency(getAveragePrice(coverPreview.prices)) }}
                      </div>
                    </div>
                  </div>
                </div>
              </div>
              <div class="step-navigation">
                <button @click="skipSearch" class="skip-button" :disabled="searchLoading">
                  <span class="icon">➡️</span>
                  <span>Saisir manuellement</span>
                </button>
                <button 
                  v-if="coverPreview.found" 
                  @click="proceedToForm" 
                  class="continue-button"
                  :disabled="searchLoading"
                  ref="continueButton"
                >
                  <span>Continuer avec la sélection</span>
                  <span class="icon">➡️</span>
                </button>
              </div>
            </div>
            <div v-if="!showSearchStep || discData" class="form-step">
              <div class="step-header" v-if="!discData">
                <div class="step-number" :class="{ 'active': !showSearchStep }">2</div>
                <h3 class="step-title">📝 Informations du disque</h3>
              </div>
              <div class="form-container">
                <div v-if="apiError" class="error-message">
                  <span class="error-icon">⚠️</span>
                  <span class="error-text">{{ apiError }}</span>
                </div>
                <form @submit.prevent="handleSubmit" class="disc-form">
                  <div v-if="coverPreview.found || discData" class="selection-summary">
                    <div class="summary-content">
                      <div class="cover-container">
                        <img v-if="coverPreview.url || currentCoverUrl"
                          :src="coverPreview.url ? getImageUrl(coverPreview.url, coverCacheBuster) : currentCoverUrl"
                          :key="coverCacheBuster"
                          alt="Pochette"
                          class="summary-image" 
                          @error="handleImageError" />
                        <div v-else class="summary-image summary-placeholder">
                          <span>?</span>
                        </div>
                        <div class="cover-actions">
                          <button type="button" @click="openFilePicker" class="cover-action-btn" title="Changer la pochette">
                            📁
                          </button>
                          <button type="button" v-if="currentCoverUrl || coverPreview.url" @click="removeCover"
                            class="cover-action-btn remove" title="Supprimer la pochette">
                            🗑️
                          </button>
                        </div>
                      </div>
                      <div class="summary-details">
                        <h4>{{ coverPreview.title || formData.title }}</h4>
                        <p>{{ coverPreview.artist || getArtistName(formData.artist_id) }} •
                          {{ coverPreview.year || formData.release_year || 'Année inconnue' }}</p>
                        <p v-if="coverPreview.label" class="summary-label">
                          <small>Label: {{ coverPreview.label }}</small>
                        </p>
                      </div>
                    </div>
                  </div>
                  <input ref="fileInput" type="file" accept="image/*" style="display: none"
                    @change="handleFileUpload" />
                  <div v-if="uploadPreview.url" class="upload-preview">
                    <div class="preview-header">
                      <h5>📷 Aperçu de l'image</h5>
                      <button type="button" @click="cancelUpload" class="btn secondary small">
                        ❌ Annuler
                      </button>
                    </div>
                    <div class="preview-content">
                      <img :src="uploadPreview.url" alt="Aperçu upload" class="preview-image" />
                      <div class="preview-info">
                        <div class="file-info">
                          <strong>Taille:</strong> {{ formatFileSize(uploadPreview.size) }}
                        </div>
                        <div class="file-info">
                          <strong>Dimensions:</strong> {{ uploadPreview.width }}×{{ uploadPreview.height }}px
                        </div>
                        <button type="button" @click="confirmUpload" class="btn primary small" :disabled="uploadPreview.loading">
                          <span v-if="uploadPreview.loading" class="btn-spinner"></span>
                          {{ uploadPreview.loading ? 'Upload...' : '✅ Utiliser cette image' }}
                        </button>
                      </div>
                    </div>
                  </div>
                  <div class="upload-section"
                    v-if="!coverPreview.found && !discData && !currentCoverUrl && !uploadPreview.url">
                    <div class="upload-zone" @click="openFilePicker" @drop="handleDrop" @dragover="handleDragOver"
                      @dragleave="handleDragLeave" :class="{ 'drag-over': isDragOver }">
                      <div class="upload-content">
                        <div class="upload-icon">📁</div>
                        <h5>Glissez-déposez votre image ici</h5>
                        <p>ou cliquez pour parcourir vos fichiers</p>
                        <p class="upload-help">Formats supportés: JPG, PNG, WebP (max. 5MB)</p>
                      </div>
                    </div>
                  </div>
                  <div class="form-section">
                    <h3 class="section-title">📀 Informations du disque</h3>
                    <div class="form-grid compact">
                      <!-- Ligne 1: Titre + Artiste -->
                      <div class="form-group">
                        <label for="title" class="required">Titre</label>
                        <input id="title" v-model="formData.title" type="text" required placeholder="Ex: Nevermind"
                          class="form-input" :disabled="isSaving" />
                      </div>
                      <div class="form-group">
                        <label for="artist_id" class="required">Artiste</label>
                        <div class="artist-input-group">
                          <select id="artist_id" v-model="formData.artist_id" required class="form-select"
                            :disabled="isSaving">
                            <option value="">Sélectionnez un artiste</option>
                            <option v-for="artist in artists" :key="artist.id" :value="artist.id">
                              {{ artist.name }}
                            </option>
                          </select>
                          <button type="button" @click="openArtistModal" class="create-artist-button"
                            title="Créer un nouvel artiste" :disabled="isSaving">
                            <span class="icon">+</span>
                          </button>
                        </div>
                        <div class="form-help"
                          v-if="coverPreview.artist && !formData.artist_id && !isArtistAlreadyExists(coverPreview.artist)">
                          <button type="button" @click="openArtistModalWithPrefill(coverPreview.artist)" class="inline-button">
                            🎨 Créer "{{ coverPreview.artist }}"
                          </button>
                        </div>
                      </div>
                      <!-- Ligne 2: Genre + Format -->
                      <div class="form-group">
                        <label for="genre_id">Genre</label>
                        <div class="select-with-button">
                          <select id="genre_id" v-model="formData.genre_id" class="form-select" :disabled="isSaving">
                            <option value="">Sélectionnez un genre</option>
                            <option v-for="genre in genres" :key="genre.id" :value="genre.id">
                              {{ genre.name }}
                            </option>
                          </select>
                          <button type="button" @click="openGenreModal" class="create-artist-button"
                            title="Créer un nouveau genre" :disabled="isSaving">
                            <span class="icon">+</span>
                          </button>
                        </div>
                        <div class="form-help"
                          v-if="coverPreview.genres && coverPreview.genres.length && !formData.genre_id && !isGenreAlreadyExists(coverPreview.genres[0])">
                          <button type="button" @click="openGenreModalWithPrefill(coverPreview.genres[0])"
                            class="inline-button">
                            🎵 Créer "{{ coverPreview.genres[0] }}"
                          </button>
                        </div>
                      </div>
                      <div class="form-group">
                        <label for="format_id">Format</label>
                        <div class="select-with-button">
                          <select id="format_id" v-model="formData.format_id" class="form-select" :disabled="isSaving">
                            <option value="">Sélectionnez un format</option>
                            <option v-for="format in formats" :key="format.id" :value="format.id">
                              {{ format.name }}
                            </option>
                          </select>
                          <button type="button" @click="openFormatModal" class="create-artist-button"
                            title="Créer un nouveau format" :disabled="isSaving">
                            <span class="icon">+</span>
                          </button>
                        </div>
                        <div class="form-help"
                          v-if="coverPreview.formats && coverPreview.formats.length && !formData.format_id && !isFormatAlreadyExists(coverPreview.formats[0])">
                          <button type="button" @click="openFormatModalWithPrefill(coverPreview.formats[0])"
                            class="inline-button">
                            💿 Créer "{{ coverPreview.formats[0] }}"
                          </button>
                        </div>
                      </div>
                      <!-- Ligne 3: Pays + Label -->
                      <div class="form-group">
                        <label for="country_id">Pays</label>
                        <div class="select-with-button">
                          <select id="country_id" v-model="formData.country_id" class="form-select"
                            :disabled="isSaving">
                            <option value="">Sélectionnez un pays</option>
                            <option v-for="country in countries" :key="country.id" :value="country.id">
                              {{ country.name }}
                            </option>
                          </select>
                          <button type="button" @click="openCountryModal" class="create-artist-button"
                            title="Créer un nouveau pays" :disabled="isSaving">
                            <span class="icon">+</span>
                          </button>
                        </div>
                        <div class="form-help"
                          v-if="coverPreview.country && !formData.country_id && !isCountryAlreadyExists(coverPreview.country)">
                          <button type="button" @click="openCountryModalWithPrefill(coverPreview.country)"
                            class="inline-button">
                            🌍 Créer "{{ coverPreview.country }}"
                          </button>
                        </div>
                      </div>
                      <div class="form-group">
                        <label for="label_id">Label</label>
                        <div class="select-with-button">
                          <select id="label_id" v-model="formData.label_id" class="form-select" :disabled="isSaving">
                            <option value="">Sélectionnez un label</option>
                            <option v-for="label in labels" :key="label.id" :value="label.id">
                              {{ label.name }}
                            </option>
                          </select>
                          <button type="button" @click="openLabelModal" class="create-artist-button"
                            title="Créer un nouveau label" :disabled="isSaving">
                            <span class="icon">+</span>
                          </button>
                        </div>
                        <div class="form-help"
                          v-if="coverPreview.label && !formData.label_id && !isLabelAlreadyExists(coverPreview.label)">
                          <button type="button" @click="openLabelModalWithPrefill(coverPreview.label)"
                            class="inline-button">
                            🏷️ Créer "{{ coverPreview.label }}"
                          </button>
                        </div>
                      </div>
                      <!-- Ligne 4: Année + Code-barres -->
                      <div class="form-group">
                        <label for="release_year">Année</label>
                        <input id="release_year" v-model="formData.release_year" type="number" min="1900"
                          :max="new Date().getFullYear()" placeholder="Ex: 1991" class="form-input"
                          :disabled="isSaving" />
                      </div>
                      <div class="form-group">
                        <label for="barcode">Code-barres</label>
                        <input id="barcode" v-model="formData.barcode" type="text" placeholder="Ex: 5099748933725"
                          class="form-input" :disabled="isSaving" />
                      </div>
                    </div>
                  </div>
                  <div class="form-section">
                    <h3 class="section-title">💰 Commercial</h3>
                    <div class="form-grid compact">
                      <div class="form-group">
                        <label for="price">Prix (€)</label>
                        <div class="price-input-wrapper">
                          <input
                            id="price"
                            v-model.number="formData.price"
                            type="number"
                            step="0.01"
                            min="0"
                            placeholder="0.00"
                            class="form-input"
                            :class="{ 'has-discogs-price': getAveragePrice(coverPreview.prices) }"
                            :disabled="isSaving"
                            @blur="autoFormatPrice"
                          />
                          <!-- Suggestion prix Discogs -->
                          <div v-if="getAveragePrice(coverPreview.prices)" class="discogs-price-suggestion">
                            <span class="suggestion-label">💶 Discogs:</span>
                            <span class="suggestion-value">{{ formatCurrency(getAveragePrice(coverPreview.prices)) }}</span>
                            <button 
                              type="button" 
                              class="apply-price-btn"
                              @click="formData.price = getAveragePrice(coverPreview.prices)"
                              :disabled="isSaving"
                              title="Appliquer le prix Discogs"
                            >
                              ✓
                            </button>
                          </div>
                        </div>
                      </div>
                      <div class="form-group">
                        <label for="quantity">Quantité</label>
                        <input
                          id="quantity"
                          v-model.number="formData.quantity"
                          type="number"
                          step="1"
                          min="1"
                          placeholder="1"
                          class="form-input"
                          :disabled="isSaving"
                        />
                      </div>
                      <div class="form-group full-width">
                        <label for="notes">Notes</label>
                        <textarea id="notes" v-model="formData.notes"
                          placeholder="État du disque, édition spéciale, anecdotes..." class="form-textarea" rows="2"
                          :disabled="isSaving"></textarea>
                        <div class="char-counter">
                          {{ formData.notes?.length || 0 }}/500
                        </div>
                      </div>
                    </div>
                  </div>
                  
      <!-- Identifiants -->
      <div class="form-section streaming-section">
        <h3 class="section-title">
          <span class="icon">🔢</span> Identifiants
          <span class="optional-badge">Optionnel</span>
        </h3>

        <div class="form-grid compact">
          <div class="form-group full-width">
            <label for="isrc">🔢 Code ISRC</label>
            <input
              id="isrc"
              v-model="formData.isrc"
              type="text"
              placeholder="FR-123-45-67890"
              class="form-input"
              maxlength="20"
              :disabled="isSaving"
            />
            <small class="form-help">Code international d'identification des enregistrements</small>
          </div>
        </div>
      </div>

      <div class="form-actions">
                    <div v-if="!discData" class="back-action">
                      <button type="button" @click="backToSearch" class="back-button" :disabled="isSaving">
                        <span class="icon">⬅️</span>
                        <span>Retour à la recherche</span>
                      </button>
                    </div>
                    <div class="main-actions">
                      <div class="action-group">
                        <button type="button" @click="handleCloseRequest" class="cancel-button" :disabled="isSaving">
                          Annuler
                        </button>
                        <button type="submit" class="save-button" :disabled="isSaving || !isFormValid">
                          <span v-if="isSaving" class="spinner"></span>
                          <span v-else>
                            {{ discData ? 'Mettre à jour' : 'Enregistrer' }}
                          </span>
                        </button>
                      </div>
                    </div>
                  </div>
                </form>
              </div>
            </div>
          </div>
        </div>
      </div>
    </Transition>
    <ArtistsModal :is-open="isArtistModalOpen" :artist-data="null" :prefill-name="artistToCreate" :api-error="null"
      @close="isArtistModalOpen = false" @artist-saved="handleArtistSaved" />
    <GenresModal :is-open="isGenreModalOpen" :genre-data="null" :prefill-name="genreToCreate" :api-error="null"
      @close="isGenreModalOpen = false" @genre-saved="handleGenreSaved" />
    <FormatsModal :is-open="isFormatModalOpen" :format-data="null" :prefill-name="formatToCreate" :api-error="null"
      @close="isFormatModalOpen = false" @format-saved="handleFormatSaved" />
    <CountriesModal :is-open="isCountryModalOpen" :country-data="countryToCreate" :api-error="null"
      @close="isCountryModalOpen = false" @country-saved="handleCountrySaved" />
    <LabelsModal :is-open="isLabelModalOpen" :label-data="null" :prefill-name="labelToCreate" :api-error="null"
      @close="handleLabelModalClose" @label-saved="handleLabelSaved" />
    <BarcodeScanner :is-open="isBarcodeScannerOpen" @close="isBarcodeScannerOpen = false"
      @barcode-detected="handleBarcodeScanned" />
    <Teleport to="body">
      <div v-if="isBarcodeModalOpen" class="modal-overlay" @click.self="closeBarcodeModal">
        <div class="modal-card">
          <div class="modal-header">
            <div class="header-title-container">
              <span class="title-icon">⚠️</span>
              <div class="header-text">
                <h2>Code-barres déjà utilisé</h2>
                <p class="header-subtitle">
                  Un disque avec ce code-barres existe déjà dans la base de données
                </p>
              </div>
            </div>
            <button class="icon-action-btn" @click="closeBarcodeModal" aria-label="Fermer">
              <span>&times;</span>
            </button>
          </div>
          <div v-if="existingDisc && existingDisc.title" class="existing-disc-info">
            <div class="info-section">
              <h4>📋 Informations du disque existant</h4>
              <div class="disc-details">
                <div class="detail-row">
                  <strong>Titre:</strong> {{ existingDisc.title }}
                </div>
                <div class="detail-row">
                  <strong>Artiste:</strong> {{ existingDisc.artist_name }}
                </div>
                <div class="detail-row">
                  <strong>Code-barres:</strong> {{ existingDisc.barcode }}
                </div>
                <div v-if="existingDisc.id" class="detail-row">
                  <strong>ID:</strong> {{ existingDisc.id }}
                </div>
              </div>
            </div>
          </div>
          <div v-else class="existing-disc-info">
            <div class="info-section">
              <h4>⚠️ Disque existant trouvé</h4>
              <p>Un disque avec le code-barres <strong>{{ formData.barcode }}</strong> existe déjà.</p>
            </div>
          </div>
          <div class="modal-actions">
            <button @click="resetAndRetry" class="ghost-btn">
              🔄 Saisir un nouveau code-barres
            </button>
            <button @click="editExistingDisc" class="primary-btn" :disabled="!existingDisc || !existingDisc.id">
              ✏️ Modifier le disque existant
            </button>
          </div>
        </div>
      </div>
    </Teleport>
    <Teleport to="body">
      <div v-if="isUnsavedChangesModalOpen" class="modal-overlay" @click.self="cancelClose">
        <div class="modal-card">
          <div class="modal-header">
            <div class="header-title-container">
              <span class="title-icon">⚠️</span>
              <div class="header-text">
                <h2>Modifications non enregistrées</h2>
                <p class="header-subtitle">
                  Les modifications que vous avez effectuées seront perdues si vous continuez
                </p>
              </div>
            </div>
            <button class="icon-action-btn" @click="cancelClose" aria-label="Fermer">
              <span>&times;</span>
            </button>
          </div>
          <div class="existing-disc-info">
            <div class="info-section">
              <h4>📝 Que souhaitez-vous faire ?</h4>
              <p>Vous avez commencé à remplir le formulaire. Si vous fermez maintenant, ces informations seront perdues.</p>
            </div>
          </div>
          <div class="modal-actions">
            <button @click="confirmCloseWithoutSaving" class="ghost-btn">
              ❌ Fermer sans enregistrer
            </button>
            <button @click="cancelClose" class="primary-btn">
              ↩️ Continuer l'édition
            </button>
          </div>
        </div>
      </div>
    </Teleport>
  </Teleport>
</template>

<script setup>
import { ref, computed, watch, onMounted, onBeforeUnmount, nextTick } from 'vue'
import { useApi } from '@/composables/useApi'
import { formatCurrency } from '@/utils/format'
import { discogsApi } from '@/services/discogsApi.js'
import ArtistsModal from '@/components/ArtistsModal.vue'
import GenresModal from '@/components/GenresModal.vue'
import FormatsModal from '@/components/FormatsModal.vue'
import CountriesModal from '@/components/CountriesModal.vue'
import LabelsModal from '@/components/LabelsModal.vue'
import BarcodeScanner from '@/components/BarcodeScanner.vue'

// === Props et Emits ===
const props = defineProps({
  isOpen: Boolean,
  discData: Object,
  apiError: String,
  isSaving: Boolean
})
const emit = defineEmits(['close', 'save', 'edit-existing-disc', 'createArtist'])

const { apiFetch, upload } = useApi()

// === États UI ===
const isSaving = ref(false)
const showSearchStep = ref(true)
const searchLoading = ref(false)
const searchError = ref('')
const searchResults = ref([])
const selectedResultId = ref(null)
const searchAttempted = ref(false)
const continueButton = ref(null)
const barcodeSearchTimeout = ref(null)
const isBarcodeSearchLocked = ref(false)

// === Barcode avancé ===
const barcodeInput = ref(null)
const fileInput = ref(null)
const fileImportInput = ref(null)
const scannerDetected = ref(false)
const scanBuffer = ref('')
const lastKeyTime = ref(0)
const isWatchingClipboard = ref(false)
const clipboardWatchInterval = ref(null)
const lastClipboardValue = ref('')
const isDragOver = ref(false)

// === Upload ===
const currentCoverUrl = ref(null)
const coverCacheBuster = ref(Date.now())
const uploadPreview = ref({
  url: null,
  file: null,
  size: 0,
  width: 0,
  height: 0,
  loading: false
})

// === Modales ===
const isArtistModalOpen = ref(false)
const isGenreModalOpen = ref(false)
const isFormatModalOpen = ref(false)
const isCountryModalOpen = ref(false)
const isLabelModalOpen = ref(false)
const isBarcodeScannerOpen = ref(false)
const isBarcodeModalOpen = ref(false)
const isUnsavedChangesModalOpen = ref(false)
const overlayMouseDownTarget = ref(null)

// === Données modales ===
const artistToCreate = ref('')
const genreToCreate = ref('')
const formatToCreate = ref('')
const labelToCreate = ref('')
const countryToCreate = ref(null)
const existingDisc = ref({})

// === Options ===
const artists = ref([])
const genres = ref([])
const formats = ref([])
const countries = ref([])
const labels = ref([])

// === Recherche ===
const searchBarcode = ref('')
const manualSearch = ref({ title: '', artist: '', country: '', year: '' })
const coverPreview = ref({
  url: null,
  loading: false,
  found: false,
  title: '',
  artist: '',
  year: '',
  genres: [],
  formats: [],
  country: '',
  label: '',
  prices: null,
  isUploaded: false
})
const formData = ref({
  id: null,
  title: '',
  artist_id: '',
  genre_id: '',
  format_id: '',
  country_id: '',
  label_id: '',
  release_year: '',
  barcode: '',
  price: null,
  quantity: 1,
  notes: '',
  cover_image: null
})
const originalFormData = ref(null)

// ✅ getImageUrl robuste
const getImageUrl = (path, cacheBuster = null) => {
  if (!path || path === 'null' || path === 'undefined') return ''
  let url = path
  if (path.startsWith('http')) return url
  const BASE = import.meta.env.VITE_SERVER_BASE_URL || ''
  if (path.startsWith('/uploads')) url = BASE + path
  else url = BASE + '/uploads/' + (path.startsWith('/') ? path.slice(1) : path)
  if (cacheBuster) url += (url.includes('?') ? '&' : '?') + `t=${cacheBuster}`
  return url
}

// ✅ Fonction utilitaire : prix moyen (compatible avec le format backend)
const getAveragePrice = (prices) => {
  if (!prices || prices.noPrice) {
    return null
  }
  // Format backend: { lowest_price, median_price, highest_price }
  if (prices.median_price && prices.median_price > 0) {
    return Math.round(prices.median_price * 100) / 100
  }
  if (prices.lowest_price && prices.highest_price) {
    const avg = (prices.lowest_price + prices.highest_price) / 2
    return Math.round(avg * 100) / 100
  }
  if (prices.lowest_price && prices.lowest_price > 0) {
    return Math.round(prices.lowest_price * 100) / 100
  }
  // Format legacy: tableau [{currency, value}]
  if (Array.isArray(prices) && prices.length > 0) {
    const eurPrices = prices
      .filter(p => p.currency === 'EUR' && p.value > 0 && isFinite(p.value))
      .map(p => parseFloat(p.value))
    if (eurPrices.length === 0) return null
    const avg = eurPrices.reduce((sum, price) => sum + price, 0) / eurPrices.length
    return Math.round(avg * 100) / 100
  }
  return null
}

// ✅ Fonction pour tronquer les genres
const truncateGenres = (genres, maxLength = 25) => {
  if (!genres || !genres.length) return ''
  const joined = genres.join(', ')
  if (joined.length <= maxLength) return joined
  let truncated = ''
  for (let i = 0; i < genres.length; i++) {
    const candidate = (truncated ? truncated + ', ' : '') + genres[i]
    if (candidate.length > maxLength - 4) break // -4 pour " ..."
    truncated = candidate
  }
  return truncated ? truncated + ', …' : genres[0].slice(0, maxLength - 3) + '…'
}

// === Computed ===
const isEditing = computed(() => !!props.discData?.id)
const isFormValid = computed(() => formData.value.title?.trim() && formData.value.artist_id)
const hasValidSearchTerms = computed(() => {
  const hasTitle = manualSearch.value.title && manualSearch.value.title.length >= 2
  const hasArtist = manualSearch.value.artist && manualSearch.value.artist.length >= 2
  return hasTitle || hasArtist
})
const isArtistAlreadyExists = computed(() => (artistName) => {
  if (!artistName?.trim()) return false
  return artists.value.some(a => a.name.toLowerCase() === artistName.toLowerCase().trim())
})
const isLabelAlreadyExists = computed(() => (labelName) => {
  if (!labelName?.trim()) return false
  return labels.value.some(l => l.name.toLowerCase() === labelName.toLowerCase().trim())
})
const isGenreAlreadyExists = computed(() => (genreName) => {
  if (!genreName?.trim()) return false
  return genres.value.some(g => g.name.toLowerCase() === genreName.toLowerCase().trim())
})
const isFormatAlreadyExists = computed(() => (formatName) => { // ✅ CORRIGÉ
  if (!formatName?.trim()) return false
  return formats.value.some(f => f.name.toLowerCase() === formatName.toLowerCase().trim())
})
const isCountryAlreadyExists = computed(() => (countryName) => {
  if (!countryName?.trim()) return false
  return countries.value.some(c => c.name.toLowerCase() === countryName.toLowerCase().trim())
})

// === Helpers ===
const decodeAsciiNumeric = (input) => {
  if (!input || !/^\d{3,}$/.test(input) || input.length % 3 !== 0) return null
  let decoded = ''
  for (let i = 0; i < input.length; i += 3) {
    const code = parseInt(input.substr(i, 3), 10)
    if (code >= 48 && code <= 57) decoded += String.fromCharCode(code)
    else return null
  }
  return decoded
}
const cleanBarcode = (input) => input?.trim() ? input.trim().replace(/\D/g, '').slice(-13) : ''
const syncBarcodeToForm = (barcode) => {
  let cleaned = cleanBarcode(barcode)
  const decoded = decodeAsciiNumeric(barcode)
  if (decoded && decoded.length >= 8 && /^[0-9]{8,13}$/.test(decoded)) cleaned = decoded
  searchBarcode.value = cleaned
  formData.value.barcode = cleaned
}
const formatFileSize = (bytes) => {
  if (!bytes) return '0 Bytes'
  const k = 1024, sizes = ['Bytes', 'KB', 'MB']
  const i = Math.min(Math.floor(Math.log(bytes) / Math.log(k)), 2)
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}
const autoFormatPrice = () => {
  if (typeof formData.value.price === 'number') {
    formData.value.price = Math.round(formData.value.price * 100) / 100
  }
}

// === Initialisation ===
const initForm = () => {
  coverCacheBuster.value = Date.now()
  if (props.discData) {
    showSearchStep.value = false
    const disc = props.discData
    formData.value = {
      id: disc.id,
      title: disc.title || '',
      artist_id: disc.artist_id || '',
      genre_id: disc.genre_id || '',
      format_id: disc.format_id || '',
      country_id: disc.country_id || '',
      label_id: disc.label_id || '',
      release_year: disc.release_year || '',
      barcode: disc.barcode || '',
      price: disc.price != null ? parseFloat(disc.price) : null,
      quantity: disc.quantity != null ? parseInt(disc.quantity) : 1,
      notes: disc.notes || '',
      cover_image: disc.cover_image || disc.cover_url || null,
      isrc: disc.isrc || null
    }

    originalFormData.value = JSON.parse(JSON.stringify(formData.value))
    if (disc.cover_url || disc.cover_image) {
      const timestamp = disc.updated_at ? new Date(disc.updated_at).getTime() : coverCacheBuster.value
      currentCoverUrl.value = getImageUrl(disc.cover_url || disc.cover_image, timestamp)
    }
  } else {
    showSearchStep.value = true
    resetForm()
    originalFormData.value = null
    nextTick(() => {
      if (barcodeInput.value) {
        barcodeInput.value.focus()
        searchBarcode.value = ''
      }
    })
  }
  loadOptions()
}

const resetForm = () => {
  searchBarcode.value = ''
  manualSearch.value = { title: '', artist: '', country: '', year: '' }
  searchResults.value = []
  selectedResultId.value = null
  formData.value = {
    id: null,
    title: '',
    artist_id: '',
    genre_id: '',
    format_id: '',
    country_id: '',
    label_id: '',
    release_year: '',
    barcode: '',
    price: null,
    quantity: 1,
    notes: '',
    cover_image: null
  }
  currentCoverUrl.value = null
  resetPreview()
  stopClipboardWatch()
}

const resetPreview = () => {
  coverPreview.value = {
    url: null,
    loading: false,
    found: false,
    title: '',
    artist: '',
    year: '',
    genres: [],
    formats: [],
    country: '',
    label: '',
    prices: null,
    isUploaded: false
  }
}

const loadOptions = async () => {
  try {
    const [g, f, a, c, l] = await Promise.all([
      apiFetch('/genres'),
      apiFetch('/formats'),
      apiFetch('/artists'),
      apiFetch('/countries'),
      apiFetch('/labels')
    ])
    genres.value = Array.isArray(g) ? g : []
    formats.value = Array.isArray(f) ? f : []
    artists.value = Array.isArray(a) ? a : []
    countries.value = Array.isArray(c) ? c : []
    labels.value = Array.isArray(l) ? l : []
  } catch (error) {
    console.error('Erreur chargement options:', error)
  }
}

// === Recherche ===
const handleBarcodeEnter = async () => {
  if (!searchBarcode.value?.trim()) return
  syncBarcodeToForm(searchBarcode.value)
  if (formData.value.barcode?.length >= 8) await handleBarcodeSearch()
}

const handleBarcodeFocus = () => {
  if (scannerDetected.value) {
    scanBuffer.value = ''
  }
}

const handleBarcodeSearch = async () => {
  if (!formData.value.barcode || formData.value.barcode.length < 8) return
  searchLoading.value = true; searchError.value = ''; searchResults.value = []; searchAttempted.value = true
  try {
    const result = await discogsApi.searchByBarcode(formData.value.barcode)
    if (result?.found) {
      coverPreview.value = {
        url: result.cover_url || '',
        loading: false,
        found: true,
        title: result.title || '',
        artist: result.artist || result.artists?.[0]?.name || '',
        year: result.year || '',
        genres: result.genres || [],
        formats: result.formats || [],
        country: result.country || '',
        label: result.label || '',
        prices: result.prices || null,
        isUploaded: false
      }
      searchResults.value = result.results?.length ? result.results : (result.found && result.id ? [result] : [])
      if (searchResults.value.length === 1) {
        selectedResultId.value = searchResults.value[0].id
        await selectSearchResult(searchResults.value[0])
      }
      if (searchResults.value.length > 0) {
        loadPricesForResults()
      }
    } else searchError.value = 'Aucun résultat trouvé'
  } catch (error) { searchError.value = error.message || 'Erreur réseau' }
  finally { 
    searchLoading.value = false
    nextTick(() => {
      if (barcodeInput.value && showSearchStep.value) {
        barcodeInput.value.focus()
      }
    })
  }
}

// ✅ Scoring de pertinence pour trier les résultats
const scoreSearchResult = (result, searchTitle, searchArtist) => {
  let score = 0
  const normalize = (str) => (str || '').toLowerCase().normalize('NFD').replace(/[\u0300-\u036f]/g, '').trim()
  
  const resultTitle = normalize(result.title)
  const resultArtist = normalize(result.artist || result.artists?.[0]?.name || '')
  const queryTitle = normalize(searchTitle)
  const queryArtist = normalize(searchArtist)
  
  // Correspondance exacte du titre (+50)
  if (queryTitle && resultTitle === queryTitle) score += 50
  // Titre commence par la recherche (+30)
  else if (queryTitle && resultTitle.startsWith(queryTitle)) score += 30
  // Titre contient la recherche (+15)
  else if (queryTitle && resultTitle.includes(queryTitle)) score += 15
  
  // Correspondance exacte de l'artiste (+40)
  if (queryArtist && resultArtist === queryArtist) score += 40
  // Artiste commence par la recherche (+25)
  else if (queryArtist && resultArtist.startsWith(queryArtist)) score += 25
  // Artiste contient la recherche (+10)
  else if (queryArtist && resultArtist.includes(queryArtist)) score += 10
  
  // Bonus si année présente (+5)
  if (result.year) score += 5
  // Bonus si pays présent (+3)
  if (result.country) score += 3
  // Bonus si pochette présente (+2)
  if (result.cover_url) score += 2
  // Bonus format Vinyl (+8)
  if (result.formats?.some(f => f.toLowerCase().includes('vinyl'))) score += 8
  
  return score
}

const searchManual = async () => {
  if (!hasValidSearchTerms.value) return
  searchLoading.value = true; searchError.value = ''; searchResults.value = []; searchAttempted.value = true
  try {
    const results = await discogsApi.searchByTitleArtist(
      manualSearch.value.title, 
      manualSearch.value.artist,
      manualSearch.value.country,
      manualSearch.value.year
    )
    // ✅ Trier les résultats par score de pertinence
    if (results?.length) {
      const scored = results.map(r => ({
        ...r,
        _score: scoreSearchResult(r, manualSearch.value.title, manualSearch.value.artist)
      }))
      scored.sort((a, b) => b._score - a._score)
      searchResults.value = scored
    } else {
      searchResults.value = []
    }
    loadPricesForResults()
  } catch (error) { searchError.value = error.message || 'Erreur réseau' }
  finally { searchLoading.value = false }
}

const loadPricesForResults = async () => {
  // ✅ Limite à 8 résultats pour éviter les 500 / rate limit
  const resultsToLoad = searchResults.value.slice(0, 8);

  for (let i = 0; i < resultsToLoad.length; i++) {
    const result = resultsToLoad[i];
    if (result.prices && (result.prices.lowest_price || result.prices.median_price)) continue;

    // Mettre à jour état loading
    searchResults.value = searchResults.value.map(r =>
      r.id === result.id ? { ...r, pricesLoading: true } : r
    );

    try {
      const detailed = await discogsApi.getRelease(result.id);
      const hasPrices = detailed?.prices && (
        detailed.prices.lowest_price > 0 ||
        detailed.prices.median_price > 0 ||
        detailed.prices.highest_price > 0
      );
      searchResults.value = searchResults.value.map(r =>
        r.id === result.id
          ? {
              ...r,
              prices: hasPrices ? detailed.prices : { noPrice: true },
              pricesLoading: false
            }
          : r
      );
    } catch (e) {
      searchResults.value = searchResults.value.map(r =>
        r.id === result.id
          ? { ...r, prices: { noPrice: true }, pricesLoading: false }
          : r
      );
    }

    // ✅ Respect rate limit Discogs (~1 req/sec)
    if (i < resultsToLoad.length - 1) {
      await new Promise(r => setTimeout(r, 700));
    }
  }
};

const selectSearchResult = async (result) => {
  selectedResultId.value = result.id
  try {
    const detailed = await discogsApi.getRelease(result.id)
    if (detailed?.found) {
      const artistName = detailed.artists?.[0]?.name || detailed.artist || result.artist || result.artists?.[0]?.name || ''
      const year = detailed.year || result.year || ''
      coverPreview.value = {
        url: detailed.cover_url || result.cover_url || '',
        loading: false,
        found: true,
        title: detailed.title || result.title,
        artist: artistName,
        year,
        genres: detailed.genres || result.genres || [],
        formats: detailed.formats || result.formats || [],
        country: detailed.country || result.country,
        label: detailed.label || result.label,
        prices: detailed.prices || result.prices,
        isUploaded: false
      }
      const resultIndex = searchResults.value.findIndex(r => r.id === result.id)
      if (resultIndex !== -1) {
        const hasPrices = detailed.prices && (detailed.prices.lowest_price || detailed.prices.median_price)
        searchResults.value[resultIndex] = {
          ...searchResults.value[resultIndex],
          prices: hasPrices ? detailed.prices : { noPrice: true },
          pricesLoading: false
        }
      }
      formData.value.title = coverPreview.value.title
      formData.value.release_year = coverPreview.value.year || ''
      
      // ✅ Alimenter le prix avec le prix moyen Discogs
      const avgPrice = getAveragePrice(coverPreview.value.prices)
      if (avgPrice && !formData.value.price) {
        formData.value.price = avgPrice
      }
      
      const setIfExists = (field, list, name) => {
        const obj = list.value.find(x => x.name.toLowerCase() === name.toLowerCase())
        if (obj) formData.value[field] = obj.id
      }
      if (coverPreview.value.artist) setIfExists('artist_id', artists, coverPreview.value.artist)
      if (coverPreview.value.label) setIfExists('label_id', labels, coverPreview.value.label)
      if (coverPreview.value.genres?.[0]) setIfExists('genre_id', genres, coverPreview.value.genres[0])
      if (coverPreview.value.formats?.[0]) setIfExists('format_id', formats, coverPreview.value.formats[0])
      if (coverPreview.value.country) setIfExists('country_id', countries, coverPreview.value.country)
    }
  } catch (error) {
    console.error('Erreur récupération détails:', error)
  }
  nextTick(() => {
    if (continueButton.value) continueButton.value.focus()
  })
}

const checkBarcodeExists = async () => {
  const bc = formData.value.barcode
  if (!bc || bc.length < 8 || isEditing.value) return
  try {
    const existing = await discogsApi.checkBarcodeExists(bc)
    if (existing) {
      existingDisc.value = existing
      isBarcodeModalOpen.value = true
    }
  } catch (e) { /* ignore */ }
}

// === Upload / Modales / Fermeture / Lifecycle ===
const openFilePicker = () => fileInput.value?.click()

const handleFileUpload = (event) => {
  const file = event.target.files[0]
  if (!file) return
  const validTypes = ['image/jpeg', 'image/jpg', 'image/png', 'image/webp']
  if (!validTypes.includes(file.type) || file.size > 5 * 1024 * 1024) return
  const reader = new FileReader()
  reader.onload = (e) => {
    const img = new Image()
    img.onload = () => {
      Object.assign(uploadPreview.value, {
        url: e.target.result,
        file,
        size: file.size,
        width: img.width,
        height: img.height
      })
    }
    img.src = e.target.result
  }
  reader.readAsDataURL(file)
}

const handleDrop = (e) => { e.preventDefault(); isDragOver.value = false }
const handleDragOver = (e) => { e.preventDefault(); isDragOver.value = true }
const handleDragLeave = (e) => { e.preventDefault(); isDragOver.value = false }
const cancelUpload = () => uploadPreview.value = { url: null, file: null, size: 0, width: 0, height: 0, loading: false }
const confirmUpload = async () => {
  if (!uploadPreview.value.file) return
  uploadPreview.value.loading = true
  try {
    const fd = new FormData()
    fd.append('cover', uploadPreview.value.file)
    fd.append('artist', artists.value.find(a => a.id === formData.value.artist_id)?.name || coverPreview.value.artist || 'Unknown')
    fd.append('title', formData.value.title || coverPreview.value.title || 'Unknown')
    if (formData.value.id) fd.append('disc_id', formData.value.id)
    const res = await upload('/upload-cover', fd)
    if (res.filePath) {
      coverCacheBuster.value = Date.now()
      const fullUrl = getImageUrl(res.filePath, coverCacheBuster.value)
      coverPreview.value = { ...coverPreview.value, url: res.filePath, isUploaded: true, found: true }
      currentCoverUrl.value = fullUrl
      formData.value.cover_image = res.filePath
      cancelUpload()
    }
  } catch (error) { console.error('❌ Erreur upload:', error) }
  finally { uploadPreview.value.loading = false }
}

const removeCover = () => {
  coverPreview.value = { ...coverPreview.value, url: null, isUploaded: false, found: false }
  currentCoverUrl.value = null
  formData.value.cover_image = ''
}

const toggleClipboardWatch = () => isWatchingClipboard.value ? stopClipboardWatch() : startClipboardWatch()

const startClipboardWatch = async () => {
  try {
    const permission = await navigator.permissions.query({ name: 'clipboard-read' })
    if (permission.state === 'denied') return
    const text = await navigator.clipboard.readText()
    lastClipboardValue.value = text || ''
    isWatchingClipboard.value = true
    clipboardWatchInterval.value = setInterval(async () => {
      try {
        const text = await navigator.clipboard.readText()
        if (text && text !== lastClipboardValue.value && /^[0-9]{12,13}$/.test(text) && !isBarcodeSearchLocked.value) {
          lastClipboardValue.value = text
          syncBarcodeToForm(text)
          queueMicrotask(handleBarcodeSearch)
        }
      } catch (e) { /* ignore */ }
    }, 500)
  } catch (e) { console.error('Clipboard error:', e) }
}

const stopClipboardWatch = () => {
  if (clipboardWatchInterval.value) clearInterval(clipboardWatchInterval.value)
  isWatchingClipboard.value = false
}

const openBarcodeScanner = () => isBarcodeScannerOpen.value = true
const handleBarcodeScanned = (barcode) => {
  isBarcodeScannerOpen.value = false
  syncBarcodeToForm(barcode)
  queueMicrotask(handleBarcodeSearch)
}

const triggerFileImport = () => fileImportInput.value?.click()
const handleFileImport = async (e) => {
  const file = e.target.files[0]
  if (!file) return
  const text = await file.text()
  syncBarcodeToForm(text)
  if ([8, 12, 13].includes(formData.value.barcode.length)) await handleBarcodeSearch()
  e.target.value = ''
}

const openArtistModal = () => { artistToCreate.value = ''; isArtistModalOpen.value = true }
const openGenreModal = () => { genreToCreate.value = ''; isGenreModalOpen.value = true }
const openFormatModal = () => { formatToCreate.value = ''; isFormatModalOpen.value = true }
const openCountryModal = () => { countryToCreate.value = null; isCountryModalOpen.value = true }
const openLabelModal = () => { labelToCreate.value = ''; isLabelModalOpen.value = true }
const openArtistModalWithPrefill = (name) => { artistToCreate.value = name; isArtistModalOpen.value = true }
const openGenreModalWithPrefill = (name) => { genreToCreate.value = name; isGenreModalOpen.value = true }
const openFormatModalWithPrefill = (name) => { formatToCreate.value = name; isFormatModalOpen.value = true }
const openCountryModalWithPrefill = (name) => { countryToCreate.value = { name }; isCountryModalOpen.value = true }
const openLabelModalWithPrefill = (name) => { labelToCreate.value = name; isLabelModalOpen.value = true }

const handleArtistSaved = async (artist) => {
  isArtistModalOpen.value = false; artistToCreate.value = ''
  await new Promise(r => setTimeout(r, 300)); await loadOptions()
  if (artist?.name) {
    const id = artist.id || artists.value.find(a => a.name.toLowerCase() === artist.name.toLowerCase())?.id
    if (id) formData.value.artist_id = id
  }
}

const handleGenreSaved = async (genre) => {
  isGenreModalOpen.value = false; genreToCreate.value = ''
  await new Promise(r => setTimeout(r, 300)); await loadOptions()
  if (genre?.name) {
    const id = genre.id || genres.value.find(g => g.name.toLowerCase() === genre.name.toLowerCase())?.id
    if (id) formData.value.genre_id = id
  }
}

const handleFormatSaved = async (format) => {
  isFormatModalOpen.value = false; formatToCreate.value = ''
  await new Promise(r => setTimeout(r, 300)); await loadOptions()
  if (format?.name) {
    const id = format.id || formats.value.find(f => f.name.toLowerCase() === format.name.toLowerCase())?.id
    if (id) formData.value.format_id = id
  }
}

const handleCountrySaved = async (country) => {
  isCountryModalOpen.value = false; countryToCreate.value = null
  await new Promise(r => setTimeout(r, 300)); await loadOptions()
  if (country?.name) {
    const id = country.id || countries.value.find(c => c.name.toLowerCase() === country.name.toLowerCase())?.id
    if (id) formData.value.country_id = id
  }
}

const handleLabelSaved = async (label) => {
  isLabelModalOpen.value = false; labelToCreate.value = ''
  await new Promise(r => setTimeout(r, 300)); await loadOptions()
  if (label?.name) {
    const id = label.id || labels.value.find(l => l.name.toLowerCase() === label.name.toLowerCase())?.id
    if (id) formData.value.label_id = id
  }
}

const handleLabelModalClose = () => { isLabelModalOpen.value = false; labelToCreate.value = '' }
const closeBarcodeModal = () => isBarcodeModalOpen.value = false
const editExistingDisc = () => {
  if (existingDisc.value?.id) {
    closeBarcodeModal()
    emit('edit-existing-disc', existingDisc.value.id)
  }
}
const resetAndRetry = () => {
  closeBarcodeModal()
  formData.value.barcode = ''
  searchBarcode.value = ''
  if (barcodeInput.value) barcodeInput.value.focus()
}
const skipSearch = () => showSearchStep.value = false
const proceedToForm = () => {
  if (coverPreview.value.found) {
    if (!formData.value.title && coverPreview.value.title) {
      formData.value.title = coverPreview.value.title
    }
    if (!formData.value.release_year && coverPreview.value.year) {
      formData.value.release_year = coverPreview.value.year
    }
    const setIfExists = (field, list, name) => {
      if (!name) return
      const obj = list.value.find(x => x.name.toLowerCase() === name.toLowerCase())
      if (obj) formData.value[field] = obj.id
    }
    if (coverPreview.value.artist && !formData.value.artist_id) {
      setIfExists('artist_id', artists, coverPreview.value.artist)
    }
    if (coverPreview.value.label && !formData.value.label_id) {
      setIfExists('label_id', labels, coverPreview.value.label)
    }
    if (coverPreview.value.genres?.[0] && !formData.value.genre_id) {
      setIfExists('genre_id', genres, coverPreview.value.genres[0])
    }
    if (coverPreview.value.formats?.[0] && !formData.value.format_id) {
      setIfExists('format_id', formats, coverPreview.value.formats[0])
    }
    if (coverPreview.value.country && !formData.value.country_id) {
      setIfExists('country_id', countries, coverPreview.value.country)
    }
  }
  showSearchStep.value = false
}
const backToSearch = () => showSearchStep.value = true
const handleSubmit = async () => {
  if (!isFormValid.value) return
  isSaving.value = true
  try {
    if (!props.discData && formData.value.barcode) {
      try {
        const existing = await discogsApi.checkBarcodeExists(formData.value.barcode)
        if (existing) {
          existingDisc.value = existing
          isBarcodeModalOpen.value = true
          isSaving.value = false
          return
        }
      } catch (e) { /* ignore */ }
    }
    
    const data = { ...formData.value }
    
    const isExplicitDeletion = formData.value.cover_image === ''
    const isEditMode = !!props.discData
    
    if (isExplicitDeletion) {
      data.cover_image = ''
    } else if (coverPreview.value.isUploaded && coverPreview.value.url) {
      data.cover_image = coverPreview.value.url
    } else if (coverPreview.value.found && coverPreview.value.url) {
      data.cover_image = coverPreview.value.url.startsWith('http') ? coverPreview.value.url : coverPreview.value.url
    } else if (isEditMode && (currentCoverUrl.value || props.discData.cover_image || props.discData.cover_url)) {
      let path = currentCoverUrl.value || props.discData.cover_image || props.discData.cover_url
      path = path.split('?')[0].split('#')[0]
      try { const url = new URL(path); path = url.pathname } catch (e) {
        if (!path.startsWith('/uploads/')) {
          const idx = path.indexOf('/uploads/')
          if (idx !== -1) path = path.slice(idx)
          else path = '/uploads/' + path.replace(/^\/?/, '')
        }
      }
      path = decodeURIComponent(path)
      data.cover_image = path
    } else {
      data.cover_image = null
    }
    
    // ✅ Nettoyer les champs vides (transformer '' en null)
    Object.keys(data).forEach(k => {
      if (k !== 'cover_image' && k !== 'price' && (data[k] === '' || data[k] === undefined)) {
        data[k] = null
      }
    })
    
    // ✅ Prix : null ou nombre valide
    if (data.price == null || isNaN(data.price)) data.price = null
    else data.price = parseFloat(data.price)
    
    data.isrc = data.isrc || null

    emit('save', data)
  } catch (error) { 
    console.error('❌ Erreur handleSubmit:', error) 
  } finally { 
    isSaving.value = false 
  }
}
const onOverlayMouseDown = (e) => overlayMouseDownTarget.value = e.target
const onOverlayClick = (e) => {
  if (e.target === e.currentTarget && overlayMouseDownTarget.value === e.currentTarget) handleCloseRequest()
  overlayMouseDownTarget.value = null
}
const handleCloseRequest = () => {
  if (isSaving.value) return
  let hasChanges = false
  if (props.discData && originalFormData.value) {
    const cur = formData.value, orig = originalFormData.value
    const norm = v => (v == null || v === '') ? '' : (typeof v === 'number' ? v.toString() : v)
    hasChanges = 
      norm(cur.title) !== norm(orig.title) ||
      norm(cur.artist_id) !== norm(orig.artist_id) ||
      norm(cur.genre_id) !== norm(orig.genre_id) ||
      norm(cur.format_id) !== norm(orig.format_id) ||
      norm(cur.country_id) !== norm(orig.country_id) ||
      norm(cur.label_id) !== norm(orig.label_id) ||
      norm(cur.release_year) !== norm(orig.release_year) ||
      norm(cur.barcode) !== norm(orig.barcode) ||
      norm(cur.price) !== norm(orig.price) ||
      norm(cur.quantity) !== norm(orig.quantity) ||
      norm(cur.notes) !== norm(orig.notes) ||
      !!uploadPreview.value.url
  } else {
    const pure = showSearchStep.value && !coverPreview.value.found && !formData.value.title && !formData.value.artist_id
    hasChanges = !pure && (formData.value.title || formData.value.artist_id || coverPreview.value.found || formData.value.price)
  }
  if (hasChanges) {
    isUnsavedChangesModalOpen.value = true
    return
  }
  resetForm()
  emit('close')
}
const confirmCloseWithoutSaving = () => {
  isUnsavedChangesModalOpen.value = false
  resetForm()
  emit('close')
}
const cancelClose = () => isUnsavedChangesModalOpen.value = false
const handleImageError = (e) => {
  e.target.src = 'image/svg+xml;charset=UTF-8,%3Csvg xmlns="http://www.w3.org/2000/svg" width="200" height="200"%3E%3Crect fill="%23e9ecef" width="200" height="200"/%3E%3Ctext x="50%25" y="50%25" dominant-baseline="middle" text-anchor="middle" font-family="sans-serif" font-size="14" fill="%236c757d"%3EPas de pochette%3C/text%3E%3C/svg%3E'
}
const getArtistName = (id) => artists.value.find(a => a.id === id)?.name || ''

watch(() => props.isOpen, (open) => { if (open) initForm(); else stopClipboardWatch() })
onMounted(() => document.addEventListener('keydown', handleKeyInput))
onBeforeUnmount(() => {
  document.removeEventListener('keydown', handleKeyInput)
  stopClipboardWatch()
})

const handleKeyInput = (e) => {
  if (!showSearchStep.value || !barcodeInput.value || document.activeElement !== barcodeInput.value) return
  if (searchLoading.value || isBarcodeSearchLocked.value) return
  if (e.key === 'Enter') {
    e.preventDefault()
    handleBarcodeEnter()
    return
  }
  if (/^\d$/.test(e.key)) {
    const now = Date.now()
    const timeSinceLastKey = now - lastKeyTime.value
    lastKeyTime.value = now
    if (timeSinceLastKey < 50) {
      scannerDetected.value = true
      scanBuffer.value += e.key
    } else {
      scanBuffer.value = e.key
      scannerDetected.value = false
    }
    if (barcodeSearchTimeout.value) clearTimeout(barcodeSearchTimeout.value)
    barcodeSearchTimeout.value = setTimeout(async () => {
      const currentVal = searchBarcode.value.trim()
      if (scannerDetected.value && /^[0-9]{8,13}$/.test(currentVal)) {
        syncBarcodeToForm(currentVal)
        isBarcodeSearchLocked.value = true
        try { 
          await handleBarcodeSearch() 
        } finally { 
          isBarcodeSearchLocked.value = false
          scannerDetected.value = false
          scanBuffer.value = ''
          nextTick(() => {
            if (barcodeInput.value && showSearchStep.value) {
              barcodeInput.value.focus()
            }
          })
        }
      }
    }, 100)
  }
}
</script>

<style scoped>
/* ✅ Optimisations UI + style prix */
.discs-modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: var(--modal-overlay);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 9999;
  padding: 20px;
  overflow-y: auto;
}
.discs-modal {
  background: var(--modal-bg);
  border-radius: 20px;
  width: 100%;
  max-width: 1200px;
  max-height: 95vh;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  box-shadow: 0 25px 80px rgba(0, 0, 0, 0.5);
}
.discs-modal.is-saving {
  pointer-events: none;
  opacity: 0.9;
}
.modal-header {
  padding: 24px 32px;
  background: linear-gradient(135deg, var(--accent), var(--accent-blue));
  color: white;
  display: flex;
  justify-content: space-between;
  align-items: center;
  border-bottom: 3px solid rgba(255, 255, 255, 0.2);
}
.header-content { flex: 1; }
.modal-title { margin: 0 0 6px 0; font-size: 1.8em; font-weight: 700; }
.modal-subtitle { margin: 0; font-size: 1em; opacity: 0.95; }
.close-button {
  width: 44px; height: 44px; border: none; background: rgba(255, 255, 255, 0.15);
  color: white; border-radius: 50%; cursor: pointer; font-size: 1.8em;
  display: flex; align-items: center; justify-content: center; transition: all 0.2s;
}
.close-button:hover:not(:disabled) {
  background: rgba(255, 255, 255, 0.25); transform: rotate(90deg);
}
.close-button:disabled { opacity: 0.5; cursor: not-allowed; }
.modal-body { flex: 1; overflow-y: auto; padding: 0; }
.search-step { padding: 32px; }
.step-header {
  display: flex;
  align-items: center;
  gap: 16px;
  margin-bottom: 24px;
}
.step-number {
  width: 36px; height: 36px; border-radius: 50%; background: var(--line); color: var(--text-soft);
  display: flex; align-items: center; justify-content: center; font-weight: bold; font-size: 18px;
}
.step-number.active {
  background: linear-gradient(135deg, var(--accent), var(--accent-blue));
  color: white;
  box-shadow: 0 4px 12px rgba(59, 130, 246, 0.3);
}
.step-title { margin: 0; font-size: 1.4em; color: var(--text); font-weight: 600; }
.search-method {
  margin-bottom: 24px;
  padding: 24px;
  background: rgba(var(--tint-rgb), 0.03);
  border-radius: 12px;
  box-shadow: 0 2px 8px rgba(0,0,0,0.08);
  border: 1px solid var(--line-soft);
}
.method-title { margin: 0 0 16px 0; font-size: 1.1em; color: var(--accent); display: flex; align-items: center; gap: 8px; }
.barcode-input-container { margin-bottom: 8px; }
.barcode-input-group { display: flex; gap: 12px; flex-wrap: wrap; }
.barcode-input {
  flex: 1; min-width: 200px;
  padding: 12px 16px;
  border: 2px solid var(--line);
  border-radius: 8px;
  font-size: 15px;
  transition: all 0.2s;
  background: rgba(var(--tint-rgb), 0.04);
  color: var(--text);
}
.barcode-input:focus {
  outline: none;
  border-color: var(--accent);
  box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
}
.barcode-input.scanner-input {
  border-color: #22c55e !important;
  box-shadow: 0 0 0 3px rgba(34, 197, 94, 0.1) !important;
  background: rgba(34, 197, 94, 0.05);
}
.scan-button {
  padding: 12px 16px;
  background: #22c55e;
  color: white;
  border: none;
  border-radius: 8px;
  cursor: pointer;
  font-size: 1.3em;
  transition: all 0.2s;
  min-width: 44px;
  display: flex;
  align-items: center;
  justify-content: center;
}
.scan-button:hover { background: #16a34a; transform: scale(1.05); }
.scan-button:disabled { opacity: 0.5; cursor: not-allowed; }
.scan-button.active {
  background: #22c55e;
  animation: pulse-green 1.5s ease-in-out infinite;
}
@keyframes pulse-green {
  0%, 100% { box-shadow: 0 0 0 0 rgba(34, 197, 94, 0.7); }
  50% { box-shadow: 0 0 0 10px rgba(34, 197, 94, 0); }
}
.input-help { margin-top: 8px; font-size: 0.9em; color: var(--text-soft); }
.input-help.scanner-detected {
  color: #22c55e;
  font-weight: 600;
  animation: pulse-text 1.5s ease-in-out infinite;
}
.reset-scan-zone {
  margin-top: 12px;
  display: flex;
  align-items: center;
  gap: 12px;
}
.btn-reset-scan {
  padding: 8px 16px;
  background: rgba(var(--tint-rgb), 0.04);
  border: 2px solid var(--line);
  border-radius: 8px;
  color: var(--text-soft);
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
}
.btn-reset-scan:hover:not(:disabled) { background: rgba(var(--tint-rgb), 0.08); }
.or-separator {
  display: flex;
  align-items: center;
  margin: 24px 0;
  color: var(--text-soft);
}
.or-line { flex: 1; height: 1px; background: var(--line); }
.or-text { margin: 0 16px; font-weight: 600; }
.manual-search-form { margin-top: 16px; }
.form-row { display: flex; gap: 12px; flex-wrap: wrap; }
.form-row.centered {
  justify-content: center;
  margin-top: 4px;
}
.search-input {
  flex: 1;
  min-width: 150px;
  padding: 12px 16px;
  border: 2px solid var(--line);
  border-radius: 8px;
  font-size: 15px;
  background: rgba(var(--tint-rgb), 0.04);
  color: var(--text);
}
.search-button {
  padding: 12px 24px;
  background: linear-gradient(135deg, var(--accent), var(--accent-blue));
  color: white;
  border: none;
  border-radius: 8px;
  font-weight: 600;
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
  transition: all 0.2s;
}
.search-button:hover:not(:disabled) {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.2);
}
.search-button:disabled { opacity: 0.5; cursor: not-allowed; }
.search-status {
  margin-top: 8px;
  padding: 8px 12px;
  border-radius: 8px;
  font-size: 0.9em;
  text-align: center;
}
.search-status.success {
  background: rgba(34, 197, 94, 0.14);
  border: 1px solid rgba(34, 197, 94, 0.28);
  color: var(--positive-text);
}
.search-status.warning {
  background: rgba(242, 168, 120, 0.14);
  border: 1px solid rgba(242, 168, 120, 0.32);
  color: var(--accent-sand);
}
.search-results { margin-top: 32px; }
.results-title {
  margin: 0 0 20px 0;
  font-size: 1.2em;
  color: var(--text);
  display: flex;
  align-items: center;
  gap: 8px;
}

/* ✅ 3 cartes par ligne — largeur fixe */
.results-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
  gap: 12px;
  max-width: 100%;
}

/* ═══════════════════════════════════════════════════════════════
   RESULT CARD - Design épuré et moderne
   ═══════════════════════════════════════════════════════════════ */
.result-card {
  background: var(--bg-elevated);
  border: 1px solid var(--line);
  border-radius: 10px;
  padding: 12px;
  cursor: pointer;
  display: flex;
  gap: 12px;
  align-items: flex-start;
  transition: all 0.2s ease;
}
.result-card:hover {
  border-color: var(--accent);
  box-shadow: 0 4px 12px rgba(59, 130, 246, 0.15);
  transform: translateY(-2px);
}
.result-card.selected {
  border-color: var(--accent);
  background: linear-gradient(135deg, rgba(59, 130, 246, 0.05) 0%, rgba(242, 168, 120, 0.05) 100%);
  box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
}

/* Pochette */
.result-cover {
  position: relative;
  width: 72px;
  height: 72px;
  border-radius: 6px;
  overflow: hidden;
  flex-shrink: 0;
  background: linear-gradient(135deg, rgba(var(--tint-rgb), 0.06) 0%, var(--line) 100%);
}
.cover-image {
  width: 100%;
  height: 100%;
  object-fit: cover;
}
.cover-placeholder {
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, var(--line) 0%, var(--line) 100%);
}
.placeholder-icon {
  font-size: 28px;
  opacity: 0.5;
}

/* Badge prix sur la pochette */
.price-badge {
  position: absolute;
  bottom: 4px;
  right: 4px;
  background: linear-gradient(135deg, #22c55e 0%, #16a34a 100%);
  color: white;
  font-size: 10px;
  font-weight: 700;
  padding: 2px 6px;
  border-radius: 4px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.2);
}
.price-badge.loading {
  background: var(--text-dim);
  animation: pulse 1.5s infinite;
}
@keyframes pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.6; }
}

/* Infos */
.result-info {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 4px;
}
.result-title {
  margin: 0;
  font-size: 13px;
  font-weight: 600;
  color: var(--text);
  line-height: 1.3;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}
.result-artist {
  margin: 0;
  font-size: 12px;
  color: var(--text-soft);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

/* Meta tags */
.result-meta {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
  margin-top: 4px;
}
.meta-item {
  font-size: 10px;
  color: var(--text-soft);
  background: rgba(var(--tint-rgb), 0.06);
  padding: 2px 6px;
  border-radius: 4px;
  white-space: nowrap;
  max-width: 120px;
  overflow: hidden;
  text-overflow: ellipsis;
}

/* Footer: ID Discogs + Bouton */
.result-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-top: auto;
  padding-top: 8px;
}
.discogs-link {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  font-size: 11px;
  color: var(--accent);
  text-decoration: none;
  font-weight: 500;
  padding: 3px 8px;
  border-radius: 4px;
  background: rgba(var(--tint-rgb), 0.04);
  transition: all 0.2s;
}
.discogs-link:hover {
  background: var(--accent);
  color: white;
}
.discogs-icon {
  width: 14px;
  height: 14px;
  background: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 24 24'%3E%3Ccircle cx='12' cy='12' r='10' fill='%233b82f6'/%3E%3Ccircle cx='12' cy='12' r='7' fill='%23fff'/%3E%3Ccircle cx='12' cy='12' r='3' fill='%233b82f6'/%3E%3C/svg%3E") no-repeat center/contain;
}
.discogs-link:hover .discogs-icon {
  background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 24 24'%3E%3Ccircle cx='12' cy='12' r='10' fill='%23fff'/%3E%3Ccircle cx='12' cy='12' r='7' fill='%233b82f6'/%3E%3Ccircle cx='12' cy='12' r='3' fill='%23fff'/%3E%3C/svg%3E");
}

/* Bouton sélection */
.select-button {
  padding: 5px 12px;
  background: linear-gradient(135deg, var(--accent), var(--accent-blue));
  color: white;
  border: none;
  border-radius: 6px;
  font-size: 11px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
}
.select-button:hover {
  background: linear-gradient(135deg, var(--accent-blue), var(--accent));
  transform: scale(1.02);
}
.select-button.is-selected {
  background: linear-gradient(135deg, #22c55e 0%, #16a34a 100%);
}
.result-card.selected .select-button {
  background: linear-gradient(135deg, #22c55e 0%, #16a34a 100%);
}

/* ✅ Reste inchangé */
.preview-section {
  background: rgba(34, 197, 94, 0.14);
  border: 2px solid #22c55e;
  border-radius: 12px;
  padding: 16px;
  margin-top: 20px;
}
.preview-header {
  display: flex;
  justify-content: space-between;
  margin-bottom: 12px;
  align-items: center;
}
.preview-header h4 {
  margin: 0;
  font-size: 1.1em;
  color: var(--positive-text);
  font-weight: 600;
}
.preview-content {
  display: flex;
  gap: 12px;
  align-items: flex-start;
}
.preview-image {
  width: 80px;
  height: 80px;
  object-fit: cover;
  border-radius: 8px;
  flex-shrink: 0;
  box-shadow: 0 2px 6px rgba(0,0,0,0.1);
}
.preview-placeholder {
  width: 80px;
  height: 80px;
  background: rgba(var(--tint-rgb), 0.08);
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 32px;
  color: var(--text-soft);
}
.preview-info {
  flex: 1;
  min-width: 0;
}
.preview-info h5 {
  margin: 0 0 4px 0;
  font-size: 1.15em;
  font-weight: 700;
  color: var(--text);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.preview-details {
  display: flex;
  flex-direction: column;
  gap: 3px;
  font-size: 0.95em;
}
.detail-item {
  margin: 0;
  color: var(--text-soft);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.detail-item strong {
  font-weight: 600;
  color: #22c55e;
  margin-right: 4px;
}
.preview-header .btn.secondary.small {
  padding: 4px 10px;
  font-size: 0.85em;
}

.step-navigation {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-top: 32px;
  padding-top: 24px;
  border-top: 1px solid var(--line-soft);
}
.skip-button,
.continue-button,
.back-button {
  padding: 12px 24px;
  border: none;
  border-radius: 8px;
  font-weight: 600;
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 8px;
  transition: all 0.2s;
}
.skip-button,
.back-button {
  background: var(--text-soft);
  color: white;
}
.skip-button:hover:not(:disabled),
.back-button:hover:not(:disabled) {
  background: var(--text-soft);
}
.continue-button {
  background: linear-gradient(135deg, var(--accent), var(--accent-blue));
  color: white;
}
.continue-button:hover:not(:disabled) {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.2);
}
.form-step {
  padding: 32px;
}
.selection-summary {
  background: rgba(59, 130, 246, 0.08);
  border: 2px solid rgba(59, 130, 246, 0.24);
  border-radius: 12px;
  padding: 20px;
  margin-bottom: 24px;
}
.summary-content {
  display: flex;
  gap: 16px;
}
.summary-label {
  margin-top: 4px;
  color: var(--text-soft);
  font-size: 0.9em;
}
.cover-container {
  position: relative;
  display: inline-block;
}
.cover-container:hover .cover-actions {
  opacity: 1;
}
.cover-actions {
  position: absolute;
  top: 4px;
  right: 4px;
  display: flex;
  gap: 4px;
  opacity: 0;
  transition: opacity 0.2s;
}
.cover-action-btn {
  background: rgba(0, 0, 0, 0.7);
  border: none;
  border-radius: 4px;
  color: white;
  cursor: pointer;
  padding: 4px 8px;
  transition: all 0.2s;
}
.cover-action-btn:hover {
  background: rgba(0, 0, 0, 0.9);
  transform: scale(1.1);
}
.cover-action-btn.remove {
  background: rgba(220, 53, 69, 0.7);
}
.cover-action-btn.remove:hover {
  background: rgba(220, 53, 69, 0.9);
}
.summary-image {
  width: 60px;
  height: 60px;
  object-fit: cover;
  border-radius: 8px;
}
.summary-placeholder {
  width: 60px;
  height: 60px;
  background: rgba(var(--tint-rgb), 0.08);
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 24px;
  color: var(--text-soft);
}
.summary-details h4 {
  margin: 0 0 4px 0;
  color: var(--text-soft);
}
.summary-details p {
  margin: 0;
  color: var(--text-soft);
}
.upload-preview {
  background: rgba(var(--tint-rgb), 0.04);
  border: 2px solid rgba(var(--tint-rgb), 0.08);
  border-radius: 12px;
  padding: 20px;
  margin-top: 16px;
}
.upload-preview .preview-content {
  display: flex;
  gap: 20px;
}
.upload-preview .preview-image {
  width: 80px;
  height: 80px;
  object-fit: cover;
  border-radius: 8px;
}
.file-info {
  font-size: 0.9em;
  color: var(--text-soft);
}
.btn {
  padding: 8px 16px;
  border: none;
  border-radius: 8px;
  font-weight: 600;
  cursor: pointer;
  font-size: 0.9em;
  transition: all 0.2s;
}
.btn.primary {
  background: #22c55e;
  color: white;
}
.btn.primary:hover:not(:disabled) {
  background: #16a34a;
}
.btn.secondary {
  background: var(--text-soft);
  color: white;
}
.btn.secondary:hover:not(:disabled) {
  background: var(--text);
}
.btn.small {
  padding: 6px 12px;
  font-size: 0.85em;
}
.btn.large {
  padding: 14px 28px;
  font-size: 1.1em;
}
.btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}
.btn-spinner {
  width: 16px;
  height: 16px;
  border: 2px solid transparent;
  border-top: 2px solid currentColor;
  border-radius: 50%;
  animation: spin 1s linear infinite;
  display: inline-block;
  margin-right: 8px;
}
.upload-section {
  margin-bottom: 24px;
}
.upload-zone {
  border: 2px dashed var(--line);
  border-radius: 12px;
  padding: 40px 20px;
  text-align: center;
  cursor: pointer;
  background: rgba(var(--tint-rgb), 0.04);
  transition: all 0.2s;
}
.upload-zone:hover {
  border-color: var(--accent);
  background: rgba(var(--tint-rgb), 0.08);
}
.upload-zone.drag-over {
  border-color: var(--accent);
  background: rgba(59, 130, 246, 0.08);
  transform: scale(1.02);
}
.upload-content {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 12px;
}
.upload-icon {
  font-size: 3em;
  color: var(--text-soft);
}
.upload-content h5 {
  margin: 0;
  color: var(--text-soft);
}
.upload-content p {
  margin: 0;
  color: var(--text-soft);
}
.upload-help {
  font-size: 0.85em;
}
.form-section {
  background: rgba(var(--tint-rgb), 0.03);
  padding: 16px 20px;
  border-radius: 10px;
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.05);
  border: 1px solid var(--line);
  margin-bottom: 12px;
}
.section-title {
  margin: 0 0 12px 0;
  font-size: 0.95em;
  font-weight: 600;
  color: var(--text-soft);
  padding-bottom: 8px;
  border-bottom: 1px solid rgba(var(--tint-rgb), 0.06);
  display: flex;
  align-items: center;
  gap: 6px;
}
.form-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 16px;
}
.form-grid.compact {
  gap: 12px 16px;
}
.form-grid.compact .form-group {
  gap: 6px;
}
.form-grid.compact .form-group label {
  font-size: 13px;
}
.form-grid.compact .form-input,
.form-grid.compact .form-select {
  padding: 10px 12px;
  font-size: 14px;
}
.form-grid.compact .form-help {
  font-size: 11px;
  margin-top: 4px;
}
.form-group {
  display: flex;
  flex-direction: column;
  gap: 8px;
}
.form-group.full-width {
  grid-column: 1 / -1;
}
.form-group label {
  font-weight: 600;
  color: var(--text-soft);
}
.form-group label.required::after {
  content: ' *';
  color: var(--negative-text);
}
.form-input,
.form-select,
.form-textarea {
  width: 100%;
  padding: 12px 16px;
  border: 2px solid var(--line);
  border-radius: 8px;
  font-size: 15px;
  transition: all 0.2s;
  box-sizing: border-box;
  background: rgba(var(--tint-rgb), 0.04);
  color: var(--text);
}
.form-input:focus,
.form-select:focus,
.form-textarea:focus {
  outline: none;
  border-color: var(--accent);
  box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
}

/* ═══════════════════════════════════════════════════════════════
   PRIX DISCOGS - Suggestion et application
   ═══════════════════════════════════════════════════════════════ */
.price-input-wrapper {
  display: flex;
  flex-direction: column;
  gap: 6px;
}
.price-input-wrapper .form-input {
  flex: 1;
}
.form-input.has-discogs-price {
  border-color: #22c55e;
}
.discogs-price-suggestion {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 6px 10px;
  background: linear-gradient(135deg, rgba(34, 197, 94, 0.1) 0%, rgba(22, 163, 74, 0.1) 100%);
  border: 1px solid rgba(34, 197, 94, 0.3);
  border-radius: 6px;
  font-size: 12px;
}
.suggestion-label {
  color: #16a34a;
  font-weight: 500;
}
.suggestion-value {
  color: #15803d;
  font-weight: 700;
}
.apply-price-btn {
  margin-left: auto;
  padding: 3px 8px;
  background: linear-gradient(135deg, #22c55e 0%, #16a34a 100%);
  color: white;
  border: none;
  border-radius: 4px;
  font-size: 11px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
}
.apply-price-btn:hover:not(:disabled) {
  background: linear-gradient(135deg, #16a34a 0%, #15803d 100%);
  transform: scale(1.05);
}
.apply-price-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.form-textarea {
  resize: vertical;
  min-height: 100px;
}
.form-select {
  cursor: pointer;
  -webkit-appearance: none;
  -moz-appearance: none;
  appearance: none;
  background-image: url("image/svg+xml;charset=UTF-8,%3csvg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 24 24' fill='none' stroke='currentColor' stroke-width='2' stroke-linecap='round' stroke-linejoin='round'%3e%3cpolyline points='6 9 12 15 18 9'%3e%3c/polyline%3e%3c/svg%3e");
  background-repeat: no-repeat;
  background-position: right 16px center;
  background-size: 16px;
  padding-right: 40px;
}
.artist-input-group,
.select-with-button {
  display: flex;
  gap: 8px;
  width: 100%;
}
.artist-input-group .form-select,
.select-with-button .form-select {
  flex: 1;
  min-width: 0;
  width: auto;
}
.create-artist-button {
  padding: 0 14px;
  background: var(--accent);
  color: white;
  border: none;
  border-radius: 8px;
  cursor: pointer;
  width: 44px;
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.2s;
}
.create-artist-button:hover:not(:disabled) {
  background: var(--accent-blue);
}
.create-artist-button:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}
.form-help {
  margin-top: 8px;
  font-size: 0.85em;
  color: var(--text-soft);
}
.inline-button {
  background: none;
  border: none;
  color: var(--accent);
  cursor: pointer;
  font-weight: 600;
  margin-left: 8px;
  text-decoration: underline;
  padding: 0;
  font-size: inherit;
  transition: all 0.2s;
}
.inline-button:hover {
  color: var(--accent-blue);
}
.char-counter {
  text-align: right;
  font-size: 0.8em;
  color: var(--text-soft);
}
.form-actions {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 16px;
  padding-top: 24px;
  border-top: 1px solid var(--line-soft);
  margin-top: 16px;
}
.back-action {
  flex-shrink: 0;
}
.main-actions {
  display: flex;
  justify-content: flex-end;
  flex: 1;
}
.action-group {
  display: flex;
  gap: 12px;
}
.cancel-button,
.save-button {
  padding: 14px 32px;
  border: none;
  border-radius: 8px;
  font-weight: 600;
  font-size: 1em;
  cursor: pointer;
  min-width: 120px;
  text-align: center;
  transition: all 0.2s;
}
.cancel-button {
  background: var(--text-soft);
  color: white;
}
.cancel-button:hover:not(:disabled) {
  background: var(--text-soft);
}
.cancel-button:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}
.save-button {
  background: linear-gradient(135deg, var(--accent), var(--accent-blue));
  color: white;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
}
.save-button:hover:not(:disabled) {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.2);
}
.save-button:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}
.spinner {
  width: 20px;
  height: 20px;
  border: 2px solid rgba(255, 255, 255, 0.3);
  border-top: 2px solid white;
  border-radius: 50%;
  animation: spin 1s linear infinite;
}
@keyframes spin {
  0% { transform: rotate(0deg); }
  100% { transform: rotate(360deg); }
}
/* Les deux mini-modals (code-barres existant / modifications non
   enregistrées) utilisent désormais .modal-overlay/.modal-card/.modal-actions
   globaux (App.vue) — seuls les styles de contenu spécifiques restent ici. */
.header-title-container {
  display: flex;
  gap: 16px;
  flex: 1;
}
.title-icon {
  font-size: 2em;
}
.header-text h2 {
  margin: 0 0 8px 0;
  font-size: 1.5em;
}
.header-subtitle {
  margin: 0;
  font-size: 0.9em;
  opacity: 0.9;
}
.existing-disc-info {
  background: rgba(242, 168, 120, 0.14);
  border: 1px solid rgba(242, 168, 120, 0.32);
  border-radius: 12px;
  padding: 20px;
  margin-bottom: 24px;
}
.info-section h4 {
  margin: 0 0 12px 0;
  color: var(--accent-sand);
}
.info-section p {
  margin: 0;
  color: var(--text-soft);
}
.disc-details {
  display: flex;
  flex-direction: column;
  gap: 12px;
}
.detail-row {
  display: flex;
  justify-content: space-between;
  padding: 8px 0;
  border-bottom: 1px solid var(--line-soft);
  color: var(--text-soft);
}
.detail-row:last-child {
  border-bottom: none;
}
.modal-fade-enter-active,
.modal-fade-leave-active {
  transition: all 0.3s ease;
}
.modal-fade-enter-from,
.modal-fade-leave-to {
  opacity: 0;
}
.modal-fade-enter-from .discs-modal,
.modal-fade-leave-to .discs-modal {
  transform: scale(0.9) translateY(20px);
}
@media (max-width: 768px) {
  .barcode-input-group,
  .form-row {
    flex-direction: column;
  }
  .form-row.centered {
    flex-direction: row;
  }
  .form-grid {
    grid-template-columns: 1fr;
  }
  .results-grid {
    grid-template-columns: repeat(auto-fill, minmax(160px, 1fr));
    gap: 12px;
  }
  @media (max-width: 576px) {
    .results-grid {
      grid-template-columns: 1fr;
    }
  }
  .preview-content,
  .upload-preview .preview-content,
  .summary-content {
    flex-direction: column;
    text-align: center;
  }
  .step-navigation,
  .form-actions,
  .action-buttons {
    flex-direction: column;
    gap: 12px;
  }
  .skip-button,
  .continue-button,
  .back-button,
  .cancel-button,
  .save-button {
    width: 100%;
  }
  .artist-input-group,
  .select-with-button {
    flex-direction: column;
  }
  .create-artist-button {
    width: 100%;
    padding: 8px;
    margin-top: 4px;
  }
  .back-action,
  .main-actions {
    width: 100%;
  }
  .action-group {
    width: 100%;
    display: flex;
    gap: 8px;
  }
}
</style>