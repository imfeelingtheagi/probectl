import { useQuery } from '@tanstack/react-query'
import { apiFetch } from './client'

/** The tenant lifecycle view (S-T5, core): retention control + residency/
 *  isolation visibility. Export streams from /v1/lifecycle/export. */

export interface LifecycleStatus {
  tenant_id?: string
  flow_retention_days: number | null
  isolation_model: string
  residency?: string
}

export function useLifecycle() {
  return useQuery({
    queryKey: ['lifecycle'],
    queryFn: () => apiFetch<LifecycleStatus>('/lifecycle/retention'),
  })
}
