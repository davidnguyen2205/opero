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
// derive the richer live statuses from that plus the shift's scheduled start.
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

// Urgency order for the timeline: what needs attention first.
const STATUS_ORDER: Record<LiveStatus, number> = {
  late: 0,
  working: 1,
  break: 2,
  done: 3,
  upcoming: 4,
  off: 5,
};

type Layout = "timeline" | "list" | "map";

function Seg({
  value,
  onChange,
}: {
  value: Layout;
  onChange: (v: Layout) => void;
}) {
  const options: { id: Layout; label: string; icon: "activity" | "list" | "map" }[] = [
    { id: "timeline", label: "Timeline", icon: "activity" },
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

// One pill per status: the count and the filter in a single element.
function StatChip({
  label,
  value,
  dot,
  active,
  onClick,
}: {
  label: string;
  value: number;
  dot?: string;
  active?: boolean;
  onClick?: () => void;
}) {
  return (
    <button
      onClick={onClick}
      disabled={!onClick}
      style={{
        display: "inline-flex",
        alignItems: "center",
        gap: 7,
        padding: "5px 13px",
        borderRadius: 9999,
        border: `1px solid ${active ? "var(--primary-300)" : "var(--adaptive-200)"}`,
        background: "var(--card)",
        boxShadow: active ? "0 0 0 3px var(--primary-100)" : "none",
        cursor: onClick ? "pointer" : "default",
        fontFamily: "inherit",
        fontSize: 12.5,
        color: "var(--adaptive-600)",
        fontFeatureSettings: "'tnum'",
        transition: "border-color .15s, box-shadow .15s",
      }}
    >
      {dot && <span style={{ width: 8, height: 8, borderRadius: "50%", background: dot }} />}
      <span style={{ fontWeight: 700, color: "var(--adaptive-900)" }}>{value}</span>
      {label}
    </button>
  );
}

// Exceptions a manager must act on, pinned above the timeline.
function AttentionBanner({
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
  if (rows.length === 0) {
    return null;
  }
  return (
    <div
      style={{
        border: "1.5px dashed var(--red-200)",
        background: "var(--red-50)",
        borderRadius: 10,
        padding: "4px 14px",
      }}
    >
      {rows.map((r, i) => {
        const lateMins = minutesBetween(r.shift.starts_at, new Date(nowMs).toISOString());
        return (
          <div
            key={r.shift.id}
            style={{
              display: "flex",
              alignItems: "center",
              gap: 9,
              padding: "7px 0",
              borderTop: i > 0 ? "1px solid var(--red-100)" : "none",
              fontSize: 13,
            }}
          >
            <Icon name="alert" size={15} color="var(--red-500)" />
            <span style={{ color: "var(--adaptive-800)", minWidth: 0 }}>
              <strong style={{ fontWeight: 600 }}>{r.employee_name}</strong> hasn&rsquo;t checked
              in — shift started {formatTime(r.shift.starts_at)}
              {lateMins > 0 ? ` (${fmtDur(lateMins)} ago)` : ""} · {locationName(r.shift.location_id)}
            </span>
            <button
              onClick={() => onSelect(r)}
              style={{
                marginLeft: "auto",
                flexShrink: 0,
                border: "1px solid var(--adaptive-200)",
                borderRadius: 6,
                background: "var(--card)",
                padding: "3px 11px",
                fontFamily: "inherit",
                fontSize: 12,
                fontWeight: 600,
                color: "var(--adaptive-600)",
                cursor: "pointer",
              }}
            >
              View
            </button>
          </div>
        );
      })}
    </div>
  );
}

const HOUR_MS = 3_600_000;
const RAIL_W = 216;
// Cap the axis so one outlier shift (e.g. crossing far past midnight) can't
// stretch every bar into illegibility; clipped bars get a "→" marker.
const MAX_AXIS_MS = 18 * HOUR_MS;

function floorToLocalHour(ms: number): number {
  const d = new Date(ms);
  d.setMinutes(0, 0, 0);
  return d.getTime();
}

// Per-status one-liner under the employee name.
function railMeta(r: Row, nowMs: number): string {
  switch (r._status) {
    case "working": {
      const mins = r.check_in_at ? minutesBetween(r.check_in_at, new Date(nowMs).toISOString()) : null;
      return mins != null ? `On shift ${fmtDur(mins)}` : "On shift";
    }
    case "break":
      return "On break";
    case "late":
      return "Not checked in";
    case "done":
      return `Done ${r.check_in_at ? formatTime(r.check_in_at) : "—"}–${
        r.check_out_at ? formatTime(r.check_out_at) : "—"
      }`;
    default:
      return `Starts ${formatTime(r.shift.starts_at)}`;
  }
}

function Timeline({
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
  if (rows.length === 0) {
    return (
      <Card style={{ padding: "40px 24px", textAlign: "center", color: "var(--adaptive-500)", fontSize: 13 }}>
        No one matches this filter.
      </Card>
    );
  }

  // Axis window: fit today's shifts (plus "now") with an hour of margin,
  // rounded to whole local hours, capped at MAX_AXIS_MS.
  const startTimes = rows.map((r) => new Date(r.shift.starts_at).getTime());
  const endTimes = rows.map((r) => new Date(r.shift.ends_at).getTime());
  const axisStart = floorToLocalHour(Math.min(...startTimes, nowMs) - HOUR_MS);
  let axisEnd = floorToLocalHour(Math.max(...endTimes, nowMs)) + HOUR_MS;
  if (axisEnd - axisStart > MAX_AXIS_MS) {
    axisEnd = axisStart + MAX_AXIS_MS;
  }
  const span = axisEnd - axisStart;
  const pct = (ms: number) => Math.min(100, Math.max(0, ((ms - axisStart) / span) * 100));

  // Hour ticks: pick a step that yields at most ~8 labels (4h blocks for a
  // full-day span; never 3h).
  const spanHours = span / HOUR_MS;
  const stepH = [1, 2, 4, 6].find((s) => spanHours / s <= 8) ?? 6;
  const ticks: number[] = [];
  for (let t = axisStart; t <= axisEnd; t += stepH * HOUR_MS) {
    ticks.push(t);
  }

  const nowPct = pct(nowMs);
  const sorted = [...rows].sort(
    (a, b) =>
      STATUS_ORDER[a._status] - STATUS_ORDER[b._status] ||
      new Date(a.shift.starts_at).getTime() - new Date(b.shift.starts_at).getTime() ||
      a.employee_name.localeCompare(b.employee_name),
  );

  const gridLines = (
    <>
      {ticks.map((t) => (
        <span
          key={t}
          style={{
            position: "absolute",
            left: `${pct(t)}%`,
            top: 0,
            bottom: 0,
            width: 1,
            background: "var(--adaptive-100)",
          }}
        />
      ))}
      <span
        style={{
          position: "absolute",
          left: `${nowPct}%`,
          top: 0,
          bottom: 0,
          width: 2,
          background: "var(--primary-500)",
          opacity: 0.85,
        }}
      />
    </>
  );

  return (
    <Card style={{ padding: "4px 18px 6px", overflowX: "auto" }}>
      <div style={{ minWidth: 720 }}>
        {/* axis */}
        <div
          style={{
            display: "grid",
            gridTemplateColumns: `${RAIL_W}px 1fr`,
            borderBottom: "1px solid var(--adaptive-200)",
          }}
        >
          <div />
          <div style={{ position: "relative", height: 30 }}>
            {/* hide hour labels the NOW pill would collide with */}
            {ticks.filter((t) => Math.abs(pct(t) - nowPct) > 5).map((t) => (
              <span
                key={t}
                style={{
                  position: "absolute",
                  left: `${pct(t)}%`,
                  bottom: 5,
                  transform: "translateX(-50%)",
                  fontSize: 10.5,
                  color: "var(--adaptive-400)",
                  fontFeatureSettings: "'tnum'",
                  whiteSpace: "nowrap",
                }}
              >
                {formatTime(new Date(t).toISOString())}
              </span>
            ))}
            <span
              style={{
                position: "absolute",
                left: `${nowPct}%`,
                top: 3,
                transform: "translateX(-50%)",
                background: "var(--primary-500)",
                color: "#fff",
                fontSize: 10,
                fontWeight: 700,
                padding: "1px 7px",
                borderRadius: 9999,
                whiteSpace: "nowrap",
              }}
            >
              {formatTime(new Date(nowMs).toISOString())}
            </span>
          </div>
        </div>

        {/* rows */}
        {sorted.map((r, i) => {
          const s = STATUS[r._status];
          const shiftStart = new Date(r.shift.starts_at).getTime();
          const shiftEnd = new Date(r.shift.ends_at).getTime();
          const clipped = shiftEnd > axisEnd;
          const schedLeft = pct(shiftStart);
          const schedWidth = Math.max(pct(shiftEnd) - schedLeft, 0.5);
          const inMs = r.check_in_at ? new Date(r.check_in_at).getTime() : null;
          const outMs = r.check_out_at ? new Date(r.check_out_at).getTime() : null;
          const actualEnd = outMs ?? nowMs;
          return (
            <div
              key={r.shift.id}
              onClick={() => onSelect(r)}
              style={{
                display: "grid",
                gridTemplateColumns: `${RAIL_W}px 1fr`,
                borderBottom: i < sorted.length - 1 ? "1px solid var(--adaptive-100)" : "none",
                cursor: "pointer",
                opacity: r._status === "upcoming" ? 0.6 : 1,
              }}
              onMouseEnter={(e) => (e.currentTarget.style.background = "var(--adaptive-50)")}
              onMouseLeave={(e) => (e.currentTarget.style.background = "transparent")}
            >
              <div style={{ display: "flex", alignItems: "center", gap: 10, padding: "9px 12px 9px 2px", minWidth: 0 }}>
                <Avatar person={{ id: r.employee_id, name: r.employee_name }} size={30} />
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
                    {r.employee_name}
                  </div>
                  <div
                    style={{
                      fontSize: 11,
                      color: "var(--adaptive-500)",
                      whiteSpace: "nowrap",
                      overflow: "hidden",
                      textOverflow: "ellipsis",
                      fontFeatureSettings: "'tnum'",
                    }}
                  >
                    <span style={{ fontWeight: 600, color: s.fg }}>{railMeta(r, nowMs)}</span>
                    {" · "}
                    {locationName(r.shift.location_id)}
                  </div>
                </div>
              </div>

              <div style={{ position: "relative", height: 48 }}>
                {gridLines}
                {/* scheduled window */}
                <span
                  style={{
                    position: "absolute",
                    top: 10,
                    height: 26,
                    left: `${schedLeft}%`,
                    width: `${schedWidth}%`,
                    border: "1.5px dashed var(--adaptive-300)",
                    borderRadius: 6,
                    boxSizing: "border-box",
                  }}
                />
                {/* actual worked segment */}
                {inMs != null && actualEnd > inMs && (
                  <span
                    style={{
                      position: "absolute",
                      top: 10,
                      height: 26,
                      left: `${pct(inMs)}%`,
                      width: `${Math.max(pct(actualEnd) - pct(inMs), 0.4)}%`,
                      background: s.dot,
                      opacity: 0.75,
                      borderRadius: 6,
                    }}
                  />
                )}
                {/* missed-start marker */}
                {r._status === "late" && (
                  <span
                    style={{
                      position: "absolute",
                      top: 7,
                      height: 32,
                      left: `${schedLeft}%`,
                      width: 2.5,
                      background: "var(--red-500)",
                      borderRadius: 2,
                    }}
                  />
                )}
                {clipped && (
                  <span
                    style={{
                      position: "absolute",
                      top: 15,
                      right: 2,
                      fontSize: 12,
                      color: "var(--adaptive-400)",
                    }}
                  >
                    →
                  </span>
                )}
              </div>
            </div>
          );
        })}
      </div>
    </Card>
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
  const [layout, setLayout] = useState<Layout>("timeline");
  const [filter, setFilter] = useState<LiveStatus | null>(null);
  const [sel, setSel] = useState<Row | null>(null);
  const nowMs = Date.now();
  const today = new Date();

  const rows: Row[] = entries.map((e) => ({ ...e, _status: deriveStatus(e, nowMs) }));
  const count = (st: LiveStatus) => rows.filter((r) => r._status === st).length;
  const shown = filter ? rows.filter((r) => r._status === filter) : rows;
  const locationName = (id?: string | null) => (id ? locationNames.get(id) ?? "Unknown" : "Unassigned");
  const lateRows = rows.filter((r) => r._status === "late");

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

  const chips: { status: LiveStatus; label: string }[] = [
    { status: "working", label: "on shift" },
    { status: "break", label: "on break" },
    { status: "late", label: "not checked in" },
    { status: "upcoming", label: "upcoming" },
    { status: "done", label: "done" },
  ];

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

      <div style={{ display: "flex", gap: 8, flexWrap: "wrap", alignItems: "center" }}>
        {chips.map((c) => (
          <StatChip
            key={c.status}
            label={c.label}
            value={count(c.status)}
            dot={STATUS[c.status].dot}
            active={filter === c.status}
            onClick={() => setFilter(filter === c.status ? null : c.status)}
          />
        ))}
        <StatChip label="on leave" value={onLeaveCount} />
      </div>

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
      ) : (
        <>
          {layout === "timeline" && filter !== "late" && (
            <AttentionBanner rows={lateRows} nowMs={nowMs} locationName={locationName} onSelect={setSel} />
          )}
          {layout === "timeline" ? (
            <Timeline rows={shown} nowMs={nowMs} locationName={locationName} onSelect={setSel} />
          ) : layout === "map" ? (
            <MapView rows={shown} locationName={locationName} onSelect={setSel} />
          ) : (
            <ListView rows={shown} nowMs={nowMs} locationName={locationName} onSelect={setSel} />
          )}
        </>
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
