import { useEffect, useState } from 'react'
import { Button, Popconfirm, Space, Table, Tag, Tooltip, Typography, message } from 'antd'
import { PlusOutlined } from '@ant-design/icons'
import { useNavigate } from 'react-router-dom'
import type { ColumnsType } from 'antd/es/table'
import { applicationApi } from '@/api/application'
import AppFormModal from '@/components/application/AppFormModal'
import { useTabs } from '@/components/layout/TabsContext'
import type { ApplicationInfo } from '@/types'

export default function ApplicationList() {
  const navigate = useNavigate()
  const { openTab } = useTabs()
  const [data, setData] = useState<ApplicationInfo[]>([])
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(1)
  const [loading, setLoading] = useState(false)
  const [modalOpen, setModalOpen] = useState(false)
  const [editing, setEditing] = useState<ApplicationInfo | undefined>()

  async function load(p = page) {
    setLoading(true)
    try {
      const resp = await applicationApi.list(p, 20)
      setData(resp.items ?? [])
      setTotal(resp.total ?? 0)
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => { openTab('/', '应用列表') }, [openTab])
  useEffect(() => { load() }, [page]) // eslint-disable-line react-hooks/exhaustive-deps

  async function handleSave(values: { name: string; upstream: string; description: string }) {
    if (editing) {
      await applicationApi.update({ ...values, id: editing.id })
      message.success('更新成功')
    } else {
      await applicationApi.create(values)
      message.success('创建成功')
    }
    setModalOpen(false)
    setEditing(undefined)
    load()
  }

  async function handleDelete(id: string) {
    await applicationApi.delete(id)
    message.success('已删除')
    load()
  }

  const columns: ColumnsType<ApplicationInfo> = [
    {
      title: '应用名称',
      dataIndex: 'name',
      render: (name, record) => (
        <Button
          type="link"
          onClick={() => {
            const path = `/app/${record.id}`
            openTab(path, record.name)
            navigate(path)
          }}
        >
          {name}
        </Button>
      ),
    },
    {
      title: '下游地址',
      dataIndex: 'upstream',
      render: (v) => <Tag>{v}</Tag>,
    },
    {
      title: '描述',
      dataIndex: 'description',
      width: 300,
      render: (v: string) => (
        <Tooltip title={v}>
          <span style={{ display: 'block', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap', maxWidth: 280 }}>
            {v}
          </span>
        </Tooltip>
      ),
    },
    {
      title: '创建时间',
      dataIndex: 'createdAt',
      width: 280,
      // render: (v) => new Date(v).toLocaleString(),
    },
    {
      title: '操作',
      width: 120,
      render: (_, record) => (
        <Space>
          <Button
            size="small"
            onClick={() => { setEditing(record); setModalOpen(true) }}
          >
            编辑
          </Button>
          <Popconfirm
            title="确认删除该应用？"
            onConfirm={() => handleDelete(record.id)}
          >
            <Button size="small" danger>删除</Button>
          </Popconfirm>
        </Space>
      ),
    },
  ]

  return (
    <>
      <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 16 }}>
        <Typography.Title level={4} style={{ margin: 0 }}>应用列表</Typography.Title>
        <Button
          type="primary"
          icon={<PlusOutlined />}
          onClick={() => { setEditing(undefined); setModalOpen(true) }}
        >
          新建应用
        </Button>
      </div>

      <Table
        rowKey="id"
        loading={loading}
        columns={columns}
        dataSource={data}
        pagination={{ total, current: page, pageSize: 20, onChange: setPage }}
      />

      <AppFormModal
        open={modalOpen}
        initialValues={editing}
        onOk={handleSave}
        onCancel={() => { setModalOpen(false); setEditing(undefined) }}
      />
    </>
  )
}
