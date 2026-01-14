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

export const tablesService = {
  async traffic(from: UnixSeconds, to: UnixSeconds) {
    const res = await api.get<TrafficTableResponse>('/tables/traffic', { params: { from, to } })
    return res.data
  },

  async domains(from: UnixSeconds, to: UnixSeconds) {
    const res = await api.get<DomainTableResponse>('/tables/domains', { params: { from, to } })
    return res.data
  },

  async countries(from: UnixSeconds, to: UnixSeconds) {
    const res = await api.get<CountryTableResponse>('/tables/countries', { params: { from, to } })
    return res.data
  },

  async protos(from: UnixSeconds, to: UnixSeconds) {
    const res = await api.get<ProtoTableResponse>('/tables/protos', { params: { from, to } })
    return res.data
  }
}
