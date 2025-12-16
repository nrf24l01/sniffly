import api from '@/service/axios'

export type UnixSeconds = number

export interface DeviceResponse {
  mac: string
  ip: string
  label: string
  hostname: string
}

export interface DeviceTrafficBucket {
  bucket: number
  up_bytes: number
  down_bytes: number
}

export interface DeviceTrafficItem {
  device: DeviceResponse
  stats: DeviceTrafficBucket[]
}

export interface DeviceDomainBucket {
  bucket: number
  domains: Record<string, number>
  req_count: number
}

export interface DeviceDomainItem {
  device: DeviceResponse
  stats: DeviceDomainBucket[]
}

export interface DeviceCountryBucket {
  bucket: number
  countries: string[]
  companies: string[]
  req_count: number
}

export interface DeviceCountryItem {
  device: DeviceResponse
  stats: DeviceCountryBucket[]
}

export interface DeviceProtoBucket {
  bucket: number
  protos: string[]
  companies: string[]
  req_count: number
}

export interface DeviceProtoItem {
  device: DeviceResponse
  stats: DeviceProtoBucket[]
}

export const chartsService = {
  async traffic(from: UnixSeconds, to: UnixSeconds) {
    const res = await api.get<DeviceTrafficItem[]>('/charts/traffic', { params: { from, to } })
    return res.data
  },

  async domains(from: UnixSeconds, to: UnixSeconds) {
    const res = await api.get<DeviceDomainItem[]>('/charts/domains', { params: { from, to } })
    return res.data
  },

  async countries(from: UnixSeconds, to: UnixSeconds) {
    const res = await api.get<DeviceCountryItem[]>('/charts/countries', { params: { from, to } })
    return res.data
  },

  async protos(from: UnixSeconds, to: UnixSeconds) {
    const res = await api.get<DeviceProtoItem[]>('/charts/protos', { params: { from, to } })
    return res.data
  }
}
