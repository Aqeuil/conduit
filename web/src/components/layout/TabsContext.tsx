import { createContext, useContext, useState, useCallback, useEffect, type ReactNode } from 'react'

export interface TabItem {
  key: string    // pathname, e.g. "/" or "/app/abc123"
  label: string
}

interface TabsCtx {
  tabs: TabItem[]
  activeKey: string
  openTab: (key: string, label: string) => void
  closeTab: (key: string, fallback: () => void) => void
  setActiveKey: (key: string) => void
}

const COOKIE_TABS = 'conduit_tabs'
const COOKIE_ACTIVE = 'conduit_active_tab'
const COOKIE_MAX_AGE = 60 * 60 * 24 * 30 // 30 days

function readCookie(name: string): string | null {
  const match = document.cookie.match(new RegExp('(?:^|; )' + name + '=([^;]*)'))
  return match ? decodeURIComponent(match[1]) : null
}

function writeCookie(name: string, value: string) {
  document.cookie = `${name}=${encodeURIComponent(value)}; path=/; max-age=${COOKIE_MAX_AGE}; SameSite=Lax`
}

function loadTabs(): TabItem[] {
  try {
    const raw = readCookie(COOKIE_TABS)
    if (raw) {
      const parsed = JSON.parse(raw)
      if (Array.isArray(parsed) && parsed.length > 0) return parsed
    }
  } catch { /* ignore */ }
  return [{ key: '/', label: '应用列表' }]
}

function loadActiveKey(tabs: TabItem[]): string {
  const raw = readCookie(COOKIE_ACTIVE)
  if (raw && tabs.some((t) => t.key === raw)) return raw
  return tabs[0].key
}

const Ctx = createContext<TabsCtx>(null!)

export function TabsProvider({ children }: { children: ReactNode }) {
  const [tabs, setTabs] = useState<TabItem[]>(() => loadTabs())
  const [activeKey, setActiveKeyState] = useState<string>(() => {
    const initial = loadTabs()
    return loadActiveKey(initial)
  })

  // Persist whenever tabs or activeKey change
  useEffect(() => {
    writeCookie(COOKIE_TABS, JSON.stringify(tabs))
  }, [tabs])

  useEffect(() => {
    writeCookie(COOKIE_ACTIVE, activeKey)
  }, [activeKey])

  const setActiveKey = useCallback((key: string) => {
    setActiveKeyState(key)
  }, [])

  const openTab = useCallback((key: string, label: string) => {
    setTabs((prev) => {
      if (prev.find((t) => t.key === key)) {
        // Update label in case it changed
        return prev.map((t) => (t.key === key ? { ...t, label } : t))
      }
      return [...prev, { key, label }]
    })
    setActiveKeyState(key)
  }, [])

  const closeTab = useCallback((key: string, fallback: () => void) => {
    setTabs((prev) => {
      const idx = prev.findIndex((t) => t.key === key)
      if (idx === -1) return prev
      const next = prev.filter((t) => t.key !== key)
      if (key === activeKey && next.length > 0) {
        const target = next[Math.min(idx, next.length - 1)]
        setActiveKeyState(target.key)
        fallback()
      }
      return next
    })
  }, [activeKey])

  return (
    <Ctx.Provider value={{ tabs, activeKey, openTab, closeTab, setActiveKey }}>
      {children}
    </Ctx.Provider>
  )
}

export function useTabs() {
  return useContext(Ctx)
}
