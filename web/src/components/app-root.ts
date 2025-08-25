import { LitElement, html, css } from 'lit'
import { customElement, state } from 'lit/decorators.js'
import './app-header.ts'
import './bookmark-list.ts'
import './bookmark-dialog.ts'

@customElement('app-root')
export class AppRoot extends LitElement {
  @state() private _showDialog = false

  static styles = css`
    :host {
      display: block;
      min-height: 100vh;
      background: 
        radial-gradient(circle at 20% 50%, rgba(0, 255, 255, 0.1) 0%, transparent 50%),
        radial-gradient(circle at 80% 20%, rgba(255, 0, 128, 0.1) 0%, transparent 50%),
        linear-gradient(135deg, #0a0a0a 0%, #1a1a1a 100%);
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
      background: rgba(10, 10, 10, 0.8);
      border: 1px solid rgba(0, 255, 255, 0.2);
      border-radius: 0.5rem;
      padding: 1.5rem;
      backdrop-filter: blur(10px);
    }

    .content {
      background: rgba(10, 10, 10, 0.8);
      border: 1px solid rgba(255, 0, 128, 0.2);
      border-radius: 0.5rem;
      padding: 1.5rem;
      backdrop-filter: blur(10px);
    }

    .welcome-message {
      text-align: center;
      padding: 3rem 1rem;
      color: #a0a0a0;
    }

    .welcome-title {
      font-size: 2rem;
      font-weight: bold;
      margin-bottom: 1rem;
      background: linear-gradient(45deg, #00ffff, #ff0080);
      -webkit-background-clip: text;
      -webkit-text-fill-color: transparent;
      background-clip: text;
    }
  `

  render() {
    return html`
      <div class="container">
        <app-header @add-bookmark=${this._handleAddBookmark}></app-header>
        
        <div class="main-content">
          <aside class="sidebar">
            <h3 class="neon-cyan">Filters</h3>
            <p class="text-gray-400">Tags and filters will go here</p>
          </aside>
          
          <main class="content">
            <div class="welcome-message">
              <h1 class="welcome-title">とりメモ (Torimemo)</h1>
              <p>Your cyberpunk bookmark manager is ready!</p>
              <p class="text-sm">Drop a link and watch the AI magic happen ✨</p>
            </div>
            <bookmark-list></bookmark-list>
          </main>
        </div>
        
        ${this._showDialog ? html`
          <bookmark-dialog 
            @close=${this._handleCloseDialog}
            @save=${this._handleSaveBookmark}>
          </bookmark-dialog>
        ` : ''}
      </div>
    `
  }

  private _handleAddBookmark() {
    this._showDialog = true
  }

  private _handleCloseDialog() {
    this._showDialog = false
  }

  private _handleSaveBookmark(e: CustomEvent) {
    console.log('Save bookmark:', e.detail)
    this._showDialog = false
    // TODO: Save to API
  }
}