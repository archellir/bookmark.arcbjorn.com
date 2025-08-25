import { LitElement, html, css } from 'lit'
import { customElement, state } from 'lit/decorators.js'

@customElement('app-header')
export class AppHeader extends LitElement {
  @state() private _searchQuery = ''

  static styles = css`
    :host {
      display: block;
    }

    .header {
      display: flex;
      align-items: center;
      justify-content: space-between;
      padding: 1rem 0;
      border-bottom: 1px solid rgba(0, 255, 255, 0.2);
      margin-bottom: 1rem;
    }

    .logo {
      font-size: 1.5rem;
      font-weight: bold;
      background: linear-gradient(45deg, #00ffff, #ff0080);
      -webkit-background-clip: text;
      -webkit-text-fill-color: transparent;
      background-clip: text;
      text-shadow: 0 0 20px rgba(0, 255, 255, 0.3);
    }

    .search-container {
      flex: 1;
      max-width: 500px;
      margin: 0 2rem;
      position: relative;
    }

    .search-input {
      width: 100%;
      background: rgba(0, 0, 0, 0.8);
      border: 1px solid rgba(0, 255, 255, 0.3);
      color: white;
      padding: 0.75rem 1rem;
      border-radius: 0.5rem;
      font-family: 'Courier New', monospace;
      transition: all 0.3s ease;
    }

    .search-input:focus {
      outline: none;
      border-color: #00ffff;
      box-shadow: 0 0 15px rgba(0, 255, 255, 0.3);
    }

    .search-input::placeholder {
      color: #666;
    }

    .add-button {
      background: linear-gradient(45deg, rgba(0, 255, 255, 0.1) 0%, transparent 50%);
      border: 1px solid #00ffff;
      color: #00ffff;
      padding: 0.75rem 1.5rem;
      border-radius: 0.5rem;
      font-family: 'Courier New', monospace;
      font-weight: bold;
      text-transform: uppercase;
      letter-spacing: 1px;
      cursor: pointer;
      transition: all 0.3s ease;
      position: relative;
      overflow: hidden;
    }

    .add-button:hover {
      background: #00ffff;
      color: black;
      box-shadow: 0 0 20px rgba(0, 255, 255, 0.5);
    }

    .add-button:before {
      content: '';
      position: absolute;
      top: 0;
      left: -100%;
      width: 100%;
      height: 100%;
      background: linear-gradient(90deg, transparent, rgba(255, 255, 255, 0.2), transparent);
      transition: left 0.5s;
    }

    .add-button:hover:before {
      left: 100%;
    }

    @media (max-width: 768px) {
      .header {
        flex-direction: column;
        gap: 1rem;
        text-align: center;
      }

      .search-container {
        margin: 0;
        max-width: 100%;
      }
    }
  `

  render() {
    return html`
      <header class="header">
        <div class="logo">とりメモ</div>
        
        <div class="search-container">
          <input 
            type="text" 
            class="search-input"
            placeholder="Search bookmarks... 🔍"
            .value=${this._searchQuery}
            @input=${this._handleSearch}
            @keydown=${this._handleKeydown}
          />
        </div>
        
        <button class="add-button" @click=${this._handleAdd}>
          + Add Link
        </button>
      </header>
    `
  }

  private _handleSearch(e: Event) {
    const input = e.target as HTMLInputElement
    this._searchQuery = input.value
    // TODO: Dispatch search event
    this.dispatchEvent(new CustomEvent('search', {
      detail: { query: this._searchQuery }
    }))
  }

  private _handleKeydown(e: KeyboardEvent) {
    if (e.key === 'Escape') {
      this._searchQuery = ''
      ;(e.target as HTMLInputElement).blur()
    }
  }

  private _handleAdd() {
    this.dispatchEvent(new CustomEvent('add-bookmark'))
  }
}