export class UnauthorizedError extends Error {
  constructor() {
    super('authentication required')
    this.name = 'UnauthorizedError'
  }
}

export function isUnauthorizedError(error: unknown): error is UnauthorizedError {
  return error instanceof UnauthorizedError
}

export async function parseAPIError(response: Response): Promise<string> {
  try {
    const payload = (await response.json()) as { error?: string }
    if (typeof payload.error === 'string' && payload.error.trim() !== '') {
      return payload.error
    }
  } catch {
    // ignore parsing failures and fall back to status text
  }
  return `API ${response.status}: ${response.statusText}`
}

export async function request(url: string, init: RequestInit = {}): Promise<Response> {
  const headers = new Headers(init.headers ?? {})
  if (!headers.has('Accept')) {
    headers.set('Accept', 'application/json')
  }

  const response = await fetch(url, {
    ...init,
    headers,
    credentials: 'same-origin',
  })

  if (response.status === 401) {
    throw new UnauthorizedError()
  }

  return response
}
