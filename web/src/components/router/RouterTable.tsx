import { useEffect, useState } from 'react'
import { Button, Popconfirm, Space, Table, Tag, message } from 'antd'
import { PlusOutlined } from '@ant-design/icons'
import type { ColumnsType } from 'antd/es/table'
import { routerApi } from '@/api/router'
import RouterFormModal from './RouterFormModal'
import type { RouterNode } from '@/types'

const METHOD_COLORS: Record<string, string> = {
  GET: 'green',
  POST: 'blue',
  PUT: 'orange',
  DELETE: 'red',
  PATCH: 'purple',
}

interface Props {
  appId: string
  parentId: number | null  // selected directory id; null = nothing selected
  onSaved?: () => void     // callback to reload the tree
}

export default function RouterTable({ appId, parentId, onSaved }: Props) {
  const [data, setData] = useState<RouterNode[]>([])
  const [loading, setLoading] = useState(false)
  const [modalOpen, setModalOpen] = useState(false)
  const [editing, setEditing] = useState<RouterNode | undefined>()

  async function load() {
    if (parentId === null) return
    setLoading(true)
    try {
      // Fetch the full tree and flatten only the direct router children of this directory
      const resp = await routerApi.list(appId)
      const routers = collectRouters(resp.routers ?? [], parentId)
      setData(routers)
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => { load() }, [appId, parentId]) // eslint-disable-line react-hooks/exhaustive-deps

  async function handleSave(values: Omit<RouterNode, 'id' | 'appId' | 'createdAt' | 'updatedAt' | 'children'>) {
    if (editing) {
      await routerApi.update({ ...values, id: editing.id })
      message.success('更新成功')
    } else {
      await routerApi.create({ ...values, appId, type: 'router' })
      message.success('创建成功')
    }
    setModalOpen(false)
    setEditing(undefined)
    load()
    onSaved?.()
  }

  async function handleDelete(id: number) {
    await routerApi.delete(id)
    message.success('已删除')
    load()
    onSaved?.()
  }

  const columns: ColumnsType<RouterNode> = [
    {
      title: '方法',
      dataIndex: 'method',
      width: 90,
      render: (m) => <Tag color={METHOD_COLORS[m] ?? 'default'}>{m}</Tag>,
    },
    { title: '路径', dataIndex: 'path' },
    { title: '描述', dataIndex: 'description' },
    {
      title: '操作',
      width: 120,
      render: (_, record) => (
        <Space>
          <Button size="small" onClick={() => { setEditing(record); setModalOpen(true) }}>
            编辑
          </Button>
          <Popconfirm title="确认删除该路由？" onConfirm={() => handleDelete(record.id)}>
            <Button size="small" danger>删除</Button>
          </Popconfirm>
        </Space>
      ),
    },
  ]

  if (parentId === null) {
    return (
      <div style={{ color: '#999', paddingTop: 60, textAlign: 'center' }}>
        请在左侧选择一个目录
      </div>
    )
  }

  return (
    <>
      <div style={{ display: 'flex', justifyContent: 'flex-end', marginBottom: 12 }}>
        <Button
          type="primary"
          icon={<PlusOutlined />}
          onClick={() => { setEditing(undefined); setModalOpen(true) }}
        >
          新建路由
        </Button>
      </div>
      <Table rowKey="id" loading={loading} columns={columns} dataSource={data} pagination={false} />

      <RouterFormModal
        open={modalOpen}
        appId={appId}
        parentId={parentId}
        initialValues={editing}
        onOk={handleSave}
        onCancel={() => { setModalOpen(false); setEditing(undefined) }}
      />
    </>
  )
}

// Walk the tree to find the node with the given id, then collect its direct router children.
function collectRouters(nodes: RouterNode[], parentId: number): RouterNode[] {
  for (const node of nodes) {
    if (node.id === parentId) {
      return (node.children ?? []).filter((c) => c.type === 'router')
    }
    if (node.children?.length) {
      const found = collectRouters(node.children, parentId)
      if (found.length > 0 || nodeContains(node, parentId)) return found
    }
  }
  return []
}

function nodeContains(node: RouterNode, id: number): boolean {
  if (node.id === id) return true
  return (node.children ?? []).some((c) => nodeContains(c, id))
}
