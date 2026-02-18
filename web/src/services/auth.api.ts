import { parseAPIError, request } from '@/services/http'

export class InvalidCredentialsError extends Error {
  constructor() {
    super('invalid credentials')
    this.name = 'InvalidCredentialsError'
  }
}

export function isInvalidCredentialsError(error: unknown): error is InvalidCredentialsError {
  return error instanceof InvalidCredentialsError
}

export async function loginAPI(username: string, password: string): Promise<void> {
  const response = await fetch('/api/v1/console/login', {
    method: 'POST',
    headers: {
      Accept: 'application/json',
      'Content-Type': 'application/json',
    },
    credentials: 'same-origin',
    body: JSON.stringify({ username, password }),
  })

  if (response.status === 401) {
    throw new InvalidCredentialsError()
  }
  if (!response.ok) {
    throw new Error(await parseAPIError(response))
  }
}

export async function logoutAPI(): Promise<void> {
  await fetch('/api/v1/console/logout', {
    method: 'POST',
    headers: {
      Accept: 'application/json',
    },
    credentials: 'same-origin',
  })
}

export async function probeSessionAPI(): Promise<void> {
  const response = await request('/api/v1/workers/stats?stale_after_sec=30')
  if (!response.ok) {
    throw new Error(await parseAPIError(response))
  }
}
