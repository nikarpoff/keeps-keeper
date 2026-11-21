import { useState } from 'react'
import { api, Folder, NoteSummary } from './api'

export type FolderFilter = 'all' | 'none' | number

type Props = {
  user: string
  folders: Folder[]
  notes: NoteSummary[]
  folder: FolderFilter
  setFolder: (f: FolderFilter) => void
  activeId: number
  reload: () => void
  onLogout: () => void
}

const formatDate = (s: string) =>
  new Date(s).toLocaleString('ru', { day: 'numeric', month: 'short', hour: '2-digit', minute: '2-digit' })

const attempt = (fn: () => Promise<unknown>) => () => fn().catch((e) => alert(e.message))

export default function Sidebar({ user, folders, notes, folder, setFolder, activeId, reload, onLogout }: Props) {
  const [query, setQuery] = useState('')
  const current = folders.find((f) => f.id === folder)
  const visible = notes.filter(
    (n) =>
      (folder === 'all' || n.folder_id === (folder === 'none' ? null : folder)) &&
      n.title.toLowerCase().includes(query.toLowerCase()),
  )

  const addFolder = attempt(async () => {
    const name = prompt('Название папки')?.trim()
    if (!name) return
    setFolder((await api.saveFolder(name)).id)
    reload()
  })

  const renameFolder = attempt(async () => {
    const name = current && prompt('Новое название', current.name)?.trim()
    if (!current || !name) return
    await api.saveFolder(name, current.id)
    reload()
  })

  const removeFolder = attempt(async () => {
    if (!current || !confirm(`Удалить папку «${current.name}»? Заметки останутся без папки.`)) return
    await api.deleteFolder(current.id)
    setFolder('all')
    reload()
  })

  const addNote = attempt(async () => {
    const note = await api.saveNote({ title: 'Без названия', body: '', folder_id: current?.id ?? null })
    reload()
    location.hash = `#/n/${note.id}/edit`
  })

  const chip = (value: FolderFilter, label: string) => (
    <button key={label + value} className={folder === value ? 'chip active' : 'chip'} onClick={() => setFolder(value)}>
      {label}
    </button>
  )

  return (
    <aside className="sidebar">
      <header className="bar">
        <h1>Заметки</h1>
        <button className="link" onClick={onLogout}>
          {user} · выйти
        </button>
      </header>
      <nav className="chips">
        {chip('all', 'Все')}
        {folders.map((f) => chip(f.id, f.name))}
        {chip('none', 'Без папки')}
        <button className="chip" onClick={addFolder} title="Новая папка">
          +
        </button>
      </nav>
      {current && (
        <div className="folder-actions">
          <button className="link" onClick={renameFolder}>Переименовать</button>
          <button className="link danger" onClick={removeFolder}>Удалить папку</button>
        </div>
      )}
      <input className="search" type="search" placeholder="Поиск" value={query} onChange={(e) => setQuery(e.target.value)} />
      <ul className="list">
        {visible.map((n) => (
          <li key={n.id}>
            <a href={`#/n/${n.id}`} className={n.id === activeId ? 'active' : undefined}>
              <span className="title">{n.title}</span>
              <span className="meta">
                {formatDate(n.updated_at)} · {n.updated_by}
              </span>
            </a>
          </li>
        ))}
        {!visible.length && <li className="empty">Пусто</li>}
      </ul>
      <button className="fab" onClick={addNote} aria-label="Новая заметка">
        +
      </button>
    </aside>
  )
}
