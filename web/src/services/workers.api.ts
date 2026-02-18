import { parseAPIError, request } from '@/services/http'
import type {
  WorkerListResponse,
  WorkerStartupCommandResponse,
  WorkerStatsResponse,
  WorkerStatus,
} from '@/types/workers'

export async function fetchWorkersAPI(
  status: WorkerStatus,
  page: number,
  pageSize: number,
  signal: AbortSignal,
): Promise<WorkerListResponse> {
  const query = new URLSearchParams({
    status,
    page: String(page),
    page_size: String(pageSize),
  })

  const response = await request(`/api/v1/workers?${query.toString()}`, { signal })
  if (!response.ok) {
    throw new Error(await parseAPIError(response))
  }

  const payload = (await response.json()) as WorkerListResponse
  return {
    items: payload.items ?? [],
    total: payload.total ?? 0,
    page: payload.page ?? page,
    page_size: payload.page_size ?? pageSize,
  }
}

export async function fetchWorkerStatsAPI(
  staleAfterSec: number,
  signal: AbortSignal,
): Promise<WorkerStatsResponse> {
  const query = new URLSearchParams({ stale_after_sec: String(staleAfterSec) })

  const response = await request(`/api/v1/workers/stats?${query.toString()}`, { signal })
  if (!response.ok) {
    throw new Error(await parseAPIError(response))
  }

  const payload = (await response.json()) as WorkerStatsResponse
  return {
    total: payload.total ?? 0,
    online: payload.online ?? 0,
    offline: payload.offline ?? 0,
    stale: payload.stale ?? 0,
    stale_after_sec: payload.stale_after_sec ?? staleAfterSec,
    generated_at: payload.generated_at ?? '',
  }
}

export async function fetchWorkerStartupCommandAPI(nodeID: string): Promise<string> {
  const response = await request(`/api/v1/workers/${encodeURIComponent(nodeID)}/startup-command`)
  if (!response.ok) {
    throw new Error(await parseAPIError(response))
  }

  const payload = (await response.json()) as WorkerStartupCommandResponse
  const command = payload.command?.trim()
  if (!command) {
    throw new Error('API returned empty startup command.')
  }
  return command
}

export async function createWorkerAPI(): Promise<WorkerStartupCommandResponse> {
  const response = await request('/api/v1/workers', {
    method: 'POST',
    headers: {
      Accept: 'application/json',
      'Content-Type': 'application/json',
    },
    body: '{}',
  })

  if (!response.ok) {
    throw new Error(await parseAPIError(response))
  }

  const payload = (await response.json()) as WorkerStartupCommandResponse
  const nodeID = payload.node_id?.trim()
  const command = payload.command?.trim()
  if (!nodeID || !command) {
    throw new Error('API returned invalid worker startup payload.')
  }

  return {
    node_id: nodeID,
    command,
  }
}

export async function deleteWorkerAPI(nodeID: string): Promise<void> {
  const response = await request(`/api/v1/workers/${encodeURIComponent(nodeID)}`, {
    method: 'DELETE',
  })

  if (response.status === 204) {
    return
  }

  if (!response.ok) {
    throw new Error(await parseAPIError(response))
  }
}
