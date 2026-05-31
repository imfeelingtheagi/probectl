import styles from './SkipLink.module.css'

/** A keyboard skip-to-content link (WCAG 2.4.1), visible only when focused. */
export function SkipLink() {
  return (
    <a className={styles.skip} href="#main-content">
      Skip to content
    </a>
  )
}
