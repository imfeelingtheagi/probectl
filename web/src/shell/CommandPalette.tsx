import { useEffect, useMemo, useRef, useState, useId, type KeyboardEvent } from 'react'
import { createPortal } from 'react-dom'
import { useNavigate } from 'react-router-dom'
import styles from './CommandPalette.module.css'
import { NAV } from '../nav/ia'
import { useTheme } from '../theme/useTheme'
import { useAuth } from '../auth/useAuth'
import { Icon, type IconName } from '../components/Icon'

interface Command {
  id: string
  label: string
  hint: string
  icon: IconName
  run: () => void
}

/**
 * The command palette (⌘K) — the keyboard-first spine of the app. It is a
 * combobox: focus stays in the input, options are tracked with
 * aria-activedescendant, arrows move, Enter runs, Escape closes.
 */
export function CommandPalette({ open, onClose }: { open: boolean; onClose: () => void }) {
  const navigate = useNavigate()
  const { setTheme, themes } = useTheme()
  const { tenants, switchTenant } = useAuth()
  const inputRef = useRef<HTMLInputElement>(null)
  const [query, setQuery] = useState('')
  const [active, setActive] = useState(0)
  const listId = useId()

  const commands = useMemo<Command[]>(() => {
    const go = NAV.map<Command>((n) => ({
      id: `go:${n.to}`,
      label: `Go to ${n.label}`,
      hint: 'Navigate',
      icon: n.icon,
      run: () => navigate(n.to),
    }))
    const theme = themes.map<Command>((t) => ({
      id: `theme:${t}`,
      label: `Theme: ${t}`,
      hint: 'Appearance',
      icon: t === 'aurora' ? 'sun' : 'moon',
      run: () => setTheme(t),
    }))
    const tenant = tenants.map<Command>((t) => ({
      id: `tenant:${t.id}`,
      label: `Switch tenant: ${t.name}`,
      hint: 'Tenant',
      icon: 'targets',
      run: () => switchTenant(t.id),
    }))
    return [...go, ...theme, ...tenant]
  }, [navigate, setTheme, themes, tenants, switchTenant])

  const filtered = useMemo(() => {
    const q = query.trim().toLowerCase()
    return q ? commands.filter((c) => c.label.toLowerCase().includes(q)) : commands
  }, [commands, query])

  useEffect(() => {
    setActive(0)
  }, [query, open])

  useEffect(() => {
    if (!open) return
    const prev = document.activeElement as HTMLElement | null
    inputRef.current?.focus()
    return () => prev?.focus?.()
  }, [open])

  if (!open) return null

  function run(cmd?: Command) {
    if (!cmd) return
    cmd.run()
    setQuery('')
    onClose()
  }

  function onKeyDown(e: KeyboardEvent<HTMLInputElement>) {
    if (e.key === 'Escape') {
      e.preventDefault()
      onClose()
    } else if (e.key === 'ArrowDown') {
      e.preventDefault()
      setActive((i) => Math.min(i + 1, filtered.length - 1))
    } else if (e.key === 'ArrowUp') {
      e.preventDefault()
      setActive((i) => Math.max(i - 1, 0))
    } else if (e.key === 'Enter') {
      e.preventDefault()
      run(filtered[active])
    }
  }

  const activeOptionId = filtered[active] ? `${listId}-${filtered[active].id}` : undefined

  return createPortal(
    <div className={styles.overlay} onMouseDown={onClose}>
      <div
        className={styles.palette}
        role="dialog"
        aria-modal="true"
        aria-label="Command palette"
        onMouseDown={(e) => e.stopPropagation()}
      >
        <div className={styles.search}>
          <Icon name="search" />
          <input
            ref={inputRef}
            className={styles.input}
            type="text"
            placeholder="Search commands…"
            aria-label="Search commands"
            role="combobox"
            aria-expanded={true}
            aria-controls={listId}
            aria-activedescendant={activeOptionId}
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            onKeyDown={onKeyDown}
          />
          <kbd className={styles.kbd}>Esc</kbd>
        </div>
        <ul className={styles.list} role="listbox" id={listId} aria-label="Commands">
          {filtered.length === 0 ? (
            <li className={styles.none}>No matching commands</li>
          ) : (
            filtered.map((c, i) => (
              <li
                key={c.id}
                id={`${listId}-${c.id}`}
                role="option"
                aria-selected={i === active}
                className={[styles.item, i === active ? styles.active : ''].join(' ')}
                onMouseEnter={() => setActive(i)}
                onMouseDown={(e) => {
                  e.preventDefault()
                  run(c)
                }}
              >
                <Icon name={c.icon} />
                <span className={styles.label}>{c.label}</span>
                <span className={styles.hint}>{c.hint}</span>
              </li>
            ))
          )}
        </ul>
      </div>
    </div>,
    document.body,
  )
}
