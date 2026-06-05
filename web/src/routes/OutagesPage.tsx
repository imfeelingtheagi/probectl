import styles from './outages.module.css'
import { Page } from './pages'
import {
  Badge,
  Card,
  CardBody,
  CardHeader,
  EmptyState,
  ErrorState,
  LoadingState,
  Table,
  type Column,
} from '../components'
import { useOutages, type FeedHealth, type OutageEvent } from '../api/outages'

/** OutagesPage (S47a): the collective internet-outage view — public outage
 * feeds + the tenant's own vantage points, correlated with affected tests.
 * The coverage notes keep the view honest: this is NOT a global probe fleet. */
export function OutagesPage() {
  const { data, isPending, isError } = useOutages()

  return (
    <Page
      title="Internet outages"
      subtitle="Collective view: public outage signals + your own vantage points, correlated with your affected tests."
    >
      <Card>
        <CardHeader
          title="Collective outage view"
          description="External events from opt-in open-data feeds (IODA, Cloudflare Radar), joined with outages detected from your own synthetic vantage points."
        />
        <CardBody>
          {isPending ? (
            <LoadingState label="Loading outage view…" />
          ) : isError ? (
            <ErrorState description="Could not load the outage view." />
          ) : !data?.outage_running ? (
            <EmptyState
              icon="outage"
              title="Outage view not wired"
              description="The control plane started with the outage engine disabled (PROBECTL_OUTAGE_ENABLED)."
            />
          ) : (
            <>
              {(data.coverage_notes?.length ?? 0) > 0 && (
                <div className={styles.coverage} role="note" aria-label="coverage notes">
                  {data.coverage_notes?.map((n) => <span key={n}>{n}</span>)}
                </div>
              )}
              {(data.events?.length ?? 0) === 0 && (data.vantage_events?.length ?? 0) === 0 ? (
                <EmptyState
                  icon="outage"
                  title="No outage signals"
                  description={
                    data.feeds_enabled
                      ? 'No external events in the window and nothing detected from your vantage points.'
                      : 'External feeds are off (PROBECTL_OUTAGE_FEEDS_ENABLED) and nothing was detected from your vantage points.'
                  }
                />
              ) : (
                <>
                  {(data.events?.length ?? 0) > 0 && (
                    <Table
                      caption="External outage events"
                      columns={eventColumns}
                      rows={data.events ?? []}
                      rowKey={(e) => e.id}
                      empty={<EmptyState icon="outage" title="No external events" description="—" />}
                    />
                  )}
                  {(data.vantage_events?.length ?? 0) > 0 && (
                    <div className={styles.sectionGap}>
                      <Table
                        caption="Vantage-detected outages (your agents)"
                        columns={eventColumns}
                        rows={data.vantage_events ?? []}
                        rowKey={(e) => e.id}
                        empty={<EmptyState icon="outage" title="No vantage detections" description="—" />}
                      />
                    </div>
                  )}
                </>
              )}
            </>
          )}
        </CardBody>
      </Card>

      {data?.outage_running && data.feeds_enabled && (data.feeds?.length ?? 0) > 0 && (
        <Card>
          <CardHeader
            title="Feed health & provenance"
            description="Per-source status and acceptable-use terms — a down feed keeps its last-good events."
          />
          <CardBody>
            <Table
              caption="Outage feeds"
              columns={feedColumns}
              rows={data.feeds ?? []}
              rowKey={(f) => f.name}
              empty={<EmptyState icon="outage" title="No feeds" description="—" />}
            />
          </CardBody>
        </Card>
      )}
    </Page>
  )
}

const eventColumns: Column<OutageEvent>[] = [
  {
    key: 'what',
    header: 'Outage',
    render: (e) => (
      <div>
        <strong>{e.title}</strong>
        <div className={styles.meta}>
          {e.summary ?? ''}
          {e.evidence_url ? (
            <>
              {e.summary ? ' · ' : ''}
              <a href={e.evidence_url} target="_blank" rel="noreferrer">
                evidence
              </a>
            </>
          ) : null}
        </div>
      </div>
    ),
  },
  {
    key: 'source',
    header: 'Source',
    render: (e) => <Badge tone={e.source === 'vantage' ? 'info' : 'neutral'}>{e.source}</Badge>,
  },
  {
    key: 'scope',
    header: 'Scope',
    render: (e) => (
      <div>
        {e.scope.code}
        {e.scope.name ? <div className={styles.scopeName}>{e.scope.name}</div> : null}
      </div>
    ),
  },
  { key: 'severity', header: 'Severity', render: (e) => severityBadge(e.severity) },
  {
    key: 'state',
    header: 'Status',
    render: (e) =>
      e.ongoing ? <Badge tone="warning">ongoing</Badge> : <Badge tone="neutral">ended</Badge>,
  },
  { key: 'start', header: 'Started', render: (e) => new Date(e.start).toLocaleString() },
  {
    key: 'impact',
    header: 'Your impact',
    render: (e) =>
      (e.affected_tests?.length ?? 0) === 0 ? (
        '—'
      ) : (
        <div className={styles.affected}>
          {e.affected_tests?.map((t) => (
            <span key={t.target}>
              {t.canary_type} {t.target} ({t.failures} failures)
            </span>
          ))}
        </div>
      ),
  },
]

const feedColumns: Column<FeedHealth>[] = [
  { key: 'name', header: 'Feed', render: (f) => <strong>{f.name}</strong> },
  {
    key: 'status',
    header: 'Status',
    render: (f) =>
      f.status === 'ok' ? (
        <Badge tone="success">ok</Badge>
      ) : f.status === 'failed' ? (
        <Badge tone="danger">failed</Badge>
      ) : (
        <Badge tone="neutral">pending</Badge>
      ),
  },
  { key: 'events', header: 'Events', render: (f) => f.events },
  {
    key: 'refreshed',
    header: 'Last refresh',
    render: (f) => (f.last_success ? new Date(f.last_success).toLocaleString() : '—'),
  },
  {
    key: 'aup',
    header: 'License / attribution',
    render: (f) => (
      <div>
        {f.license}
        <div className={styles.feedMeta}>
          {f.attribution ?? ''}
          {f.attribution ? ' · ' : ''}commercial use: {f.commercial_use}
        </div>
      </div>
    ),
  },
]

function severityBadge(sev: OutageEvent['severity']) {
  switch (sev) {
    case 'critical':
      return <Badge tone="danger">critical</Badge>
    case 'warning':
      return <Badge tone="warning">warning</Badge>
    default:
      return <Badge tone="neutral">info</Badge>
  }
}
