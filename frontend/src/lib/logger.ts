const DEBUG_ENABLED = import.meta.env.VITE_DEBUG === 'true'

type LogLevel = 'log' | 'info' | 'warn' | 'error'

const createLogger = (module: string) => {
  const log = (level: LogLevel, message: string, data?: unknown) => {
    if (!DEBUG_ENABLED) return

    const timestamp = new Date().toISOString()
    const prefix = `[${timestamp}] [${module}]`

    switch (level) {
      case 'log':
        console.log(`${prefix} 📝`, message, data ?? '')
        break
      case 'info':
        console.info(`${prefix} ℹ️`, message, data ?? '')
        break
      case 'warn':
        console.warn(`${prefix} ⚠️`, message, data ?? '')
        break
      case 'error':
        console.error(`${prefix} ❌`, message, data ?? '')
        break
    }
  }

  return {
    log: (message: string, data?: unknown) => log('log', message, data),
    info: (message: string, data?: unknown) => log('info', message, data),
    warn: (message: string, data?: unknown) => log('warn', message, data),
    error: (message: string, data?: unknown) => log('error', message, data),
  }
}

export const logger = {
  oauth: createLogger('OAuth'),
  api: createLogger('API'),
  auth: createLogger('Auth'),
  account: createLogger('Account'),
  app: createLogger('App'),
}
