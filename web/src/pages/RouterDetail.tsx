import { useEffect, useState } from 'react'
import {
  Breadcrumb,
  Button,
  Descriptions,
  Form,
  Input,
  Select,
  Space,
  Tag,
  Typography,
  message,
} from 'antd'
import { useNavigate, useParams } from 'react-router-dom'
import { routerApi } from '@/api/router'
import { applicationApi } from '@/api/application'
import { useTabs } from '@/components/layout/TabsContext'
import type { RouterNode } from '@/types'

const HTTP_METHODS = ['GET', 'POST', 'PUT', 'DELETE', 'PATCH', 'HEAD', 'OPTIONS']

const METHOD_COLORS: Record<string, string> = {
  GET: 'green',
  POST: 'blue',
  PUT: 'orange',
  DELETE: 'red',
  PATCH: 'purple',
}

export default function RouterDetail() {
  const { appId, routerId } = useParams<{ appId: string; routerId: string }>()
  const navigate = useNavigate()
  const { openTab } = useTabs()
  const [router, setRouter] = useState<RouterNode | null>(null)
  const [appName, setAppName] = useState<string>('')
  const [editing, setEditing] = useState(false)
  const [form] = Form.useForm()

  useEffect(() => {
    if (!appId || !routerId) return
    Promise.all([
      routerApi.get(Number(routerId)),
      applicationApi.get(appId),
    ]).then(([r, app]) => {
      setRouter(r)
      setAppName(app.name)
      openTab(`/app/${appId}/router/${routerId}`, `${r.method} ${r.name || r.path}`)
    }).catch(() => {
      message.error('路由不存在')
      navigate(`/app/${appId}`)
    })
  }, [appId, routerId, navigate, openTab])

  if (!appId || !routerId || !router) return null

  function startEdit() {
    form.setFieldsValue({
      method: router!.method,
      name: router!.name,
      path: router!.path,
      description: router!.description,
      headers: router!.headers,
      requestSchema: router!.requestSchema,
      responseSchema: router!.responseSchema,
    })
    setEditing(true)
  }

  async function handleSave() {
    const values = await form.validateFields()
    await routerApi.update({ id: router!.id, parentId: router!.parentId, ...values })
    message.success('更新成功')
    const updated = await routerApi.get(Number(routerId))
    setRouter(updated)
    openTab(`/app/${appId}/router/${routerId}`, `${updated.method} ${updated.name || updated.path}`)
    setEditing(false)
  }

  return (
    <>
      <Breadcrumb
        style={{ marginBottom: 16 }}
        items={[
          { title: <a onClick={() => navigate('/')}>应用列表</a> },
          { title: <a onClick={() => navigate(`/app/${appId}`)}>{appName || appId}</a> },
          { title: `${router.method} ${router.name || router.path}` },
        ]}
      />

      <div style={{ display: 'flex', alignItems: 'center', gap: 12, marginBottom: 20 }}>
        <Tag color={METHOD_COLORS[router.method] ?? 'default'} style={{ fontSize: 13, padding: '2px 8px' }}>
          {router.method}
        </Tag>
        <Typography.Title level={4} style={{ margin: 0 }}>
          {router.name || router.path}
        </Typography.Title>
        {!editing && (
          <Button onClick={startEdit} style={{ marginLeft: 'auto' }}>编辑</Button>
        )}
      </div>

      {editing ? (
        <Form form={form} layout="vertical" style={{ maxWidth: 640 }}>
          <Form.Item name="method" label="HTTP 方法" rules={[{ required: true }]}>
            <Select>
              {HTTP_METHODS.map((m) => (
                <Select.Option key={m} value={m}>
                  <span style={{ color: METHOD_COLORS[m] ?? '#000', fontWeight: 600 }}>{m}</span>
                </Select.Option>
              ))}
            </Select>
          </Form.Item>
          <Form.Item name="name" label="名称" rules={[{ required: true }]}>
            <Input placeholder="获取用户信息" />
          </Form.Item>
          <Form.Item name="path" label="路径" rules={[{ required: true }]}>
            <Input placeholder="/api/v1/resource" />
          </Form.Item>
          <Form.Item name="description" label="描述">
            <Input />
          </Form.Item>
          <Form.Item name="headers" label="请求头 (JSON Array)" rules={[{ validator: validateJson }]}>
            <Input.TextArea rows={3} />
          </Form.Item>
          <Form.Item name="requestSchema" label="Request Schema (JSON)" rules={[{ validator: validateJson }]}>
            <Input.TextArea rows={4} />
          </Form.Item>
          <Form.Item name="responseSchema" label="Response Schema (JSON)" rules={[{ validator: validateJson }]}>
            <Input.TextArea rows={4} />
          </Form.Item>
          <Form.Item>
            <Space>
              <Button type="primary" onClick={handleSave}>保存</Button>
              <Button onClick={() => setEditing(false)}>取消</Button>
            </Space>
          </Form.Item>
        </Form>
      ) : (
        <Descriptions bordered column={1} style={{ maxWidth: 640 }}>
          <Descriptions.Item label="路径">{router.path}</Descriptions.Item>
          <Descriptions.Item label="描述">{router.description || '—'}</Descriptions.Item>
          <Descriptions.Item label="请求头">
            <pre style={{ margin: 0, fontSize: 12 }}>{formatJson(router.headers)}</pre>
          </Descriptions.Item>
          <Descriptions.Item label="Request Schema">
            <pre style={{ margin: 0, fontSize: 12 }}>{formatJson(router.requestSchema)}</pre>
          </Descriptions.Item>
          <Descriptions.Item label="Response Schema">
            <pre style={{ margin: 0, fontSize: 12 }}>{formatJson(router.responseSchema)}</pre>
          </Descriptions.Item>
          <Descriptions.Item label="创建时间">{router.createdAt}</Descriptions.Item>
          <Descriptions.Item label="更新时间">{router.updatedAt}</Descriptions.Item>
        </Descriptions>
      )}
    </>
  )
}

function formatJson(val: string) {
  try {
    return JSON.stringify(JSON.parse(val), null, 2)
  } catch {
    return val || '—'
  }
}

function validateJson(_: unknown, value: string) {
  if (!value) return Promise.resolve()
  try {
    JSON.parse(value)
    return Promise.resolve()
  } catch {
    return Promise.reject(new Error('请输入合法的 JSON'))
  }
}
