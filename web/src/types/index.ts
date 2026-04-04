// ---- Plugin ----

export interface PluginParamRule {
  type: string
  name: string
  children?: PluginParamRule[]
}

export interface PluginInfo {
  key: string
  desc: string
  rule: PluginParamRule[]
}

// ---- Application ----

export interface ApplicationInfo {
  id: string
  name: string
  upstream: string
  description: string
  created_at: string
  updated_at: string
}

export interface CreateApplicationReq {
  name: string
  upstream: string
  description: string
}

export interface UpdateApplicationReq {
  id: string
  name: string
  upstream: string
  description: string
}

export interface ListApplicationsResp {
  items: ApplicationInfo[]
  total: number
}

// ---- Router (unified directory + route node) ----

export type RouterNodeType = 'directory' | 'router'

export interface RouterNode {
  id: number
  appId: string
  parentId: number
  type: RouterNodeType
  // directory fields
  name: string
  sortOrder: number
  // router fields
  path: string
  method: string
  headers: string        // JSON array string
  requestSchema: string  // JSON object string
  responseSchema: string
  description: string
  createdAt: string
  updatedAt: string
  children?: RouterNode[]
}

export interface ListRoutersResp {
  routers: RouterNode[]
}

export interface CreateRouterReq {
  appId: string
  parentId: number
  type: RouterNodeType
  // directory fields
  name?: string
  sortOrder?: number
  // router fields
  path?: string
  method?: string
  headers?: string
  requestSchema?: string
  responseSchema?: string
  description?: string
}

export interface UpdateRouterReq {
  id: number
  parentId: number
  // directory fields
  name?: string
  sortOrder?: number
  // router fields
  path?: string
  method?: string
  headers?: string
  requestSchema?: string
  responseSchema?: string
  description?: string
}

// ---- Workflow ----

export type WorkflowType = 'pre_work' | 'post_work'

export interface WorkflowInfo {
  id: number
  appId: string
  workflowType: WorkflowType
  funcKey: string
  params: string  // JSON object string
  sortOrder: number
  createdAt: string
  updatedAt: string
}

export interface CreateWorkflowReq {
  appId: string
  workflowType: WorkflowType
  funcKey: string
  params: string
  sortOrder: number
}

export interface UpdateWorkflowReq {
  id: number
  workflowType: WorkflowType
  funcKey: string
  params: string
  sortOrder: number
}
