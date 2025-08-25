import { LitElement, html, css } from 'lit'
import { customElement, state } from 'lit/decorators.js'
// import { searchSuggestionsService } from '../services/search-suggestions.ts' // TODO: Integrate suggestions service
import './search-suggestions.ts'
import './auth-dialog.ts'

interface User {
  id: number
  username: string
  email: string
  full_name?: string
  is_admin: boolean
}

@customElement('app-header-new')
export class AppHeaderNew extends LitElement {
  @state() private _searchQuery = ''
  @state() private _showSuggestions = false
  @state() private _selectedSuggestionIndex = -1
  @state() private _user: User | null = null
  @state() private _showUserMenu = false

  static styles = css`
    :host {
      display: block;
    }

    .user-menu {
      position: absolute;
      right: 0;
      top: 100%;
      margin-top: 0.5rem;
      min-width: 200px;
      z-index: 50;
    }
  `

  connectedCallback() {
    super.connectedCallback()
    this._loadUser()
    document.addEventListener('click', this._handleOutsideClick.bind(this))
  }

  disconnectedCallback() {
    super.disconnectedCallback()
    document.removeEventListener('click', this._handleOutsideClick.bind(this))
  }

  private _loadUser() {
    const token = localStorage.getItem('auth_token')
    const userStr = localStorage.getItem('user')
    
    if (token && userStr) {
      try {
        this._user = JSON.parse(userStr)
      } catch (e) {
        console.error('Failed to parse user data:', e)
        this._logout()
      }
    }
  }

  private _handleOutsideClick(e: Event) {
    const target = e.target as Element
    if (!target.closest('app-header-new')) {
      this._showUserMenu = false
    }
  }

  private _handleSearch(e: CustomEvent) {
    this._searchQuery = e.detail.query
    this._showSuggestions = false
    
    // Dispatch search event
    this.dispatchEvent(new CustomEvent('search', {
      detail: { query: this._searchQuery }
    }))
  }

  private _handleAuthSuccess(e: CustomEvent) {
    this._user = e.detail.user
    this._showUserMenu = false
  }

  private _showAuthDialog() {
    const authDialog = this.shadowRoot?.querySelector('auth-dialog')
    if (authDialog) {
      (authDialog as any).open()
    }
  }

  private async _logout() {
    try {
      const token = localStorage.getItem('auth_token')
      if (token) {
        await fetch('/api/auth/logout', {
          method: 'POST',
          headers: {
            'Authorization': `Bearer ${token}`
          }
        })
      }
    } catch (e) {
      console.error('Logout request failed:', e)
    }

    localStorage.removeItem('auth_token')
    localStorage.removeItem('user')
    this._user = null
    this._showUserMenu = false

    // Dispatch logout event
    this.dispatchEvent(new CustomEvent('auth-logout'))
  }

  private _toggleUserMenu() {
    this._showUserMenu = !this._showUserMenu
  }

  private _showAddBookmark() {
    this.dispatchEvent(new CustomEvent('add-bookmark'))
  }

  render() {
    return html`
      <div class="header-layout">
        <div class="container">
          <div class="flex-between py-4">
            <!-- Logo -->
            <div class="text-2xl font-bold bg-gradient-to-r from-cyan-400 to-pink-500 bg-clip-text text-transparent">
              📚 Torimemo
            </div>

            <!-- Search Container -->
            <div class="flex-1 max-w-md mx-8 relative">
              <input
                type="text"
                .value=${this._searchQuery}
                @input=${(e: Event) => {
                  const target = e.target as HTMLInputElement
                  this._searchQuery = target.value
                  this._showSuggestions = target.value.length > 0
                }}
                @focus=${() => this._showSuggestions = this._searchQuery.length > 0}
                @blur=${() => setTimeout(() => this._showSuggestions = false, 200)}
                @keydown=${this._handleSearchKeydown}
                class="cyber-input w-full"
                placeholder="Search bookmarks..."
              />
              
              ${this._showSuggestions ? html`
                <search-suggestions
                  .query=${this._searchQuery}
                  .selectedIndex=${this._selectedSuggestionIndex}
                  @suggestion-select=${this._handleSearch}
                  class="absolute w-full z-10 mt-1">
                </search-suggestions>
              ` : ''}
            </div>

            <!-- Actions -->
            <div class="flex items-center space-x-4">
              <!-- Add Bookmark Button -->
              <button @click=${this._showAddBookmark} class="cyber-btn">
                <svg class="w-5 h-5 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4"></path>
                </svg>
                Add
              </button>

              <!-- User Menu -->
              ${this._user ? html`
                <div class="relative">
                  <button @click=${this._toggleUserMenu}
                          class="flex items-center space-x-2 text-cyan-400 hover:text-cyan-300 transition-colors">
                    <div class="w-8 h-8 bg-gradient-to-r from-cyan-500 to-blue-500 rounded-full 
                               flex items-center justify-center text-black font-bold">
                      ${this._user.username.charAt(0).toUpperCase()}
                    </div>
                    <span class="hidden md:block">${this._user.username}</span>
                    ${this._user.is_admin ? html`
                      <span class="cyber-badge text-xs">Admin</span>
                    ` : ''}
                    <svg class="w-4 h-4 transform transition-transform ${this._showUserMenu ? 'rotate-180' : ''}" 
                         fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7"></path>
                    </svg>
                  </button>

                  ${this._showUserMenu ? html`
                    <div class="user-menu cyber-dropdown">
                      <div class="px-4 py-3 border-b border-gray-700">
                        <div class="text-sm font-medium text-white">${this._user.username}</div>
                        <div class="text-xs text-gray-400">${this._user.email}</div>
                      </div>
                      
                      <div class="py-1">
                        <button class="cyber-dropdown-item w-full text-left flex items-center">
                          <svg class="w-4 h-4 mr-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" 
                                  d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z"></path>
                          </svg>
                          Profile
                        </button>
                        
                        <button class="cyber-dropdown-item w-full text-left flex items-center">
                          <svg class="w-4 h-4 mr-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" 
                                  d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z"></path>
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"></path>
                          </svg>
                          Settings
                        </button>
                        
                        <div class="border-t border-gray-700 my-1"></div>
                        
                        <button @click=${this._logout}
                                class="cyber-dropdown-item w-full text-left flex items-center text-red-400 hover:text-red-300">
                          <svg class="w-4 h-4 mr-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" 
                                  d="M17 16l4-4m0 0l-4-4m4 4H7m6 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h4a3 3 0 013 3v1"></path>
                          </svg>
                          Logout
                        </button>
                      </div>
                    </div>
                  ` : ''}
                </div>
              ` : html`
                <button @click=${this._showAuthDialog} class="cyber-btn">
                  <svg class="w-5 h-5 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" 
                          d="M11 16l-4-4m0 0l4-4m-4 4h14m-5 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h4a3 3 0 013 3v1"></path>
                  </svg>
                  Login
                </button>
              `}
            </div>
          </div>
        </div>
      </div>

      <auth-dialog @auth-success=${this._handleAuthSuccess}></auth-dialog>
    `
  }

  private _handleSearchKeydown = (e: KeyboardEvent) => {
    if (e.key === 'Escape') {
      this._showSuggestions = false
    } else if (e.key === 'Enter') {
      this._showSuggestions = false
      this.dispatchEvent(new CustomEvent('search', {
        detail: { query: this._searchQuery }
      }))
    }
  }
}

declare global {
  interface HTMLElementTagNameMap {
    'app-header-new': AppHeaderNew
  }
}