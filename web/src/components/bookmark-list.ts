import { LitElement, html, css } from 'lit'
import { customElement, state, property } from 'lit/decorators.js'
import { apiService, type Bookmark, type BookmarkListResponse } from '../services/api.ts'
import './bookmark-item.ts'

@customElement('bookmark-list')
export class BookmarkList extends LitElement {
  @state() private _bookmarks: Bookmark[] = []
  @state() private _loading = true
  @state() private _error: string | null = null
  @state() private _stats = {
    total: 0,
    favorites: 0,
    tags: 0,
    hasMore: false,
    page: 1
  }

  @property() searchQuery = ''
  @property() tagFilter = ''
  @property() favoritesOnly = false

  static styles = css`
    :host {
      display: block;
    }

    .loading {
      text-align: center;
      padding: 2rem;
      color: #00ffff;
      display: flex;
      flex-direction: column;
      align-items: center;
      gap: 1rem;
    }

    .loading-spinner {
      width: 40px;
      height: 40px;
      border: 3px solid rgba(0, 255, 255, 0.2);
      border-top: 3px solid #00ffff;
      border-radius: 50%;
      animation: spin 1s linear infinite;
    }

    @keyframes spin {
      0% { transform: rotate(0deg); }
      100% { transform: rotate(360deg); }
    }

    .error {
      text-align: center;
      padding: 2rem;
      color: #ff1744;
      background: rgba(255, 23, 68, 0.1);
      border: 1px solid rgba(255, 23, 68, 0.3);
      border-radius: 0.5rem;
      margin-bottom: 1rem;
    }

    .retry-button {
      background: transparent;
      border: 1px solid #ff1744;
      color: #ff1744;
      padding: 0.5rem 1rem;
      border-radius: 0.25rem;
      cursor: pointer;
      margin-top: 0.5rem;
      transition: all 0.3s ease;
    }

    .retry-button:hover {
      background: #ff1744;
      color: white;
    }

    .empty-state {
      text-align: center;
      padding: 3rem 1rem;
      color: #666;
    }

    .empty-icon {
      font-size: 3rem;
      margin-bottom: 1rem;
      opacity: 0.5;
    }

    .bookmark-grid {
      display: grid;
      gap: 1rem;
      grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
    }

    .stats {
      display: flex;
      justify-content: space-between;
      align-items: center;
      margin-bottom: 1.5rem;
      padding: 1rem;
      background: rgba(0, 255, 255, 0.05);
      border: 1px solid rgba(0, 255, 255, 0.1);
      border-radius: 0.5rem;
    }

    .stats-item {
      display: flex;
      flex-direction: column;
      align-items: center;
      gap: 0.25rem;
    }

    .stats-number {
      font-size: 1.25rem;
      font-weight: bold;
      color: #00ffff;
    }

    .stats-label {
      font-size: 0.75rem;
      color: #a0a0a0;
      text-transform: uppercase;
      letter-spacing: 1px;
    }

    .load-more {
      text-align: center;
      margin-top: 2rem;
    }

    .load-more-button {
      background: linear-gradient(45deg, rgba(255, 0, 128, 0.1) 0%, transparent 50%);
      border: 1px solid #ff0080;
      color: #ff0080;
      padding: 0.75rem 1.5rem;
      border-radius: 0.5rem;
      cursor: pointer;
      transition: all 0.3s ease;
      position: relative;
      overflow: hidden;
    }

    .load-more-button:hover {
      background: #ff0080;
      color: black;
      box-shadow: 0 0 20px rgba(255, 0, 128, 0.5);
    }

    .load-more-button:disabled {
      opacity: 0.5;
      cursor: not-allowed;
    }

    @media (max-width: 768px) {
      .bookmark-grid {
        grid-template-columns: 1fr;
      }

      .stats {
        flex-direction: column;
        gap: 1rem;
        text-align: center;
      }
    }
  `

  connectedCallback() {
    super.connectedCallback()
    this.loadBookmarks()
  }

  updated(changedProperties: Map<string, any>) {
    if (changedProperties.has('searchQuery') || 
        changedProperties.has('tagFilter') || 
        changedProperties.has('favoritesOnly')) {
      this.loadBookmarks(true) // Reset to first page
    }
  }

  async loadBookmarks(reset = false) {
    if (reset) {
      this._stats.page = 1
      this._bookmarks = []
    }

    this._loading = true
    this._error = null

    try {
      let response: BookmarkListResponse

      // Use full-text search if search query exists
      if (this.searchQuery && this.searchQuery.trim()) {
        const searchResponse = await apiService.searchBookmarks(this.searchQuery, 20)
        // Convert SearchResult[] to BookmarkListResponse format
        response = {
          bookmarks: searchResponse.results,
          total: searchResponse.count,
          page: 1,
          limit: 20,
          has_more: false,
          total_pages: 1,
          tag_count: 0,
          favorite_count: searchResponse.results.filter(b => b.is_favorite).length
        }
      } else {
        response = await apiService.getBookmarks({
          page: this._stats.page,
          limit: 20,
          tag: this.tagFilter || undefined,
          favorites: this.favoritesOnly || undefined
        })
      }

      if (reset) {
        this._bookmarks = response.bookmarks
      } else {
        this._bookmarks = [...this._bookmarks, ...response.bookmarks]
      }

      this._stats = {
        total: response.total,
        favorites: response.favorite_count,
        tags: response.tag_count,
        hasMore: response.has_more,
        page: response.page
      }
    } catch (error) {
      this._error = error instanceof Error ? error.message : 'Failed to load bookmarks'
    } finally {
      this._loading = false
    }
  }

  async loadMore() {
    if (this._loading || !this._stats.hasMore) return
    
    this._stats.page++
    await this.loadBookmarks()
  }

  render() {
    if (this._loading && this._bookmarks.length === 0) {
      return html`
        <div class="loading">
          <div class="loading-spinner"></div>
          <div class="neon-cyan">Loading bookmarks...</div>
        </div>
      `
    }

    if (this._error) {
      return html`
        <div class="error">
          <div>⚠️ ${this._error}</div>
          <button class="retry-button" @click=${() => this.loadBookmarks(true)}>
            Retry
          </button>
        </div>
      `
    }

    if (this._bookmarks.length === 0) {
      return html`
        <div class="empty-state">
          <div class="empty-icon">${this.searchQuery ? '🔍' : '🔖'}</div>
          <h3>${this.searchQuery ? 'No search results' : 'No bookmarks found'}</h3>
          ${this.searchQuery ? html`
            <p>No results for <strong>"${this.searchQuery}"</strong></p>
            <p>Try different keywords or check your spelling.</p>
          ` : this.tagFilter ? html`
            <p>No bookmarks found with the tag <strong>"${this.tagFilter}"</strong></p>
          ` : this.favoritesOnly ? html`
            <p>No favorite bookmarks yet. Star some bookmarks to see them here!</p>
          ` : html`
            <p>Add your first bookmark to get started!</p>
          `}
        </div>
      `
    }

    return html`
      <div class="stats">
        <div class="stats-item">
          <div class="stats-number">${this._stats.total}</div>
          <div class="stats-label">Total</div>
        </div>
        <div class="stats-item">
          <div class="stats-number">${this._stats.favorites}</div>
          <div class="stats-label">Favorites</div>
        </div>
        <div class="stats-item">
          <div class="stats-number">${this._stats.tags}</div>
          <div class="stats-label">Tags</div>
        </div>
      </div>

      <div class="bookmark-grid">
        ${this._bookmarks.map(bookmark => html`
          <bookmark-item 
            .bookmark=${bookmark}
            @toggle-favorite=${this._handleToggleFavorite}
            @delete=${this._handleDelete}
            @edit=${this._handleEdit}>
          </bookmark-item>
        `)}
      </div>

      ${this._stats.hasMore && !this.searchQuery ? html`
        <div class="load-more">
          <button 
            class="load-more-button"
            ?disabled=${this._loading}
            @click=${this.loadMore}>
            ${this._loading ? 'Loading...' : 'Load More'}
          </button>
        </div>
      ` : ''}
    `
  }

  private async _handleToggleFavorite(e: CustomEvent) {
    const bookmarkId = e.detail.id
    const bookmark = this._bookmarks.find(b => b.id === bookmarkId)
    if (!bookmark) return

    try {
      const updatedBookmark = await apiService.updateBookmark(bookmarkId, {
        is_favorite: !bookmark.is_favorite
      })

      // Update local state
      this._bookmarks = this._bookmarks.map(b => 
        b.id === bookmarkId ? updatedBookmark : b
      )

      // Update stats
      if (updatedBookmark.is_favorite) {
        this._stats.favorites++
      } else {
        this._stats.favorites--
      }
    } catch (error) {
      console.error('Failed to toggle favorite:', error)
      // TODO: Show toast notification
    }
  }

  private async _handleDelete(e: CustomEvent) {
    const bookmarkId = e.detail.id
    
    try {
      await apiService.deleteBookmark(bookmarkId)
      
      // Remove from local state
      const bookmark = this._bookmarks.find(b => b.id === bookmarkId)
      this._bookmarks = this._bookmarks.filter(b => b.id !== bookmarkId)
      
      // Update stats
      this._stats.total--
      if (bookmark?.is_favorite) {
        this._stats.favorites--
      }

      // Notify parent to refresh tag cloud
      this.dispatchEvent(new CustomEvent('bookmark-deleted', {
        bubbles: true,
        detail: { bookmark }
      }))
    } catch (error) {
      console.error('Failed to delete bookmark:', error)
      // TODO: Show toast notification
    }
  }

  private _handleEdit(e: CustomEvent) {
    // Dispatch event to parent component to open edit dialog
    this.dispatchEvent(new CustomEvent('edit-bookmark', {
      detail: e.detail,
      bubbles: true
    }))
  }
}