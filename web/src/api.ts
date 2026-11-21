export type Folder = { id: number; name: string }
export type NoteSummary = { id: number; title: string; folder_id: number | null; updated_by: string; updated_at: string }
export type Note = NoteSummary & { body: string }
export type NoteInput = Pick<Note, 'title' | 'body' | 'folder_id'>

export const UNAUTHORIZED = 'keeper:unauthorized'

async function request<T>(method: string, path: string, body?: unknown): Promise<T> {
  const res = await fetch(`/api${path}`, {
    method,
    headers: body === undefined ? undefined : { 'Content-Type': 'application/json' },
    body: body === undefined ? undefined : JSON.stringify(body),
  })
  if (res.status === 401 && path !== '/login') dispatchEvent(new Event(UNAUTHORIZED))
  if (!res.ok) throw new Error((await res.json().catch(() => null))?.error ?? res.statusText)
  return res.status === 204 ? (undefined as T) : res.json()
}

const save = (path: string, id: number | undefined, body: unknown) =>
  request<any>(id ? 'PUT' : 'POST', id ? `${path}/${id}` : path, body)

export const api = {
  me: () => request<{ login: string }>('GET', '/me'),
  login: (login: string, password: string) => request<{ login: string }>('POST', '/login', { login, password }),
  logout: () => request<void>('POST', '/logout'),
  folders: () => request<Folder[]>('GET', '/folders'),
  saveFolder: (name: string, id?: number): Promise<Folder> => save('/folders', id, { name }),
  deleteFolder: (id: number) => request<void>('DELETE', `/folders/${id}`),
  notes: () => request<NoteSummary[]>('GET', '/notes'),
  note: (id: number) => request<Note>('GET', `/notes/${id}`),
  saveNote: (note: NoteInput, id?: number): Promise<Note> => save('/notes', id, note),
  deleteNote: (id: number) => request<void>('DELETE', `/notes/${id}`),
}
