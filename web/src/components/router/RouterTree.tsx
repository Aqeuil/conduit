import { useEffect, useRef, useState } from 'react'
import {
  Button,
  Input,
  Modal,
  Space,
  Spin,
  Tag,
  Tree,
  Typography,
  theme,
  message,
  type InputRef,
} from 'antd'
import {
  ApiOutlined,
  DeleteOutlined,
  EditOutlined,
  FolderAddOutlined,
  FolderOutlined,
  PlusOutlined,
} from '@ant-design/icons'
import type { DataNode } from 'antd/es/tree'
import { routerApi } from '@/api/router'
import RouterFormModal from './RouterFormModal'
import type { RouterNode } from '@/types'

const METHOD_COLORS: Record<string, string> = {
  GET: 'success',
  POST: 'processing',
  PUT: 'warning',
  DELETE: 'error',
  PATCH: 'purple',
}

interface Props {
  appId: string
  selectedKey: string | null
  onSelect: (node: RouterNode) => void
  onReload?: () => void
}

interface TreeNode extends DataNode {
  raw: RouterNode
  children: TreeNode[]
}

// ── NodeTitle ──────────────────────────────────────────────────────────────────

interface NodeTitleProps {
  node: RouterNode
  onCreateDir: (parentId: number) => void
  onCreateRouter: (parentId: number) => void
  onEdit: (node: RouterNode) => void
  onDelete: (node: RouterNode) => void
}

function NodeTitle({ node, onCreateDir, onCreateRouter, onEdit, onDelete }: NodeTitleProps) {
  const { token } = theme.useToken()
  const isDir = node.type === 'directory'

  return (
    <div
      className="router-tree-node"
      style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', width: '100%' }}
    >
      {/* ── Left: icon + label ── */}
      <Space size={6} style={{ flex: 1, minWidth: 0 }}>
        {isDir ? (
          <FolderOutlined style={{ color: token.colorTextSecondary, fontSize: 14 }} />
        ) : (
          <ApiOutlined style={{ color: token.colorPrimary, fontSize: 13 }} />
        )}
        {!isDir && (
          <Tag
            color={METHOD_COLORS[node.method] ?? 'default'}
            style={{ fontSize: 10, lineHeight: '16px', padding: '0 4px', margin: 0, borderRadius: 3 }}
          >
            {node.method}
          </Tag>
        )}
        <span style={{ overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap', fontSize: 13 }}>
          {node.name || node.path}
        </span>
      </Space>

      {/* ── Right: action buttons, shown on hover ── */}
      <Space
        size={0}
        className="router-tree-actions"
        style={{ flexShrink: 0, opacity: 0, transition: 'opacity 0.15s' }}
        onClick={(e) => e.stopPropagation()}
      >
        {isDir && (
          <>
            <Button
              type="text" size="small"
              icon={<FolderAddOutlined />}
              title="新建子目录"
              onClick={() => onCreateDir(node.id)}
            />
            <Button
              type="text" size="small"
              icon={<PlusOutlined />}
              title="新建接口"
              onClick={() => onCreateRouter(node.id)}
            />
          </>
        )}
        <Button
          type="text" size="small"
          icon={<EditOutlined />}
          title="编辑"
          onClick={() => onEdit(node)}
        />
        <Button
          type="text" size="small" danger
          icon={<DeleteOutlined />}
          title="删除"
          onClick={() => onDelete(node)}
        />
      </Space>
    </div>
  )
}

// ── RouterTree ─────────────────────────────────────────────────────────────────

/** 只存原始数据，不存 JSX，避免 stale closure */
function toTreeNodes(nodes: RouterNode[]): TreeNode[] {
  return nodes.map((n) => ({
    key: `${n.type}-${n.id}`,
    title: n.name || n.path,   // antd 要求 title 存在；实际渲染由 titleRender 覆盖
    isLeaf: n.type === 'router',
    raw: n,
    children: n.children ? toTreeNodes(n.children) : [],
  }))
}

export default function RouterTree({ appId, selectedKey, onSelect, onReload }: Props) {
  const [treeData, setTreeData] = useState<TreeNode[]>([])
  const [loading, setLoading] = useState(false)
  const [expandedKeys, setExpandedKeys] = useState<React.Key[]>([])

  // directory modal
  const [dirModalMode, setDirModalMode] = useState<'create' | 'rename' | null>(null)
  const [dirModalParent, setDirModalParent] = useState<number>(0)
  const [dirModalTarget, setDirModalTarget] = useState<RouterNode | null>(null)
  const [dirInputVal, setDirInputVal] = useState('')
  const dirInputRef = useRef<InputRef>(null)

  // router form modal
  const [routerModalOpen, setRouterModalOpen] = useState(false)
  const [routerModalParent, setRouterModalParent] = useState<number>(0)
  const [routerEditing, setRouterEditing] = useState<RouterNode | undefined>()

  async function load() {
    setLoading(true)
    try {
      const resp = await routerApi.list(appId)
      setTreeData(toTreeNodes(resp.routers ?? []))
      setExpandedKeys((prev) =>
        prev.length > 0 ? prev : collectDirKeys(resp.routers ?? []),
      )
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => { load() }, [appId]) // eslint-disable-line react-hooks/exhaustive-deps

  useEffect(() => {
    if (dirModalMode) setTimeout(() => dirInputRef.current?.focus(), 100)
  }, [dirModalMode])

  // ── Directory actions ────────────────────────────────────────────────────────

  function openCreateDir(parentId: number) {
    setDirModalParent(parentId)
    setDirModalTarget(null)
    setDirInputVal('')
    setDirModalMode('create')
  }

  function openRenameDir(node: RouterNode) {
    setDirModalTarget(node)
    setDirInputVal(node.name)
    setDirModalMode('rename')
  }

  async function handleDirModalOk() {
    const name = dirInputVal.trim()
    if (!name) return
    if (dirModalMode === 'create') {
      await routerApi.create({ appId, parentId: dirModalParent, type: 'directory', name, sortOrder: 0 })
      message.success('目录已创建')
    } else if (dirModalMode === 'rename' && dirModalTarget) {
      await routerApi.update({
        id: dirModalTarget.id,
        parentId: dirModalTarget.parentId,
        name,
        sortOrder: dirModalTarget.sortOrder,
      })
      message.success('已重命名')
    }
    setDirModalMode(null)
    setDirInputVal('')
    load()
    onReload?.()
  }

  async function handleDeleteNode(node: RouterNode) {
    await routerApi.delete(node.id)
    message.success('已删除')
    load()
    onReload?.()
  }

  function confirmDelete(node: RouterNode) {
    Modal.confirm({
      title: `删除"${node.name || node.path}"？`,
      content: node.type === 'directory' ? '目录下的路由也会一并删除，不可恢复。' : '路由将被永久删除。',
      okType: 'danger',
      okText: '删除',
      cancelText: '取消',
      onOk: () => handleDeleteNode(node),
    })
  }

  // ── Router actions ───────────────────────────────────────────────────────────

  function openCreateRouter(parentId: number) {
    setRouterEditing(undefined)
    setRouterModalParent(parentId)
    setRouterModalOpen(true)
  }

  function openEditRouter(node: RouterNode) {
    setRouterEditing(node)
    setRouterModalParent(node.parentId)
    setRouterModalOpen(true)
  }

  async function handleRouterSave(
    values: Omit<RouterNode, 'id' | 'appId' | 'createdAt' | 'updatedAt' | 'children'>,
  ) {
    if (routerEditing) {
      await routerApi.update({ ...values, id: routerEditing.id })
      message.success('更新成功')
    } else {
      await routerApi.create({
        ...values,
        appId,
        type: 'router',
        parentId: routerModalParent
      })
      message.success('创建成功')
    }
    setRouterModalOpen(false)
    setRouterEditing(undefined)
    load()
    onReload?.()
  }

  // ── titleRender: called fresh on every render, no stale closure ──────────────

  function titleRender(node: DataNode) {
    const raw = (node as TreeNode).raw
    return (
      <NodeTitle
        node={raw}
        onCreateDir={openCreateDir}
        onCreateRouter={openCreateRouter}
        onEdit={(n) => n.type === 'directory' ? openRenameDir(n) : openEditRouter(n)}
        onDelete={confirmDelete}
      />
    )
  }

  // ── Render ───────────────────────────────────────────────────────────────────

  return (
    <>
      <style>{`
        .router-tree-node:hover .router-tree-actions,
        .ant-tree-treenode-selected .router-tree-actions {
          opacity: 1 !important;
        }
      `}</style>

      <Spin spinning={loading}>
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 8 }}>
          <Typography.Text strong style={{ fontSize: 13 }}>路由树</Typography.Text>
          <Space size={4}>
            <Button size="small" icon={<FolderAddOutlined />} onClick={() => openCreateDir(0)}>
              新建目录
            </Button>
            <Button size="small" icon={<PlusOutlined />} onClick={() => openCreateRouter(0)}>
              新建接口
            </Button>
          </Space>
        </div>

        {treeData.length === 0 && !loading ? (
          <div style={{ color: '#999', fontSize: 12, paddingLeft: 4, paddingTop: 8 }}>暂无数据</div>
        ) : (
          <Tree
            blockNode
            treeData={treeData}
            titleRender={titleRender}
            expandedKeys={expandedKeys}
            selectedKeys={selectedKey !== null ? [selectedKey] : []}
            onExpand={(keys) => setExpandedKeys(keys)}
            onSelect={(keys, { node }) => {
              const raw = (node as unknown as TreeNode).raw
              if (raw.type === 'directory') {
                const k = `directory-${raw.id}`
                setExpandedKeys((prev) =>
                  prev.includes(k) ? prev.filter((x) => x !== k) : [...prev, k],
                )
              } else {
                if (keys.length) onSelect(raw)
              }
            }}
          />
        )}
      </Spin>

      <Modal
        title={dirModalMode === 'create' ? '新建目录' : '重命名目录'}
        open={!!dirModalMode}
        onOk={handleDirModalOk}
        onCancel={() => setDirModalMode(null)}
        destroyOnClose
      >
        <Input
          ref={dirInputRef}
          value={dirInputVal}
          onChange={(e) => setDirInputVal(e.target.value)}
          onPressEnter={handleDirModalOk}
          placeholder="目录名称"
          style={{ marginTop: 16 }}
        />
      </Modal>

      <RouterFormModal
        open={routerModalOpen}
        appId={appId}
        parentId={routerModalParent}
        initialValues={routerEditing}
        onOk={handleRouterSave}
        onCancel={() => { setRouterModalOpen(false); setRouterEditing(undefined) }}
      />
    </>
  )
}

function collectDirKeys(nodes: RouterNode[]): string[] {
  const keys: string[] = []
  for (const n of nodes) {
    if (n.type === 'directory') {
      keys.push(`directory-${n.id}`)
      if (n.children?.length) keys.push(...collectDirKeys(n.children))
    }
  }
  return keys
}
