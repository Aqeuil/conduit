import { Form, Input, Modal, Select } from 'antd'
import { useEffect } from 'react'
import type { RouterNode } from '@/types'

const HTTP_METHODS = ['GET', 'POST', 'PUT', 'DELETE', 'PATCH', 'HEAD', 'OPTIONS']

const METHOD_COLORS: Record<string, string> = {
  GET: '#52c41a',
  POST: '#1677ff',
  PUT: '#fa8c16',
  DELETE: '#f5222d',
  PATCH: '#722ed1',
}

interface Props {
  open: boolean
  appId: string
  parentId: number  // directory node id that this router belongs to
  initialValues?: RouterNode
  onOk: (values: Omit<RouterNode, 'id' | 'appId' | 'createdAt' | 'updatedAt' | 'children'>) => Promise<void>
  onCancel: () => void
}

export default function RouterFormModal({
  open, appId, parentId, initialValues, onOk, onCancel,
}: Props) {
  const [form] = Form.useForm()

  useEffect(() => {
    if (open) {
      form.setFieldsValue(
        initialValues
          ? {
              parentId: initialValues.parentId,
              name: initialValues.name,
              path: initialValues.path,
              method: initialValues.method,
              headers: initialValues.headers,
              requestSchema: initialValues.requestSchema,
              responseSchema: initialValues.responseSchema,
              description: initialValues.description,
            }
          : {
              parentId: parentId,
              method: 'GET',
              headers: '[]',
              requestSchema: '{}',
              responseSchema: '{}',
            },
      )
    }
  }, [open, initialValues, parentId, form])

  async function handleOk() {
    const values = await form.validateFields()
    await onOk({ ...values, appId: appId })
    form.resetFields()
  }

  return (
    <Modal
      title={initialValues ? '编辑路由' : '新建路由'}
      open={open}
      onOk={handleOk}
      onCancel={onCancel}
      width={640}
      destroyOnClose
    >
      <Form form={form} layout="vertical" style={{ marginTop: 16 }}>
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
        <Form.Item
          name="headers"
          label='请求头 (JSON Array)'
          rules={[{ validator: validateJson }]}
        >
          <Input.TextArea rows={3} placeholder='[{"key":"Authorization","value":"Bearer token"}]' />
        </Form.Item>
        <Form.Item
          name="requestSchema"
          label="Request Schema (JSON)"
          rules={[{ validator: validateJson }]}
        >
          <Input.TextArea rows={4} placeholder='{"type":"object","properties":{}}' />
        </Form.Item>
        <Form.Item
          name="responseSchema"
          label="Response Schema (JSON)"
          rules={[{ validator: validateJson }]}
        >
          <Input.TextArea rows={4} placeholder='{"type":"object","properties":{}}' />
        </Form.Item>
      </Form>
    </Modal>
  )
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
