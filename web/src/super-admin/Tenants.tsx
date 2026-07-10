import { useState } from "react";
import type { PlatformTenant, PlatformUpdateTenantRequest } from "../api/platform";
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
  humanize,
  initials,
  sortRows,
  useSort,
} from "../ui";
import { AuditNote, EmptyState, KV, StatusPill, TableShell, tdStyle, tenantTone } from "./parts";

export function Tenants({
  tenants,
  onUpdate,
  loading,
}: {
  tenants: PlatformTenant[];
  onUpdate: (id: string, body: PlatformUpdateTenantRequest, message: string) => Promise<void>;
  loading: boolean;
}) {
  const [sort, toggleSort] = useSort("name");
  const [query, setQuery] = useState("");
  const [selectedId, setSelectedId] = useState<string | null>(null);

  const filtered = tenants.filter((t) => {
    const q = query.trim().toLowerCase();
    if (!q) return true;
    return t.name.toLowerCase().includes(q) || t.slug.toLowerCase().includes(q);
  });
  const rows = sortRows(filtered, sort, {
    name: (t) => t.name,
    slug: (t) => t.slug,
    status: (t) => t.status,
    plan: (t) => t.plan,
    created_at: (t) => t.created_at,
  });

  const selected = tenants.find((t) => t.id === selectedId) ?? null;

  return (
    <div style={{ padding: 24, display: "flex", flexDirection: "column", gap: 18 }}>
      <PageHeader
        title="Tenants"
        subtitle={`${tenants.length} tenant${tenants.length === 1 ? "" : "s"} in the control-plane registry`}
        actions={
          <div
            style={{
              display: "flex",
              alignItems: "center",
              gap: 8,
              height: 36,
              padding: "0 10px",
              background: "var(--adaptive-100)",
              borderRadius: 6,
              minWidth: 220,
            }}
          >
            <Icon name="search" size={15} color="var(--adaptive-500)" />
            <input
              value={query}
              onChange={(e) => setQuery(e.target.value)}
              placeholder="Search name or slug…"
              style={{
                border: 0,
                background: "transparent",
                fontFamily: "inherit",
                fontSize: 13,
                outline: "none",
                width: "100%",
              }}
            />
          </div>
        }
      />

      <TableShell>
        <thead>
          <tr>
            <SortTh label="Tenant" sortKey="name" sort={sort} onSort={toggleSort} />
            <SortTh label="Slug" sortKey="slug" sort={sort} onSort={toggleSort} />
            <SortTh label="Status" sortKey="status" sort={sort} onSort={toggleSort} />
            <SortTh label="Plan" sortKey="plan" sort={sort} onSort={toggleSort} />
            <SortTh label="Created" sortKey="created_at" sort={sort} onSort={toggleSort} />
            <SortTh label="" sortKey={null} sort={sort} onSort={toggleSort} align="right" />
          </tr>
        </thead>
        <tbody>
          {rows.map((t) => (
            <tr
              key={t.id}
              onClick={() => setSelectedId(t.id)}
              style={{ cursor: "pointer" }}
              onMouseEnter={(e) => (e.currentTarget.style.background = "var(--adaptive-50)")}
              onMouseLeave={(e) => (e.currentTarget.style.background = "transparent")}
            >
              <td style={tdStyle}>
                <div style={{ display: "flex", alignItems: "center", gap: 10 }}>
                  <div
                    style={{
                      width: 28,
                      height: 28,
                      borderRadius: 6,
                      background: "var(--blue-600)",
                      color: "#fff",
                      display: "flex",
                      alignItems: "center",
                      justifyContent: "center",
                      fontSize: 11,
                      fontWeight: 700,
                    }}
                  >
                    {initials(t.name)}
                  </div>
                  <span style={{ fontWeight: 600, color: "var(--adaptive-900)" }}>{t.name}</span>
                </div>
              </td>
              <td style={{ ...tdStyle, color: "var(--adaptive-500)" }}>{t.slug}</td>
              <td style={tdStyle}>
                <StatusPill tone={tenantTone(t.status)} label={humanize(t.status)} />
              </td>
              <td style={tdStyle}>{t.plan}</td>
              <td style={{ ...tdStyle, color: "var(--adaptive-500)" }}>{formatDateTime(t.created_at)}</td>
              <td style={{ ...tdStyle, textAlign: "right" }}>
                <Icon name="chevron" size={16} color="var(--adaptive-400)" />
              </td>
            </tr>
          ))}
        </tbody>
      </TableShell>
      {rows.length === 0 && (
        <EmptyState text={loading ? "Loading tenants…" : "No tenants match your search."} />
      )}

      {selected && (
        <TenantDrawer
          tenant={selected}
          onClose={() => setSelectedId(null)}
          onUpdate={onUpdate}
        />
      )}
    </div>
  );
}

function TenantDrawer({
  tenant,
  onClose,
  onUpdate,
}: {
  tenant: PlatformTenant;
  onClose: () => void;
  onUpdate: (id: string, body: PlatformUpdateTenantRequest, message: string) => Promise<void>;
}) {
  const [plan, setPlan] = useState(tenant.plan);
  const [busy, setBusy] = useState(false);

  async function act(body: PlatformUpdateTenantRequest, message: string) {
    setBusy(true);
    await onUpdate(tenant.id, body, message);
    setBusy(false);
  }

  const isSuspended = tenant.status === "suspended";
  const isProvisioning = tenant.status === "provisioning";

  return (
    <Drawer
      onClose={onClose}
      width={440}
      header={
        <>
          <div
            style={{
              width: 36,
              height: 36,
              borderRadius: 8,
              background: "var(--blue-600)",
              color: "#fff",
              display: "flex",
              alignItems: "center",
              justifyContent: "center",
              fontWeight: 700,
              fontSize: 13,
            }}
          >
            {initials(tenant.name)}
          </div>
          <div style={{ minWidth: 0 }}>
            <div style={{ fontSize: 15, fontWeight: 700, color: "var(--adaptive-900)" }}>{tenant.name}</div>
            <div style={{ fontSize: 12.5, color: "var(--adaptive-500)" }}>{tenant.slug}</div>
          </div>
        </>
      }
    >
      <AuditNote>
        <Icon name="alert" size={15} /> Status and plan changes write a Super Admin audit event.
      </AuditNote>

      <div>
        <DrawerSectionLabel>Registry</DrawerSectionLabel>
        <div style={{ display: "grid", gridTemplateColumns: "1fr 1fr", gap: 14 }}>
          <KV label="Status" value={<StatusPill tone={tenantTone(tenant.status)} label={humanize(tenant.status)} />} />
          <KV label="Plan" value={tenant.plan} />
          <KV label="Database" value={tenant.db_name} />
          <KV label="Tenant ID" value={<span style={{ fontSize: 11.5 }}>{tenant.id}</span>} />
          <KV label="Created" value={formatDateTime(tenant.created_at)} />
          <KV label="Updated" value={formatDateTime(tenant.updated_at)} />
        </div>
      </div>

      <div>
        <DrawerSectionLabel>Lifecycle</DrawerSectionLabel>
        <div style={{ display: "flex", flexDirection: "column", gap: 10 }}>
          {isSuspended ? (
            <Btn
              variant="primary"
              icon="check"
              disabled={busy}
              onClick={() => void act({ status: "active" }, `Reactivated ${tenant.name}.`)}
            >
              Reactivate tenant
            </Btn>
          ) : (
            <Btn
              variant="secondary"
              icon="alert"
              disabled={busy || isProvisioning}
              onClick={() => void act({ status: "suspended" }, `Suspended ${tenant.name}.`)}
              style={{ borderColor: "var(--red-200)", color: "var(--red-700)" }}
            >
              Suspend tenant
            </Btn>
          )}
          {isProvisioning && (
            <div style={{ fontSize: 12.5, color: "var(--adaptive-500)" }}>
              Tenant is still provisioning; suspend is unavailable until it is active.
            </div>
          )}
        </div>
      </div>

      <div>
        <DrawerSectionLabel>Change plan</DrawerSectionLabel>
        <div style={{ display: "flex", gap: 8, alignItems: "flex-end" }}>
          <div style={{ flex: 1 }}>
            <Field label="Plan key">
              <input value={plan} onChange={(e) => setPlan(e.target.value)} style={controlStyle} />
            </Field>
          </div>
          <Btn
            variant="secondary"
            disabled={busy || plan.trim() === "" || plan === tenant.plan}
            onClick={() => void act({ plan: plan.trim() }, `Updated plan for ${tenant.name}.`)}
          >
            Save
          </Btn>
        </div>
      </div>
    </Drawer>
  );
}
