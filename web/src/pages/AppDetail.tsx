import { useEffect, useState } from 'react'
import { Breadcrumb, Tabs, Typography, message } from 'antd'
import { useNavigate, useParams } from 'react-router-dom'
import { applicationApi } from '@/api/application'
import RouterTree from '@/components/router/RouterTree'
import WorkflowList from '@/components/workflow/WorkflowList'
import { useTabs } from '@/components/layout/TabsContext'
import type { ApplicationInfo, RouterNode } from '@/types'

export default function AppDetail() {
  const { appId } = useParams<{ appId: string }>()
  const navigate = useNavigate()
  const { openTab } = useTabs()
  const [app, setApp] = useState<ApplicationInfo | null>(null)
  const [selectedKey, setSelectedKey] = useState<string | null>(null)

  useEffect(() => {
    if (!appId) return
    applicationApi.get(appId).then((info) => {
      setApp(info)
      openTab(`/app/${appId}`, info.name)
    }).catch(() => {
      message.error('应用不存在')
      navigate('/')
    })
  }, [appId, navigate, openTab])

  if (!appId) return null

  function handleNodeSelect(node: RouterNode) {
    setSelectedKey(`${node.type}-${node.id}`)
    if (node.type === 'router') {
      navigate(`/app/${appId}/router/${node.id}`)
    }
  }

  return (
    <>
      <Breadcrumb
        style={{ marginBottom: 16 }}
        items={[
          { title: <a onClick={() => navigate('/')}>应用列表</a> },
          { title: app?.name ?? appId },
        ]}
      />

      <Typography.Title level={4} style={{ marginBottom: 4 }}>
        {app?.name}
      </Typography.Title>
      {app?.upstream && (
        <Typography.Text type="secondary" style={{ display: 'block', marginBottom: 16 }}>
          上游: {app.upstream}
        </Typography.Text>
      )}

      <Tabs
        defaultActiveKey="routers"
        items={[
          {
            key: 'routers',
            label: '路由管理',
            children: (
              <RouterTree
                appId={appId}
                selectedKey={selectedKey}
                onSelect={handleNodeSelect}
              />
            ),
          },
          {
            key: 'workflows',
            label: '工作流',
            children: <WorkflowList appId={appId} />,
          },
        ]}
      />
    </>
  )
}
