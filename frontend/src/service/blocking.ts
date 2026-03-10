import api from '@/service/axios'

export type BlockRuleTargetType = 'ip' | 'cidr' | 'sni'

export interface DeviceBlockRuleItem {
  uuid: string
  device_uuid: string
  device_mac: string
  device_ip: string
  device_label: string
  target_type: BlockRuleTargetType
  target_value: string
  enabled: boolean
  created_at: string
  updated_at: string
}

export const blockingService = {
  async list(): Promise<DeviceBlockRuleItem[]> {
    const res = await api.get<DeviceBlockRuleItem[]>('/devices/block-rules')
    return res.data
  },

  async create(deviceId: string, payload: { target_type: BlockRuleTargetType; target_value: string; enabled?: boolean }): Promise<DeviceBlockRuleItem> {
    const res = await api.post<DeviceBlockRuleItem>(`/devices/${deviceId}/block-rules`, payload)
    return res.data
  },

  async update(deviceId: string, ruleId: string, payload: Partial<{ target_type: BlockRuleTargetType; target_value: string; enabled: boolean }>): Promise<DeviceBlockRuleItem> {
    const res = await api.patch<DeviceBlockRuleItem>(`/devices/${deviceId}/block-rules/${ruleId}`, payload)
    return res.data
  },

  async remove(deviceId: string, ruleId: string): Promise<void> {
    await api.delete(`/devices/${deviceId}/block-rules/${ruleId}`)
  }
}