import { useState } from "react";
import type {
  AccessLevel,
  CreateRoleRequest,
  Department,
  Employee,
  Role,
  UpdateRoleRequest,
} from "../api/resources";
import {
  ACCESS_LABEL,
  ACCESS_TONE,
  Avatar,
  AvatarStack,
  Btn,
  Card,
  Chip,
  Drawer,
  DrawerSectionLabel,
  Field,
  Icon,
  IconButton,
  PageHeader,
  SortTh,
  colorForId,
  controlStyle,
  sortRows,
  useSort,
} from "../ui";
import type { Person } from "../ui";

const ACCESS_LEVELS: AccessLevel[] = ["mobile", "web_manager", "web_admin"];

function toPeople(employees: Employee[]): Person[] {
  return employees.map((e) => ({ id: e.id, name: e.full_name }));
}

function RoleDetailDrawer({
  role,
  members,
  departmentName,
  onClose,
  onEdit,
  onDelete,
}: {
  role: Role;
  members: Employee[];
  departmentName: string | null;
  onClose: () => void;
  onEdit: () => void;
  onDelete: () => void;
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
            <div style={{ display: "flex", alignItems: "center", gap: 8, marginTop: 3 }}>
              <Chip tone={ACCESS_TONE[role.access_level] ?? "neutral"}>
                {ACCESS_LABEL[role.access_level] ?? role.access_level}
              </Chip>
              {departmentName && <Chip>{departmentName}</Chip>}
            </div>
          </div>
        </>
      }
      footer={
        <>
          <Btn
            variant="secondary"
            icon="x"
            style={{ color: "var(--red-700)", borderColor: "var(--red-200)" }}
            onClick={() => {
              onDelete();
              onClose();
            }}
          >
            Delete
          </Btn>
          <Btn variant="primary" icon="pencil" style={{ marginLeft: "auto" }} onClick={onEdit}>
            Edit Role
          </Btn>
        </>
      }
    >
      {role.description && (
        <p style={{ margin: 0, fontSize: 13.5, lineHeight: 1.55, color: "var(--adaptive-600)" }}>{role.description}</p>
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

function RoleFormDrawer({
  role,
  departments,
  onClose,
  onCreate,
  onUpdate,
}: {
  role: Role | null;
  departments: Department[];
  onClose: () => void;
  onCreate: (body: CreateRoleRequest) => Promise<void>;
  onUpdate: (id: string, body: UpdateRoleRequest) => Promise<void>;
}) {
  const isEdit = Boolean(role);
  const [name, setName] = useState(role?.name ?? "");
  const [description, setDescription] = useState(role?.description ?? "");
  const [departmentId, setDepartmentId] = useState(role?.department_id ?? "");
  const [accessLevel, setAccessLevel] = useState<AccessLevel>(role?.access_level ?? "mobile");
  const [permissions, setPermissions] = useState((role?.permissions ?? []).join(", "));
  const [submitting, setSubmitting] = useState(false);
  const canSubmit = name.trim().length > 0 && !submitting;

  async function submit() {
    if (!canSubmit) return;
    setSubmitting(true);
    try {
      const body = {
        name: name.trim(),
        description: description.trim() || undefined,
        department_id: departmentId || undefined,
        access_level: accessLevel,
        permissions: permissions
          .split(",")
          .map((p) => p.trim())
          .filter((p) => p.length > 0),
      };
      if (role) {
        await onUpdate(role.id, body);
      } else {
        await onCreate(body);
      }
      onClose();
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <Drawer
      onClose={onClose}
      width={420}
      header={<div style={{ fontSize: 16, fontWeight: 600 }}>{isEdit ? "Edit role" : "Add role"}</div>}
      footer={
        <>
          <Btn variant="tertiary" onClick={onClose} style={{ marginRight: "auto" }}>
            Cancel
          </Btn>
          <Btn variant="primary" icon={isEdit ? "check" : "plus"} disabled={!canSubmit} onClick={() => void submit()}>
            {isEdit ? "Save changes" : "Create role"}
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
        <div style={{ display: "flex", gap: 12 }}>
          <div style={{ flex: 1 }}>
            <Field label="Department">
              <select value={departmentId} onChange={(e) => setDepartmentId(e.target.value)} style={controlStyle}>
                <option value="">None</option>
                {departments.map((d) => (
                  <option key={d.id} value={d.id}>
                    {d.name}
                  </option>
                ))}
              </select>
            </Field>
          </div>
          <div style={{ flex: 1 }}>
            <Field label="Access level">
              <select
                value={accessLevel}
                onChange={(e) => setAccessLevel(e.target.value as AccessLevel)}
                style={controlStyle}
              >
                {ACCESS_LEVELS.map((a) => (
                  <option key={a} value={a}>
                    {ACCESS_LABEL[a]}
                  </option>
                ))}
              </select>
            </Field>
          </div>
        </div>
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
  departments,
  onCreate,
  onUpdate,
  onDelete,
}: {
  roles: Role[];
  employees: Employee[];
  departments: Department[];
  onCreate: (body: CreateRoleRequest) => Promise<void>;
  onUpdate: (id: string, body: UpdateRoleRequest) => Promise<void>;
  onDelete: (id: string) => void;
}) {
  const [sel, setSel] = useState<Role | null>(null);
  const [editing, setEditing] = useState<Role | "new" | null>(null);
  const [sort, toggleSort] = useSort("name");
  const membersOf = (r: Role) => employees.filter((e) => e.role_id === r.id);
  const deptName = new Map(departments.map((d) => [d.id, d.name]));

  const sorted = sortRows(roles, sort, {
    name: (r) => r.name,
    department: (r) => (r.department_id ? deptName.get(r.department_id) ?? "" : ""),
    access: (r) => ACCESS_LABEL[r.access_level] ?? r.access_level,
    people: (r) => membersOf(r).length,
    perms: (r) => r.permissions.length,
  });

  return (
    <div style={{ padding: "20px 24px 32px", display: "flex", flexDirection: "column", gap: 18 }}>
      <PageHeader
        title="Roles"
        subtitle={`${roles.length} role${roles.length === 1 ? "" : "s"}`}
        actions={
          <Btn variant="primary" icon="plus" onClick={() => setEditing("new")}>
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
                <tr>
                  {(
                    [
                      ["Role", "name"],
                      ["Department", "department"],
                      ["Access", "access"],
                      ["People", "people"],
                      ["Permissions", "perms"],
                      ["", null],
                    ] as const
                  ).map(([label, key], i) => (
                    <SortTh key={i} label={label} sortKey={key} sort={sort} onSort={toggleSort} align={key === null ? "right" : "left"} />
                  ))}
                </tr>
              </thead>
              <tbody>
                {sorted.map((r, i) => {
                  const members = membersOf(r);
                  return (
                    <tr
                      key={r.id}
                      onClick={() => setSel(r)}
                      style={{
                        cursor: "pointer",
                        borderBottom: i < sorted.length - 1 ? "1px solid var(--adaptive-100)" : "none",
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
                      <td style={{ padding: "12px 16px", color: "var(--adaptive-700)" }}>
                        {r.department_id ? deptName.get(r.department_id) ?? "Unknown" : "—"}
                      </td>
                      <td style={{ padding: "12px 16px" }}>
                        <Chip tone={ACCESS_TONE[r.access_level] ?? "neutral"}>
                          {ACCESS_LABEL[r.access_level] ?? r.access_level}
                        </Chip>
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
                        <div style={{ display: "inline-flex", gap: 6 }}>
                          <IconButton icon="pencil" title="Edit" onClick={() => setEditing(r)} />
                          <IconButton icon="x" title="Delete" tone="danger" onClick={() => onDelete(r.id)} />
                        </div>
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
        <RoleDetailDrawer
          role={sel}
          members={membersOf(sel)}
          departmentName={sel.department_id ? deptName.get(sel.department_id) ?? null : null}
          onClose={() => setSel(null)}
          onEdit={() => {
            setEditing(sel);
            setSel(null);
          }}
          onDelete={() => onDelete(sel.id)}
        />
      )}
      {editing && (
        <RoleFormDrawer
          role={editing === "new" ? null : editing}
          departments={departments}
          onClose={() => setEditing(null)}
          onCreate={onCreate}
          onUpdate={onUpdate}
        />
      )}
    </div>
  );
}
