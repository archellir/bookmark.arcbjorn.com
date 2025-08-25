import { LitElement, html, css } from 'lit'
import { customElement, state } from 'lit/decorators.js'
import type { Bookmark } from '../services/api.ts'
import './app-header.ts'
import './bookmark-list.ts'
import './bookmark-dialog.ts'
import './tag-cloud.ts'
import './theme-toggle.ts'

@customElement('app-root')
export class AppRoot extends LitElement {
  @state() private _showDialog = false
  @state() private _editBookmark: Bookmark | null = null
  @state() private _searchQuery = ''
  @state() private _tagFilter = ''
  @state() private _favoritesOnly = false

  connectedCallback() {
    super.connectedCallback();
    // Initialize theme manager to apply theme classes
    import('../services/theme.js').then(({ themeManager }) => {
      const updateTheme = () => {
        const theme = themeManager.getCurrentTheme();
        const actualTheme = theme === 'auto' 
          ? (window.matchMedia('(prefers-color-scheme: light)').matches ? 'light' : 'dark')
          : theme;
        this.className = actualTheme;
      };
      
      updateTheme();
      themeManager.subscribe(updateTheme);
    });
  }

  static styles = css`
    :host {
      display: block;
      min-height: 100vh;
      background: var(--bg-primary);
      color: var(--text-primary);
      transition: background-color 0.3s ease, color 0.3s ease;
    }

    /* Dark theme backgrounds (cyberpunk) */
    :host(.dark) {
      background: 
        radial-gradient(circle at 20% 50%, rgba(0, 255, 255, 0.1) 0%, transparent 50%),
        radial-gradient(circle at 80% 20%, rgba(255, 0, 128, 0.1) 0%, transparent 50%),
        linear-gradient(135deg, #0a0a0a 0%, #1a1a1a 100%);
    }

    /* Light theme backgrounds */
    :host(.light) {
      background: 
        radial-gradient(circle at 20% 50%, rgba(13, 110, 253, 0.05) 0%, transparent 50%),
        radial-gradient(circle at 80% 20%, rgba(111, 66, 193, 0.05) 0%, transparent 50%),
        linear-gradient(135deg, #f8f9fa 0%, #ffffff 100%);
    }

    .container {
      max-width: 1200px;
      margin: 0 auto;
      padding: 2rem;
    }

    .main-content {
      display: grid;
      grid-template-columns: 1fr;
      gap: 2rem;
      margin-top: 2rem;
    }

    @media (min-width: 768px) {
      .main-content {
        grid-template-columns: 300px 1fr;
      }
    }

    .sidebar {
      background: var(--bg-card);
      border: 1px solid var(--border-color);
      border-radius: 0.5rem;
      padding: 1.5rem;
      backdrop-filter: blur(10px);
      height: fit-content;
      box-shadow: var(--shadow-md);
    }

    .sidebar-title {
      color: var(--accent-primary);
      font-size: 1.25rem;
      font-weight: bold;
      margin-bottom: 1rem;
      text-shadow: var(--shadow-sm);
    }

    .theme-container {
      margin-bottom: 1.5rem;
      padding-bottom: 1rem;
      border-bottom: 1px solid var(--border-color);
    }

    .filter-group {
      margin-bottom: 1.5rem;
    }

    .filter-label {
      display: block;
      color: var(--text-secondary);
      font-size: 0.875rem;
      font-weight: 500;
      text-transform: uppercase;
      letter-spacing: 0.5px;
      margin-bottom: 0.5rem;
    }

    .filter-toggle {
      display: flex;
      align-items: center;
      gap: 0.5rem;
      cursor: pointer;
      padding: 0.5rem;
      border-radius: 0.25rem;
      transition: background-color 0.3s ease;
      color: var(--text-primary);
    }

    .filter-toggle:hover {
      background: var(--bg-card-hover);
    }

    .filter-toggle input[type="checkbox"] {
      width: 16px;
      height: 16px;
      accent-color: var(--accent-primary);
    }

    .content {
      background: var(--bg-card);
      border: 1px solid var(--border-color);
      border-radius: 0.5rem;
      padding: 1.5rem;
      backdrop-filter: blur(10px);
      box-shadow: var(--shadow-md);
    }

    .welcome-message {
      text-align: center;
      padding: 3rem 1rem;
      color: var(--text-muted);
    }

    .welcome-title {
      font-size: 2rem;
      font-weight: bold;
      margin-bottom: 1rem;
      background: linear-gradient(45deg, var(--accent-primary), var(--accent-secondary));
      -webkit-background-clip: text;
      -webkit-text-fill-color: transparent;
      background-clip: text;
    }

    .quick-actions {
      background: rgba(var(--accent-warning), 0.05);
      border: 1px solid rgba(var(--accent-warning), 0.2);
      border-radius: 0.5rem;
      padding: 1rem;
      margin-bottom: 1.5rem;
    }

    .quick-actions-title {
      color: var(--accent-warning);
      font-size: 1rem;
      font-weight: bold;
      margin-bottom: 0.5rem;
    }

    .quick-actions-text {
      color: var(--text-muted);
      font-size: 0.875rem;
      line-height: 1.4;
    }

    @media (max-width: 768px) {
      .main-content {
        grid-template-columns: 1fr;
      }
      
      .sidebar {
        order: 2;
      }
      
      .content {
        order: 1;
      }
    }
  `

  render() {
    return html`
      <div class="container">
        <app-header 
          @search=${this._handleSearch}
          @add-bookmark=${this._handleAddBookmark}>
        </app-header>
        
        <div class="main-content">
          <aside class="sidebar">
            <div class="theme-container">
              <label class="filter-label">Theme</label>
              <theme-toggle></theme-toggle>
            </div>
            
            <h3 class="sidebar-title">Filters</h3>
            
            <div class="filter-group">
              <label class="filter-label">Quick Filters</label>
              <label class="filter-toggle">
                <input 
                  type="checkbox" 
                  .checked=${this._favoritesOnly}
                  @change=${this._handleFavoritesToggle}
                />
                <span>⭐ Favorites Only</span>
              </label>
            </div>

            <div class="filter-group">
              <label class="filter-label">Tags</label>
              <tag-cloud 
                .selectedTag=${this._tagFilter}
                @tag-selected=${this._handleTagSelected}
                @tag-cleared=${this._handleTagCleared}>
              </tag-cloud>
            </div>

            <div class="quick-actions">
              <div class="quick-actions-title">💡 Pro Tip</div>
              <div class="quick-actions-text">
                Use the search box to find bookmarks instantly, or click the "+" button to add new ones with AI-powered tagging!
              </div>
            </div>
          </aside>
          
          <main class="content">
            <bookmark-list 
              .searchQuery=${this._searchQuery}
              .tagFilter=${this._tagFilter}
              .favoritesOnly=${this._favoritesOnly}
              @edit-bookmark=${this._handleEditBookmark}
              @bookmark-deleted=${this._handleBookmarkDeleted}>
            </bookmark-list>
          </main>
        </div>
        
        ${this._showDialog ? html`
          <bookmark-dialog 
            .editBookmark=${this._editBookmark}
            @close=${this._handleCloseDialog}
            @save=${this._handleSaveBookmark}>
          </bookmark-dialog>
        ` : ''}
      </div>
    `
  }

  private _handleSearch(e: CustomEvent) {
    this._searchQuery = e.detail.query || ''
  }

  private _handleFavoritesToggle(e: Event) {
    const checkbox = e.target as HTMLInputElement
    this._favoritesOnly = checkbox.checked
  }

  private _handleAddBookmark() {
    this._editBookmark = null
    this._showDialog = true
  }

  private _handleEditBookmark(e: CustomEvent) {
    this._editBookmark = e.detail.bookmark
    this._showDialog = true
  }

  private _handleCloseDialog() {
    this._showDialog = false
    this._editBookmark = null
  }

  private _handleSaveBookmark(e: CustomEvent) {
    const { bookmark, isEdit } = e.detail
    
    // Close dialog
    this._showDialog = false
    this._editBookmark = null
    
    // Force bookmark list to refresh by dispatching an event
    this.shadowRoot?.querySelector('bookmark-list')?.dispatchEvent(
      new CustomEvent('refresh-needed')
    )
    
    // Refresh tag cloud to update counts
    const tagCloud = this.shadowRoot?.querySelector('tag-cloud') as any
    tagCloud?.loadTags()
    
    // Could show a success toast here
    console.log(isEdit ? 'Bookmark updated:' : 'Bookmark created:', bookmark)
  }

  private _handleTagSelected(e: CustomEvent) {
    this._tagFilter = e.detail.tag
  }

  private _handleTagCleared() {
    this._tagFilter = ''
  }

  private _handleBookmarkDeleted() {
    // Refresh tag cloud when bookmark is deleted
    const tagCloud = this.shadowRoot?.querySelector('tag-cloud') as any
    tagCloud?.loadTags()
  }
}