import { useMemo, useState } from 'react'
import styles from './topology.module.css'
import { Page } from './pages'
import {
  Badge,
  Button,
  Card,
  CardBody,
  CardHeader,
  EmptyState,
  ErrorState,
  Field,
  LoadingState,
} from '../components'
import { useTopology, useWhatIf, type TopoNode, type WhatIfImpact } from '../api/topology'
import { layoutTopology, T_NODE_H, T_NODE_W } from '../viz/topoLayout'

/** TopologyPage (S43, PR1): the tenant's dependency graph — agents, hops,
 * devices, hosts, services, prefixes — with temporal time travel (?at) and
 * the what-if failure simulation. The functional view; PR2+ iterates layout/
 * drill-down/change-overlay polish (design-led, multi-PR). */
export function TopologyPage() {
  const [at, setAt] = useState('') // '' = live
  const { data, isPending, isError } = useTopology(at || undefined)
  const whatIf = useWhatIf()
  const [selected, setSelected] = useState<TopoNode | null>(null)

  const layout = useMemo(() => layoutTopology(data?.nodes ?? [], data?.edges ?? []), [data])
  const impact = whatIf.data ?? null
  const impacted = useMemo(() => impactedNodeIDs(impact), [impact])

  const simulate = (target: string) => {
    whatIf.mutate({ target, at: at || undefined })
  }

  return (
    <Page
      title="Topology"
      subtitle="The dependency graph across planes — and what breaks if an element fails."
    >
      <div className={styles.toolbar}>
        <Field
          label="As of"
          hint="Empty = live; pick a time to view the graph as it was."
          type="datetime-local"
          value={at ? at.slice(0, 16) : ''}
          onChange={(e) => {
            setSelected(null)
            whatIf.reset()
            setAt(e.target.value ? new Date(e.target.value).toISOString() : '')
          }}
        />
        {at !== '' && (
          <Button variant="ghost" onClick={() => setAt('')}>
            Back to live
          </Button>
        )}
      </div>

      {isPending || isError || !data?.topology_running || layout.nodes.length === 0 ? (
        <Card>
          <CardHeader
            title="Dependency graph"
            description="Click a node to inspect it, then simulate its failure."
          />
          <CardBody>
            {isPending ? (
              <LoadingState label="Loading topology…" />
            ) : isError ? (
              <ErrorState description="Could not load the topology graph." />
            ) : !data?.topology_running ? (
              <EmptyState
                icon="path"
                title="Topology not wired"
                description="The control plane started without a topology store."
              />
            ) : (
              <EmptyState
                icon="path"
                title="No topology observed yet"
                description="Run a path discovery, or let eBPF/BGP/device telemetry stream in."
              />
            )}
          </CardBody>
        </Card>
      ) : (
        <div className={styles.grid}>
          <Card className={styles.graphCard}>
            <CardHeader
              title="Dependency graph"
              description="Click a node to inspect it, then simulate its failure."
            />
            <CardBody>
              {(data.coverage?.notes?.length ?? 0) > 0 && (
                <div className={styles.coverage} role="note" aria-label="coverage gaps">
                  {data.coverage?.notes?.map((n) => (
                    <span key={n}>{n}</span>
                  ))}
                </div>
              )}
              {layout.truncated && (
                <p className={styles.truncated}>
                  Showing {layout.nodes.length} of {layout.total} nodes (densest view is capped for
                  legibility).
                </p>
              )}
              <div className={styles.graphWrap}>
                <svg
                  role="group"
                  aria-label="Topology graph"
                  width={layout.width}
                  height={layout.height}
                  viewBox={`0 0 ${layout.width} ${layout.height}`}
                >
                  {layout.edges.map((e) => (
                    <line
                      key={e.id}
                      className={[
                        styles.edge,
                        e.kind === 'flow' ? styles.edgeFlow : '',
                        e.kind === 'routing' ? styles.edgeRouting : '',
                        e.kind === 'device' ? styles.edgeDevice : '',
                        impacted.edges.has(e.id) ? styles.edgeImpacted : '',
                      ]
                        .filter(Boolean)
                        .join(' ')}
                      x1={e.x1}
                      y1={e.y1}
                      x2={e.x2}
                      y2={e.y2}
                    />
                  ))}
                  {layout.nodes.map((n) => (
                    <g
                      key={n.id}
                      role="button"
                      tabIndex={0}
                      aria-label={`${n.kind} ${n.label}`}
                      className={[
                        styles.node,
                        selected?.id === n.id ? styles.nodeSelected : '',
                        impact?.target === n.id ? styles.nodeFailed : '',
                        impacted.nodes.has(n.id) ? styles.nodeImpacted : '',
                      ]
                        .filter(Boolean)
                        .join(' ')}
                      transform={`translate(${n.x}, ${n.y})`}
                      onClick={() => setSelected(n)}
                      onKeyDown={(e) => {
                        if (e.key === 'Enter' || e.key === ' ') {
                          e.preventDefault()
                          setSelected(n)
                        }
                      }}
                    >
                      <rect className={styles.nodeBox} width={T_NODE_W} height={T_NODE_H} rx={8} />
                      <text className={styles.nodeKind} x={10} y={15}>
                        {n.kind}
                      </text>
                      <text className={styles.nodeLabel} x={10} y={30}>
                        {n.label.length > 20 ? `${n.label.slice(0, 19)}…` : n.label}
                      </text>
                    </g>
                  ))}
                </svg>
              </div>
            </CardBody>
          </Card>

          <div className={styles.side}>
            <Card>
              <CardHeader title="Inspector" />
              <CardBody>
                {!selected ? (
                  <EmptyState
                    icon="path"
                    title="No node selected"
                    description="Click a node in the graph."
                  />
                ) : (
                  <>
                    <dl className={styles.detailList}>
                      <dt>Node</dt>
                      <dd>
                        <code>{selected.id}</code>
                      </dd>
                      <dt>Kind</dt>
                      <dd>
                        <Badge tone="info">{selected.kind}</Badge>
                      </dd>
                      <dt>Label</dt>
                      <dd>{selected.label}</dd>
                    </dl>
                    <p>
                      <Button onClick={() => simulate(selected.id)} disabled={whatIf.isPending}>
                        {whatIf.isPending ? 'Simulating…' : 'Simulate failure'}
                      </Button>
                    </p>
                  </>
                )}
              </CardBody>
            </Card>

            {impact && <ImpactCard impact={impact} />}
            {whatIf.isError && (
              <Card>
                <CardBody>
                  <ErrorState description="Simulation failed — the element may not exist at that time." />
                </CardBody>
              </Card>
            )}
          </div>
        </div>
      )}
    </Page>
  )
}

/** ImpactCard renders the what-if prediction: broken/rerouted paths with
 * routes, impacted services/prefixes, and the coverage honesty notes. */
function ImpactCard({ impact }: { impact: WhatIfImpact }) {
  return (
    <Card>
      <CardHeader
        title="Predicted impact"
        description={`If ${impact.target} fails — a simulation, nothing was touched.`}
      />
      <CardBody>
        {(impact.coverage.notes?.length ?? 0) > 0 && (
          <div className={styles.coverage} role="note" aria-label="simulation coverage gaps">
            {impact.coverage.notes?.map((n) => (
              <span key={n}>{n}</span>
            ))}
          </div>
        )}
        <dl className={styles.detailList}>
          <dt>Broken paths</dt>
          <dd>
            {impact.broken_paths.length === 0 ? (
              '0'
            ) : (
              <ul className={styles.impactList} aria-label="broken paths">
                {impact.broken_paths.map((p) => (
                  <li key={`${p.from}-${p.to}`}>
                    <Badge tone="danger">broken</Badge> {p.from} → {p.to}
                  </li>
                ))}
              </ul>
            )}
          </dd>
          <dt>Rerouted</dt>
          <dd>
            {impact.rerouted_paths.length === 0 ? (
              '0'
            ) : (
              <ul className={styles.impactList} aria-label="rerouted paths">
                {impact.rerouted_paths.map((p) => (
                  <li key={`${p.from}-${p.to}`}>
                    <Badge tone="warning">rerouted</Badge> {p.from} → {p.to}
                    <div className={styles.route}>via {p.alt_route?.join(' → ')}</div>
                  </li>
                ))}
              </ul>
            )}
          </dd>
          <dt>Services</dt>
          <dd>{impact.impacted_services.length ? impact.impacted_services.join(', ') : '0'}</dd>
          <dt>Prefixes</dt>
          <dd>{impact.impacted_prefixes.length ? impact.impacted_prefixes.join(', ') : '0'}</dd>
          <dt>Disconnected</dt>
          <dd>{impact.disconnected.length ? impact.disconnected.join(', ') : '0'}</dd>
          <dt>SLOs</dt>
          <dd>{impact.impacted_slos.length ? impact.impacted_slos.join(', ') : '—'}</dd>
        </dl>
      </CardBody>
    </Card>
  )
}

/** impactedNodeIDs derives the overlay sets from a simulation result. */
function impactedNodeIDs(impact: WhatIfImpact | null): {
  nodes: Set<string>
  edges: Set<string>
} {
  const nodes = new Set<string>()
  const edges = new Set<string>()
  if (!impact) return { nodes, edges }
  for (const p of [...impact.broken_paths, ...impact.rerouted_paths]) {
    nodes.add(p.from)
    nodes.add(p.to)
    for (let i = 0; i + 1 < p.route.length; i++) {
      edges.add(`${p.route[i]}|path|${p.route[i + 1]}`)
    }
  }
  for (const s of impact.impacted_services) nodes.add(s)
  for (const s of impact.impacted_prefixes) nodes.add(s)
  for (const s of impact.disconnected) nodes.add(s)
  return { nodes, edges }
}
