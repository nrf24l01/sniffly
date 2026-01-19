import api from '@/service/axios'

export type UnixSeconds = number

export interface TrafficTableResponse {
  stats: {
    up_bytes: number
    down_bytes: number
  }
}

export interface DomainTableResponse {
  stats: Record<string, number>
}

export interface CountryTableResponse {
  stats: Record<string, number>
}

export interface ProtoTableResponse {
  stats: Record<string, number>
}

export interface CompanyTableResponse {
  stats: Record<string, number>
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

export const tablesService = {
  async traffic(from: UnixSeconds, to: UnixSeconds, deviceIds?: string[]) {
    const res = await api.get<TrafficTableResponse>('/tables/traffic', { params: buildParams(from, to, deviceIds) })
    return res.data
  },

  async domains(from: UnixSeconds, to: UnixSeconds, deviceIds?: string[]) {
    const res = await api.get<DomainTableResponse>('/tables/domains', { params: buildParams(from, to, deviceIds) })
    return res.data
  },

  async countries(from: UnixSeconds, to: UnixSeconds, deviceIds?: string[]) {
    const res = await api.get<CountryTableResponse>('/tables/countries', { params: buildParams(from, to, deviceIds) })
    return res.data
  },

  async protos(from: UnixSeconds, to: UnixSeconds, deviceIds?: string[]) {
    const res = await api.get<ProtoTableResponse>('/tables/protos', { params: buildParams(from, to, deviceIds) })
    return res.data
  },

  async companies(from: UnixSeconds, to: UnixSeconds, deviceIds?: string[]) {
    const res = await api.get<CompanyTableResponse>('/tables/companies', { params: buildParams(from, to, deviceIds) })
    return res.data
  }
}
