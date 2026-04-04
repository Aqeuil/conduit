import { useEffect, useState } from 'react'
import { Button, Card, Popconfirm, Segmented, Space, Spin, Tag, Typography, message } from 'antd'
import { PlusOutlined } from '@ant-design/icons'
import { workflowApi } from '@/api/workflow'
import WorkflowFormModal from './WorkflowFormModal'
import type { WorkflowInfo, WorkflowType } from '@/types'

interface Props {
  appId: string
}

export default function WorkflowList({ appId }: Props) {
  const [wfType, setWfType] = useState<WorkflowType>('pre_work')
  const [data, setData] = useState<WorkflowInfo[]>([])
  const [loading, setLoading] = useState(false)
  const [modalOpen, setModalOpen] = useState(false)
  const [editing, setEditing] = useState<WorkflowInfo | undefined>()

  async function load(type: WorkflowType = wfType) {
    setLoading(true)
    try {
      const resp = await workflowApi.list(appId, type)
      setData(resp.items ?? [])
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => { load() }, [appId, wfType]) // eslint-disable-line react-hooks/exhaustive-deps

  async function handleSave(values: Omit<WorkflowInfo, 'id' | 'appId' | 'createdAt' | 'updatedAt'>) {
    if (editing) {
      await workflowApi.update({ ...values, id: editing.id })
      message.success('更新成功')
    } else {
      await workflowApi.create({ ...values, appId: appId, workflowType: wfType })
      message.success('创建成功')
    }
    setModalOpen(false)
    setEditing(undefined)
    load()
  }

  async function handleDelete(id: number) {
    await workflowApi.delete(id)
    message.success('已删除')
    load()
  }

  return (
    <Spin spinning={loading}>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 16 }}>
        <Segmented
          options={[
            { label: '前置工作流 (pre)', value: 'pre_work' },
            { label: '后置工作流 (post)', value: 'post_work' },
          ]}
          value={wfType}
          onChange={(v) => setWfType(v as WorkflowType)}
        />
        <Button
          type="primary"
          icon={<PlusOutlined />}
          onClick={() => { setEditing(undefined); setModalOpen(true) }}
        >
          新增
        </Button>
      </div>

      {data.length === 0 ? (
        <div style={{ color: '#999', textAlign: 'center', paddingTop: 40 }}>暂无工作流</div>
      ) : (
        <Space direction="vertical" style={{ width: '100%' }}>
          {data.map((wf) => (
            <Card
              key={wf.id}
              size="small"
              extra={
                <Space>
                  <Button size="small" onClick={() => { setEditing(wf); setModalOpen(true) }}>
                    编辑
                  </Button>
                  <Popconfirm title="确认删除？" onConfirm={() => handleDelete(wf.id)}>
                    <Button size="small" danger>删除</Button>
                  </Popconfirm>
                </Space>
              }
            >
              <Space>
                <Tag color="blue">#{wf.sortOrder}</Tag>
                <Typography.Text strong>{wf.funcKey}</Typography.Text>
                {wf.params && wf.params !== '{}' && (
                  <Typography.Text type="secondary" style={{ fontSize: 12 }}>
                    params: {wf.params}
                  </Typography.Text>
                )}
              </Space>
            </Card>
          ))}
        </Space>
      )}

      <WorkflowFormModal
        open={modalOpen}
        appId={appId}
        workflowType={wfType}
        initialValues={editing}
        onOk={handleSave}
        onCancel={() => { setModalOpen(false); setEditing(undefined) }}
      />
    </Spin>
  )
}
