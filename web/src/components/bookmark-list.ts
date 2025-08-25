import { LitElement, html, css } from 'lit'
import { customElement, state } from 'lit/decorators.js'
import './bookmark-item.ts'

interface Bookmark {
  id: number
  title: string
  url: string
  description?: string
  tags: string[]
  favicon?: string
  createdAt: string
  isFavorite: boolean
}

@customElement('bookmark-list')
export class BookmarkList extends LitElement {
  @state() private _bookmarks: Bookmark[] = [
    {
      id: 1,
      title: 'GitHub',
      url: 'https://github.com',
      description: 'Where the world builds software',
      tags: ['Development', 'Code'],
      favicon: 'https://github.com/favicon.ico',
      createdAt: new Date().toISOString(),
      isFavorite: true
    },
    {
      id: 2,
      title: 'Cyberpunk Design Inspiration',
      url: 'https://example.com/cyberpunk',
      description: 'Neon colors and futuristic UI patterns',
      tags: ['Design', 'Cyberpunk', 'UI/UX'],
      createdAt: new Date().toISOString(),
      isFavorite: false
    }
  ]

  @state() private _loading = false

  static styles = css`
    :host {
      display: block;
    }

    .loading {
      text-align: center;
      padding: 2rem;
      color: #00ffff;
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

    @media (max-width: 768px) {
      .bookmark-grid {
        grid-template-columns: 1fr;
      }
    }
  `

  render() {
    if (this._loading) {
      return html`
        <div class="loading">
          <div class="neon-cyan">Loading bookmarks...</div>
        </div>
      `
    }

    if (this._bookmarks.length === 0) {
      return html`
        <div class="empty-state">
          <div class="empty-icon">🔖</div>
          <h3>No bookmarks yet</h3>
          <p>Add your first bookmark to get started!</p>
        </div>
      `
    }

    return html`
      <div class="stats">
        <div class="stats-item">
          <div class="stats-number">${this._bookmarks.length}</div>
          <div class="stats-label">Total</div>
        </div>
        <div class="stats-item">
          <div class="stats-number">${this._bookmarks.filter(b => b.isFavorite).length}</div>
          <div class="stats-label">Favorites</div>
        </div>
        <div class="stats-item">
          <div class="stats-number">${new Set(this._bookmarks.flatMap(b => b.tags)).size}</div>
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
    `
  }

  private _handleToggleFavorite(e: CustomEvent) {
    const bookmarkId = e.detail.id
    this._bookmarks = this._bookmarks.map(bookmark => 
      bookmark.id === bookmarkId 
        ? { ...bookmark, isFavorite: !bookmark.isFavorite }
        : bookmark
    )
  }

  private _handleDelete(e: CustomEvent) {
    const bookmarkId = e.detail.id
    this._bookmarks = this._bookmarks.filter(bookmark => bookmark.id !== bookmarkId)
  }

  private _handleEdit(e: CustomEvent) {
    console.log('Edit bookmark:', e.detail)
    // TODO: Open edit dialog
  }
}