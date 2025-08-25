import { LitElement, html, css } from 'lit'
import { customElement, property, state } from 'lit/decorators.js'
import type { Bookmark } from '../services/api.ts'

export interface BookmarkHealth {
  id: number
  url: string
  status: 'healthy' | 'broken' | 'slow' | 'redirect' | 'unknown'
  status_code: number
  response_time_ms: number
  redirect_url?: string
  error?: string
  last_checked: string
}

@customElement('bookmark-item')
export class BookmarkItem extends LitElement {
  @property({ type: Object }) bookmark!: Bookmark
  @state() private _health: BookmarkHealth | null = null
  @state() private _loadingHealth = false

  static styles = css`
    :host {
      display: block;
    }

    .bookmark-card {
      background: var(--bg-card);
      border: 1px solid var(--border-color);
      border-radius: 0.75rem;
      padding: 1.5rem;
      backdrop-filter: blur(10px);
      transition: all 0.3s ease;
      position: relative;
      overflow: hidden;
      box-shadow: var(--shadow-sm);
    }

    .bookmark-card:hover {
      border-color: var(--accent-primary);
      box-shadow: var(--shadow-lg);
      background: var(--bg-card-hover);
      transform: translateY(-2px);
    }

    .bookmark-card:before {
      content: '';
      position: absolute;
      top: 0;
      left: 0;
      right: 0;
      height: 2px;
      background: linear-gradient(90deg, var(--accent-primary), var(--accent-secondary), var(--accent-warning));
      opacity: 0;
      transition: opacity 0.3s ease;
    }

    .bookmark-card:hover:before {
      opacity: 1;
    }

    .bookmark-header {
      display: flex;
      align-items: flex-start;
      gap: 0.75rem;
      margin-bottom: 1rem;
    }

    .favicon {
      width: 20px;
      height: 20px;
      border-radius: 3px;
      flex-shrink: 0;
      margin-top: 2px;
    }

    .favicon-placeholder {
      width: 20px;
      height: 20px;
      background: linear-gradient(45deg, var(--accent-primary), var(--accent-secondary));
      border-radius: 3px;
      display: flex;
      align-items: center;
      justify-content: center;
      font-size: 10px;
      color: var(--bg-primary);
      font-weight: bold;
      flex-shrink: 0;
      margin-top: 2px;
    }

    .bookmark-info {
      flex: 1;
      min-width: 0;
    }

    .bookmark-title {
      font-size: 1rem;
      font-weight: 600;
      color: var(--text-primary);
      margin: 0 0 0.25rem 0;
      line-height: 1.4;
      overflow: hidden;
      text-overflow: ellipsis;
      white-space: nowrap;
    }

    .bookmark-url {
      font-size: 0.8rem;
      color: var(--accent-primary);
      text-decoration: none;
      opacity: 0.8;
      overflow: hidden;
      text-overflow: ellipsis;
      white-space: nowrap;
      display: block;
      transition: opacity 0.3s ease;
    }

    .bookmark-url:hover {
      opacity: 1;
      text-shadow: var(--shadow-sm);
    }

    .bookmark-description {
      font-size: 0.875rem;
      color: var(--text-secondary);
      line-height: 1.4;
      margin: 0.75rem 0;
      display: -webkit-box;
      -webkit-line-clamp: 2;
      -webkit-box-orient: vertical;
      overflow: hidden;
    }

    .bookmark-tags {
      display: flex;
      flex-wrap: wrap;
      gap: 0.5rem;
      margin: 1rem 0;
    }

    .tag {
      background: rgba(var(--accent-secondary), 0.1);
      border: 1px solid rgba(var(--accent-secondary), 0.3);
      color: var(--accent-secondary);
      font-size: 0.7rem;
      padding: 0.25rem 0.5rem;
      border-radius: 0.25rem;
      font-family: 'Courier New', monospace;
      text-transform: uppercase;
      letter-spacing: 0.5px;
      font-weight: 500;
    }

    .bookmark-footer {
      display: flex;
      align-items: center;
      justify-content: space-between;
      margin-top: 1rem;
      padding-top: 1rem;
      border-top: 1px solid var(--border-color);
    }

    .bookmark-date {
      font-size: 0.7rem;
      color: var(--text-muted);
      font-family: 'Courier New', monospace;
    }

    .bookmark-actions {
      display: flex;
      gap: 0.5rem;
    }

    .action-button {
      background: none;
      border: none;
      color: var(--text-muted);
      cursor: pointer;
      padding: 0.25rem;
      border-radius: 0.25rem;
      transition: all 0.3s ease;
      font-size: 1rem;
    }

    .action-button:hover {
      color: var(--accent-primary);
      background: rgba(var(--accent-primary), 0.1);
    }

    .favorite-button.active {
      color: var(--accent-warning);
      text-shadow: var(--shadow-sm);
    }

    .delete-button:hover {
      color: var(--accent-danger);
      background: rgba(var(--accent-danger), 0.1);
    }

    .health-indicator {
      position: absolute;
      top: 0.5rem;
      right: 0.5rem;
      width: 12px;
      height: 12px;
      border-radius: 50%;
      transition: all 0.3s ease;
    }

    .health-healthy {
      background: var(--accent-success);
      box-shadow: 0 0 8px rgba(var(--accent-success), 0.4);
    }

    .health-broken {
      background: var(--accent-danger);
      box-shadow: 0 0 8px rgba(var(--accent-danger), 0.4);
    }

    .health-slow {
      background: var(--accent-warning);
      box-shadow: 0 0 8px rgba(var(--accent-warning), 0.4);
    }

    .health-redirect {
      background: var(--accent-secondary);
      box-shadow: 0 0 8px rgba(var(--accent-secondary), 0.4);
    }

    .health-unknown {
      background: var(--text-muted);
      box-shadow: 0 0 8px rgba(var(--text-muted), 0.4);
    }

    .health-tooltip {
      position: absolute;
      top: -2rem;
      right: 0;
      background: var(--bg-card);
      border: 1px solid var(--border-color);
      border-radius: 0.25rem;
      padding: 0.5rem;
      font-size: 0.75rem;
      white-space: nowrap;
      opacity: 0;
      pointer-events: none;
      transition: opacity 0.3s ease;
      z-index: 1000;
      box-shadow: var(--shadow-lg);
    }

    .health-indicator:hover .health-tooltip {
      opacity: 1;
    }

    .health-check-button {
      background: none;
      border: none;
      color: var(--text-muted);
      cursor: pointer;
      padding: 0.25rem;
      border-radius: 0.25rem;
      transition: all 0.3s ease;
      font-size: 0.875rem;
      opacity: 0.7;
    }

    .health-check-button:hover {
      color: var(--accent-primary);
      background: rgba(var(--accent-primary), 0.1);
      opacity: 1;
    }

    .health-loading {
      width: 12px;
      height: 12px;
      border: 2px solid rgba(var(--accent-primary), 0.2);
      border-top: 2px solid var(--accent-primary);
      border-radius: 50%;
      animation: spin 1s linear infinite;
    }

    @keyframes spin {
      0% { transform: rotate(0deg); }
      100% { transform: rotate(360deg); }
    }
  `

  connectedCallback() {
    super.connectedCallback()
    this._loadHealth()
  }

  private async _loadHealth() {
    try {
      const response = await fetch(`/api/health/bookmark/${this.bookmark.id}`)
      if (response.ok) {
        this._health = await response.json()
      }
    } catch (error) {
      // Health data is optional, don't error if it fails
      console.debug('Failed to load health data:', error)
    }
  }

  private async _checkHealthNow() {
    if (this._loadingHealth) return
    
    this._loadingHealth = true
    try {
      const response = await fetch(`/api/health/check/${this.bookmark.id}`, {
        method: 'POST'
      })
      if (response.ok) {
        this._health = await response.json()
      }
    } catch (error) {
      console.error('Failed to check health:', error)
    } finally {
      this._loadingHealth = false
    }
  }

  private _renderHealthIndicator() {
    if (this._loadingHealth) {
      return html`<div class="health-loading"></div>`
    }

    if (!this._health) return ''

    const statusClass = `health-${this._health.status}`
    const statusText = this._getHealthStatusText()

    return html`
      <div class="health-indicator ${statusClass}">
        <div class="health-tooltip">
          ${statusText}
        </div>
      </div>
    `
  }

  private _getHealthStatusText(): string {
    if (!this._health) return ''

    switch (this._health.status) {
      case 'healthy':
        return `✅ Healthy (${this._health.response_time_ms}ms)`
      case 'broken':
        return `❌ Broken: ${this._health.error || 'HTTP ' + this._health.status_code}`
      case 'slow':
        return `⚠️ Slow (${this._health.response_time_ms}ms)`
      case 'redirect':
        return `🔄 Redirects to: ${this._health.redirect_url || 'unknown'}`
      case 'unknown':
        return `❓ Unknown status`
      default:
        return this._health.status
    }
  }

  render() {
    const createdDate = new Date(this.bookmark.created_at)
    const formattedDate = new Intl.DateTimeFormat('en-US', {
      month: 'short',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit'
    }).format(createdDate)

    const tagNames = this.bookmark.tags?.map(tag => tag.name) || []

    return html`
      <div class="bookmark-card">
        ${this._renderHealthIndicator()}
        <div class="bookmark-header">
          ${this.bookmark.favicon_url ? html`
            <img class="favicon" src="${this.bookmark.favicon_url}" alt="Favicon" />
          ` : html`
            <div class="favicon-placeholder">${this.bookmark.title.charAt(0).toUpperCase()}</div>
          `}
          
          <div class="bookmark-info">
            <h3 class="bookmark-title">${this.bookmark.title}</h3>
            <a class="bookmark-url" href="${this.bookmark.url}" target="_blank" rel="noopener">
              ${this.bookmark.url}
            </a>
          </div>
        </div>

        ${this.bookmark.description ? html`
          <div class="bookmark-description">${this.bookmark.description}</div>
        ` : ''}

        ${tagNames.length > 0 ? html`
          <div class="bookmark-tags">
            ${tagNames.map(tagName => html`
              <span class="tag">${tagName}</span>
            `)}
          </div>
        ` : ''}

        <div class="bookmark-footer">
          <div class="bookmark-date">${formattedDate}</div>
          <div class="bookmark-actions">
            <button 
              class="action-button favorite-button ${this.bookmark.is_favorite ? 'active' : ''}"
              @click=${this._handleToggleFavorite}
              title=${this.bookmark.is_favorite ? 'Remove from favorites' : 'Add to favorites'}>
              ${this.bookmark.is_favorite ? '⭐' : '☆'}
            </button>
            <button 
              class="action-button edit-button"
              @click=${this._handleEdit}
              title="Edit bookmark">
              ✏️
            </button>
            <button 
              class="action-button health-check-button"
              @click=${this._checkHealthNow}
              title="Check health now"
              ?disabled=${this._loadingHealth}>
              🔗
            </button>
            <button 
              class="action-button delete-button"
              @click=${this._handleDelete}
              title="Delete bookmark">
              🗑️
            </button>
          </div>
        </div>
      </div>
    `
  }

  private _handleToggleFavorite() {
    this.dispatchEvent(new CustomEvent('toggle-favorite', {
      detail: { id: this.bookmark.id }
    }))
  }

  private _handleEdit() {
    this.dispatchEvent(new CustomEvent('edit', {
      detail: { bookmark: this.bookmark }
    }))
  }

  private _handleDelete() {
    if (confirm(`Delete "${this.bookmark.title}"?`)) {
      this.dispatchEvent(new CustomEvent('delete', {
        detail: { id: this.bookmark.id }
      }))
    }
  }
}