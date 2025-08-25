export class SecurityService {
  private static instance: SecurityService

  static getInstance(): SecurityService {
    if (!SecurityService.instance) {
      SecurityService.instance = new SecurityService()
    }
    return SecurityService.instance
  }

  private constructor() {
    this.setupCSPMonitoring()
    this.setupSecurityHeaders()
  }

  private setupCSPMonitoring() {
    document.addEventListener('securitypolicyviolation', (e) => {
      console.warn('CSP Violation:', {
        blockedURI: e.blockedURI,
        violatedDirective: e.violatedDirective,
        originalPolicy: e.originalPolicy,
        documentURI: e.documentURI,
        statusCode: e.statusCode
      })
      
      // Report CSP violations to server if endpoint exists
      this.reportSecurityViolation('csp-violation', {
        blockedURI: e.blockedURI,
        violatedDirective: e.violatedDirective,
        originalPolicy: e.originalPolicy,
        documentURI: e.documentURI
      })
    })
  }

  private setupSecurityHeaders() {
    // Add security-related meta tags if not present
    if (!document.querySelector('meta[http-equiv="X-Content-Type-Options"]')) {
      const meta = document.createElement('meta')
      meta.setAttribute('http-equiv', 'X-Content-Type-Options')
      meta.setAttribute('content', 'nosniff')
      document.head.appendChild(meta)
    }
  }

  // Sanitize user input to prevent XSS
  sanitizeHTML(html: string): string {
    const temp = document.createElement('div')
    temp.textContent = html
    return temp.innerHTML
  }

  // Validate and sanitize URLs
  sanitizeURL(url: string): string | null {
    try {
      const parsed = new URL(url)
      
      // Only allow http and https protocols
      if (!['http:', 'https:'].includes(parsed.protocol)) {
        return null
      }
      
      return parsed.href
    } catch {
      return null
    }
  }

  // Generate secure random strings
  generateSecureToken(length: number = 32): string {
    const array = new Uint8Array(length)
    crypto.getRandomValues(array)
    return Array.from(array, byte => byte.toString(16).padStart(2, '0')).join('')
  }

  // Check if running over HTTPS in production
  enforceHTTPS() {
    if (location.protocol !== 'https:' && location.hostname !== 'localhost') {
      if (import.meta.env.PROD) {
        // Redirect to HTTPS in production
        location.replace(`https:${location.href.substring(location.protocol.length)}`)
      } else {
        console.warn('Not running over HTTPS. This is only acceptable in development.')
      }
    }
  }

  // Validate authentication tokens
  isValidToken(token: string): boolean {
    try {
      // Basic JWT format validation (header.payload.signature)
      const parts = token.split('.')
      if (parts.length !== 3) return false
      
      // Check if parts are valid base64
      for (const part of parts) {
        try {
          atob(part.replace(/-/g, '+').replace(/_/g, '/'))
        } catch {
          return false
        }
      }
      
      return true
    } catch {
      return false
    }
  }

  // Clear sensitive data from memory (best effort)
  clearSensitiveData(obj: any) {
    if (typeof obj === 'object' && obj !== null) {
      for (const key in obj) {
        if (obj.hasOwnProperty(key)) {
          if (typeof obj[key] === 'string') {
            obj[key] = '\0'.repeat(obj[key].length)
          } else if (typeof obj[key] === 'object') {
            this.clearSensitiveData(obj[key])
          }
        }
      }
    }
  }

  // Rate limiting for API calls (client-side)
  private rateLimiter = new Map<string, { count: number, resetTime: number }>()
  
  checkRateLimit(key: string, maxRequests: number = 100, windowMs: number = 60000): boolean {
    const now = Date.now()
    const window = this.rateLimiter.get(key)
    
    if (!window || now > window.resetTime) {
      this.rateLimiter.set(key, { count: 1, resetTime: now + windowMs })
      return true
    }
    
    if (window.count >= maxRequests) {
      return false
    }
    
    window.count++
    return true
  }

  // Report security violations
  private async reportSecurityViolation(type: string, details: any) {
    try {
      const token = localStorage.getItem('auth_token')
      const headers: Record<string, string> = {
        'Content-Type': 'application/json'
      }
      
      if (token) {
        headers['Authorization'] = `Bearer ${token}`
      }

      await fetch('/api/security-violations', {
        method: 'POST',
        headers,
        body: JSON.stringify({
          type,
          details,
          timestamp: Date.now(),
          userAgent: navigator.userAgent,
          url: window.location.href
        })
      })
    } catch (error) {
      console.warn('Failed to report security violation:', error)
    }
  }

  // Initialize security measures
  initialize() {
    this.enforceHTTPS()
    
    // Clear clipboard when window loses focus (for sensitive data)
    window.addEventListener('blur', () => {
      if (navigator.clipboard?.writeText) {
        navigator.clipboard.writeText('').catch(() => {
          // Ignore errors - clipboard clearing is best effort
        })
      }
    })

    // Disable right-click in production (optional security measure)
    if (import.meta.env.PROD) {
      document.addEventListener('contextmenu', (e) => {
        e.preventDefault()
      })
      
      // Disable common developer shortcuts
      document.addEventListener('keydown', (e) => {
        if (
          (e.ctrlKey && e.shiftKey && (e.key === 'I' || e.key === 'J' || e.key === 'C')) ||
          (e.ctrlKey && e.key === 'U') ||
          e.key === 'F12'
        ) {
          e.preventDefault()
        }
      })
    }
  }
}

export const securityService = SecurityService.getInstance()