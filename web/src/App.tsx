import { ConfigProvider } from 'antd'
import zhCN from 'antd/locale/zh_CN'
import { BrowserRouter, Route, Routes } from 'react-router-dom'
import AppLayout from '@/components/layout/AppLayout'
import ApplicationList from '@/pages/ApplicationList'
import AppDetail from '@/pages/AppDetail'
import RouterDetail from '@/pages/RouterDetail'

export default function App() {
  return (
    <ConfigProvider locale={zhCN}>
      <BrowserRouter>
        <Routes>
          <Route element={<AppLayout />}>
            <Route path="/" element={<ApplicationList />} />
            <Route path="/app/:appId" element={<AppDetail />} />
            <Route path="/app/:appId/router/:routerId" element={<RouterDetail />} />
          </Route>
        </Routes>
      </BrowserRouter>
    </ConfigProvider>
  )
}
