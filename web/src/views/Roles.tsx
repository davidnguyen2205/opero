import { useState } from "react";
import type { CreateRoleRequest, Employee, Role } from "../api/resources";
import {
  Avatar,
  AvatarStack,
  Btn,
  Card,
  Drawer,
  DrawerSectionLabel,
  Field,
  Icon,
  PageHeader,
  colorForId,
  controlStyle,
} from "../ui";
import type { Person } from "../ui";

function toPeople(employees: Employee[]): Person[] {
  return employees.map((e) => ({ id: e.id, name: e.full_name }));
}

function RoleDrawer({
  role,
  members,
  onClose,
  onDelete,
}: {
  role: Role;
  members: Employee[];
  onClose: () => void;
  onDelete: (id: string) => void;
}) {
  return (
    <Drawer
      onClose={onClose}
      width={420}
      header={
        <>
          <span style={{ width: 12, height: 12, borderRadius: 4, background: colorForId(role.id), flexShrink: 0 }} />
          <div style={{ minWidth: 0 }}>
            <div style={{ fontSize: 16, fontWeight: 600, color: "var(--adaptive-900)" }}>{role.name}</div>
            <div style={{ fontSize: 12.5, color: "var(--adaptive-500)" }}>
              {members.length} {members.length === 1 ? "person" : "people"} · {role.permissions.length} permission
              {role.permissions.length === 1 ? "" : "s"}
            </div>
          </div>
        </>
      }
      footer={
        <Btn
          variant="secondary"
          icon="x"
          style={{ flex: 1, color: "var(--red-700)", borderColor: "var(--red-200)" }}
          onClick={() => {
            onDelete(role.id);
            onClose();
          }}
        >
          Delete role
        </Btn>
      }
    >
      {role.description && (
        <p style={{ margin: 0, fontSize: 13.5, lineHeight: 1.55, color: "var(--adaptive-600)" }}>
          {role.description}
        </p>
      )}

      <div>
        <DrawerSectionLabel>Permissions</DrawerSectionLabel>
        {role.permissions.length === 0 ? (
          <div style={{ fontSize: 13, color: "var(--adaptive-500)" }}>No permissions granted.</div>
        ) : (
          <div style={{ display: "flex", flexDirection: "column", gap: 2 }}>
            {role.permissions.map((p, i) => (
              <div
                key={i}
                style={{
                  display: "flex",
                  alignItems: "center",
                  gap: 10,
                  padding: "9px 0",
                  borderBottom: i < role.permissions.length - 1 ? "1px solid var(--adaptive-100)" : "none",
                }}
              >
                <span
                  style={{
                    width: 20,
                    height: 20,
                    borderRadius: 5,
                    background: "var(--green-50)",
                    border: "1px solid var(--green-200)",
                    display: "flex",
                    alignItems: "center",
                    justifyContent: "center",
                    flexShrink: 0,
                  }}
                >
                  <Icon name="check" size={13} color="var(--green-600)" />
                </span>
                <span style={{ fontSize: 13.5, color: "var(--adaptive-800)", fontFamily: "var(--font-mono)" }}>{p}</span>
              </div>
            ))}
          </div>
        )}
      </div>

      <div>
        <DrawerSectionLabel>People with this role · {members.length}</DrawerSectionLabel>
        {members.length === 0 ? (
          <div style={{ fontSize: 13, color: "var(--adaptive-500)" }}>No one assigned yet.</div>
        ) : (
          <div style={{ display: "flex", flexDirection: "column", gap: 2 }}>
            {members.map((s) => (
              <div key={s.id} style={{ display: "flex", alignItems: "center", gap: 10, padding: "7px 0" }}>
                <Avatar person={{ id: s.id, name: s.full_name }} size={30} />
                <div style={{ flex: 1, minWidth: 0 }}>
                  <div style={{ fontSize: 13, fontWeight: 600, color: "var(--adaptive-900)" }}>{s.full_name}</div>
                </div>
              </div>
            ))}
          </div>
        )}
      </div>
    </Drawer>
  );
}

function AddRoleDrawer({
  onClose,
  onCreate,
}: {
  onClose: () => void;
  onCreate: (body: CreateRoleRequest) => Promise<void>;
}) {
  const [name, setName] = useState("");
  const [description, setDescription] = useState("");
  const [permissions, setPermissions] = useState("");
  const [submitting, setSubmitting] = useState(false);
  const canSubmit = name.trim().length > 0 && !submitting;

  async function submit() {
    if (!canSubmit) return;
    setSubmitting(true);
    try {
      await onCreate({
        name: name.trim(),
        description: description.trim() || undefined,
        permissions: permissions
          .split(",")
          .map((p) => p.trim())
          .filter((p) => p.length > 0),
      });
      onClose();
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <Drawer
      onClose={onClose}
      width={420}
      header={<div style={{ fontSize: 16, fontWeight: 600 }}>Add role</div>}
      footer={
        <>
          <Btn variant="tertiary" onClick={onClose} style={{ marginRight: "auto" }}>
            Cancel
          </Btn>
          <Btn variant="primary" icon="plus" disabled={!canSubmit} onClick={() => void submit()}>
            Create role
          </Btn>
        </>
      }
    >
      <div style={{ display: "flex", flexDirection: "column", gap: 16 }}>
        <Field label="Name">
          <input value={name} onChange={(e) => setName(e.target.value)} style={controlStyle} required />
        </Field>
        <Field label="Description">
          <textarea
            value={description}
            onChange={(e) => setDescription(e.target.value)}
            style={{ ...controlStyle, minHeight: 80, resize: "vertical" }}
          />
        </Field>
        <Field label="Permissions (comma-separated)">
          <input
            value={permissions}
            onChange={(e) => setPermissions(e.target.value)}
            placeholder="employees.read, shifts.publish"
            style={controlStyle}
          />
        </Field>
      </div>
    </Drawer>
  );
}

export function Roles({
  roles,
  employees,
  onCreate,
  onDelete,
}: {
  roles: Role[];
  employees: Employee[];
  onCreate: (body: CreateRoleRequest) => Promise<void>;
  onDelete: (id: string) => void;
}) {
  const [sel, setSel] = useState<Role | null>(null);
  const [adding, setAdding] = useState(false);
  const membersOf = (r: Role) => employees.filter((e) => e.role_id === r.id);

  return (
    <div style={{ padding: "20px 24px 32px", display: "flex", flexDirection: "column", gap: 18 }}>
      <PageHeader
        title="Roles"
        subtitle={`${roles.length} role${roles.length === 1 ? "" : "s"}`}
        actions={
          <Btn variant="primary" icon="plus" onClick={() => setAdding(true)}>
            Add role
          </Btn>
        }
      />

      {roles.length === 0 ? (
        <Card style={{ padding: 28 }}>
          <div style={{ textAlign: "center", color: "var(--adaptive-500)" }}>No roles yet.</div>
        </Card>
      ) : (
        <Card style={{ overflow: "hidden" }}>
          <div style={{ overflowX: "auto" }}>
            <table style={{ width: "100%", borderCollapse: "collapse", fontSize: 13, minWidth: 720 }}>
              <thead>
                <tr style={{ background: "var(--adaptive-50)", textAlign: "left" }}>
                  {["Role", "People", "Permissions", ""].map((h, i) => (
                    <th
                      key={i}
                      style={{
                        padding: "11px 16px",
                        fontWeight: 600,
                        fontSize: 12,
                        color: "var(--adaptive-500)",
                        borderBottom: "1px solid var(--adaptive-200)",
                        whiteSpace: "nowrap",
                      }}
                    >
                      {h}
                    </th>
                  ))}
                </tr>
              </thead>
              <tbody>
                {roles.map((r, i) => {
                  const members = membersOf(r);
                  return (
                    <tr
                      key={r.id}
                      onClick={() => setSel(r)}
                      style={{
                        cursor: "pointer",
                        borderBottom: i < roles.length - 1 ? "1px solid var(--adaptive-100)" : "none",
                      }}
                      onMouseEnter={(e) => (e.currentTarget.style.background = "var(--adaptive-50)")}
                      onMouseLeave={(e) => (e.currentTarget.style.background = "transparent")}
                    >
                      <td style={{ padding: "12px 16px" }}>
                        <div style={{ display: "flex", alignItems: "center", gap: 10 }}>
                          <span
                            style={{ width: 10, height: 10, borderRadius: 3, background: colorForId(r.id), flexShrink: 0 }}
                          />
                          <div>
                            <div style={{ fontWeight: 600, color: "var(--adaptive-900)" }}>{r.name}</div>
                            {r.description && (
                              <div
                                style={{
                                  fontSize: 11.5,
                                  color: "var(--adaptive-500)",
                                  maxWidth: 320,
                                  whiteSpace: "nowrap",
                                  overflow: "hidden",
                                  textOverflow: "ellipsis",
                                }}
                              >
                                {r.description}
                              </div>
                            )}
                          </div>
                        </div>
                      </td>
                      <td style={{ padding: "12px 16px" }}>
                        {members.length ? (
                          <AvatarStack people={toPeople(members)} size={26} max={4} />
                        ) : (
                          <span style={{ color: "var(--adaptive-400)" }}>—</span>
                        )}
                      </td>
                      <td style={{ padding: "12px 16px", color: "var(--adaptive-500)" }}>
                        {r.permissions.length} granted
                      </td>
                      <td style={{ padding: "12px 16px", textAlign: "right" }}>
                        <Icon name="chevron" size={15} color="var(--adaptive-300)" />
                      </td>
                    </tr>
                  );
                })}
              </tbody>
            </table>
          </div>
        </Card>
      )}

      {sel && (
        <RoleDrawer role={sel} members={membersOf(sel)} onClose={() => setSel(null)} onDelete={onDelete} />
      )}
      {adding && <AddRoleDrawer onClose={() => setAdding(false)} onCreate={onCreate} />}
    </div>
  );
}
