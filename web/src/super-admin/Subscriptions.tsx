import { useState } from "react";
import type {
  PlatformSubscription,
  PlatformUpdateSubscriptionRequest,
} from "../api/platform";
import {
  Btn,
  Drawer,
  DrawerSectionLabel,
  Field,
  Icon,
  PageHeader,
  SortTh,
  controlStyle,
  formatDateTime,
  sortRows,
  useSort,
} from "../ui";
import { AuditNote, EmptyState, KV, StatusPill, TableShell, tdStyle } from "./parts";

export function Subscriptions({
  subscriptions,
  onUpdate,
  loading,
}: {
  subscriptions: PlatformSubscription[];
  onUpdate: (
    id: string,
    body: PlatformUpdateSubscriptionRequest,
    message: string,
  ) => Promise<void>;
  loading: boolean;
}) {
  const [sort, toggleSort] = useSort("tenant");
  const [selectedId, setSelectedId] = useState<string | null>(null);
  const rows = sortRows(subscriptions, sort, {
    tenant: (s) => s.tenant_name,
    plan: (s) => s.plan,
    status: (s) => s.status,
    updated_at: (s) => s.updated_at,
  });
  const selected = subscriptions.find((s) => s.id === selectedId) ?? null;

  return (
    <div style={{ padding: 24, display: "flex", flexDirection: "column", gap: 18 }}>
      <PageHeader
        title="Subscriptions"
        subtitle={`${subscriptions.length} subscription${subscriptions.length === 1 ? "" : "s"}`}
      />

      <TableShell>
        <thead>
          <tr>
            <SortTh label="Tenant" sortKey="tenant" sort={sort} onSort={toggleSort} />
            <SortTh label="Plan" sortKey="plan" sort={sort} onSort={toggleSort} />
            <SortTh label="Status" sortKey="status" sort={sort} onSort={toggleSort} />
            <SortTh label="Updated" sortKey="updated_at" sort={sort} onSort={toggleSort} />
            <SortTh label="" sortKey={null} sort={sort} onSort={toggleSort} align="right" />
          </tr>
        </thead>
        <tbody>
          {rows.map((s) => (
            <tr
              key={s.id}
              onClick={() => setSelectedId(s.id)}
              style={{ cursor: "pointer" }}
              onMouseEnter={(e) => (e.currentTarget.style.background = "var(--adaptive-50)")}
              onMouseLeave={(e) => (e.currentTarget.style.background = "transparent")}
            >
              <td style={tdStyle}>
                <div style={{ display: "flex", flexDirection: "column" }}>
                  <span style={{ fontWeight: 600, color: "var(--adaptive-900)" }}>{s.tenant_name}</span>
                  <span style={{ fontSize: 11.5, color: "var(--adaptive-500)" }}>{s.tenant_slug}</span>
                </div>
              </td>
              <td style={tdStyle}>{s.plan}</td>
              <td style={tdStyle}>
                <StatusPill tone={s.status === "active" ? "green" : "gray"} label={s.status} />
              </td>
              <td style={{ ...tdStyle, color: "var(--adaptive-500)" }}>{formatDateTime(s.updated_at)}</td>
              <td style={{ ...tdStyle, textAlign: "right" }}>
                <Icon name="chevron" size={16} color="var(--adaptive-400)" />
              </td>
            </tr>
          ))}
        </tbody>
      </TableShell>
      {rows.length === 0 && (
        <EmptyState text={loading ? "Loading subscriptions…" : "No subscriptions found."} />
      )}

      {selected && (
        <SubscriptionDrawer
          subscription={selected}
          onClose={() => setSelectedId(null)}
          onUpdate={onUpdate}
        />
      )}
    </div>
  );
}

function SubscriptionDrawer({
  subscription,
  onClose,
  onUpdate,
}: {
  subscription: PlatformSubscription;
  onClose: () => void;
  onUpdate: (
    id: string,
    body: PlatformUpdateSubscriptionRequest,
    message: string,
  ) => Promise<void>;
}) {
  const [plan, setPlan] = useState(subscription.plan);
  const [status, setStatus] = useState(subscription.status);
  const [busy, setBusy] = useState(false);

  const dirty = plan.trim() !== subscription.plan || status.trim() !== subscription.status;

  async function save() {
    setBusy(true);
    await onUpdate(
      subscription.id,
      { plan: plan.trim(), status: status.trim() },
      `Updated subscription for ${subscription.tenant_name}.`,
    );
    setBusy(false);
  }

  return (
    <Drawer
      onClose={onClose}
      width={420}
      header={
        <div style={{ minWidth: 0 }}>
          <div style={{ fontSize: 15, fontWeight: 700, color: "var(--adaptive-900)" }}>
            {subscription.tenant_name}
          </div>
          <div style={{ fontSize: 12.5, color: "var(--adaptive-500)" }}>Subscription</div>
        </div>
      }
      footer={
        <>
          <Btn variant="secondary" onClick={onClose} style={{ flex: 1 }}>
            Cancel
          </Btn>
          <Btn variant="primary" disabled={busy || !dirty} onClick={() => void save()} style={{ flex: 1 }}>
            Save changes
          </Btn>
        </>
      }
    >
      <AuditNote>
        <Icon name="alert" size={15} /> Subscription changes write a Super Admin audit event.
      </AuditNote>

      <div>
        <DrawerSectionLabel>Current</DrawerSectionLabel>
        <div style={{ display: "grid", gridTemplateColumns: "1fr 1fr", gap: 14 }}>
          <KV label="Tenant slug" value={subscription.tenant_slug} />
          <KV label="Created" value={formatDateTime(subscription.created_at)} />
        </div>
      </div>

      <div>
        <DrawerSectionLabel>Edit</DrawerSectionLabel>
        <div style={{ display: "flex", flexDirection: "column", gap: 12 }}>
          <Field label="Plan key">
            <input value={plan} onChange={(e) => setPlan(e.target.value)} style={controlStyle} />
          </Field>
          <Field label="Status">
            <input value={status} onChange={(e) => setStatus(e.target.value)} style={controlStyle} />
          </Field>
        </div>
      </div>
    </Drawer>
  );
}
