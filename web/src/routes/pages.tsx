import type { ReactNode } from 'react'
import { useQuery } from '@tanstack/react-query'
import styles from './pages.module.css'
import { NAV } from '../nav/ia'
import {
  Badge,
  Card,
  CardBody,
  CardHeader,
  ChartShell,
  Column,
  EmptyState,
  ErrorState,
  Icon,
  LoadingState,
  Sparkline,
  StatusDot,
  Table,
} from '../components'

export function Page({
  title,
  subtitle,
  actions,
  children,
}: {
  title: string
  subtitle?: string
  actions?: ReactNode
  children: ReactNode
}) {
  return (
    <div className={styles.page}>
      <header className={styles.header}>
        <div>
          <h1 className={styles.title}>{title}</h1>
          {subtitle ? <p className={styles.subtitle}>{subtitle}</p> : null}
        </div>
        {actions ? <div className={styles.actions}>{actions}</div> : null}
      </header>
      {children}
    </div>
  )
}

/** PlaceholderPage stands in for an IA section until its sprint lands. */
export function PlaceholderPage({ to }: { to: string }) {
  const item = NAV.find((n) => n.to === to)
  const label = item?.label ?? 'Page'
  return (
    <Page title={label} subtitle="This surface is part of the netctl information architecture.">
      <Card>
        <CardBody>
          <EmptyState
            icon={item?.icon}
            title={`${label} lands in a later sprint`}
            description="The app shell, navigation, and design system are ready — feature screens build on this foundation."
          />
        </CardBody>
      </Card>
    </Page>
  )
}

export function NotFoundPage() {
  return (
    <Page title="Not found">
      <Card>
        <CardBody>
          <ErrorState title="404 — page not found" description="That route does not exist." />
        </CardBody>
      </Card>
    </Page>
  )
}

// --- Targets & Tests: a worked example of the data + component patterns. ---

interface TargetRow {
  id: string
  name: string
  type: string
  target: string
  lossRatio: number
  rttMs: number
  up: boolean
}

const MOCK_TARGETS: TargetRow[] = [
  { id: '1', name: 'edge-dns', type: 'dns', target: '1.1.1.1', lossRatio: 0, rttMs: 8.2, up: true },
  { id: '2', name: 'api-gateway', type: 'tcp', target: 'api.example.com:443', lossRatio: 0, rttMs: 21.7, up: true },
  { id: '3', name: 'branch-vpn', type: 'icmp', target: '10.20.0.1', lossRatio: 0.12, rttMs: 64.5, up: true },
  { id: '4', name: 'legacy-billing', type: 'tcp', target: '10.30.4.9:8080', lossRatio: 1, rttMs: 0, up: false },
]

function useTargets() {
  return useQuery({
    queryKey: ['targets'],
    queryFn: async (): Promise<TargetRow[]> => {
      await new Promise((r) => setTimeout(r, 120))
      return MOCK_TARGETS
    },
  })
}

const columns: Column<TargetRow>[] = [
  { key: 'name', header: 'Test', render: (r) => <strong>{r.name}</strong> },
  { key: 'type', header: 'Type', render: (r) => <Badge tone="neutral">{r.type}</Badge> },
  { key: 'target', header: 'Target', render: (r) => <code>{r.target}</code> },
  {
    key: 'loss',
    header: 'Loss',
    numeric: true,
    render: (r) => `${(r.lossRatio * 100).toFixed(0)}%`,
  },
  {
    key: 'rtt',
    header: 'RTT',
    numeric: true,
    render: (r) => (r.up ? `${r.rttMs.toFixed(1)} ms` : '—'),
  },
  {
    key: 'status',
    header: 'Status',
    render: (r) =>
      r.up ? <StatusDot tone="success" label="Online" /> : <StatusDot tone="danger" label="Down" />,
  },
]

export function TargetsPage() {
  const { data, isPending, isError } = useTargets()

  return (
    <Page
      title="Targets & Tests"
      subtitle="Active synthetic tests across your network."
      actions={
        <Badge tone="accent">
          <Icon name="targets" size={14} /> {MOCK_TARGETS.length} tests
        </Badge>
      }
    >
      <div className={styles.statRow}>
        <ChartShell title="Avg RTT (24h)" height={120}>
          <Sparkline label="Average round-trip time, last 24 hours" data={[20, 18, 22, 19, 24, 30, 26, 21, 23, 19, 17, 20]} />
        </ChartShell>
        <ChartShell title="Packet loss (24h)" height={120}>
          <Sparkline label="Packet loss, last 24 hours" data={[0, 0, 0, 1, 0, 0, 3, 8, 2, 0, 0, 0]} />
        </ChartShell>
      </div>

      <Card>
        <CardHeader title="Tests" description="Latest result per test." />
        <CardBody>
          {isPending ? (
            <LoadingState label="Loading tests…" />
          ) : isError ? (
            <ErrorState description="Could not load tests." />
          ) : (
            <Table
              caption="Synthetic tests and their latest results"
              columns={columns}
              rows={data ?? []}
              rowKey={(r) => r.id}
              empty={<EmptyState title="No tests yet" description="Create your first test to begin." />}
            />
          )}
        </CardBody>
      </Card>
    </Page>
  )
}
