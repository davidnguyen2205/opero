import type { PlatformSystemHealth } from "../api/platform";
import { Card, Icon, PageHeader, humanize } from "../ui";
import { StatusPill, tenantTone } from "./parts";

export function Health({
  health,
  loading,
}: {
  health: PlatformSystemHealth | null;
  loading: boolean;
}) {
  const byStatus = health?.tenants_by_status ?? {};
  const statuses = Object.keys(byStatus).sort();
  const total = statuses.reduce((sum, k) => sum + byStatus[k], 0);

  return (
    <div style={{ padding: 24, display: "flex", flexDirection: "column", gap: 18 }}>
      <PageHeader title="System health" subtitle="Control-plane status and tenant counts" />

      {loading && !health ? (
        <div style={{ fontSize: 13.5, color: "var(--adaptive-500)" }}>Loading health…</div>
      ) : (
        <>
          <Card style={{ padding: 20, display: "flex", alignItems: "center", gap: 16 }}>
            <div
              style={{
                width: 44,
                height: 44,
                borderRadius: 10,
                background: health?.control_plane === "ok" ? "var(--green-50)" : "var(--adaptive-100)",
                border: `1px solid ${health?.control_plane === "ok" ? "var(--green-200)" : "var(--adaptive-200)"}`,
                display: "flex",
                alignItems: "center",
                justifyContent: "center",
              }}
            >
              <Icon
                name="activity"
                size={22}
                color={health?.control_plane === "ok" ? "var(--green-600)" : "var(--adaptive-500)"}
              />
            </div>
            <div style={{ flex: 1 }}>
              <div style={{ fontSize: 15, fontWeight: 700, color: "var(--adaptive-900)" }}>Control plane</div>
              <div style={{ fontSize: 13, color: "var(--adaptive-500)" }}>
                Registry, auth, and platform metadata
              </div>
            </div>
            <StatusPill
              tone={health?.control_plane === "ok" ? "green" : "gray"}
              label={health ? humanize(health.control_plane) : "unknown"}
            />
          </Card>

          <div>
            <div
              style={{
                fontSize: 12,
                fontWeight: 600,
                letterSpacing: ".05em",
                textTransform: "uppercase",
                color: "var(--adaptive-500)",
                marginBottom: 10,
              }}
            >
              Tenants by status · {total} total
            </div>
            <div
              style={{
                display: "grid",
                gridTemplateColumns: "repeat(auto-fill, minmax(180px, 1fr))",
                gap: 12,
              }}
            >
              {statuses.length === 0 && (
                <div style={{ fontSize: 13.5, color: "var(--adaptive-500)" }}>No tenants recorded.</div>
              )}
              {statuses.map((status) => (
                <Card key={status} style={{ padding: 16 }}>
                  <div style={{ fontSize: 30, fontWeight: 700, letterSpacing: "-0.02em", color: "var(--adaptive-900)" }}>
                    {byStatus[status]}
                  </div>
                  <div style={{ marginTop: 8 }}>
                    <StatusPill tone={tenantTone(status)} label={humanize(status)} />
                  </div>
                </Card>
              ))}
            </div>
          </div>
        </>
      )}
    </div>
  );
}
