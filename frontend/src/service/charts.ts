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
  req_count?: number
}

export interface TrafficChartResponse {
  stats: DeviceTrafficBucket[]
}

export interface DeviceDomainBucket {
  bucket: number
  domains: Record<string, number>
  req_count: number
}

export interface DomainChartResponse {
  stats: DeviceDomainBucket[]
}

export interface DeviceCountryBucket {
  bucket: number
  countries: Record<string, number>
  companies: Record<string, number>
  req_count: number
}

export interface CountryChartResponse {
  stats: DeviceCountryBucket[]
}

export interface DeviceProtoBucket {
  bucket: number
  protos: Record<string, number>
  req_count: number
}

export interface ProtoChartResponse {
  stats: DeviceProtoBucket[]
}

function buildParams(from: UnixSeconds, to: UnixSeconds, deviceIds?: string[]) {
  const params = new URLSearchParams()
  params.set('from', String(from))
  params.set('to', String(to))
  for (const id of deviceIds ?? []) {
    if (id) params.append('device_id', id)
  }
  return params
}

export const chartsService = {
  async traffic(from: UnixSeconds, to: UnixSeconds, deviceIds?: string[]) {
    const res = await api.get<TrafficChartResponse>('/charts/traffic', { params: buildParams(from, to, deviceIds) })
    return res.data
  },

  async domains(from: UnixSeconds, to: UnixSeconds, deviceIds?: string[]) {
    const res = await api.get<DomainChartResponse>('/charts/domains', { params: buildParams(from, to, deviceIds) })
    return res.data
  },

  async countries(from: UnixSeconds, to: UnixSeconds, deviceIds?: string[]) {
    const res = await api.get<CountryChartResponse>('/charts/countries', { params: buildParams(from, to, deviceIds) })
    return res.data
  },

  async protos(from: UnixSeconds, to: UnixSeconds, deviceIds?: string[]) {
    const res = await api.get<ProtoChartResponse>('/charts/protos', { params: buildParams(from, to, deviceIds) })
    return res.data
  }
}
