import { readFileSync, readdirSync } from 'node:fs'
import { join } from 'node:path'
import { describe, expect, test } from 'vitest'

// Enforces the design-system contract (F54-foundation): component styling carries
// NO hardcoded colors — every color must come from a design token (var(--...)).
// This is what makes per-tenant white-label a runtime token override rather than
// a per-screen rewrite. tokens.css is the one place color literals may live.
const COLOR_LITERAL = /#[0-9a-fA-F]{3,8}\b|\b(?:rgb|rgba|hsl|hsla)\(/

function moduleCssFiles(dir: string): string[] {
  const out: string[] = []
  for (const entry of readdirSync(dir, { withFileTypes: true })) {
    const full = join(dir, entry.name)
    if (entry.isDirectory()) {
      out.push(...moduleCssFiles(full))
    } else if (entry.name.endsWith('.module.css')) {
      out.push(full)
    }
  }
  return out
}

describe('design tokens', () => {
  test('no .module.css uses a hardcoded color literal', () => {
    const srcDir = join(process.cwd(), 'src')
    const offenders: string[] = []
    for (const file of moduleCssFiles(srcDir)) {
      const css = readFileSync(file, 'utf8')
      css.split('\n').forEach((line, i) => {
        if (COLOR_LITERAL.test(line)) {
          offenders.push(`${file.split('/src/')[1]}:${i + 1}  ${line.trim()}`)
        }
      })
    }
    expect(offenders, `hardcoded colors must move into tokens.css:\n${offenders.join('\n')}`).toEqual(
      [],
    )
  })
})
