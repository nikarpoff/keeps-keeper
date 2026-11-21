import { useCallback, useEffect, useRef, useState } from 'react'
import Markdown from 'react-markdown'
import remarkGfm from 'remark-gfm'
import { api, Folder, Note } from './api'

type Status = 'saved' | 'dirty' | 'saving' | 'error'

const statusText: Record<Status, string> = {
  saved: 'Сохранено',
  dirty: 'Изменено',
  saving: 'Сохранение…',
  error: 'Ошибка сохранения',
}

type Props = { id: number; edit: boolean; folders: Folder[]; onChange: () => void }

export default function NoteView({ id, edit, folders, onChange }: Props) {
  const [note, setNote] = useState<Note | null>(null)
  const [error, setError] = useState('')
  const [status, setStatus] = useState<Status>('saved')
  const latest = useRef({ note, status, deleted: false })
  latest.current = { ...latest.current, note, status }

  useEffect(() => {
    api.note(id).then(setNote, (e) => setError(e.message))
  }, [id])

  const save = useCallback(async () => {
    const n = latest.current.note
    if (!n || latest.current.deleted) return
    setStatus('saving')
    try {
      await api.saveNote({ title: n.title.trim() || 'Без названия', body: n.body, folder_id: n.folder_id }, n.id)
      setStatus((s) => (s === 'saving' ? 'saved' : s))
      onChange()
    } catch {
      setStatus('error')
    }
  }, [onChange])

  useEffect(() => {
    if (status !== 'dirty') return
    const timer = setTimeout(save, 800)
    return () => clearTimeout(timer)
  }, [note, status, save])

  useEffect(() => {
    const warn = (e: BeforeUnloadEvent) => latest.current.status !== 'saved' && e.preventDefault()
    addEventListener('beforeunload', warn)
    return () => {
      removeEventListener('beforeunload', warn)
      if (latest.current.status === 'dirty') save()
    }
  }, [save])

  const update = (patch: Partial<Note>) => {
    setNote((n) => n && { ...n, ...patch })
    setStatus('dirty')
  }

  async function done() {
    if (status !== 'saved') await save()
    location.hash = `#/n/${id}`
  }

  async function remove() {
    if (!note || !confirm(`Удалить «${note.title}»?`)) return
    latest.current.deleted = true
    await api.deleteNote(note.id).catch((e) => alert(e.message))
    onChange()
    location.hash = '#/'
  }

  if (error) return <p className="empty">{error}</p>
  if (!note) return null

  return (
    <article className="note">
      <header className="bar">
        <a className="back" href="#/" aria-label="Назад">
          ←
        </a>
        <span className="meta">{edit ? statusText[status] : note.updated_by}</span>
        <div className="actions">
          {edit ? (
            <button className="primary" onClick={done}>Готово</button>
          ) : (
            <a className="button" href={`#/n/${id}/edit`}>Изменить</a>
          )}
          <button className="danger" onClick={remove}>Удалить</button>
        </div>
      </header>
      {edit ? (
        <div className="editor">
          <input className="title-input" value={note.title} onChange={(e) => update({ title: e.target.value })} placeholder="Заголовок" />
          <select value={note.folder_id ?? ''} onChange={(e) => update({ folder_id: e.target.value ? Number(e.target.value) : null })}>
            <option value="">Без папки</option>
            {folders.map((f) => (
              <option key={f.id} value={f.id}>{f.name}</option>
            ))}
          </select>
          <textarea value={note.body} onChange={(e) => update({ body: e.target.value })} placeholder="Текст в Markdown…" autoFocus />
        </div>
      ) : (
        <div className="markdown">
          <h1>{note.title}</h1>
          <Markdown remarkPlugins={[remarkGfm]}>{note.body}</Markdown>
        </div>
      )}
    </article>
  )
}
