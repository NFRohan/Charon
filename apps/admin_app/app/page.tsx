const modules = [
  {
    name: 'Wallet Ops',
    description: 'Credits, refunds, approval queues, and ledger lookups.',
  },
  {
    name: 'Fleet Ops',
    description: 'Bus registry, durable QR rotation, and live service state.',
  },
  {
    name: 'Schedules',
    description: 'Routes, stops, templates, exceptions, and service instances.',
  },
  {
    name: 'Alerts',
    description: 'Operational alerts, public advisories, and rider disruptions.',
  },
  {
    name: 'Audit',
    description: 'Immutable audit events with append-only investigation notes.',
  },
  {
    name: 'System Ops',
    description: 'Outbox health, DLQ inspection, and import/export job tracking.',
  },
];

export default function HomePage() {
  return (
    <main className="page-shell">
      <section className="hero">
        <div>
          <p className="eyebrow">Charon Control</p>
          <h1>Single-campus transit operations, scaffolded and ready to build.</h1>
          <p className="lede">
            This shell mirrors the specs already locked in the repo: finance,
            fleet, schedules, alerts, and technical operations in one role-aware
            web app.
          </p>
        </div>
        <div className="hero-panel">
          <span>Scaffold status</span>
          <strong>Ready for implementation</strong>
          <p>API contracts and product specs are in the repo root.</p>
        </div>
      </section>

      <section className="module-grid">
        {modules.map((module) => (
          <article className="module-card" key={module.name}>
            <h2>{module.name}</h2>
            <p>{module.description}</p>
          </article>
        ))}
      </section>
    </main>
  );
}
