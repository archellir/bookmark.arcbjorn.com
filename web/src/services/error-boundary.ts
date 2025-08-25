export interface ErrorInfo {
  message: string
  stack?: string
  timestamp: number
  userAgent: string
  url: string
  userId?: string
}

export class ErrorBoundaryService {
  private static instance: ErrorBoundaryService
  private errorQueue: ErrorInfo[] = []
  private maxRetries = 3
  private retryDelay = 1000

  static getInstance(): ErrorBoundaryService {
    if (!ErrorBoundaryService.instance) {
      ErrorBoundaryService.instance = new ErrorBoundaryService()
    }
    return ErrorBoundaryService.instance
  }

  private constructor() {
    this.setupGlobalErrorHandlers()
  }

  private setupGlobalErrorHandlers() {
    // Handle uncaught JavaScript errors
    window.addEventListener('error', (event) => {
      this.handleError({
        message: event.message,
        stack: event.error?.stack,
        timestamp: Date.now(),
        userAgent: navigator.userAgent,
        url: window.location.href,
        userId: this.getCurrentUserId()
      })
    })

    // Handle unhandled promise rejections
    window.addEventListener('unhandledrejection', (event) => {
      this.handleError({
        message: `Unhandled Promise Rejection: ${event.reason}`,
        stack: event.reason?.stack,
        timestamp: Date.now(),
        userAgent: navigator.userAgent,
        url: window.location.href,
        userId: this.getCurrentUserId()
      })
    })
  }

  private getCurrentUserId(): string | undefined {
    try {
      const user = localStorage.getItem('user')
      return user ? JSON.parse(user).id : undefined
    } catch {
      return undefined
    }
  }

  handleError(errorInfo: ErrorInfo) {
    // Log to console in development
    if (import.meta.env.DEV) {
      console.error('Error captured by boundary:', errorInfo)
    }

    // Add to queue for reporting
    this.errorQueue.push(errorInfo)

    // Attempt to report immediately
    this.reportErrors()

    // Show user-friendly error message
    this.showErrorNotification(errorInfo)
  }

  private async reportErrors(retryCount = 0) {
    if (this.errorQueue.length === 0) return

    try {
      const token = localStorage.getItem('auth_token')
      const headers: Record<string, string> = {
        'Content-Type': 'application/json'
      }
      
      if (token) {
        headers['Authorization'] = `Bearer ${token}`
      }

      const response = await fetch('/api/errors', {
        method: 'POST',
        headers,
        body: JSON.stringify({ errors: this.errorQueue })
      })

      if (response.ok) {
        // Clear queue on successful report
        this.errorQueue = []
      } else {
        throw new Error(`Failed to report errors: ${response.status}`)
      }
    } catch (error) {
      console.warn('Failed to report errors:', error)

      // Retry with exponential backoff
      if (retryCount < this.maxRetries) {
        setTimeout(() => {
          this.reportErrors(retryCount + 1)
        }, this.retryDelay * Math.pow(2, retryCount))
      } else {
        // Store in localStorage as fallback
        this.storeErrorsLocally()
      }
    }
  }

  private storeErrorsLocally() {
    try {
      const existingErrors = localStorage.getItem('pending_errors')
      const errors = existingErrors ? JSON.parse(existingErrors) : []
      
      // Keep only last 10 errors to prevent localStorage bloat
      const allErrors = [...errors, ...this.errorQueue].slice(-10)
      
      localStorage.setItem('pending_errors', JSON.stringify(allErrors))
      this.errorQueue = []
    } catch (error) {
      console.warn('Failed to store errors locally:', error)
    }
  }

  private showErrorNotification(_errorInfo: ErrorInfo) {
    // Create toast notification
    const toast = document.createElement('div')
    toast.className = 'error-toast'
    toast.innerHTML = `
      <div class="error-toast-content">
        <div class="error-toast-icon">⚠️</div>
        <div>
          <div class="error-toast-title">Something went wrong</div>
          <div class="error-toast-message">We've been notified and are looking into it.</div>
        </div>
        <button class="error-toast-close" onclick="this.parentElement.parentElement.remove()">×</button>
      </div>
    `

    // Add styles
    const style = document.createElement('style')
    style.textContent = `
      .error-toast {
        position: fixed;
        top: 20px;
        right: 20px;
        background: rgba(255, 23, 68, 0.9);
        color: white;
        padding: 16px;
        border-radius: 8px;
        box-shadow: 0 4px 12px rgba(0, 0, 0, 0.3);
        z-index: 10000;
        max-width: 400px;
        backdrop-filter: blur(10px);
        border: 1px solid rgba(255, 23, 68, 0.3);
      }
      
      .error-toast-content {
        display: flex;
        align-items: flex-start;
        gap: 12px;
      }
      
      .error-toast-icon {
        font-size: 20px;
        flex-shrink: 0;
      }
      
      .error-toast-title {
        font-weight: bold;
        margin-bottom: 4px;
      }
      
      .error-toast-message {
        font-size: 14px;
        opacity: 0.9;
      }
      
      .error-toast-close {
        background: none;
        border: none;
        color: white;
        font-size: 20px;
        cursor: pointer;
        padding: 0;
        margin-left: auto;
        opacity: 0.7;
      }
      
      .error-toast-close:hover {
        opacity: 1;
      }
      
      @media (max-width: 480px) {
        .error-toast {
          top: 10px;
          right: 10px;
          left: 10px;
          max-width: none;
        }
      }
    `

    document.head.appendChild(style)
    document.body.appendChild(toast)

    // Auto remove after 5 seconds
    setTimeout(() => {
      toast.remove()
      style.remove()
    }, 5000)
  }

  // Method for components to manually report errors
  reportError(error: Error, context?: string) {
    this.handleError({
      message: context ? `${context}: ${error.message}` : error.message,
      stack: error.stack,
      timestamp: Date.now(),
      userAgent: navigator.userAgent,
      url: window.location.href,
      userId: this.getCurrentUserId()
    })
  }

  // Clear any stored errors (useful after successful app startup)
  clearStoredErrors() {
    localStorage.removeItem('pending_errors')
  }
}

export const errorBoundary = ErrorBoundaryService.getInstance()