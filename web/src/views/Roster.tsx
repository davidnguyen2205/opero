import { useMemo, useState } from "react";
import type { CSSProperties } from "react";
import type {
  CreateShiftRequest,
  Department,
  Employee,
  Location,
  Shift,
  UpdateShiftRequest,
} from "../api/resources";
import {
  Avatar,
  Btn,
  Card,
  Chip,
  Drawer,
  Icon,
  colorForId,
  formatTime,
  humanize,
} from "../ui";

// ── Local date helpers (week grid) ────────────────────────────────────────
function startOfWeek(date: Date): Date {
  const copy = new Date(date);
  const day = copy.getDay();
  const diff = day === 0 ? -6 : 1 - day;
  copy.setDate(copy.getDate() + diff);
  copy.setHours(0, 0, 0, 0);
  return copy;
}

function weekDays(anchor: Date): Date[] {
  const start = startOfWeek(anchor);
  return Array.from({ length: 7 }, (_, i) => {
    const d = new Date(start);
    d.setDate(start.getDate() + i);
    return d;
  });
}

function sameDay(a: Date, b: Date): boolean {
  return (
    a.getFullYear() === b.getFullYear() && a.getMonth() === b.getMonth() && a.getDate() === b.getDate()
  );
}

function toLocalInput(d: Date): string {
  const offset = d.getTimezoneOffset() * 60_000;
  return new Date(d.getTime() - offset).toISOString().slice(0, 16);
}

const FIELD: CSSProperties = {
  fontSize: 13,
  minHeight: 38,
  borderRadius: 6,
  border: "1px solid var(--adaptive-200)",
  padding: "0 10px",
  fontFamily: "inherit",
  width: "100%",
};
const LBL: CSSProperties = {
  fontSize: 12,
  fontWeight: 600,
  color: "var(--adaptive-700)",
  marginBottom: 6,
  display: "block",
};

function ShiftChip({
  shift,
  label,
  color,
  onPublish,
  onEdit,
  onDelete,
}: {
  shift: Shift;
  label: string;
  color: string;
  onPublish: (id: string) => void;
  onEdit: (shift: Shift) => void;
  onDelete: (id: string) => void;
}) {
  const draft = shift.status === "draft";
  return (
    <div
      style={{
        width: "100%",
        padding: "7px 8px",
        borderRadius: 6,
        border: draft ? `1px dashed ${color}` : "1px solid transparent",
        background: draft ? "transparent" : `color-mix(in srgb, ${color} 10%, transparent)`,
        borderLeft: `3px ${draft ? "dashed" : "solid"} ${color}`,
        display: "flex",
        flexDirection: "column",
        gap: 2,
      }}
    >
      <span
        style={{
          fontSize: 12,
          fontWeight: 600,
          color: "var(--adaptive-900)",
          whiteSpace: "nowrap",
          overflow: "hidden",
          textOverflow: "ellipsis",
        }}
      >
        {label}
      </span>
      <span
        style={{
          display: "flex",
          alignItems: "center",
          gap: 5,
          fontSize: 11,
          color: "var(--adaptive-500)",
          fontFeatureSettings: "'tnum'",
        }}
      >
        {formatTime(shift.starts_at)}–{formatTime(shift.ends_at)}
        {draft && (
          <span
            style={{
              marginLeft: "auto",
              fontSize: 9,
              fontWeight: 700,
              color,
              border: `1px solid ${color}`,
              borderRadius: 4,
              padding: "0 4px",
              letterSpacing: ".04em",
            }}
          >
            DRAFT
          </span>
        )}
      </span>
      <div style={{ display: "flex", gap: 8, marginTop: 2 }}>
        {draft && (
          <button
            onClick={() => onPublish(shift.id)}
            style={{
              border: 0,
              background: "transparent",
              color: "var(--adaptive-600)",
              fontSize: 11,
              fontWeight: 700,
              cursor: "pointer",
              fontFamily: "inherit",
              padding: 0,
            }}
          >
            Publish
          </button>
        )}
        <button
          onClick={() => onEdit(shift)}
          style={{
            border: 0,
            background: "transparent",
            color: "var(--adaptive-600)",
            fontSize: 11,
            fontWeight: 700,
            cursor: "pointer",
            fontFamily: "inherit",
            padding: 0,
          }}
        >
          Edit
        </button>
        <button
          onClick={() => onDelete(shift.id)}
          style={{
            border: 0,
            background: "transparent",
            color: "var(--adaptive-600)",
            fontSize: 11,
            fontWeight: 700,
            cursor: "pointer",
            fontFamily: "inherit",
            padding: 0,
          }}
        >
          Delete
        </button>
      </div>
    </div>
  );
}

function ShiftDrawer({
  shift,
  employees,
  locations,
  defaultEmployeeId,
  defaultDay,
  onClose,
  onCreate,
  onUpdate,
}: {
  shift?: Shift;
  employees: Employee[];
  locations: Location[];
  defaultEmployeeId: string;
  defaultDay: Date | null;
  onClose: () => void;
  onCreate: (body: CreateShiftRequest) => Promise<void>;
  onUpdate: (id: string, body: UpdateShiftRequest) => Promise<void>;
}) {
  const isEdit = Boolean(shift);
  const base = defaultDay ? new Date(defaultDay) : new Date();
  base.setHours(10, 0, 0, 0);
  const end = new Date(base);
  end.setHours(base.getHours() + 3);

  const [employeeId, setEmployeeId] = useState(
    shift?.employee_id ?? defaultEmployeeId ?? employees[0]?.id ?? "",
  );
  const [locationId, setLocationId] = useState(shift?.location_id ?? "");
  const [startsAt, setStartsAt] = useState(
    shift ? toLocalInput(new Date(shift.starts_at)) : toLocalInput(base),
  );
  const [endsAt, setEndsAt] = useState(shift ? toLocalInput(new Date(shift.ends_at)) : toLocalInput(end));
  const [notes, setNotes] = useState(shift?.notes ?? "");
  const [submitting, setSubmitting] = useState(false);

  const canSubmit = Boolean(employeeId && startsAt && endsAt) && !submitting;

  async function submit() {
    if (!canSubmit) return;
    setSubmitting(true);
    try {
      const body = {
        employee_id: employeeId,
        location_id: locationId || undefined,
        starts_at: new Date(startsAt).toISOString(),
        ends_at: new Date(endsAt).toISOString(),
        notes: notes.trim() || undefined,
      };
      if (shift) {
        await onUpdate(shift.id, body);
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
      header={<div style={{ fontSize: 16, fontWeight: 600 }}>{isEdit ? "Edit shift" : "Add shift"}</div>}
      footer={
        <>
          <Btn variant="tertiary" onClick={onClose} style={{ marginRight: "auto" }}>
            Cancel
          </Btn>
          <Btn variant="primary" icon={isEdit ? "check" : "plus"} disabled={!canSubmit} onClick={() => void submit()}>
            {isEdit ? "Save changes" : "Add draft shift"}
          </Btn>
        </>
      }
    >
      <div style={{ display: "flex", flexDirection: "column", gap: 16 }}>
        <div>
          <label style={LBL}>Staff member</label>
          <select value={employeeId} onChange={(e) => setEmployeeId(e.target.value)} style={FIELD} required>
            <option value="">Select employee</option>
            {employees.map((emp) => (
              <option key={emp.id} value={emp.id}>
                {emp.full_name}
              </option>
            ))}
          </select>
        </div>
        <div>
          <label style={LBL}>Location / assignment</label>
          <select value={locationId} onChange={(e) => setLocationId(e.target.value)} style={FIELD}>
            <option value="">Unassigned</option>
            {locations.map((loc) => (
              <option key={loc.id} value={loc.id}>
                {loc.name}
              </option>
            ))}
          </select>
        </div>
        <div style={{ display: "flex", gap: 12 }}>
          <div style={{ flex: 1 }}>
            <label style={LBL}>Starts</label>
            <input
              type="datetime-local"
              value={startsAt}
              onChange={(e) => setStartsAt(e.target.value)}
              style={FIELD}
              required
            />
          </div>
          <div style={{ flex: 1 }}>
            <label style={LBL}>Ends</label>
            <input
              type="datetime-local"
              value={endsAt}
              onChange={(e) => setEndsAt(e.target.value)}
              style={FIELD}
              required
            />
          </div>
        </div>
        <div>
          <label style={LBL}>Notes</label>
          <textarea
            value={notes}
            onChange={(e) => setNotes(e.target.value)}
            style={{ ...FIELD, minHeight: 80, padding: "8px 10px", resize: "vertical" }}
          />
        </div>
        {!isEdit && (
          <div
            style={{
              display: "flex",
              alignItems: "center",
              gap: 8,
              padding: "10px 12px",
              background: "var(--primary-50)",
              border: "1px solid var(--primary-200)",
              borderRadius: 7,
              fontSize: 12.5,
              color: "var(--primary-800)",
            }}
          >
            <Icon name="alert" size={15} color="var(--primary-600)" />
            New shifts are added as <strong style={{ fontWeight: 700 }}>drafts</strong> until you publish.
          </div>
        )}
      </div>
    </Drawer>
  );
}

type RosterGroup = { key: string; label: string; employees: Employee[] };

export function Roster({
  employees,
  locations,
  departments,
  shifts,
  locationNames,
  onCreate,
  onUpdate,
  onDelete,
  onPublish,
  onPublishMany,
}: {
  employees: Employee[];
  locations: Location[];
  departments: Department[];
  shifts: Shift[];
  locationNames: Map<string, string>;
  onCreate: (body: CreateShiftRequest) => Promise<void>;
  onUpdate: (id: string, body: UpdateShiftRequest) => Promise<void>;
  onDelete: (id: string) => void;
  onPublish: (id: string) => void;
  onPublishMany: (ids: string[]) => Promise<void>;
}) {
  const [weekOffset, setWeekOffset] = useState(0);
  const [adding, setAdding] = useState<{ employeeId: string; day: Date | null } | null>(null);
  const [editingShift, setEditingShift] = useState<Shift | null>(null);

  const anchor = useMemo(() => {
    const d = new Date();
    d.setDate(d.getDate() + weekOffset * 7);
    return d;
  }, [weekOffset]);
  const days = useMemo(() => weekDays(anchor), [anchor]);
  const today = new Date();

  const activeEmployees = employees.filter((e) => e.status === "active");
  const roster = activeEmployees.length ? activeEmployees : employees;

  // Shifts within the displayed week only.
  const weekShifts = shifts.filter((s) => {
    const t = new Date(s.starts_at);
    return t >= days[0] && t < new Date(days[6].getTime() + 86_400_000);
  });
  const draftIds = weekShifts.filter((s) => s.status === "draft").map((s) => s.id);

  // Group rows by department when departments exist; otherwise one flat group.
  const groups: RosterGroup[] = useMemo(() => {
    if (!departments.length) {
      return [{ key: "all", label: "Staff", employees: roster }];
    }
    const byDept = new Map<string, Employee[]>();
    const unassigned: Employee[] = [];
    for (const emp of roster) {
      if (emp.department_id) {
        const list = byDept.get(emp.department_id) ?? [];
        list.push(emp);
        byDept.set(emp.department_id, list);
      } else {
        unassigned.push(emp);
      }
    }
    const result: RosterGroup[] = departments
      .filter((d) => byDept.has(d.id))
      .map((d) => ({ key: d.id, label: d.name, employees: byDept.get(d.id) ?? [] }));
    if (unassigned.length) {
      result.push({ key: "unassigned", label: "Unassigned", employees: unassigned });
    }
    return result;
  }, [departments, roster]);

  const colTemplate = "200px repeat(7, minmax(96px, 1fr))";
  const weekLabel = `${new Intl.DateTimeFormat(undefined, { day: "2-digit", month: "short" }).format(
    days[0],
  )} – ${new Intl.DateTimeFormat(undefined, {
    day: "2-digit",
    month: "short",
    year: "numeric",
  }).format(days[6])}`;

  function shiftLabel(s: Shift): string {
    return s.location_id ? locationNames.get(s.location_id) ?? "Assigned shift" : "Assigned shift";
  }

  return (
    <div style={{ padding: "20px 24px 32px", display: "flex", flexDirection: "column", gap: 16 }}>
      <div style={{ display: "flex", alignItems: "flex-end", gap: 16, flexWrap: "wrap" }}>
        <div style={{ flex: 1, minWidth: 220 }}>
          <h1
            style={{
              margin: 0,
              fontSize: 24,
              fontWeight: 700,
              letterSpacing: "-0.02em",
              color: "var(--adaptive-900)",
            }}
          >
            Roster
          </h1>
          <div style={{ display: "flex", alignItems: "center", gap: 10, marginTop: 6 }}>
            <button
              onClick={() => setWeekOffset((w) => w - 1)}
              style={{
                width: 30,
                height: 30,
                borderRadius: 6,
                border: "1px solid var(--adaptive-200)",
                background: "var(--card)",
                cursor: "pointer",
                display: "flex",
                alignItems: "center",
                justifyContent: "center",
              }}
            >
              <Icon name="chevron" size={15} color="var(--adaptive-600)" style={{ transform: "rotate(180deg)" }} />
            </button>
            <span style={{ fontSize: 14, fontWeight: 600, color: "var(--adaptive-800)", fontFeatureSettings: "'tnum'" }}>
              Week of {weekLabel}
            </span>
            <button
              onClick={() => setWeekOffset((w) => w + 1)}
              style={{
                width: 30,
                height: 30,
                borderRadius: 6,
                border: "1px solid var(--adaptive-200)",
                background: "var(--card)",
                cursor: "pointer",
                display: "flex",
                alignItems: "center",
                justifyContent: "center",
              }}
            >
              <Icon name="chevron" size={15} color="var(--adaptive-600)" />
            </button>
            {weekOffset !== 0 && (
              <button
                onClick={() => setWeekOffset(0)}
                style={{
                  border: 0,
                  background: "none",
                  color: "var(--primary-600)",
                  fontWeight: 600,
                  cursor: "pointer",
                  fontFamily: "inherit",
                  fontSize: 13,
                }}
              >
                Today
              </button>
            )}
          </div>
        </div>
        <div style={{ display: "flex", alignItems: "center", gap: 10 }}>
          {draftIds.length > 0 && <Chip tone="orange">{draftIds.length} unpublished</Chip>}
          <Btn variant="secondary" icon="plus" onClick={() => setAdding({ employeeId: "", day: null })}>
            Add shift
          </Btn>
          <Btn
            variant="primary"
            icon="send"
            disabled={draftIds.length === 0}
            onClick={() => void onPublishMany(draftIds)}
          >
            {draftIds.length > 0 ? `Publish ${draftIds.length}` : "Published"}
          </Btn>
        </div>
      </div>

      {roster.length === 0 ? (
        <Card style={{ padding: 28 }}>
          <div style={{ textAlign: "center", color: "var(--adaptive-500)" }}>
            Create employees before building the roster.
          </div>
        </Card>
      ) : (
        <Card style={{ overflow: "hidden" }}>
          <div style={{ overflowX: "auto" }}>
            <div style={{ minWidth: 920 }}>
              {/* header row */}
              <div
                style={{
                  display: "grid",
                  gridTemplateColumns: colTemplate,
                  borderBottom: "1px solid var(--adaptive-200)",
                  background: "var(--adaptive-50)",
                }}
              >
                <div style={{ padding: "10px 16px", fontSize: 12, fontWeight: 600, color: "var(--adaptive-500)" }}>
                  Staff · {roster.length}
                </div>
                {days.map((d, i) => {
                  const isToday = sameDay(d, today);
                  return (
                    <div
                      key={i}
                      style={{
                        padding: "8px 10px",
                        textAlign: "center",
                        borderLeft: "1px solid var(--adaptive-100)",
                        background: isToday ? "var(--primary-50)" : "transparent",
                      }}
                    >
                      <div
                        style={{
                          fontSize: 11,
                          color: isToday ? "var(--primary-600)" : "var(--adaptive-500)",
                          fontWeight: 600,
                          textTransform: "uppercase",
                          letterSpacing: ".04em",
                        }}
                      >
                        {new Intl.DateTimeFormat(undefined, { weekday: "short" }).format(d)}
                      </div>
                      <div
                        style={{
                          fontSize: 16,
                          fontWeight: 700,
                          color: isToday ? "var(--primary-700)" : "var(--adaptive-800)",
                          fontFeatureSettings: "'tnum'",
                        }}
                      >
                        {new Intl.DateTimeFormat(undefined, { day: "2-digit" }).format(d)}
                      </div>
                    </div>
                  );
                })}
              </div>

              {groups.map((group) => (
                <div key={group.key}>
                  <div
                    style={{
                      padding: "7px 16px",
                      fontSize: 11,
                      fontWeight: 700,
                      letterSpacing: ".06em",
                      textTransform: "uppercase",
                      color: "var(--adaptive-500)",
                      background: "var(--adaptive-50)",
                      borderBottom: "1px solid var(--adaptive-100)",
                    }}
                  >
                    {group.label} · {group.employees.length}
                  </div>
                  {group.employees.map((emp) => {
                    const empShifts = weekShifts.filter((s) => s.employee_id === emp.id);
                    return (
                      <div
                        key={emp.id}
                        style={{
                          display: "grid",
                          gridTemplateColumns: colTemplate,
                          borderBottom: "1px solid var(--adaptive-100)",
                          minHeight: 68,
                        }}
                      >
                        <div style={{ display: "flex", alignItems: "center", gap: 9, padding: "8px 16px", background: "var(--card)" }}>
                          <Avatar person={{ id: emp.id, name: emp.full_name }} size={30} />
                          <div style={{ minWidth: 0 }}>
                            <div
                              style={{
                                fontSize: 13,
                                fontWeight: 600,
                                color: "var(--adaptive-900)",
                                whiteSpace: "nowrap",
                                overflow: "hidden",
                                textOverflow: "ellipsis",
                              }}
                            >
                              {emp.full_name}
                            </div>
                            <div style={{ fontSize: 11, color: "var(--adaptive-500)" }}>
                              {humanize(emp.employment_type)}
                            </div>
                          </div>
                        </div>
                        {days.map((d, di) => {
                          const isToday = sameDay(d, today);
                          const dayShifts = empShifts.filter((s) => sameDay(new Date(s.starts_at), d));
                          return (
                            <div
                              key={di}
                              style={{
                                borderLeft: "1px solid var(--adaptive-100)",
                                padding: 5,
                                background: isToday ? "rgba(234,88,12,0.03)" : "transparent",
                                display: "flex",
                                flexDirection: "column",
                                gap: 4,
                              }}
                            >
                              {dayShifts.length ? (
                                dayShifts.map((s) => (
                                  <ShiftChip
                                    key={s.id}
                                    shift={s}
                                    label={shiftLabel(s)}
                                    color={colorForId(s.location_id ?? s.id)}
                                    onPublish={onPublish}
                                    onEdit={setEditingShift}
                                    onDelete={onDelete}
                                  />
                                ))
                              ) : (
                                <button
                                  onClick={() => setAdding({ employeeId: emp.id, day: d })}
                                  className="opero-emptycell"
                                  style={{
                                    width: "100%",
                                    flex: 1,
                                    minHeight: 44,
                                    border: 0,
                                    background: "transparent",
                                    borderRadius: 6,
                                    cursor: "pointer",
                                    color: "var(--adaptive-300)",
                                    display: "flex",
                                    alignItems: "center",
                                    justifyContent: "center",
                                    fontFamily: "inherit",
                                  }}
                                >
                                  <Icon name="plus" size={15} />
                                </button>
                              )}
                            </div>
                          );
                        })}
                      </div>
                    );
                  })}
                </div>
              ))}
            </div>
          </div>
        </Card>
      )}

      {locations.length > 0 && (
        <div
          style={{
            display: "flex",
            alignItems: "center",
            gap: 16,
            flexWrap: "wrap",
            fontSize: 12,
            color: "var(--adaptive-500)",
          }}
        >
          <span style={{ fontWeight: 600 }}>Locations:</span>
          {locations.slice(0, 8).map((loc) => (
            <span key={loc.id} style={{ display: "inline-flex", alignItems: "center", gap: 6 }}>
              <span style={{ width: 9, height: 9, borderRadius: 3, background: colorForId(loc.id) }} />
              {loc.name}
            </span>
          ))}
        </div>
      )}

      {(adding || editingShift) && (
        <ShiftDrawer
          shift={editingShift ?? undefined}
          employees={roster}
          locations={locations}
          defaultEmployeeId={adding?.employeeId ?? ""}
          defaultDay={adding?.day ?? null}
          onClose={() => {
            setAdding(null);
            setEditingShift(null);
          }}
          onCreate={onCreate}
          onUpdate={onUpdate}
        />
      )}
    </div>
  );
}
