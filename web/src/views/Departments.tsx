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
function deptIcon(id: string): IconName {
  let total = 0;
  for (const c of id) total += c.charCodeAt(0);
  return DEPT_ICONS[total % DEPT_ICONS.length];
}

function toPeople(employees: Employee[]): Person[] {
  return employees.map((e) => ({ id: e.id, name: e.full_name }));
}

function DeptDetailDrawer({
  dept,
  members,
  parentName,
  onClose,
  onEdit,
  onDelete,
}: {
  dept: Department;
  members: Employee[];
  parentName: string | null;
  onClose: () => void;
  onEdit: () => void;
  onDelete: () => void;
}) {
  const color = colorForId(dept.id);
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
            <Icon name={deptIcon(dept.id)} size={20} color={color} />
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
  onClose,
  onCreate,
  onUpdate,
}: {
  dept: Department | null;
  departments: Department[];
  onClose: () => void;
  onCreate: (body: CreateDepartmentRequest) => Promise<void>;
  onUpdate: (id: string, body: UpdateDepartmentRequest) => Promise<void>;
}) {
  const isEdit = Boolean(dept);
  const [name, setName] = useState(dept?.name ?? "");
  const [parentId, setParentId] = useState(dept?.parent_id ?? "");
  const [submitting, setSubmitting] = useState(false);
  const canSubmit = name.trim().length > 0 && !submitting;

  async function submit() {
    if (!canSubmit) return;
    setSubmitting(true);
    try {
      const body = { name: name.trim(), parent_id: parentId || undefined };
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
    </Drawer>
  );
}

function DeptList({
  departments,
  membersOf,
  nameById,
  onSelect,
  onEdit,
  onDelete,
}: {
  departments: Department[];
  membersOf: (d: Department) => Employee[];
  nameById: Map<string, string>;
  onSelect: (d: Department) => void;
  onEdit: (d: Department) => void;
  onDelete: (id: string) => void;
}) {
  const [sort, toggleSort] = useSort("name");
  const sorted = sortRows(departments, sort, {
    name: (d) => d.name,
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
              const color = colorForId(dep.id);
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
                        <Icon name={deptIcon(dep.id)} size={17} color={color} />
                      </div>
                      <span style={{ fontWeight: 600, color: "var(--adaptive-900)" }}>{dep.name}</span>
                    </div>
                  </td>
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
          onSelect={setSel}
          onEdit={(d) => setEditing(d)}
          onDelete={onDelete}
        />
      ) : (
        <div style={{ display: "grid", gridTemplateColumns: "repeat(auto-fill, minmax(330px, 1fr))", gap: 16 }}>
          {departments.map((dep) => {
            const members = membersOf(dep);
            const color = colorForId(dep.id);
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
                    <Icon name={deptIcon(dep.id)} size={21} color={color} />
                  </div>
                  <div style={{ flex: 1, minWidth: 0 }}>
                    <div style={{ fontSize: 16, fontWeight: 600, color: "var(--adaptive-900)" }}>{dep.name}</div>
                    <div style={{ fontSize: 12.5, color: "var(--adaptive-500)" }}>
                      {members.length} {members.length === 1 ? "person" : "people"}
                    </div>
                  </div>
                  <IconButton icon="pencil" title="Edit" onClick={() => setEditing(dep)} />
                </div>
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
          onClose={() => setEditing(null)}
          onCreate={onCreate}
          onUpdate={onUpdate}
        />
      )}
    </div>
  );
}
