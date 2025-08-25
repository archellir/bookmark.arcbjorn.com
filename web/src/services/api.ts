// API service for Torimemo bookmark manager

export interface Bookmark {
  id: number
  title: string
  url: string
  description?: string
  favicon_url?: string
  created_at: string
  updated_at: string
  is_favorite: boolean
  tags?: Tag[]
}

export interface Tag {
  id: number
  name: string
  color: string
  created_at: string
  count?: number
}

export interface CreateBookmarkRequest {
  title: string
  url: string
  description?: string
  tags?: string[]
}

export interface UpdateBookmarkRequest {
  title?: string
  url?: string
  description?: string
  is_favorite?: boolean
  tags?: string[]
}

export interface BookmarkListResponse {
  bookmarks: Bookmark[]
  total: number
  page: number
  limit: number
  has_more: boolean
  total_pages: number
  tag_count: number
  favorite_count: number
}

export interface SearchResult extends Bookmark {
  rank: number
  snippet?: string
}

export interface ApiError {
  error: string
  status: number
}

class ApiService {
  private baseUrl = '/api'

  private async request<T>(endpoint: string, options?: RequestInit): Promise<T> {
    const url = `${this.baseUrl}${endpoint}`
    
    const config: RequestInit = {
      headers: {
        'Content-Type': 'application/json',
        ...options?.headers,
      },
      ...options,
    }

    try {
      const response = await fetch(url, config)
      
      if (!response.ok) {
        let errorData: ApiError
        try {
          errorData = await response.json()
        } catch {
          errorData = {
            error: `HTTP ${response.status}: ${response.statusText}`,
            status: response.status
          }
        }
        throw new Error(errorData.error || 'API request failed')
      }

      // Handle no content responses
      if (response.status === 204) {
        return null as T
      }

      return await response.json()
    } catch (error) {
      if (error instanceof Error) {
        throw error
      }
      throw new Error('Network error occurred')
    }
  }

  // Health check
  async getHealth(): Promise<{ status: string; message: string; version: string; features: string[] }> {
    return this.request('/health')
  }

  // Database stats
  async getStats(): Promise<Record<string, any>> {
    return this.request('/stats')
  }

  // Bookmarks
  async getBookmarks(params: {
    page?: number
    limit?: number
    search?: string
    tag?: string
    favorites?: boolean
  } = {}): Promise<BookmarkListResponse> {
    const searchParams = new URLSearchParams()
    
    if (params.page) searchParams.set('page', params.page.toString())
    if (params.limit) searchParams.set('limit', params.limit.toString())
    if (params.search) searchParams.set('search', params.search)
    if (params.tag) searchParams.set('tag', params.tag)
    if (params.favorites) searchParams.set('favorites', 'true')

    const query = searchParams.toString()
    return this.request(`/bookmarks${query ? `?${query}` : ''}`)
  }

  async getBookmark(id: number): Promise<Bookmark> {
    return this.request(`/bookmarks/${id}`)
  }

  async createBookmark(data: CreateBookmarkRequest): Promise<Bookmark> {
    return this.request('/bookmarks', {
      method: 'POST',
      body: JSON.stringify(data),
    })
  }

  async updateBookmark(id: number, data: UpdateBookmarkRequest): Promise<Bookmark> {
    return this.request(`/bookmarks/${id}`, {
      method: 'PUT',
      body: JSON.stringify(data),
    })
  }

  async deleteBookmark(id: number): Promise<void> {
    return this.request(`/bookmarks/${id}`, {
      method: 'DELETE',
    })
  }

  async searchBookmarks(query: string, limit = 20): Promise<{
    query: string
    results: SearchResult[]
    count: number
  }> {
    const searchParams = new URLSearchParams({
      q: query,
      limit: limit.toString()
    })
    return this.request(`/bookmarks/search?${searchParams.toString()}`)
  }

  // Tags
  async getTags(search?: string): Promise<{ tags: Tag[]; count: number }> {
    const searchParams = new URLSearchParams()
    if (search) searchParams.set('search', search)
    
    const query = searchParams.toString()
    return this.request(`/tags${query ? `?${query}` : ''}`)
  }

  async createTag(data: { name: string; color?: string }): Promise<Tag> {
    return this.request('/tags', {
      method: 'POST',
      body: JSON.stringify(data),
    })
  }

  async updateTag(id: number, data: { name?: string; color?: string }): Promise<Tag> {
    return this.request(`/tags/${id}`, {
      method: 'PUT',
      body: JSON.stringify(data),
    })
  }

  async deleteTag(id: number): Promise<void> {
    return this.request(`/tags/${id}`, {
      method: 'DELETE',
    })
  }

  async getTagCloud(limit = 20): Promise<{ 
    tags: Array<{ name: string; count: number; size: number; color: string }>
    count: number 
  }> {
    return this.request(`/tags/cloud?limit=${limit}`)
  }

  async getPopularTags(limit = 10): Promise<{ tags: Tag[]; count: number }> {
    return this.request(`/tags/popular?limit=${limit}`)
  }

  async cleanupUnusedTags(): Promise<{ message: string; deleted_count: number }> {
    return this.request('/tags/cleanup', {
      method: 'DELETE',
    })
  }

  // Import/Export
  async exportBookmarks(): Promise<Blob> {
    const response = await fetch(`${this.baseUrl}/export`)
    if (!response.ok) {
      throw new Error('Failed to export bookmarks')
    }
    return response.blob()
  }

  async importBookmarks(file: File): Promise<{ imported: number; skipped: number; total: number }> {
    const formData = new FormData()
    formData.append('file', file)
    
    // Read file as JSON and send as request body
    const fileContent = await file.text()
    const data = JSON.parse(fileContent)
    
    return this.request('/import', {
      method: 'POST',
      body: JSON.stringify(data),
    })
  }
}

export const apiService = new ApiService()