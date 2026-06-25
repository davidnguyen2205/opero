import { useState } from "react";
import type {
  CreateDepartmentRequest,
  Department,
  Employee,
  UpdateDepartmentRequest,
} from "../api/resources";
import {
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
  ViewToggle,
  colorForId,
  controlStyle,
  humanize,
  sortRows,
  useSort,
} from "../ui";
import type { IconName, Person } from "../ui";

const DEPT_ICONS: IconName[] = ["briefcase", "route", "activity", "users", "grid", "pin"];
function fallbackIcon(id: string): IconName {
  let total = 0;
  for (const c of id) total += c.charCodeAt(0);
  return DEPT_ICONS[total % DEPT_ICONS.length];
}
// Prefer the department's stored icon/color, falling back to a stable derived one.
function deptIcon(d: Department): IconName {
  if (d.icon && (DEPT_ICONS as string[]).includes(d.icon)) return d.icon as IconName;
  return fallbackIcon(d.id);
}
function deptColor(d: Department): string {
  return d.color ?? colorForId(d.id);
}

const DEPT_COLORS = ["#ea580c", "#2563eb", "#7c3aed", "#0d9488", "#db2777", "#d97706", "#15803d", "#9333ea"];

function toPeople(employees: Employee[]): Person[] {
  return employees.map((e) => ({ id: e.id, name: e.full_name }));
}

function DeptDetailDrawer({
  dept,
  members,
  parentName,
  leadName,
  onClose,
  onEdit,
  onDelete,
}: {
  dept: Department;
  members: Employee[];
  parentName: string | null;
  leadName: string | null;
  onClose: () => void;
  onEdit: () => void;
  onDelete: () => void;
}) {
  const color = deptColor(dept);
  return (
    <Drawer
      onClose={onClose}
      width={420}
      header={
        <>
          <div
            style={{
              width: 40,
              height: 40,
              borderRadius: 9,
              background: `color-mix(in srgb, ${color} 12%, transparent)`,
              display: "flex",
              alignItems: "center",
              justifyContent: "center",
              flexShrink: 0,
            }}
          >
            <Icon name={deptIcon(dept)} size={20} color={color} />
          </div>
          <div style={{ minWidth: 0 }}>
            <div style={{ fontSize: 16, fontWeight: 600, color: "var(--adaptive-900)" }}>{dept.name}</div>
            <div style={{ fontSize: 12.5, color: "var(--adaptive-500)" }}>
              {members.length} {members.length === 1 ? "person" : "people"}
              {parentName ? ` · in ${parentName}` : ""}
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
            Edit Department
          </Btn>
        </>
      }
    >
      {dept.description && (
        <p style={{ margin: 0, fontSize: 13.5, lineHeight: 1.55, color: "var(--adaptive-600)" }}>{dept.description}</p>
      )}

      {leadName && (
        <div
          style={{
            display: "flex",
            alignItems: "center",
            gap: 10,
            padding: "11px 14px",
            borderRadius: 8,
            background: "var(--adaptive-50)",
            border: "1px solid var(--adaptive-200)",
          }}
        >
          <Avatar person={{ id: dept.lead_employee_id ?? dept.id, name: leadName }} size={32} />
          <div>
            <div style={{ fontSize: 11.5, color: "var(--adaptive-500)" }}>Department lead</div>
            <div style={{ fontSize: 13.5, fontWeight: 600, color: "var(--adaptive-900)" }}>{leadName}</div>
          </div>
        </div>
      )}

      <div>
        <DrawerSectionLabel>Members · {members.length}</DrawerSectionLabel>
        {members.length === 0 ? (
          <div style={{ fontSize: 13, color: "var(--adaptive-500)" }}>No members assigned yet.</div>
        ) : (
          <div style={{ display: "flex", flexDirection: "column", gap: 2 }}>
            {members.map((s) => (
              <div key={s.id} style={{ display: "flex", alignItems: "center", gap: 10, padding: "7px 0" }}>
                <Avatar person={{ id: s.id, name: s.full_name }} size={30} />
                <div style={{ flex: 1, minWidth: 0 }}>
                  <div style={{ fontSize: 13, fontWeight: 600, color: "var(--adaptive-900)" }}>{s.full_name}</div>
                  {s.title && <div style={{ fontSize: 11.5, color: "var(--adaptive-500)" }}>{s.title}</div>}
                </div>
                <Chip>{humanize(s.employment_type)}</Chip>
              </div>
            ))}
          </div>
        )}
      </div>
    </Drawer>
  );
}

function DeptFormDrawer({
  dept,
  departments,
  employees,
  onClose,
  onCreate,
  onUpdate,
}: {
  dept: Department | null;
  departments: Department[];
  employees: Employee[];
  onClose: () => void;
  onCreate: (body: CreateDepartmentRequest) => Promise<void>;
  onUpdate: (id: string, body: UpdateDepartmentRequest) => Promise<void>;
}) {
  const isEdit = Boolean(dept);
  const [name, setName] = useState(dept?.name ?? "");
  const [parentId, setParentId] = useState(dept?.parent_id ?? "");
  const [description, setDescription] = useState(dept?.description ?? "");
  const [leadId, setLeadId] = useState(dept?.lead_employee_id ?? "");
  const [icon, setIcon] = useState<IconName>(
    dept?.icon && (DEPT_ICONS as string[]).includes(dept.icon)
      ? (dept.icon as IconName)
      : DEPT_ICONS[0],
  );
  const [color, setColor] = useState(dept?.color ?? DEPT_COLORS[0]);
  const [submitting, setSubmitting] = useState(false);
  const canSubmit = name.trim().length > 0 && !submitting;

  async function submit() {
    if (!canSubmit) return;
    setSubmitting(true);
    try {
      const body = {
        name: name.trim(),
        parent_id: parentId || undefined,
        description: description.trim() || undefined,
        lead_employee_id: leadId || undefined,
        icon,
        color,
      };
      if (dept) {
        await onUpdate(dept.id, body);
      } else {
        await onCreate(body);
      }
      onClose();
    } finally {
      setSubmitting(false);
    }
  }

  // A department can't be its own parent.
  const parentOptions = departments.filter((d) => d.id !== dept?.id);

  return (
    <Drawer
      onClose={onClose}
      width={420}
      header={<div style={{ fontSize: 16, fontWeight: 600 }}>{isEdit ? "Edit department" : "Add department"}</div>}
      footer={
        <>
          <Btn variant="tertiary" onClick={onClose} style={{ marginRight: "auto" }}>
            Cancel
          </Btn>
          <Btn variant="primary" icon={isEdit ? "check" : "plus"} disabled={!canSubmit} onClick={() => void submit()}>
            {isEdit ? "Save changes" : "Create department"}
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
            style={{ ...controlStyle, minHeight: 70, resize: "vertical" }}
            placeholder="What this department does…"
          />
        </Field>
        <div style={{ display: "flex", gap: 12 }}>
          <div style={{ flex: 1 }}>
            <Field label="Parent department">
              <select value={parentId} onChange={(e) => setParentId(e.target.value)} style={controlStyle}>
                <option value="">None</option>
                {parentOptions.map((d) => (
                  <option key={d.id} value={d.id}>
                    {d.name}
                  </option>
                ))}
              </select>
            </Field>
          </div>
          <div style={{ flex: 1 }}>
            <Field label="Lead">
              <select value={leadId} onChange={(e) => setLeadId(e.target.value)} style={controlStyle}>
                <option value="">None</option>
                {employees.map((e) => (
                  <option key={e.id} value={e.id}>
                    {e.full_name}
                  </option>
                ))}
              </select>
            </Field>
          </div>
        </div>
        <Field label="Icon">
          <div style={{ display: "flex", gap: 8, flexWrap: "wrap" }}>
            {DEPT_ICONS.map((ic) => {
              const on = ic === icon;
              return (
                <button
                  key={ic}
                  onClick={() => setIcon(ic)}
                  style={{
                    width: 36,
                    height: 36,
                    borderRadius: 8,
                    cursor: "pointer",
                    background: on ? `color-mix(in srgb, ${color} 14%, transparent)` : "var(--card)",
                    border: `1px solid ${on ? color : "var(--adaptive-200)"}`,
                    display: "flex",
                    alignItems: "center",
                    justifyContent: "center",
                  }}
                >
                  <Icon name={ic} size={18} color={on ? color : "var(--adaptive-500)"} />
                </button>
              );
            })}
          </div>
        </Field>
        <Field label="Accent color">
          <div style={{ display: "flex", gap: 8, flexWrap: "wrap" }}>
            {DEPT_COLORS.map((c) => (
              <button
                key={c}
                onClick={() => setColor(c)}
                style={{
                  width: 30,
                  height: 30,
                  borderRadius: 7,
                  background: c,
                  border: "2px solid var(--card)",
                  cursor: "pointer",
                  boxShadow: color === c ? `0 0 0 2px ${c}` : "0 0 0 1px var(--adaptive-200)",
                }}
              />
            ))}
          </div>
        </Field>
      </div>
    </Drawer>
  );
}

function DeptList({
  departments,
  membersOf,
  nameById,
  leadName,
  onSelect,
  onEdit,
  onDelete,
}: {
  departments: Department[];
  membersOf: (d: Department) => Employee[];
  nameById: Map<string, string>;
  leadName: (d: Department) => string | null;
  onSelect: (d: Department) => void;
  onEdit: (d: Department) => void;
  onDelete: (id: string) => void;
}) {
  const [sort, toggleSort] = useSort("name");
  const sorted = sortRows(departments, sort, {
    name: (d) => d.name,
    lead: (d) => leadName(d) ?? "",
    members: (d) => membersOf(d).length,
    parent: (d) => (d.parent_id ? nameById.get(d.parent_id) ?? "" : ""),
  });
  return (
    <Card style={{ overflow: "hidden" }}>
      <div style={{ overflowX: "auto" }}>
        <table style={{ width: "100%", borderCollapse: "collapse", fontSize: 13, minWidth: 560 }}>
          <thead>
            <tr>
              {(
                [
                  ["Department", "name"],
                  ["Lead", "lead"],
                  ["Members", "members"],
                  ["Parent", "parent"],
                  ["", null],
                ] as const
              ).map(([label, key], i) => (
                <SortTh key={i} label={label} sortKey={key} sort={sort} onSort={toggleSort} align={key === null ? "right" : "left"} />
              ))}
            </tr>
          </thead>
          <tbody>
            {sorted.map((dep, i) => {
              const members = membersOf(dep);
              const color = deptColor(dep);
              return (
                <tr
                  key={dep.id}
                  onClick={() => onSelect(dep)}
                  style={{ cursor: "pointer", borderBottom: i < sorted.length - 1 ? "1px solid var(--adaptive-100)" : "none" }}
                  onMouseEnter={(e) => (e.currentTarget.style.background = "var(--adaptive-50)")}
                  onMouseLeave={(e) => (e.currentTarget.style.background = "transparent")}
                >
                  <td style={{ padding: "11px 16px" }}>
                    <div style={{ display: "flex", alignItems: "center", gap: 11 }}>
                      <div
                        style={{
                          width: 32,
                          height: 32,
                          borderRadius: 8,
                          background: `color-mix(in srgb, ${color} 12%, transparent)`,
                          display: "flex",
                          alignItems: "center",
                          justifyContent: "center",
                          flexShrink: 0,
                        }}
                      >
                        <Icon name={deptIcon(dep)} size={17} color={color} />
                      </div>
                      <div>
                        <span style={{ fontWeight: 600, color: "var(--adaptive-900)" }}>{dep.name}</span>
                        {dep.description && (
                          <div
                            style={{
                              fontSize: 11.5,
                              color: "var(--adaptive-500)",
                              maxWidth: 280,
                              whiteSpace: "nowrap",
                              overflow: "hidden",
                              textOverflow: "ellipsis",
                            }}
                          >
                            {dep.description}
                          </div>
                        )}
                      </div>
                    </div>
                  </td>
                  <td style={{ padding: "11px 16px", color: "var(--adaptive-700)" }}>{leadName(dep) ?? "—"}</td>
                  <td style={{ padding: "11px 16px", color: "var(--adaptive-700)" }}>{members.length}</td>
                  <td style={{ padding: "11px 16px", color: "var(--adaptive-700)" }}>
                    {dep.parent_id ? nameById.get(dep.parent_id) ?? "Unknown" : "—"}
                  </td>
                  <td style={{ padding: "11px 16px", textAlign: "right" }}>
                    <div style={{ display: "inline-flex", gap: 6 }}>
                      <IconButton icon="pencil" title="Edit" onClick={() => onEdit(dep)} />
                      <IconButton icon="x" title="Delete" tone="danger" onClick={() => onDelete(dep.id)} />
                    </div>
                  </td>
                </tr>
              );
            })}
          </tbody>
        </table>
      </div>
    </Card>
  );
}

export function Departments({
  departments,
  employees,
  onCreate,
  onUpdate,
  onDelete,
}: {
  departments: Department[];
  employees: Employee[];
  onCreate: (body: CreateDepartmentRequest) => Promise<void>;
  onUpdate: (id: string, body: UpdateDepartmentRequest) => Promise<void>;
  onDelete: (id: string) => void;
}) {
  const [layout, setLayout] = useState<"grid" | "list">("grid");
  const [sel, setSel] = useState<Department | null>(null);
  const [editing, setEditing] = useState<Department | "new" | null>(null);
  const membersOf = (d: Department) => employees.filter((e) => e.department_id === d.id);
  const nameById = new Map(departments.map((d) => [d.id, d.name]));
  const empName = new Map(employees.map((e) => [e.id, e.full_name]));
  const leadName = (d: Department) =>
    d.lead_employee_id ? empName.get(d.lead_employee_id) ?? null : null;

  return (
    <div style={{ padding: "20px 24px 32px", display: "flex", flexDirection: "column", gap: 18 }}>
      <PageHeader
        title="Departments"
        subtitle={`${departments.length} department${departments.length === 1 ? "" : "s"} · ${
          employees.length
        } people`}
        actions={
          <>
            <ViewToggle value={layout} onChange={setLayout} />
            <Btn variant="primary" icon="plus" onClick={() => setEditing("new")}>
              Add department
            </Btn>
          </>
        }
      />

      {departments.length === 0 ? (
        <Card style={{ padding: 28 }}>
          <div style={{ textAlign: "center", color: "var(--adaptive-500)" }}>No departments yet.</div>
        </Card>
      ) : layout === "list" ? (
        <DeptList
          departments={departments}
          membersOf={membersOf}
          nameById={nameById}
          leadName={leadName}
          onSelect={setSel}
          onEdit={(d) => setEditing(d)}
          onDelete={onDelete}
        />
      ) : (
        <div style={{ display: "grid", gridTemplateColumns: "repeat(auto-fill, minmax(330px, 1fr))", gap: 16 }}>
          {departments.map((dep) => {
            const members = membersOf(dep);
            const color = deptColor(dep);
            const lead = leadName(dep);
            return (
              <Card key={dep.id} hover onClick={() => setSel(dep)} style={{ padding: 18, borderTop: `3px solid ${color}` }}>
                <div style={{ display: "flex", alignItems: "flex-start", gap: 12 }}>
                  <div
                    style={{
                      width: 42,
                      height: 42,
                      borderRadius: 10,
                      background: `color-mix(in srgb, ${color} 12%, transparent)`,
                      display: "flex",
                      alignItems: "center",
                      justifyContent: "center",
                      flexShrink: 0,
                    }}
                  >
                    <Icon name={deptIcon(dep)} size={21} color={color} />
                  </div>
                  <div style={{ flex: 1, minWidth: 0 }}>
                    <div style={{ fontSize: 16, fontWeight: 600, color: "var(--adaptive-900)" }}>{dep.name}</div>
                    <div style={{ fontSize: 12.5, color: "var(--adaptive-500)" }}>
                      {members.length} {members.length === 1 ? "person" : "people"}
                      {lead ? ` · Lead ${lead}` : ""}
                    </div>
                  </div>
                  <IconButton icon="pencil" title="Edit" onClick={() => setEditing(dep)} />
                </div>
                {dep.description && (
                  <div
                    style={{
                      marginTop: 10,
                      fontSize: 12.5,
                      color: "var(--adaptive-600)",
                      lineHeight: 1.45,
                      display: "-webkit-box",
                      WebkitLineClamp: 2,
                      WebkitBoxOrient: "vertical",
                      overflow: "hidden",
                    }}
                  >
                    {dep.description}
                  </div>
                )}
                {dep.parent_id && (
                  <div style={{ marginTop: 12 }}>
                    <Chip>in {nameById.get(dep.parent_id) ?? "Unknown"}</Chip>
                  </div>
                )}
                <div
                  style={{
                    display: "flex",
                    alignItems: "center",
                    gap: 10,
                    marginTop: 16,
                    paddingTop: 14,
                    borderTop: "1px solid var(--adaptive-100)",
                    minHeight: 30,
                  }}
                >
                  {members.length ? (
                    <AvatarStack people={toPeople(members)} />
                  ) : (
                    <span style={{ fontSize: 12.5, color: "var(--adaptive-400)" }}>No members yet</span>
                  )}
                </div>
              </Card>
            );
          })}
        </div>
      )}

      {sel && (
        <DeptDetailDrawer
          dept={sel}
          members={membersOf(sel)}
          parentName={sel.parent_id ? nameById.get(sel.parent_id) ?? null : null}
          leadName={leadName(sel)}
          onClose={() => setSel(null)}
          onEdit={() => {
            setEditing(sel);
            setSel(null);
          }}
          onDelete={() => onDelete(sel.id)}
        />
      )}
      {editing && (
        <DeptFormDrawer
          dept={editing === "new" ? null : editing}
          departments={departments}
          employees={employees}
          onClose={() => setEditing(null)}
          onCreate={onCreate}
          onUpdate={onUpdate}
        />
      )}
    </div>
  );
}
