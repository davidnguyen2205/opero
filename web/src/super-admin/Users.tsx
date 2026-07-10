import type { PlatformTenantUser } from "../api/platform";
import {
  Btn,
  PageHeader,
  SortTh,
  formatDateTime,
  humanize,
  sortRows,
  useSort,
} from "../ui";
import { AuditNote, EmptyState, StatusPill, TableShell, tdStyle, userTone } from "./parts";
import { Icon } from "../ui";

export function Users({
  users,
  onSetStatus,
  loading,
}: {
  users: PlatformTenantUser[];
  onSetStatus: (id: string, status: "active" | "disabled", message: string) => Promise<void>;
  loading: boolean;
}) {
  const [sort, toggleSort] = useSort("email");
  const rows = sortRows(users, sort, {
    email: (u) => u.email,
    tenant: (u) => u.tenant_name,
    role: (u) => u.role,
    status: (u) => u.status,
    created_at: (u) => u.created_at,
  });

  return (
    <div style={{ padding: 24, display: "flex", flexDirection: "column", gap: 18 }}>
      <PageHeader
        title="Tenant login users"
        subtitle={`${users.length} login user${users.length === 1 ? "" : "s"} across all tenants`}
      />
      <AuditNote>
        <Icon name="alert" size={15} /> Enabling or disabling a login user writes a Super Admin audit event.
      </AuditNote>

      <TableShell>
        <thead>
          <tr>
            <SortTh label="Email" sortKey="email" sort={sort} onSort={toggleSort} />
            <SortTh label="Tenant" sortKey="tenant" sort={sort} onSort={toggleSort} />
            <SortTh label="Role" sortKey="role" sort={sort} onSort={toggleSort} />
            <SortTh label="Status" sortKey="status" sort={sort} onSort={toggleSort} />
            <SortTh label="Created" sortKey="created_at" sort={sort} onSort={toggleSort} />
            <SortTh label="" sortKey={null} sort={sort} onSort={toggleSort} align="right" />
          </tr>
        </thead>
        <tbody>
          {rows.map((u) => {
            const disabled = u.status === "disabled";
            return (
              <tr key={u.id}>
                <td style={{ ...tdStyle, fontWeight: 600, color: "var(--adaptive-900)" }}>{u.email}</td>
                <td style={tdStyle}>
                  <div style={{ display: "flex", flexDirection: "column" }}>
                    <span>{u.tenant_name}</span>
                    <span style={{ fontSize: 11.5, color: "var(--adaptive-500)" }}>{u.tenant_slug}</span>
                  </div>
                </td>
                <td style={tdStyle}>{humanize(u.role)}</td>
                <td style={tdStyle}>
                  <StatusPill tone={userTone(u.status)} label={humanize(u.status)} />
                </td>
                <td style={{ ...tdStyle, color: "var(--adaptive-500)" }}>{formatDateTime(u.created_at)}</td>
                <td style={{ ...tdStyle, textAlign: "right" }}>
                  {disabled ? (
                    <Btn
                      size="sm"
                      variant="secondary"
                      icon="check"
                      onClick={() => void onSetStatus(u.id, "active", `Enabled ${u.email}.`)}
                    >
                      Enable
                    </Btn>
                  ) : (
                    <Btn
                      size="sm"
                      variant="secondary"
                      onClick={() => void onSetStatus(u.id, "disabled", `Disabled ${u.email}.`)}
                      style={{ borderColor: "var(--red-200)", color: "var(--red-700)" }}
                    >
                      Disable
                    </Btn>
                  )}
                </td>
              </tr>
            );
          })}
        </tbody>
      </TableShell>
      {rows.length === 0 && <EmptyState text={loading ? "Loading users…" : "No login users found."} />}
    </div>
  );
}
