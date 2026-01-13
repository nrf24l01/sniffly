import api from '@/service/axios'

export interface Capture {
  uuid: string
  name: string
  api_key: string
  enabled: boolean
}

export interface CaptureCreatePayload {
  name: string
  enabled: boolean
}

export interface CaptureUpdatePayload {
  name?: string
  enabled?: boolean
}

export const capturesService = {
  async list(): Promise<Capture[]> {
    const res = await api.get<Capture[]>('/capture')
    return res.data
  },

  async create(payload: CaptureCreatePayload): Promise<Capture> {
    const res = await api.post<Capture>('/capture', payload)
    return res.data
  },

  async get(id: string): Promise<Capture> {
    const res = await api.get<Capture>(`/capture/${id}`)
    return res.data
  },

  async update(id: string, payload: CaptureUpdatePayload): Promise<Capture> {
    const res = await api.patch<Capture>(`/capture/${id}`, payload)
    return res.data
  },

  async remove(id: string): Promise<void> {
    await api.delete(`/capture/${id}`)
  },

  async regenerate(id: string): Promise<Capture> {
    const res = await api.post<Capture>(`/capture/${id}/regenerate`)
    return res.data
  }
}
