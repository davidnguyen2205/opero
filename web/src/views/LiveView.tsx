import { useState } from "react";
import type { LeaveRequest, LiveViewEntry } from "../api/resources";
import {
  Avatar,
  Btn,
  Card,
  Drawer,
  DrawerSectionLabel,
  Icon,
  STATUS,
  StatusChip,
  fmtDur,
  formatTime,
  initials,
  leaveCoversDay,
  minutesBetween,
} from "../ui";
import type { LiveStatus } from "../ui";

// The API exposes not_checked_in / checked_in / on_break / checked_out. We
// derive the richer board statuses from that plus the shift's scheduled start.
function deriveStatus(entry: LiveViewEntry, nowMs: number): LiveStatus {
  switch (entry.attendance_status) {
    case "checked_in":
      return "working";
    case "on_break":
      return "break";
    case "checked_out":
      return "done";
    default:
      return new Date(entry.shift.starts_at).getTime() > nowMs ? "upcoming" : "late";
  }
}

type Row = LiveViewEntry & { _status: LiveStatus };

type Layout = "board" | "list" | "map";

function Seg({
  value,
  onChange,
}: {
  value: Layout;
  onChange: (v: Layout) => void;
}) {
  const options: { id: Layout; label: string; icon: "grid" | "list" | "map" }[] = [
    { id: "board", label: "Board", icon: "grid" },
    { id: "list", label: "List", icon: "list" },
    { id: "map", label: "Map", icon: "map" },
  ];
  return (
    <div
      style={{
        display: "inline-flex",
        background: "var(--adaptive-100)",
        borderRadius: 7,
        padding: 3,
        gap: 2,
      }}
    >
      {options.map((o) => {
        const on = o.id === value;
        return (
          <button
            key={o.id}
            onClick={() => onChange(o.id)}
            style={{
              display: "inline-flex",
              alignItems: "center",
              gap: 6,
              padding: "5px 11px",
              borderRadius: 5,
              border: 0,
              cursor: "pointer",
              fontFamily: "inherit",
              fontSize: 12.5,
              fontWeight: 600,
              background: on ? "var(--card)" : "transparent",
              color: on ? "var(--adaptive-900)" : "var(--adaptive-500)",
              boxShadow: on ? "var(--shadow-xs)" : "none",
              transition: "all .15s",
            }}
          >
            <Icon name={o.icon} size={15} color={on ? "var(--primary-600)" : "var(--adaptive-400)"} />
            {o.label}
          </button>
        );
      })}
    </div>
  );
}

function KpiCard({
  label,
  value,
  dot,
  active,
  onClick,
}: {
  label: string;
  value: number;
  dot: string;
  active: boolean;
  onClick: () => void;
}) {
  return (
    <Card
      hover
      onClick={onClick}
      style={{
        flex: 1,
        minWidth: 130,
        padding: "12px 14px",
        borderColor: active ? "var(--primary-300)" : "var(--adaptive-200)",
        boxShadow: active ? "0 0 0 3px var(--primary-100)" : undefined,
      }}
    >
      <div style={{ display: "flex", alignItems: "center", gap: 7, marginBottom: 8 }}>
        <span style={{ width: 9, height: 9, borderRadius: "50%", background: dot }} />
        <span style={{ fontSize: 12, fontWeight: 500, color: "var(--adaptive-600)" }}>{label}</span>
      </div>
      <div
        style={{
          fontSize: 26,
          fontWeight: 700,
          color: "var(--adaptive-900)",
          letterSpacing: "-0.02em",
          fontFeatureSettings: "'tnum'",
        }}
      >
        {value}
      </div>
    </Card>
  );
}

// Non-interactive KPI fed by approved leave covering today (not a board status,
// so it isn't a clickable filter).
function LeaveKpi({ value }: { value: number }) {
  return (
    <Card style={{ flex: 1, minWidth: 130, padding: "12px 14px", background: "var(--adaptive-50)" }}>
      <div style={{ display: "flex", alignItems: "center", gap: 7, marginBottom: 8 }}>
        <Icon name="sun" size={14} color="var(--adaptive-400)" />
        <span style={{ fontSize: 12, fontWeight: 500, color: "var(--adaptive-600)" }}>On leave</span>
      </div>
      <div
        style={{
          fontSize: 26,
          fontWeight: 700,
          color: "var(--adaptive-900)",
          letterSpacing: "-0.02em",
          fontFeatureSettings: "'tnum'",
        }}
      >
        {value}
      </div>
    </Card>
  );
}

function WorkerCard({
  row,
  nowMs,
  locationName,
  onSelect,
}: {
  row: Row;
  nowMs: number;
  locationName: string;
  onSelect: (r: Row) => void;
}) {
  const onShiftMins = row.check_in_at
    ? minutesBetween(row.check_in_at, new Date(nowMs).toISOString())
    : null;
  return (
    <Card hover onClick={() => onSelect(row)} style={{ padding: 13 }}>
      <div style={{ display: "flex", alignItems: "flex-start", gap: 10 }}>
        <Avatar person={{ id: row.employee_id, name: row.employee_name }} size={38} />
        <div style={{ flex: 1, minWidth: 0 }}>
          <div
            style={{
              fontSize: 14,
              fontWeight: 600,
              color: "var(--adaptive-900)",
              whiteSpace: "nowrap",
              overflow: "hidden",
              textOverflow: "ellipsis",
            }}
          >
            {row.employee_name}
          </div>
          <div style={{ fontSize: 12, color: "var(--adaptive-500)", fontFeatureSettings: "'tnum'" }}>
            {formatTime(row.shift.starts_at)} – {formatTime(row.shift.ends_at)}
          </div>
        </div>
      </div>

      <div
        style={{
          display: "flex",
          alignItems: "center",
          gap: 6,
          marginTop: 10,
          color: "var(--adaptive-500)",
          fontSize: 12,
        }}
      >
        <Icon name="pin" size={14} color="var(--adaptive-400)" />
        <span style={{ whiteSpace: "nowrap", overflow: "hidden", textOverflow: "ellipsis" }}>
          {locationName}
        </span>
      </div>

      <div
        style={{
          display: "flex",
          alignItems: "center",
          gap: 8,
          marginTop: 11,
          paddingTop: 11,
          borderTop: "1px solid var(--adaptive-100)",
        }}
      >
        {row._status === "late" ? (
          <span
            style={{
              fontSize: 12,
              fontWeight: 600,
              color: "var(--red-600)",
              display: "flex",
              alignItems: "center",
              gap: 5,
            }}
          >
            <Icon name="alert" size={14} color="var(--red-500)" /> Not checked in
          </span>
        ) : row._status === "upcoming" ? (
          <span style={{ fontSize: 12, color: "var(--adaptive-500)", display: "flex", alignItems: "center", gap: 5 }}>
            <Icon name="clock" size={14} color="var(--adaptive-400)" /> Starts {formatTime(row.shift.starts_at)}
          </span>
        ) : row._status === "done" ? (
          <span style={{ fontSize: 12, color: "var(--adaptive-500)", display: "flex", alignItems: "center", gap: 5 }}>
            <Icon name="check" size={14} color="var(--blue-500)" />{" "}
            {row.check_in_at ? formatTime(row.check_in_at) : "—"}–
            {row.check_out_at ? formatTime(row.check_out_at) : "—"}
          </span>
        ) : (
          <span style={{ fontSize: 12, color: "var(--adaptive-600)", display: "flex", alignItems: "center", gap: 5 }}>
            <Icon name="clock" size={14} color="var(--adaptive-400)" /> In{" "}
            {row.check_in_at ? formatTime(row.check_in_at) : "—"}
            {onShiftMins != null ? ` · ${fmtDur(onShiftMins)} on shift` : ""}
          </span>
        )}
      </div>
    </Card>
  );
}

function Board({
  rows,
  nowMs,
  locationName,
  onSelect,
}: {
  rows: Row[];
  nowMs: number;
  locationName: (id?: string | null) => string;
  onSelect: (r: Row) => void;
}) {
  const cols: { status: LiveStatus; label: string }[] = [
    { status: "working", label: "On shift" },
    { status: "break", label: "On break" },
    { status: "late", label: "Not checked in" },
    { status: "upcoming", label: "Upcoming today" },
    { status: "done", label: "Checked out" },
  ];
  return (
    <div style={{ display: "flex", gap: 14, overflowX: "auto", paddingBottom: 8, alignItems: "flex-start" }}>
      {cols.map((c) => {
        const items = rows.filter((r) => r._status === c.status);
        const s = STATUS[c.status];
        return (
          <div
            key={c.status}
            style={{ flex: "1 1 0", minWidth: 248, display: "flex", flexDirection: "column", gap: 10 }}
          >
            <div style={{ display: "flex", alignItems: "center", gap: 8, padding: "2px 2px" }}>
              <span style={{ width: 8, height: 8, borderRadius: "50%", background: s.dot }} />
              <span style={{ fontSize: 13, fontWeight: 600, color: "var(--adaptive-800)" }}>{c.label}</span>
              <span
                style={{
                  fontSize: 12,
                  fontWeight: 600,
                  color: "var(--adaptive-400)",
                  background: "var(--adaptive-100)",
                  borderRadius: 9999,
                  padding: "0 8px",
                  minWidth: 22,
                  textAlign: "center",
                }}
              >
                {items.length}
              </span>
            </div>
            {items.length === 0 ? (
              <div
                style={{
                  fontSize: 12,
                  color: "var(--adaptive-400)",
                  padding: "14px 0",
                  textAlign: "center",
                  border: "1px dashed var(--adaptive-200)",
                  borderRadius: 8,
                }}
              >
                None
              </div>
            ) : (
              items.map((r) => (
                <WorkerCard
                  key={r.shift.id}
                  row={r}
                  nowMs={nowMs}
                  locationName={locationName(r.shift.location_id)}
                  onSelect={onSelect}
                />
              ))
            )}
          </div>
        );
      })}
    </div>
  );
}

function ListView({
  rows,
  nowMs,
  locationName,
  onSelect,
}: {
  rows: Row[];
  nowMs: number;
  locationName: (id?: string | null) => string;
  onSelect: (r: Row) => void;
}) {
  return (
    <Card style={{ overflow: "hidden" }}>
      <div style={{ overflowX: "auto" }}>
        <table style={{ width: "100%", borderCollapse: "collapse", fontSize: 13, minWidth: 720 }}>
          <thead>
            <tr style={{ background: "var(--adaptive-50)", textAlign: "left" }}>
              {["Staff", "Status", "Shift", "Location", "Check-in", "On shift"].map((h) => (
                <th
                  key={h}
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
            {rows.map((r, i) => {
              const onShiftMins = r.check_in_at
                ? minutesBetween(r.check_in_at, new Date(nowMs).toISOString())
                : null;
              return (
                <tr
                  key={r.shift.id}
                  onClick={() => onSelect(r)}
                  style={{
                    cursor: "pointer",
                    borderBottom: i < rows.length - 1 ? "1px solid var(--adaptive-100)" : "none",
                  }}
                  onMouseEnter={(e) => (e.currentTarget.style.background = "var(--adaptive-50)")}
                  onMouseLeave={(e) => (e.currentTarget.style.background = "transparent")}
                >
                  <td style={{ padding: "10px 16px" }}>
                    <div style={{ display: "flex", alignItems: "center", gap: 10 }}>
                      <Avatar person={{ id: r.employee_id, name: r.employee_name }} size={30} />
                      <span style={{ fontWeight: 600, color: "var(--adaptive-900)" }}>{r.employee_name}</span>
                    </div>
                  </td>
                  <td style={{ padding: "10px 16px" }}>
                    <StatusChip status={r._status} small />
                  </td>
                  <td style={{ padding: "10px 16px", color: "var(--adaptive-700)", fontFeatureSettings: "'tnum'" }}>
                    {formatTime(r.shift.starts_at)} – {formatTime(r.shift.ends_at)}
                  </td>
                  <td style={{ padding: "10px 16px", color: "var(--adaptive-500)" }}>
                    {locationName(r.shift.location_id)}
                  </td>
                  <td style={{ padding: "10px 16px", color: "var(--adaptive-700)", fontFeatureSettings: "'tnum'" }}>
                    {r.check_in_at ? formatTime(r.check_in_at) : "—"}
                  </td>
                  <td style={{ padding: "10px 16px", color: "var(--adaptive-700)", fontFeatureSettings: "'tnum'" }}>
                    {onShiftMins != null && (r._status === "working" || r._status === "done")
                      ? fmtDur(onShiftMins)
                      : "—"}
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

// Map layout: plots check-in coordinates, normalized to the points' bounding
// box, in a stylized panel. Only in-the-field crew (with coordinates) appear.
function MapView({
  rows,
  locationName,
  onSelect,
}: {
  rows: Row[];
  locationName: (id?: string | null) => string;
  onSelect: (r: Row) => void;
}) {
  const pinned = rows.filter((r) => r.check_in_lat != null && r.check_in_lng != null);

  const lats = pinned.map((r) => r.check_in_lat as number);
  const lngs = pinned.map((r) => r.check_in_lng as number);
  const minLat = Math.min(...lats);
  const maxLat = Math.max(...lats);
  const minLng = Math.min(...lngs);
  const maxLng = Math.max(...lngs);
  const spanLat = maxLat - minLat || 1;
  const spanLng = maxLng - minLng || 1;

  // Normalize to 6%–94% so pins never sit on the panel edge. Latitude is
  // inverted (north = up).
  function pos(r: Row): { left: string; top: string } {
    const x = pinned.length === 1 ? 0.5 : (((r.check_in_lng as number) - minLng) / spanLng);
    const y = pinned.length === 1 ? 0.5 : (((r.check_in_lat as number) - minLat) / spanLat);
    return { left: `${6 + x * 88}%`, top: `${6 + (1 - y) * 88}%` };
  }

  return (
    <div style={{ display: "grid", gridTemplateColumns: "minmax(0, 1.6fr) minmax(0, 1fr)", gap: 16, alignItems: "stretch" }}>
      <Card style={{ overflow: "hidden", position: "relative", minHeight: 420 }}>
        {/* stylized map backdrop */}
        <div
          style={{
            position: "absolute",
            inset: 0,
            background:
              "linear-gradient(135deg, var(--adaptive-50), var(--adaptive-100))",
            backgroundImage:
              "linear-gradient(var(--adaptive-200) 1px, transparent 1px), linear-gradient(90deg, var(--adaptive-200) 1px, transparent 1px)",
            backgroundSize: "44px 44px",
            opacity: 0.6,
          }}
        />
        {pinned.length === 0 ? (
          <div
            style={{
              position: "relative",
              minHeight: 420,
              display: "grid",
              placeItems: "center",
              color: "var(--adaptive-500)",
              textAlign: "center",
              padding: 24,
            }}
          >
            No check-in locations to plot yet.
          </div>
        ) : (
          pinned.map((r) => {
            const s = STATUS[r._status];
            const p = pos(r);
            return (
              <button
                key={r.shift.id}
                onClick={() => onSelect(r)}
                title={`${r.employee_name} · ${s.label}`}
                style={{
                  position: "absolute",
                  left: p.left,
                  top: p.top,
                  transform: "translate(-50%, -50%)",
                  width: 38,
                  height: 38,
                  borderRadius: "50%",
                  border: 0,
                  cursor: "pointer",
                  background: colorForPin(r.employee_id),
                  color: "#fff",
                  fontWeight: 600,
                  fontSize: 13,
                  display: "flex",
                  alignItems: "center",
                  justifyContent: "center",
                  boxShadow: "0 2px 8px rgba(3,7,18,0.28)",
                }}
              >
                {initials(r.employee_name)}
                <span
                  style={{
                    position: "absolute",
                    right: -2,
                    bottom: -2,
                    width: 12,
                    height: 12,
                    borderRadius: "50%",
                    background: s.dot,
                    border: "2px solid var(--card)",
                  }}
                />
              </button>
            );
          })
        )}
      </Card>

      <Card style={{ padding: 14, display: "flex", flexDirection: "column", gap: 4 }}>
        <DrawerSectionLabel>In the field · {pinned.length}</DrawerSectionLabel>
        {pinned.length === 0 ? (
          <div style={{ fontSize: 13, color: "var(--adaptive-500)" }}>No one checked in with a location.</div>
        ) : (
          pinned.map((r) => (
            <button
              key={r.shift.id}
              onClick={() => onSelect(r)}
              style={{
                display: "flex",
                alignItems: "center",
                gap: 10,
                padding: "8px 6px",
                border: 0,
                borderBottom: "1px solid var(--adaptive-100)",
                background: "transparent",
                cursor: "pointer",
                fontFamily: "inherit",
                textAlign: "left",
              }}
            >
              <Avatar person={{ id: r.employee_id, name: r.employee_name }} size={30} />
              <div style={{ flex: 1, minWidth: 0 }}>
                <div style={{ fontSize: 13, fontWeight: 600, color: "var(--adaptive-900)" }}>{r.employee_name}</div>
                <div style={{ fontSize: 11.5, color: "var(--adaptive-500)" }}>{locationName(r.shift.location_id)}</div>
              </div>
              <StatusChip status={r._status} small />
            </button>
          ))
        )}
      </Card>
    </div>
  );
}

// Distinct pin color per employee (reuses the avatar hue logic indirectly).
function colorForPin(id: string): string {
  const palette = ["#ea580c", "#2563eb", "#7c3aed", "#0d9488", "#db2777", "#d97706", "#15803d", "#4b5563"];
  let total = 0;
  for (const c of id) total += c.charCodeAt(0);
  return palette[total % palette.length];
}

function ActivityRow({ time, label, dot }: { time: string; label: string; dot: string }) {
  return (
    <div style={{ display: "flex", gap: 10, alignItems: "flex-start" }}>
      <div
        style={{
          width: 9,
          height: 9,
          borderRadius: "50%",
          background: dot,
          marginTop: 4,
          flexShrink: 0,
          boxShadow: "0 0 0 3px var(--card)",
        }}
      />
      <div style={{ flex: 1, paddingBottom: 14 }}>
        <div style={{ fontSize: 13, color: "var(--adaptive-800)" }}>{label}</div>
        <div style={{ fontSize: 11.5, color: "var(--adaptive-400)", fontFeatureSettings: "'tnum'" }}>{time}</div>
      </div>
    </div>
  );
}

function DetailDrawer({
  row,
  nowMs,
  locationName,
  onClose,
}: {
  row: Row;
  nowMs: number;
  locationName: string;
  onClose: () => void;
}) {
  const onShiftMins = row.check_in_at
    ? minutesBetween(row.check_in_at, new Date(nowMs).toISOString())
    : null;
  const acts: { time: string; label: string; dot: string }[] = [];
  if (row.check_out_at) {
    acts.push({ time: formatTime(row.check_out_at), label: "Checked out", dot: "var(--blue-500)" });
  }
  if (row.check_in_at) {
    acts.push({ time: formatTime(row.check_in_at), label: `Checked in at ${locationName}`, dot: "var(--green-500)" });
  }
  if (row._status === "late") {
    acts.push({
      time: formatTime(row.shift.starts_at),
      label: "Scheduled start — not checked in",
      dot: "var(--red-500)",
    });
  }
  acts.push({
    time: `${formatTime(row.shift.starts_at)}–${formatTime(row.shift.ends_at)}`,
    label: "Shift assigned",
    dot: "var(--adaptive-300)",
  });

  return (
    <Drawer
      onClose={onClose}
      header={
        <>
          <Avatar person={{ id: row.employee_id, name: row.employee_name }} size={42} />
          <div style={{ minWidth: 0 }}>
            <div style={{ fontSize: 16, fontWeight: 600, color: "var(--adaptive-900)" }}>
              {row.employee_name}
            </div>
            <div style={{ fontSize: 12.5, color: "var(--adaptive-500)" }}>Field staff</div>
          </div>
        </>
      }
    >
      <div style={{ display: "flex", alignItems: "center", gap: 10 }}>
        <StatusChip status={row._status} />
        {row._status === "working" && onShiftMins != null && (
          <span style={{ fontSize: 13, color: "var(--adaptive-500)" }}>{fmtDur(onShiftMins)} on shift</span>
        )}
      </div>

      <Card style={{ padding: 14 }}>
        <div
          style={{
            display: "grid",
            gridTemplateColumns: "auto 1fr",
            gap: "8px 14px",
            fontSize: 13,
          }}
        >
          <span style={{ color: "var(--adaptive-500)" }}>Window</span>
          <span style={{ color: "var(--adaptive-800)", fontFeatureSettings: "'tnum'" }}>
            {formatTime(row.shift.starts_at)}–{formatTime(row.shift.ends_at)}
          </span>
          <span style={{ color: "var(--adaptive-500)" }}>Check-in</span>
          <span style={{ color: "var(--adaptive-800)", fontFeatureSettings: "'tnum'" }}>
            {row.check_in_at ? formatTime(row.check_in_at) : "Not yet"}
          </span>
          <span style={{ color: "var(--adaptive-500)" }}>Check-out</span>
          <span style={{ color: "var(--adaptive-800)", fontFeatureSettings: "'tnum'" }}>
            {row.check_out_at ? formatTime(row.check_out_at) : "—"}
          </span>
          <span style={{ color: "var(--adaptive-500)" }}>Location</span>
          <span style={{ color: "var(--adaptive-800)" }}>{locationName}</span>
        </div>
      </Card>

      <div>
        <DrawerSectionLabel>Today's activity</DrawerSectionLabel>
        {acts.map((a, i) => (
          <ActivityRow key={i} {...a} />
        ))}
      </div>
    </Drawer>
  );
}

export function LiveView({
  entries,
  leaveRequests,
  locationNames,
  onRefresh,
  loading,
}: {
  entries: LiveViewEntry[];
  leaveRequests: LeaveRequest[];
  locationNames: Map<string, string>;
  onRefresh: () => void;
  loading: boolean;
}) {
  const [layout, setLayout] = useState<Layout>("board");
  const [filter, setFilter] = useState<LiveStatus | null>(null);
  const [sel, setSel] = useState<Row | null>(null);
  const nowMs = Date.now();
  const today = new Date();

  const rows: Row[] = entries.map((e) => ({ ...e, _status: deriveStatus(e, nowMs) }));
  const count = (st: LiveStatus) => rows.filter((r) => r._status === st).length;
  const shown = filter ? rows.filter((r) => r._status === filter) : rows;
  const locationName = (id?: string | null) => (id ? locationNames.get(id) ?? "Unknown" : "Unassigned");

  // On leave today: distinct employees with an approved request covering today.
  const onLeaveCount = (() => {
    const ids = new Set<string>();
    for (const l of leaveRequests) {
      if (l.status === "approved" && leaveCoversDay(l.start_date, l.end_date, today)) {
        ids.add(l.employee_id);
      }
    }
    return ids.size;
  })();

  return (
    <div style={{ padding: "20px 24px 32px", display: "flex", flexDirection: "column", gap: 18 }}>
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
            Who's working now
          </h1>
          <div
            style={{
              display: "flex",
              alignItems: "center",
              gap: 8,
              marginTop: 5,
              fontSize: 13,
              color: "var(--adaptive-500)",
            }}
          >
            <span style={{ display: "inline-flex", alignItems: "center", gap: 6 }}>
              <span
                className="opero-pulse"
                style={{ width: 8, height: 8, borderRadius: "50%", background: "var(--green-500)" }}
              />
              Live
            </span>
            <span>·</span>
            <span>{count("working")} on shift</span>
            <span>·</span>
            <span>{rows.length} scheduled today</span>
          </div>
        </div>
        <div style={{ display: "flex", alignItems: "center", gap: 10 }}>
          <Seg value={layout} onChange={setLayout} />
          <Btn variant="secondary" icon="refresh" onClick={onRefresh} disabled={loading}>
            {loading ? "Refreshing" : "Refresh"}
          </Btn>
        </div>
      </div>

      <div style={{ display: "flex", gap: 12, flexWrap: "wrap" }}>
        <KpiCard label="On shift" dot="var(--green-500)" value={count("working")} active={filter === "working"} onClick={() => setFilter(filter === "working" ? null : "working")} />
        <KpiCard label="On break" dot="var(--amber-500)" value={count("break")} active={filter === "break"} onClick={() => setFilter(filter === "break" ? null : "break")} />
        <KpiCard label="Not checked in" dot="var(--red-500)" value={count("late")} active={filter === "late"} onClick={() => setFilter(filter === "late" ? null : "late")} />
        <KpiCard label="Upcoming" dot="var(--adaptive-400)" value={count("upcoming")} active={filter === "upcoming"} onClick={() => setFilter(filter === "upcoming" ? null : "upcoming")} />
        <KpiCard label="Checked out" dot="var(--blue-500)" value={count("done")} active={filter === "done"} onClick={() => setFilter(filter === "done" ? null : "done")} />
        <LeaveKpi value={onLeaveCount} />
      </div>

      {filter && (
        <div style={{ display: "flex", alignItems: "center", gap: 8, fontSize: 13, color: "var(--adaptive-600)" }}>
          Filtered by <StatusChip status={filter} small />
          <button
            onClick={() => setFilter(null)}
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
            Clear
          </button>
        </div>
      )}

      {rows.length === 0 ? (
        <Card style={{ padding: 0 }}>
          <div
            style={{
              display: "grid",
              minHeight: 200,
              placeItems: "center",
              color: "var(--adaptive-500)",
              textAlign: "center",
              padding: 24,
            }}
          >
            No published shifts in the current window.
          </div>
        </Card>
      ) : layout === "board" ? (
        <Board rows={shown} nowMs={nowMs} locationName={locationName} onSelect={setSel} />
      ) : layout === "map" ? (
        <MapView rows={shown} locationName={locationName} onSelect={setSel} />
      ) : (
        <ListView rows={shown} nowMs={nowMs} locationName={locationName} onSelect={setSel} />
      )}

      {sel && (
        <DetailDrawer
          row={sel}
          nowMs={nowMs}
          locationName={locationName(sel.shift.location_id)}
          onClose={() => setSel(null)}
        />
      )}
    </div>
  );
}
