import { post } from './client'
import type { PluginInfo } from '@/types'

export const pluginApi = {
  list: () =>
    post<{ pre_plugins: PluginInfo[]; post_plugins: PluginInfo[] }>(
      '/plugin/base/list',
    ),
}
