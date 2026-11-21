import { FormEvent, useState } from 'react'
import { api } from './api'

export default function Login({ onLogin }: { onLogin: (login: string) => void }) {
  const [login, setLogin] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState('')

  async function submit(e: FormEvent) {
    e.preventDefault()
    try {
      onLogin((await api.login(login, password)).login)
    } catch {
      setError('Неверный логин или пароль')
    }
  }

  return (
    <form className="login" onSubmit={submit}>
      <h1>Заметки</h1>
      <input placeholder="Логин" value={login} onChange={(e) => setLogin(e.target.value)} autoComplete="username" autoFocus />
      <input
        placeholder="Пароль"
        type="password"
        value={password}
        onChange={(e) => setPassword(e.target.value)}
        autoComplete="current-password"
      />
      {error && <p className="error">{error}</p>}
      <button className="primary" disabled={!login || !password}>Войти</button>
    </form>
  )
}
