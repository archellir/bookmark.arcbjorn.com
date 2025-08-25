import { LitElement, html, css } from 'lit'
import { customElement, state } from 'lit/decorators.js'

@customElement('bookmark-dialog')
export class BookmarkDialog extends LitElement {
  @state() private _url = ''
  @state() private _title = ''
  @state() private _description = ''
  @state() private _tags = ''
  @state() private _isAnalyzing = false

  static styles = css`
    :host {
      position: fixed;
      top: 0;
      left: 0;
      right: 0;
      bottom: 0;
      background: rgba(0, 0, 0, 0.8);
      backdrop-filter: blur(10px);
      display: flex;
      align-items: center;
      justify-content: center;
      z-index: 1000;
    }

    .dialog {
      background: linear-gradient(135deg, rgba(0, 255, 255, 0.05) 0%, transparent 50%);
      background-color: rgba(10, 10, 10, 0.95);
      border: 1px solid rgba(0, 255, 255, 0.3);
      border-radius: 1rem;
      padding: 2rem;
      width: 90%;
      max-width: 500px;
      max-height: 90vh;
      overflow-y: auto;
      box-shadow: 0 20px 40px rgba(0, 0, 0, 0.5);
      position: relative;
    }

    .dialog:before {
      content: '';
      position: absolute;
      top: 0;
      left: 0;
      right: 0;
      height: 2px;
      background: linear-gradient(90deg, #00ffff, #ff0080, #ffff00);
    }

    .dialog-header {
      display: flex;
      align-items: center;
      justify-content: space-between;
      margin-bottom: 1.5rem;
    }

    .dialog-title {
      font-size: 1.25rem;
      font-weight: bold;
      color: #00ffff;
      text-shadow: 0 0 10px rgba(0, 255, 255, 0.3);
    }

    .close-button {
      background: none;
      border: none;
      color: #666;
      font-size: 1.5rem;
      cursor: pointer;
      padding: 0.25rem;
      border-radius: 0.25rem;
      transition: all 0.3s ease;
    }

    .close-button:hover {
      color: #ff1744;
      background: rgba(255, 23, 68, 0.1);
    }

    .form-group {
      margin-bottom: 1.5rem;
    }

    .form-label {
      display: block;
      margin-bottom: 0.5rem;
      color: #a0a0a0;
      font-size: 0.875rem;
      font-weight: 500;
      text-transform: uppercase;
      letter-spacing: 0.5px;
    }

    .form-input, .form-textarea {
      width: 100%;
      background: rgba(0, 0, 0, 0.8);
      border: 1px solid rgba(0, 255, 255, 0.3);
      color: white;
      padding: 0.75rem;
      border-radius: 0.5rem;
      font-family: 'Courier New', monospace;
      transition: all 0.3s ease;
      box-sizing: border-box;
    }

    .form-input:focus, .form-textarea:focus {
      outline: none;
      border-color: #00ffff;
      box-shadow: 0 0 15px rgba(0, 255, 255, 0.3);
    }

    .form-textarea {
      resize: vertical;
      min-height: 80px;
    }

    .form-input::placeholder, .form-textarea::placeholder {
      color: #666;
    }

    .analyze-status {
      display: flex;
      align-items: center;
      gap: 0.5rem;
      margin-top: 0.5rem;
      color: #00ffff;
      font-size: 0.875rem;
    }

    .analyze-spinner {
      width: 16px;
      height: 16px;
      border: 2px solid rgba(0, 255, 255, 0.2);
      border-top: 2px solid #00ffff;
      border-radius: 50%;
      animation: spin 1s linear infinite;
    }

    @keyframes spin {
      0% { transform: rotate(0deg); }
      100% { transform: rotate(360deg); }
    }

    .tag-suggestions {
      display: flex;
      flex-wrap: wrap;
      gap: 0.5rem;
      margin-top: 0.5rem;
    }

    .tag-suggestion {
      background: linear-gradient(45deg, rgba(255, 0, 128, 0.2), rgba(0, 255, 255, 0.2));
      border: 1px solid rgba(255, 0, 128, 0.3);
      color: #ff0080;
      font-size: 0.7rem;
      padding: 0.25rem 0.5rem;
      border-radius: 0.25rem;
      cursor: pointer;
      transition: all 0.3s ease;
      position: relative;
      overflow: hidden;
    }

    .tag-suggestion:hover {
      background: rgba(255, 0, 128, 0.3);
      transform: translateY(-1px);
    }

    .tag-suggestion:before {
      content: '';
      position: absolute;
      top: 0;
      left: -100%;
      width: 100%;
      height: 100%;
      background: linear-gradient(90deg, transparent, rgba(255, 255, 255, 0.2), transparent);
      transition: left 0.5s;
    }

    .tag-suggestion:hover:before {
      left: 100%;
    }

    .dialog-actions {
      display: flex;
      gap: 1rem;
      margin-top: 2rem;
    }

    .button {
      flex: 1;
      background: transparent;
      border: 1px solid;
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

    .button-primary {
      border-color: #00ffff;
      color: #00ffff;
      background: linear-gradient(45deg, rgba(0, 255, 255, 0.1) 0%, transparent 50%);
    }

    .button-primary:hover {
      background: #00ffff;
      color: black;
      box-shadow: 0 0 20px rgba(0, 255, 255, 0.5);
    }

    .button-secondary {
      border-color: #666;
      color: #666;
    }

    .button-secondary:hover {
      border-color: #a0a0a0;
      color: #a0a0a0;
    }

    .button:before {
      content: '';
      position: absolute;
      top: 0;
      left: -100%;
      width: 100%;
      height: 100%;
      background: linear-gradient(90deg, transparent, rgba(255, 255, 255, 0.2), transparent);
      transition: left 0.5s;
    }

    .button:hover:before {
      left: 100%;
    }
  `

  render() {
    return html`
      <div class="dialog" @click=${this._handleDialogClick}>
        <div class="dialog-header">
          <h2 class="dialog-title">Add Bookmark</h2>
          <button class="close-button" @click=${this._handleClose}>×</button>
        </div>

        <form @submit=${this._handleSubmit}>
          <div class="form-group">
            <label class="form-label">URL</label>
            <input 
              type="url" 
              class="form-input"
              placeholder="https://example.com"
              .value=${this._url}
              @input=${this._handleUrlInput}
              required
            />
            ${this._isAnalyzing ? html`
              <div class="analyze-status">
                <div class="analyze-spinner"></div>
                AI is analyzing this link...
              </div>
            ` : ''}
          </div>

          <div class="form-group">
            <label class="form-label">Title</label>
            <input 
              type="text" 
              class="form-input"
              placeholder="Page title (auto-detected)"
              .value=${this._title}
              @input=${this._handleTitleInput}
            />
          </div>

          <div class="form-group">
            <label class="form-label">Description</label>
            <textarea 
              class="form-textarea"
              placeholder="Optional description..."
              .value=${this._description}
              @input=${this._handleDescriptionInput}
            ></textarea>
          </div>

          <div class="form-group">
            <label class="form-label">Tags</label>
            <input 
              type="text" 
              class="form-input"
              placeholder="development, react, tutorial (comma separated)"
              .value=${this._tags}
              @input=${this._handleTagsInput}
            />
            <div class="tag-suggestions">
              <span class="tag-suggestion" @click=${this._addSuggestedTag} data-tag="Development">Development</span>
              <span class="tag-suggestion" @click=${this._addSuggestedTag} data-tag="Tutorial">Tutorial</span>
              <span class="tag-suggestion" @click=${this._addSuggestedTag} data-tag="Reference">Reference</span>
            </div>
          </div>

          <div class="dialog-actions">
            <button type="button" class="button button-secondary" @click=${this._handleClose}>
              Cancel
            </button>
            <button type="submit" class="button button-primary">
              Save Bookmark
            </button>
          </div>
        </form>
      </div>
    `
  }

  private _handleDialogClick(e: Event) {
    e.stopPropagation()
  }

  private _handleClose() {
    this.dispatchEvent(new CustomEvent('close'))
  }

  private _handleSubmit(e: Event) {
    e.preventDefault()
    
    const bookmark = {
      url: this._url,
      title: this._title || this._url,
      description: this._description,
      tags: this._tags.split(',').map(tag => tag.trim()).filter(tag => tag)
    }

    this.dispatchEvent(new CustomEvent('save', {
      detail: bookmark
    }))
  }

  private _handleUrlInput(e: Event) {
    const input = e.target as HTMLInputElement
    this._url = input.value
    
    // Simulate AI analysis
    if (this._url && this._url.startsWith('http')) {
      this._simulateAnalysis()
    }
  }

  private _handleTitleInput(e: Event) {
    const input = e.target as HTMLInputElement
    this._title = input.value
  }

  private _handleDescriptionInput(e: Event) {
    const textarea = e.target as HTMLTextAreaElement
    this._description = textarea.value
  }

  private _handleTagsInput(e: Event) {
    const input = e.target as HTMLInputElement
    this._tags = input.value
  }

  private _addSuggestedTag(e: Event) {
    const button = e.target as HTMLElement
    const tag = button.dataset.tag
    if (tag) {
      const currentTags = this._tags.split(',').map(t => t.trim()).filter(t => t)
      if (!currentTags.includes(tag)) {
        this._tags = [...currentTags, tag].join(', ')
      }
    }
  }

  private _simulateAnalysis() {
    this._isAnalyzing = true
    
    setTimeout(() => {
      this._isAnalyzing = false
      // Simulate fetched metadata
      if (this._url.includes('github.com')) {
        this._title = 'GitHub Repository'
        this._tags = 'Development, Code, Git'
      } else if (this._url.includes('youtube.com')) {
        this._title = 'YouTube Video'
        this._tags = 'Video, Tutorial, Media'
      } else {
        this._title = 'Analyzed Website'
        this._tags = 'Web, Reference'
      }
    }, 1500)
  }
}