import api from '@/service/axios'

export interface DeviceListItem {
  uuid: string
  mac: string
  ip: string
  user_label: string
}

export const devicesService = {
  async list(): Promise<DeviceListItem[]> {
    const res = await api.get<DeviceListItem[]>('/devices')
    return res.data
  },

  async updateLabel(id: string, userLabel: string): Promise<DeviceListItem> {
    const res = await api.patch<DeviceListItem>(`/devices/${id}`, { user_label: userLabel })
    return res.data
  }
}
