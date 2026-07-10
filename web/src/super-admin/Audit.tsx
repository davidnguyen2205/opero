import { useState } from "react";
import type { ReactNode } from "react";
import type { SuperAdminAuditEvent } from "../api/platform";
import { Btn, Icon, PageHeader, controlStyle, formatDateTime } from "../ui";
import { EmptyState, tdStyle, TableShell } from "./parts";

// The list endpoint returns events newest-first. Filters map to the query
// params the endpoint declares (action substring + limit here).
export function Audit({
  events,
  onFilter,
  loading,
}: {
  events: SuperAdminAuditEvent[];
  onFilter: (filters: { action?: string; limit?: number }) => Promise<void>;
  loading: boolean;
}) {
  const [action, setAction] = useState("");

  return (
    <div style={{ padding: 24, display: "flex", flexDirection: "column", gap: 18 }}>
      <PageHeader
        title="Audit log"
        subtitle="Audited platform actions, newest first"
        actions={
          <div style={{ display: "flex", alignItems: "center", gap: 8 }}>
            <input
              value={action}
              onChange={(e) => setAction(e.target.value)}
              placeholder="Filter by action…"
              style={{ ...controlStyle, width: 200, minHeight: 36 }}
              onKeyDown={(e) => e.key === "Enter" && void onFilter({ action: action.trim() || undefined })}
            />
            <Btn
              variant="secondary"
              icon="filter"
              onClick={() => void onFilter({ action: action.trim() || undefined })}
            >
              Apply
            </Btn>
          </div>
        }
      />

      <TableShell>
        <thead>
          <tr>
            <Th>When</Th>
            <Th>Actor</Th>
            <Th>Action</Th>
            <Th>Target</Th>
            <Th>Tenant</Th>
          </tr>
        </thead>
        <tbody>
          {events.map((e) => (
            <tr key={e.id}>
              <td style={{ ...tdStyle, color: "var(--adaptive-500)", whiteSpace: "nowrap" }}>
                {formatDateTime(e.created_at)}
              </td>
              <td style={tdStyle}>{e.actor_email}</td>
              <td style={tdStyle}>
                <span
                  style={{
                    fontFamily: "var(--font-mono, monospace)",
                    fontSize: 12,
                    fontWeight: 600,
                    padding: "2px 8px",
                    borderRadius: 6,
                    background: "var(--blue-50)",
                    color: "var(--blue-700)",
                    border: "1px solid var(--blue-200)",
                  }}
                >
                  {e.action}
                </span>
              </td>
              <td style={{ ...tdStyle, color: "var(--adaptive-600)" }}>
                {e.target_type}
                {e.target_id ? (
                  <span style={{ color: "var(--adaptive-400)", fontSize: 11.5 }}> · {e.target_id.slice(0, 8)}</span>
                ) : null}
              </td>
              <td style={{ ...tdStyle, color: "var(--adaptive-600)" }}>
                {e.tenant_name ?? (e.tenant_slug ?? "—")}
              </td>
            </tr>
          ))}
        </tbody>
      </TableShell>
      {events.length === 0 && (
        <EmptyState text={loading ? "Loading audit events…" : "No audit events recorded yet."} />
      )}

      <div style={{ display: "flex", alignItems: "center", gap: 8, fontSize: 12.5, color: "var(--adaptive-500)" }}>
        <Icon name="list" size={14} /> Showing {events.length} event{events.length === 1 ? "" : "s"}.
      </div>
    </div>
  );
}

function Th({ children }: { children: ReactNode }) {
  return (
    <th
      style={{
        padding: "11px 16px",
        fontWeight: 600,
        fontSize: 12,
        textAlign: "left",
        color: "var(--adaptive-500)",
        background: "var(--adaptive-50)",
        borderBottom: "1px solid var(--adaptive-200)",
        whiteSpace: "nowrap",
      }}
    >
      {children}
    </th>
  );
}
