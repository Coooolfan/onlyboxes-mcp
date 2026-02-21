const buildConsoleVersion =
  typeof import.meta.env.VITE_CONSOLE_VERSION === 'string'
    ? import.meta.env.VITE_CONSOLE_VERSION.trim()
    : ''

export const defaultConsoleVersion = buildConsoleVersion || 'dev'
export const defaultConsoleRepoURL = 'https://github.com/Coooolfan/onlyboxes'
