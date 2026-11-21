import { useCallback, useEffect, useState } from 'react'
import { api, Folder, NoteSummary, UNAUTHORIZED } from './api'
import Login from './Login'
import NoteView from './NoteView'
import Sidebar, { FolderFilter } from './Sidebar'

function parseHash() {
  const m = location.hash.match(/^#\/n\/(\d+)(\/edit)?$/)
  return { id: m ? Number(m[1]) : 0, edit: !!m?.[2] }
}

function useRoute() {
  const [route, setRoute] = useState(parseHash)
  useEffect(() => {
    const onChange = () => setRoute(parseHash())
    addEventListener('hashchange', onChange)
    return () => removeEventListener('hashchange', onChange)
  }, [])
  return route
}

export default function App() {
  const [user, setUser] = useState<string | null>()
  const [folders, setFolders] = useState<Folder[]>([])
  const [notes, setNotes] = useState<NoteSummary[]>([])
  const [folder, setFolder] = useState<FolderFilter>('all')
  const route = useRoute()

  useEffect(() => {
    api.me().then((u) => setUser(u.login), () => setUser(null))
    const onUnauthorized = () => setUser(null)
    addEventListener(UNAUTHORIZED, onUnauthorized)
    return () => removeEventListener(UNAUTHORIZED, onUnauthorized)
  }, [])

  const reload = useCallback(() => {
    api.folders().then(setFolders, () => {})
    api.notes().then(setNotes, () => {})
  }, [])

  useEffect(() => {
    if (user) reload()
  }, [user, reload])

  if (user === undefined) return null
  if (user === null) return <Login onLogin={setUser} />

  return (
    <div className={route.id ? 'app has-note' : 'app'}>
      <Sidebar
        user={user}
        folders={folders}
        notes={notes}
        folder={folder}
        setFolder={setFolder}
        activeId={route.id}
        reload={reload}
        onLogout={() => api.logout().then(() => setUser(null))}
      />
      <main className="content">
        {route.id ? (
          <NoteView key={route.id} id={route.id} edit={route.edit} folders={folders} onChange={reload} />
        ) : (
          <p className="empty">Выберите заметку</p>
        )}
      </main>
    </div>
  )
}
