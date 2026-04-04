import { post } from './client'
import type {
  ApplicationInfo,
  CreateApplicationReq,
  UpdateApplicationReq,
  ListApplicationsResp,
} from '@/types'

export const applicationApi = {
  create: (req: CreateApplicationReq) =>
    post<ApplicationInfo>('/application/create', req),

  update: (req: UpdateApplicationReq) =>
    post<ApplicationInfo>('/application/update', req),

  delete: (id: string) =>
    post<Record<string, never>>('/application/delete', { id }),

  get: (id: string) =>
    post<ApplicationInfo>('/application/get', { id }),

  list: (page = 1, page_size = 20) =>
    post<ListApplicationsResp>('/application/list', { page, page_size }),
}
