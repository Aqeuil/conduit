import { post } from './client'
import type { RouterNode, ListRoutersResp, CreateRouterReq, UpdateRouterReq } from '@/types'

export const routerApi = {
  create: (req: CreateRouterReq) =>
    post<RouterNode>('/router/create', req),

  update: (req: UpdateRouterReq) =>
    post<RouterNode>('/router/update', req),

  delete: (id: number) =>
    post<Record<string, never>>('/router/delete', { id }),

  get: (id: number) =>
    post<RouterNode>('/router/get', { id }),

  list: (appId: string) =>
    post<ListRoutersResp>('/router/list', { appId }),
}
