import '@/style.css'
import '@components/app-root.ts'
import { errorBoundary } from '@services/error-boundary.ts'
import { performanceMonitor } from '@services/performance.ts'
import { securityService } from '@services/security.ts'

// Initialize production services
errorBoundary.clearStoredErrors()
performanceMonitor.startAutoReporting()
securityService.initialize()

// Track app initialization time
const initComplete = performanceMonitor.trackInteraction('AppInit')

document.querySelector<HTMLDivElement>('#app')!.innerHTML = `
  <app-root></app-root>
`

// Mark initialization complete
setTimeout(() => {
  initComplete()
}, 0)