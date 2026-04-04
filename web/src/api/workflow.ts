import { post } from './client'
import type {
  WorkflowInfo,
  WorkflowType,
  CreateWorkflowReq,
  UpdateWorkflowReq,
} from '@/types'

export const workflowApi = {
  create: (req: CreateWorkflowReq) =>
    post<WorkflowInfo>('/workflow/create', req),

  update: (req: UpdateWorkflowReq) =>
    post<WorkflowInfo>('/workflow/update', req),

  delete: (id: number) =>
    post<Record<string, never>>('/workflow/delete', { id }),

  list: (appId: string, workflowType?: WorkflowType) =>
    post<{ items: WorkflowInfo[] }>('/workflow/list', { appId, workflowType }),
}
