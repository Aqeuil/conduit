import { useEffect } from 'react'
import { Layout, Menu, Tabs, Typography, theme } from 'antd'
import { ApiOutlined, UnorderedListOutlined, AppstoreOutlined } from '@ant-design/icons'
import { Outlet, useNavigate, useLocation } from 'react-router-dom'
import { TabsProvider, useTabs } from './TabsContext'

const { Content, Sider } = Layout
const { useToken } = theme

function Inner() {
  const navigate = useNavigate()
  const location = useLocation()
  const { tabs, activeKey, closeTab, setActiveKey } = useTabs()
  const { token } = useToken()

  // On first mount, navigate to the last active tab if we're on the root
  useEffect(() => {
    if (location.pathname === '/' && activeKey !== '/') {
      navigate(activeKey, { replace: true })
    }
  }, []) // eslint-disable-line react-hooks/exhaustive-deps

  const isDetail = location.pathname.startsWith('/app/')
  const selectedMenuKey = isDetail ? '/app' : '/'

  function handleTabChange(key: string) {
    setActiveKey(key)
    navigate(key)
  }

  function handleTabEdit(
    targetKey: React.MouseEvent | React.KeyboardEvent | string,
    action: 'add' | 'remove',
  ) {
    if (action === 'remove' && typeof targetKey === 'string') {
      closeTab(targetKey, () => {
        const remaining = tabs.filter((t) => t.key !== targetKey)
        navigate(remaining[remaining.length - 1].key)
      })
    }
  }

  return (
    <Layout style={{ minHeight: '100vh' }}>
      {/* ── Sidebar ── */}
      <Sider
        width={220}
        style={{
          background: token.colorBgContainer,
          borderRight: `1px solid ${token.colorBorderSecondary}`,
          display: 'flex',
          flexDirection: 'column',
        }}
      >
        {/* Logo / title */}
        <div
          style={{
            height: 48,
            display: 'flex',
            alignItems: 'center',
            gap: 10,
            padding: '0 20px',
            borderBottom: `1px solid ${token.colorBorderSecondary}`,
            flexShrink: 0,
          }}
        >
          <ApiOutlined style={{ fontSize: 18, color: token.colorPrimary }} />
          <Typography.Text strong style={{ fontSize: 15, letterSpacing: 0.3 }}>
            Conduit Admin
          </Typography.Text>
        </div>

        <Menu
          mode="inline"
          selectedKeys={[selectedMenuKey]}
          defaultOpenKeys={['app-mgmt']}
          style={{ flex: 1, borderRight: 0, paddingTop: 8 }}
          items={[
            {
              key: 'app-mgmt',
              label: '应用管理',
              icon: <AppstoreOutlined />,
              children: [
                {
                  key: '/',
                  label: '应用列表',
                  icon: <UnorderedListOutlined />,
                  onClick: () => navigate('/'),
                },
                {
                  key: '/app',
                  label: '应用详情',
                  icon: <ApiOutlined />,
                  disabled: !isDetail,
                  onClick: () => isDetail && navigate(location.pathname),
                },
              ],
            },
          ]}
        />
      </Sider>

      {/* ── Main ── */}
      <Layout style={{ background: token.colorBgLayout }}>
        {/* Tab bar — flush to top */}
        <div
          style={{
            background: token.colorBgContainer,
            borderBottom: `1px solid ${token.colorBorderSecondary}`,
            padding: '6px 16px 0',
            lineHeight: 1,
          }}
        >
          <Tabs
            type="editable-card"
            hideAdd
            size="small"
            activeKey={activeKey}
            onChange={handleTabChange}
            onEdit={handleTabEdit}
            style={{ marginBottom: 0 }}
            items={tabs.map((t) => ({
              key: t.key,
              label: t.label,
              closable: tabs.length > 1,
            }))}
          />
        </div>

        {/* Content */}
        <Content style={{ padding: 24, overflow: 'auto' }}>
          <div
            style={{
              background: token.colorBgContainer,
              borderRadius: token.borderRadiusLG,
              padding: 24,
              minHeight: 360,
            }}
          >
            <Outlet />
          </div>
        </Content>
      </Layout>
    </Layout>
  )
}

export default function AppLayout() {
  return (
    <TabsProvider>
      <Inner />
    </TabsProvider>
  )
}
