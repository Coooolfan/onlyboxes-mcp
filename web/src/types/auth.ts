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
  console_version: string
  console_repo_url: string
}

export interface RegisterAccountPayload {
  account: AccountProfile
  created_at: string
  updated_at: string
}

export interface AccountListItem extends AccountProfile {
  created_at: string
  updated_at: string
}

export interface AccountListResponse {
  items: AccountListItem[]
  total: number
  page: number
  page_size: number
}
