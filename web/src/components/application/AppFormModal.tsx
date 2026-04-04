import { Form, Input, Modal } from 'antd'
import { useEffect } from 'react'
import type { ApplicationInfo } from '@/types'

interface Props {
  open: boolean
  initialValues?: ApplicationInfo
  onOk: (values: { name: string; upstream: string; description: string }) => Promise<void>
  onCancel: () => void
}

export default function AppFormModal({ open, initialValues, onOk, onCancel }: Props) {
  const [form] = Form.useForm()

  useEffect(() => {
    if (open) {
      form.setFieldsValue(initialValues ?? { name: '', upstream: '', description: '' })
    }
  }, [open, initialValues, form])

  async function handleOk() {
    const values = await form.validateFields()
    await onOk(values)
    form.resetFields()
  }

  return (
    <Modal
      title={initialValues ? '编辑应用' : '新建应用'}
      open={open}
      onOk={handleOk}
      onCancel={onCancel}
      destroyOnClose
    >
      <Form form={form} layout="vertical" style={{ marginTop: 16 }}>
        <Form.Item name="name" label="应用名称" rules={[{ required: true }]}>
          <Input placeholder="my-service" />
        </Form.Item>
        <Form.Item name="upstream" label="下游地址" rules={[{ required: true }]}>
          <Input placeholder="localhost:7890" />
        </Form.Item>
        <Form.Item name="description" label="描述">
          <Input.TextArea rows={3} />
        </Form.Item>
      </Form>
    </Modal>
  )
}
