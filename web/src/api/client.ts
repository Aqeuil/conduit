import { API_BASE } from '@/config/env'
import { message } from 'antd'

interface ApiResponse<T> {
  code: number
  data: T
  message: string
}

export async function post<T>(path: string, body: unknown = {}): Promise<T> {
  const res = await fetch(`${API_BASE}${path}`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(body),
  })

  if (!res.ok) {
    const text = await res.text().catch(() => res.statusText)
    message.error(`请求失败: ${text}`)
    throw new Error(text)
  }

  const json: ApiResponse<T> = await res.json()

  if (json.code !== 200) {
    message.error(json.message)
    throw new Error(json.message)
  }

  return json.data
}
