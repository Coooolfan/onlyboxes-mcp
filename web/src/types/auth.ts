export type AuthState = 'loading' | 'unauthenticated' | 'authenticated'

export interface AccountProfile {
  account_id: string
  username: string
  is_admin: boolean
}

export interface ConsoleSessionPayload {
  authenticated: boolean
  account: AccountProfile
  registration_enabled: boolean
}

export interface RegisterAccountPayload {
  account: AccountProfile
  created_at: string
  updated_at: string
}
