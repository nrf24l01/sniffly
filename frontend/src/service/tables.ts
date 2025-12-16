import api from '@/service/axios'

export type UnixSeconds = number

export interface DeviceResponse {
  mac: string
  ip: string
  label: string
  hostname: string
}

export interface DeviceTrafficSummary {
  device: DeviceResponse
  stats: {
    up_bytes: number
    down_bytes: number
  }
}

export interface DeviceDomainSummary {
  device: DeviceResponse
  stats: Record<string, number>
}

export interface DeviceCountrySummary {
  device: DeviceResponse
  stats: Record<string, number>
}

export interface DeviceProtoSummary {
  device: DeviceResponse
  stats: Record<string, number>
}

export const tablesService = {
  async traffic(from: UnixSeconds, to: UnixSeconds) {
    const res = await api.get<DeviceTrafficSummary[]>('/tables/traffic', { params: { from, to } })
    return res.data
  },

  async domains(from: UnixSeconds, to: UnixSeconds) {
    const res = await api.get<DeviceDomainSummary[]>('/tables/domains', { params: { from, to } })
    return res.data
  },

  async countries(from: UnixSeconds, to: UnixSeconds) {
    const res = await api.get<DeviceCountrySummary[]>('/tables/countries', { params: { from, to } })
    return res.data
  },

  async protos(from: UnixSeconds, to: UnixSeconds) {
    const res = await api.get<DeviceProtoSummary[]>('/tables/protos', { params: { from, to } })
    return res.data
  }
}
