import type { ReactNode } from 'react'
import styles from './Table.module.css'

export interface Column<Row> {
  key: string
  header: ReactNode
  render: (row: Row) => ReactNode
  align?: 'start' | 'end'
  numeric?: boolean
}

export interface TableProps<Row> {
  caption: string
  columns: Column<Row>[]
  rows: Row[]
  rowKey: (row: Row) => string
  empty?: ReactNode
}

/** A semantic, accessible data table (the base for the data-dense screens). */
export function Table<Row>({ caption, columns, rows, rowKey, empty }: TableProps<Row>) {
  return (
    <div className={styles.scroll}>
      <table className={styles.table}>
        <caption className="sr-only">{caption}</caption>
        <thead>
          <tr>
            {columns.map((c) => (
              <th
                key={c.key}
                scope="col"
                className={c.align === 'end' || c.numeric ? styles.end : undefined}
              >
                {c.header}
              </th>
            ))}
          </tr>
        </thead>
        <tbody>
          {rows.length === 0 ? (
            <tr>
              <td className={styles.emptyCell} colSpan={columns.length}>
                {empty ?? 'No data.'}
              </td>
            </tr>
          ) : (
            rows.map((row) => (
              <tr key={rowKey(row)}>
                {columns.map((c) => (
                  <td
                    key={c.key}
                    className={[
                      c.align === 'end' || c.numeric ? styles.end : '',
                      c.numeric ? styles.numeric : '',
                    ]
                      .filter(Boolean)
                      .join(' ')}
                  >
                    {c.render(row)}
                  </td>
                ))}
              </tr>
            ))
          )}
        </tbody>
      </table>
    </div>
  )
}
