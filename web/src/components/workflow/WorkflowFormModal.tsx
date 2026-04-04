import { Form, Input, InputNumber, Modal, Select, Spin } from 'antd'
import { useEffect, useState } from 'react'
import { pluginApi } from '@/api/plugin'
import type { PluginInfo, WorkflowInfo, WorkflowType } from '@/types'

interface Props {
  open: boolean
  appId: string
  workflowType: WorkflowType
  initialValues?: WorkflowInfo
  onOk: (values: Omit<WorkflowInfo, 'id' | 'appId' | 'createdAt' | 'updatedAt'>) => Promise<void>
  onCancel: () => void
}

export default function WorkflowFormModal({
  open, appId: _appId, workflowType, initialValues, onOk, onCancel,
}: Props) {
  const [form] = Form.useForm()
  const [plugins, setPlugins] = useState<PluginInfo[]>([])
  const [loadingPlugins, setLoadingPlugins] = useState(false)

  useEffect(() => {
    if (!open) return
    setLoadingPlugins(true)
    pluginApi.list().then((resp) => {
      setPlugins(workflowType === 'pre_work' ? (resp.pre_plugins ?? []) : (resp.post_plugins ?? []))
    }).finally(() => setLoadingPlugins(false))
  }, [open, workflowType])

  useEffect(() => {
    if (open) {
      form.setFieldsValue(
        initialValues
          ? {
              workflowType: initialValues.workflowType,
              funcKey: initialValues.funcKey,
              params: initialValues.params,
              sortOrder: initialValues.sortOrder,
            }
          : { workflowType: workflowType, params: '{}', sortOrder: 0 },
      )
    }
  }, [open, initialValues, workflowType, form])

  async function handleOk() {
    const values = await form.validateFields()
    await onOk(values)
    form.resetFields()
  }

  return (
    <Modal
      title={initialValues ? '编辑工作流' : '新建工作流'}
      open={open}
      onOk={handleOk}
      onCancel={onCancel}
      destroyOnClose
    >
      <Spin spinning={loadingPlugins}>
        <Form form={form} layout="vertical" style={{ marginTop: 16 }}>
          <Form.Item name="funcKey" label="插件" rules={[{ required: true }]}>
            <Select placeholder="选择插件" showSearch optionFilterProp="label">
              {plugins.map((p) => (
                <Select.Option key={p.key} value={p.key} label={p.key}>
                  <span style={{ fontWeight: 600 }}>{p.key}</span>
                  {p.desc && <span style={{ color: '#888', marginLeft: 8, fontSize: 12 }}>{p.desc}</span>}
                </Select.Option>
              ))}
            </Select>
          </Form.Item>
          <Form.Item
            name="params"
            label="参数 (JSON)"
            rules={[{ validator: validateJson }]}
          >
            <Input.TextArea rows={4} placeholder='{"key":"value"}' />
          </Form.Item>
          <Form.Item name="sortOrder" label="排序">
            <InputNumber min={0} style={{ width: '100%' }} />
          </Form.Item>
        </Form>
      </Spin>
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
