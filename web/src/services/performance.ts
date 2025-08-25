interface PerformanceMetric {
  name: string
  value: number
  timestamp: number
  url: string
  userId?: string
}

export class PerformanceMonitor {
  private static instance: PerformanceMonitor
  private metrics: PerformanceMetric[] = []
  private observer?: PerformanceObserver

  static getInstance(): PerformanceMonitor {
    if (!PerformanceMonitor.instance) {
      PerformanceMonitor.instance = new PerformanceMonitor()
    }
    return PerformanceMonitor.instance
  }

  private constructor() {
    this.setupPerformanceObserver()
    this.measureInitialMetrics()
  }

  private setupPerformanceObserver() {
    if ('PerformanceObserver' in window) {
      this.observer = new PerformanceObserver((list) => {
        const entries = list.getEntries()
        
        for (const entry of entries) {
          // Track Core Web Vitals
          if (entry.entryType === 'largest-contentful-paint') {
            this.recordMetric('LCP', entry.startTime)
          }
          
          if (entry.entryType === 'first-input') {
            this.recordMetric('FID', (entry as any).processingStart - entry.startTime)
          }
          
          if (entry.entryType === 'layout-shift' && !(entry as any).hadRecentInput) {
            this.recordMetric('CLS', (entry as any).value)
          }

          // Track navigation timing
          if (entry.entryType === 'navigation') {
            const nav = entry as PerformanceNavigationTiming
            this.recordMetric('TTFB', nav.responseStart - nav.requestStart)
            this.recordMetric('DOMContentLoaded', nav.domContentLoadedEventEnd - nav.domContentLoadedEventStart)
            this.recordMetric('LoadComplete', nav.loadEventEnd - nav.loadEventStart)
          }

          // Track resource timing for critical assets
          if (entry.entryType === 'resource') {
            const resource = entry as PerformanceResourceTiming
            if (resource.name.includes('index.js') || resource.name.includes('index.css')) {
              this.recordMetric(`Resource_${resource.name.split('/').pop()}`, resource.responseEnd - resource.requestStart)
            }
          }
        }
      })

      try {
        this.observer.observe({ entryTypes: ['largest-contentful-paint', 'first-input', 'layout-shift', 'navigation', 'resource'] })
      } catch (e) {
        console.warn('Performance Observer setup failed:', e)
      }
    }
  }

  private measureInitialMetrics() {
    // Measure First Contentful Paint
    if ('performance' in window && 'getEntriesByType' in performance) {
      const paintEntries = performance.getEntriesByType('paint')
      const fcpEntry = paintEntries.find(entry => entry.name === 'first-contentful-paint')
      if (fcpEntry) {
        this.recordMetric('FCP', fcpEntry.startTime)
      }
    }

    // Measure memory usage if available
    if ('memory' in performance) {
      const memory = (performance as any).memory
      this.recordMetric('MemoryUsed', memory.usedJSHeapSize / 1024 / 1024) // MB
      this.recordMetric('MemoryTotal', memory.totalJSHeapSize / 1024 / 1024) // MB
    }
  }

  private recordMetric(name: string, value: number) {
    const metric: PerformanceMetric = {
      name,
      value: Math.round(value * 100) / 100, // Round to 2 decimal places
      timestamp: Date.now(),
      url: window.location.pathname,
      userId: this.getCurrentUserId()
    }

    this.metrics.push(metric)

    // Log in development
    if (import.meta.env.DEV) {
      console.log(`📊 ${name}: ${value}ms`)
    }

    // Keep only last 100 metrics to prevent memory bloat
    if (this.metrics.length > 100) {
      this.metrics = this.metrics.slice(-100)
    }
  }

  private getCurrentUserId(): string | undefined {
    try {
      const user = localStorage.getItem('user')
      return user ? JSON.parse(user).id : undefined
    } catch {
      return undefined
    }
  }

  // Method to manually track custom metrics
  measureAndRecord<T>(name: string, fn: () => T | Promise<T>): T | Promise<T> {
    const startTime = performance.now()
    
    try {
      const result = fn()
      
      if (result instanceof Promise) {
        return result.then((value) => {
          this.recordMetric(name, performance.now() - startTime)
          return value
        }).catch((error) => {
          this.recordMetric(name, performance.now() - startTime)
          throw error
        })
      } else {
        this.recordMetric(name, performance.now() - startTime)
        return result
      }
    } catch (error) {
      this.recordMetric(name, performance.now() - startTime)
      throw error
    }
  }

  // Method to track user interactions
  trackInteraction(action: string, element?: string) {
    const startTime = performance.now()
    const interactionName = `Interaction_${action}${element ? `_${element}` : ''}`
    
    // Return a function to call when interaction completes
    return () => {
      this.recordMetric(interactionName, performance.now() - startTime)
    }
  }

  // Get metrics summary
  getMetricsSummary() {
    const summary: { [key: string]: { avg: number, max: number, min: number, count: number } } = {}
    
    for (const metric of this.metrics) {
      if (!summary[metric.name]) {
        summary[metric.name] = { avg: 0, max: 0, min: Infinity, count: 0 }
      }
      
      const s = summary[metric.name]
      s.max = Math.max(s.max, metric.value)
      s.min = Math.min(s.min, metric.value)
      s.avg = (s.avg * s.count + metric.value) / (s.count + 1)
      s.count++
    }
    
    return summary
  }

  // Report metrics to server
  async reportMetrics() {
    if (this.metrics.length === 0) return

    try {
      const token = localStorage.getItem('auth_token')
      const headers: Record<string, string> = {
        'Content-Type': 'application/json'
      }
      
      if (token) {
        headers['Authorization'] = `Bearer ${token}`
      }

      const response = await fetch('/api/metrics', {
        method: 'POST',
        headers,
        body: JSON.stringify({ 
          metrics: this.metrics,
          summary: this.getMetricsSummary()
        })
      })

      if (response.ok) {
        // Clear metrics after successful report
        this.metrics = []
      }
    } catch (error) {
      console.warn('Failed to report metrics:', error)
    }
  }

  // Setup automatic reporting
  startAutoReporting(intervalMinutes = 5) {
    setInterval(() => {
      this.reportMetrics()
    }, intervalMinutes * 60 * 1000)
  }

  disconnect() {
    if (this.observer) {
      this.observer.disconnect()
    }
  }
}

export const performanceMonitor = PerformanceMonitor.getInstance()