// Shared UI primitives, ported from the Opero design system (Blazeup tokens).
// These mirror the Claude Design prototype: inline styles driven by CSS
// variables defined in styles.css. Kept framework-light and fully typed.
import { useState } from "react";
import type { CSSProperties, ReactNode } from "react";

// ── Icons: 24×24, 1.75 stroke, currentColor (Lucide-style) ──────────────
export type IconName =
  | "activity"
  | "calendar"
  | "users"
  | "phone"
  | "grid"
  | "list"
  | "map"
  | "clock"
  | "pin"
  | "camera"
  | "check"
  | "plus"
  | "search"
  | "bell"
  | "send"
  | "chevron"
  | "chevronDown"
  | "filter"
  | "download"
  | "more"
  | "alert"
  | "sun"
  | "briefcase"
  | "refresh"
  | "x"
  | "route"
  | "pencil"
  | "wifi";

const ICON_PATHS: Record<IconName, string> = {
  activity: "M22 12h-4l-3 9L9 3l-3 9H2",
  calendar:
    "M8 2v4M16 2v4M3 10h18M5 4h14a2 2 0 0 1 2 2v14a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V6a2 2 0 0 1 2-2z",
  users:
    "M16 21v-2a4 4 0 0 0-4-4H6a4 4 0 0 0-4 4v2M9 11a4 4 0 1 0 0-8 4 4 0 0 0 0 8zM22 21v-2a4 4 0 0 0-3-3.87M16 3.13a4 4 0 0 1 0 7.75",
  phone: "M7 2h10a1 1 0 0 1 1 1v18a1 1 0 0 1-1 1H7a1 1 0 0 1-1-1V3a1 1 0 0 1 1-1zM11 19h2",
  grid: "M3 3h7v7H3zM14 3h7v7h-7zM14 14h7v7h-7zM3 14h7v7H3z",
  list: "M8 6h13M8 12h13M8 18h13M3 6h.01M3 12h.01M3 18h.01",
  map: "M9 4 3 6v15l6-2 6 2 6-2V4l-6 2-6-2zM9 4v15M15 6v15",
  clock: "M12 22a10 10 0 1 0 0-20 10 10 0 0 0 0 20zM12 6v6l4 2",
  pin: "M20 10c0 6-8 12-8 12s-8-6-8-12a8 8 0 0 1 16 0zM12 10a2 2 0 1 0 0-4 2 2 0 0 0 0 4z",
  camera:
    "M23 19a2 2 0 0 1-2 2H3a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h4l2-3h6l2 3h4a2 2 0 0 1 2 2zM12 17a4 4 0 1 0 0-8 4 4 0 0 0 0 8z",
  check: "M20 6 9 17l-5-5",
  plus: "M12 5v14M5 12h14",
  search: "M11 19a8 8 0 1 0 0-16 8 8 0 0 0 0 16zM21 21l-4.3-4.3",
  bell: "M18 8a6 6 0 0 0-12 0c0 7-3 9-3 9h18s-3-2-3-9zM13.7 21a2 2 0 0 1-3.4 0",
  send: "M22 2 11 13M22 2l-7 20-4-9-9-4 20-7z",
  chevron: "M9 18l6-6-6-6",
  chevronDown: "M6 9l6 6 6-6",
  filter: "M22 3H2l8 9.46V19l4 2v-8.54L22 3z",
  download: "M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4M7 10l5 5 5-5M12 15V3",
  more: "M12 13a1 1 0 1 0 0-2 1 1 0 0 0 0 2zM19 13a1 1 0 1 0 0-2 1 1 0 0 0 0 2zM5 13a1 1 0 1 0 0-2 1 1 0 0 0 0 2z",
  alert:
    "M10.29 3.86 1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0zM12 9v4M12 17h.01",
  sun: "M12 17a5 5 0 1 0 0-10 5 5 0 0 0 0 10zM12 1v2M12 21v2M4.2 4.2l1.4 1.4M18.4 18.4l1.4 1.4M1 12h2M21 12h2M4.2 19.8l1.4-1.4M18.4 5.6l1.4-1.4",
  briefcase:
    "M20 7H4a2 2 0 0 0-2 2v9a2 2 0 0 0 2 2h16a2 2 0 0 0 2-2V9a2 2 0 0 0-2-2zM16 7V5a2 2 0 0 0-2-2h-4a2 2 0 0 0-2 2v2",
  refresh:
    "M23 4v6h-6M1 20v-6h6M3.51 9a9 9 0 0 1 14.85-3.36L23 10M1 14l4.64 4.36A9 9 0 0 0 20.49 15",
  x: "M18 6 6 18M6 6l12 12",
  route:
    "M6 19a3 3 0 1 0 0-6 3 3 0 0 0 0 6zM18 11a3 3 0 1 0 0-6 3 3 0 0 0 0 6zM9 16h6a3 3 0 0 0 0-6H9a3 3 0 0 1 0-6h0",
  pencil: "M12 20h9M16.5 3.5a2.12 2.12 0 0 1 3 3L7 19l-4 1 1-4z",
  wifi: "M5 12.55a11 11 0 0 1 14 0M8.5 16.1a6 6 0 0 1 7 0M2 8.82a15 15 0 0 1 20 0M12 20h.01",
};

export function Icon({
  name,
  size = 18,
  color = "currentColor",
  strokeWidth = 1.75,
  style,
}: {
  name: IconName;
  size?: number;
  color?: string;
  strokeWidth?: number;
  style?: CSSProperties;
}) {
  const d = ICON_PATHS[name];
  return (
    <svg
      aria-hidden="true"
      width={size}
      height={size}
      viewBox="0 0 24 24"
      fill="none"
      stroke={color}
      strokeWidth={strokeWidth}
      strokeLinecap="round"
      strokeLinejoin="round"
      style={{ flexShrink: 0, ...style }}
    >
      {d
        .split("M")
        .filter(Boolean)
        .map((seg, i) => (
          <path d={`M${seg}`} key={i} />
        ))}
    </svg>
  );
}

// ── Status system ────────────────────────────────────────────────────────
// The API exposes only not_checked_in / checked_in / checked_out. These are
// the richer presentational statuses we derive from that + shift timing.
export type LiveStatus = "working" | "break" | "late" | "upcoming" | "done" | "off";

export const STATUS: Record<
  LiveStatus,
  { label: string; dot: string; fg: string; bg: string; bd: string }
> = {
  working: { label: "On shift", dot: "var(--green-500)", fg: "var(--green-700)", bg: "var(--green-50)", bd: "var(--green-200)" },
  break: { label: "On break", dot: "var(--amber-500)", fg: "var(--amber-700)", bg: "var(--amber-50)", bd: "var(--amber-200)" },
  late: { label: "Running late", dot: "var(--red-500)", fg: "var(--red-700)", bg: "var(--red-50)", bd: "var(--red-200)" },
  upcoming: { label: "Upcoming", dot: "var(--adaptive-400)", fg: "var(--adaptive-600)", bg: "var(--adaptive-100)", bd: "var(--adaptive-200)" },
  done: { label: "Checked out", dot: "var(--blue-500)", fg: "var(--blue-700)", bg: "var(--blue-50)", bd: "var(--blue-200)" },
  off: { label: "Off today", dot: "var(--adaptive-300)", fg: "var(--adaptive-500)", bg: "transparent", bd: "var(--adaptive-200)" },
};

// Avatar color palette (from the Blazeup hues used in the prototype).
export const AVATAR_COLORS = [
  "#ea580c",
  "#2563eb",
  "#7c3aed",
  "#0d9488",
  "#db2777",
  "#d97706",
  "#15803d",
  "#4b5563",
];

export function colorForId(id: string): string {
  let total = 0;
  for (const char of id) {
    total += char.charCodeAt(0);
  }
  return AVATAR_COLORS[total % AVATAR_COLORS.length];
}

export function initials(name: string): string {
  const parts = name.trim().split(/\s+/).filter(Boolean);
  if (parts.length === 0) {
    return "?";
  }
  if (parts.length === 1) {
    return parts[0].slice(0, 2).toUpperCase();
  }
  return `${parts[0][0]}${parts[parts.length - 1][0]}`.toUpperCase();
}

export type Person = { id: string; name: string };

export function Avatar({
  person,
  size = 32,
  ring,
}: {
  person: Person;
  size?: number;
  ring?: string;
}) {
  return (
    <div
      style={{
        width: size,
        height: size,
        borderRadius: "50%",
        flexShrink: 0,
        background: colorForId(person.id),
        color: "#fff",
        display: "flex",
        alignItems: "center",
        justifyContent: "center",
        fontWeight: 600,
        fontSize: size * 0.38,
        letterSpacing: "-0.01em",
        boxShadow: ring ? `0 0 0 2px var(--card), 0 0 0 4px ${ring}` : "none",
      }}
    >
      {initials(person.name)}
    </div>
  );
}

export function AvatarStack({
  people,
  max = 5,
  size = 30,
}: {
  people: Person[];
  max?: number;
  size?: number;
}) {
  const shown = people.slice(0, max);
  const extra = people.length - shown.length;
  return (
    <div style={{ display: "flex", alignItems: "center" }}>
      {shown.map((p, i) => (
        <div
          key={p.id}
          style={{ marginLeft: i === 0 ? 0 : -10, borderRadius: "50%", boxShadow: "0 0 0 2px var(--card)" }}
        >
          <Avatar person={p} size={size} />
        </div>
      ))}
      {extra > 0 && (
        <div
          style={{
            marginLeft: -10,
            width: size,
            height: size,
            borderRadius: "50%",
            background: "var(--adaptive-100)",
            border: "1px solid var(--adaptive-200)",
            boxShadow: "0 0 0 2px var(--card)",
            display: "flex",
            alignItems: "center",
            justifyContent: "center",
            fontSize: size * 0.34,
            fontWeight: 600,
            color: "var(--adaptive-600)",
          }}
        >
          +{extra}
        </div>
      )}
    </div>
  );
}

export function StatusChip({ status, small }: { status: LiveStatus; small?: boolean }) {
  const s = STATUS[status];
  return (
    <span
      style={{
        display: "inline-flex",
        alignItems: "center",
        gap: 6,
        padding: small ? "2px 8px" : "3px 10px",
        borderRadius: 9999,
        background: s.bg,
        border: `1px solid ${s.bd}`,
        fontSize: small ? 11 : 12,
        fontWeight: 600,
        color: s.fg,
        whiteSpace: "nowrap",
      }}
    >
      <span style={{ width: 7, height: 7, borderRadius: "50%", background: s.dot }} />
      {s.label}
    </span>
  );
}

export type ChipTone = "neutral" | "orange" | "blue";

export function Chip({ children, tone = "neutral" }: { children: ReactNode; tone?: ChipTone }) {
  const tones: Record<ChipTone, { bg: string; fg: string; bd: string }> = {
    neutral: { bg: "var(--adaptive-100)", fg: "var(--adaptive-700)", bd: "var(--adaptive-200)" },
    orange: { bg: "var(--primary-50)", fg: "var(--primary-700)", bd: "var(--primary-200)" },
    blue: { bg: "var(--blue-50)", fg: "var(--blue-700)", bd: "var(--blue-200)" },
  };
  const t = tones[tone];
  return (
    <span
      style={{
        display: "inline-flex",
        alignItems: "center",
        gap: 5,
        padding: "2px 8px",
        borderRadius: 6,
        background: t.bg,
        border: `1px solid ${t.bd}`,
        fontSize: 11,
        fontWeight: 600,
        color: t.fg,
        whiteSpace: "nowrap",
      }}
    >
      {children}
    </span>
  );
}

export type BtnVariant = "primary" | "secondary" | "tertiary";
export type BtnSize = "sm" | "md" | "lg";

export function Btn({
  children,
  variant = "secondary",
  size = "md",
  icon,
  onClick,
  style,
  disabled,
  type = "button",
}: {
  children?: ReactNode;
  variant?: BtnVariant;
  size?: BtnSize;
  icon?: IconName;
  onClick?: () => void;
  style?: CSSProperties;
  disabled?: boolean;
  type?: "button" | "submit";
}) {
  const [hover, setHover] = useState(false);
  const sizes: Record<BtnSize, { h: number; px: number; fs: number }> = {
    sm: { h: 32, px: 12, fs: 13 },
    md: { h: 36, px: 14, fs: 13 },
    lg: { h: 42, px: 18, fs: 14 },
  };
  const z = sizes[size];
  const base: CSSProperties = {
    height: z.h,
    padding: `0 ${z.px}px`,
    borderRadius: 6,
    fontSize: z.fs,
    fontWeight: 600,
    fontFamily: "inherit",
    cursor: disabled ? "not-allowed" : "pointer",
    display: "inline-flex",
    alignItems: "center",
    justifyContent: "center",
    gap: 7,
    transition: "background .15s, box-shadow .15s, border-color .15s",
    whiteSpace: "nowrap",
  };
  let v: CSSProperties;
  if (variant === "primary") {
    v = {
      background: disabled ? "var(--adaptive-200)" : hover ? "var(--primary-500)" : "var(--primary-600)",
      color: disabled ? "var(--adaptive-400)" : "#fff",
      border: "1px solid transparent",
      boxShadow: hover && !disabled ? "0 0 0 3px var(--adaptive-300)" : "none",
    };
  } else if (variant === "tertiary") {
    v = {
      background: "transparent",
      color: hover ? "var(--adaptive-900)" : "var(--adaptive-600)",
      border: "1px solid transparent",
    };
  } else {
    v = {
      background: "var(--card)",
      color: "var(--adaptive-800)",
      border: `1px solid ${hover && !disabled ? "var(--adaptive-950)" : "var(--adaptive-200)"}`,
      boxShadow: hover && !disabled ? "0 0 0 3px var(--adaptive-300)" : "none",
    };
  }
  return (
    <button
      type={type}
      onClick={disabled ? undefined : onClick}
      disabled={disabled}
      onMouseEnter={() => setHover(true)}
      onMouseLeave={() => setHover(false)}
      style={{ ...base, ...v, ...style }}
    >
      {icon && <Icon name={icon} size={size === "lg" ? 18 : 15} />}
      {children}
    </button>
  );
}

export function Card({
  children,
  style,
  hover,
  onClick,
}: {
  children: ReactNode;
  style?: CSSProperties;
  hover?: boolean;
  onClick?: () => void;
}) {
  const [h, setH] = useState(false);
  return (
    <div
      onClick={onClick}
      onMouseEnter={() => hover && setH(true)}
      onMouseLeave={() => hover && setH(false)}
      style={{
        background: "var(--card)",
        border: "1px solid var(--adaptive-200)",
        borderRadius: 8,
        boxShadow: h ? "0 0 0 3px var(--adaptive-300)" : "none",
        transition: "box-shadow .15s",
        cursor: onClick ? "pointer" : "default",
        ...style,
      }}
    >
      {children}
    </div>
  );
}

// ── Drawer: shared right-side slide-over (scrim + panel) ──────────────────
export function Drawer({
  onClose,
  header,
  footer,
  children,
  width = 400,
}: {
  onClose: () => void;
  header: ReactNode;
  footer?: ReactNode;
  children: ReactNode;
  width?: number;
}) {
  return (
    <>
      <div
        onClick={onClose}
        style={{ position: "fixed", inset: 0, background: "rgba(3,7,18,0.32)", zIndex: 80 }}
      />
      <div
        style={{
          position: "fixed",
          top: 0,
          right: 0,
          bottom: 0,
          width,
          maxWidth: "92vw",
          zIndex: 81,
          background: "var(--card)",
          borderLeft: "1px solid var(--adaptive-200)",
          boxShadow: "var(--shadow-lg)",
          display: "flex",
          flexDirection: "column",
          fontFamily: "var(--font-sans)",
          animation: "opero-slide .2s ease-out",
        }}
      >
        <div
          style={{
            display: "flex",
            alignItems: "center",
            gap: 12,
            padding: "16px 20px",
            borderBottom: "1px solid var(--adaptive-200)",
          }}
        >
          <div style={{ flex: 1, minWidth: 0, display: "flex", alignItems: "center", gap: 12 }}>
            {header}
          </div>
          <button
            onClick={onClose}
            style={{
              width: 32,
              height: 32,
              borderRadius: 6,
              border: "1px solid var(--adaptive-200)",
              background: "var(--card)",
              cursor: "pointer",
              display: "flex",
              alignItems: "center",
              justifyContent: "center",
              flexShrink: 0,
            }}
          >
            <Icon name="x" size={16} color="var(--adaptive-600)" />
          </button>
        </div>
        <div
          style={{
            flex: 1,
            overflow: "auto",
            padding: 20,
            display: "flex",
            flexDirection: "column",
            gap: 18,
          }}
        >
          {children}
        </div>
        {footer && (
          <div style={{ display: "flex", gap: 8, padding: 16, borderTop: "1px solid var(--adaptive-200)" }}>
            {footer}
          </div>
        )}
      </div>
    </>
  );
}

export function DrawerSectionLabel({ children }: { children: ReactNode }) {
  return (
    <div
      style={{
        fontSize: 12,
        fontWeight: 600,
        color: "var(--adaptive-500)",
        textTransform: "uppercase",
        letterSpacing: ".05em",
        marginBottom: 10,
      }}
    >
      {children}
    </div>
  );
}

// ── Time / text helpers ───────────────────────────────────────────────────
export function humanize(value: string): string {
  return value.replaceAll("_", " ");
}

export function fmtDur(mins: number): string {
  if (mins < 0) mins = 0;
  const h = Math.floor(mins / 60);
  const m = mins % 60;
  return h > 0 ? `${h}h ${m}m` : `${m}m`;
}

export function minutesBetween(fromIso: string, toIso: string): number {
  return Math.round((new Date(toIso).getTime() - new Date(fromIso).getTime()) / 60000);
}

export function formatTime(value: string): string {
  return new Intl.DateTimeFormat(undefined, { hour: "2-digit", minute: "2-digit" }).format(
    new Date(value),
  );
}

export function formatDateTime(value: string): string {
  return new Intl.DateTimeFormat(undefined, { dateStyle: "medium", timeStyle: "short" }).format(
    new Date(value),
  );
}

// ── Form helpers ──────────────────────────────────────────────────────────
export const controlStyle: CSSProperties = {
  fontSize: 13,
  minHeight: 38,
  borderRadius: 6,
  border: "1px solid var(--adaptive-200)",
  padding: "8px 10px",
  fontFamily: "inherit",
  width: "100%",
};

export const labelStyle: CSSProperties = {
  fontSize: 12,
  fontWeight: 600,
  color: "var(--adaptive-700)",
  marginBottom: 6,
  display: "block",
};

export function Field({ label, children }: { label: string; children: ReactNode }) {
  return (
    <div>
      <label style={labelStyle}>{label}</label>
      {children}
    </div>
  );
}

// ── Sorting (mirrors the prototype's useSort/sortRows/SortTh) ──────────────
export type SortState = { key: string | null; dir: "asc" | "desc" };

export function useSort(
  defaultKey: string | null,
  defaultDir: "asc" | "desc" = "asc",
): [SortState, (key: string) => void] {
  const [sort, setSort] = useState<SortState>({ key: defaultKey, dir: defaultDir });
  const toggle = (key: string) =>
    setSort((s) => (s.key === key ? { key, dir: s.dir === "asc" ? "desc" : "asc" } : { key, dir: "asc" }));
  return [sort, toggle];
}

export function sortRows<T>(
  rows: T[],
  sort: SortState,
  accessors: Record<string, (row: T) => string | number | null | undefined>,
): T[] {
  if (!sort.key || !accessors[sort.key]) return rows;
  const acc = accessors[sort.key];
  const out = [...rows].sort((a, b) => {
    const va = acc(a);
    const vb = acc(b);
    if (va == null && vb == null) return 0;
    if (va == null) return 1;
    if (vb == null) return -1;
    if (typeof va === "number" && typeof vb === "number") return va - vb;
    return String(va).localeCompare(String(vb), undefined, { numeric: true });
  });
  return sort.dir === "desc" ? out.reverse() : out;
}

// Sortable <th>. Pass sortKey=null for non-sortable columns (e.g. action cells).
export function SortTh({
  label,
  sortKey,
  sort,
  onSort,
  align = "left",
}: {
  label: string;
  sortKey: string | null;
  sort: SortState;
  onSort: (key: string) => void;
  align?: "left" | "right";
}) {
  const active = sort.key === sortKey;
  const sortable = sortKey != null;
  return (
    <th
      onClick={sortable ? () => onSort(sortKey) : undefined}
      style={{
        padding: "11px 16px",
        fontWeight: 600,
        fontSize: 12,
        borderBottom: "1px solid var(--adaptive-200)",
        whiteSpace: "nowrap",
        textAlign: align,
        userSelect: "none",
        cursor: sortable ? "pointer" : "default",
        color: active ? "var(--adaptive-800)" : "var(--adaptive-500)",
        background: "var(--adaptive-50)",
      }}
    >
      <span
        style={{
          display: "inline-flex",
          alignItems: "center",
          gap: 5,
          flexDirection: align === "right" ? "row-reverse" : "row",
        }}
      >
        {label}
        {sortable && (
          <svg
            width="12"
            height="12"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            strokeWidth="2.2"
            strokeLinecap="round"
            strokeLinejoin="round"
            style={{ opacity: active ? 1 : 0.28, color: active ? "var(--primary-600)" : "currentColor" }}
          >
            {active ? (
              sort.dir === "asc" ? (
                <path d="M18 15l-6-6-6 6" />
              ) : (
                <path d="M6 9l6 6 6-6" />
              )
            ) : (
              <>
                <path d="M8 9l4-4 4 4" />
                <path d="M16 15l-4 4-4-4" />
              </>
            )}
          </svg>
        )}
      </span>
    </th>
  );
}

// ── Grid/list segmented toggle (shared by Tours, Departments) ─────────────
export function ViewToggle({ value, onChange }: { value: "grid" | "list"; onChange: (v: "grid" | "list") => void }) {
  return (
    <div style={{ display: "inline-flex", background: "var(--adaptive-100)", borderRadius: 7, padding: 3, gap: 2 }}>
      {(["grid", "list"] as const).map((o) => {
        const on = o === value;
        return (
          <button
            key={o}
            onClick={() => onChange(o)}
            style={{
              display: "inline-flex",
              alignItems: "center",
              gap: 6,
              padding: "6px 12px",
              borderRadius: 5,
              border: 0,
              cursor: "pointer",
              fontFamily: "inherit",
              fontSize: 12.5,
              fontWeight: 600,
              textTransform: "capitalize",
              background: on ? "var(--card)" : "transparent",
              color: on ? "var(--adaptive-900)" : "var(--adaptive-500)",
              boxShadow: on ? "var(--shadow-xs)" : "none",
            }}
          >
            <Icon name={o === "grid" ? "grid" : "list"} size={15} color={on ? "var(--primary-600)" : "var(--adaptive-400)"} />
            {o}
          </button>
        );
      })}
    </div>
  );
}

// ── Square icon action buttons for table rows ─────────────────────────────
export function IconButton({
  icon,
  title,
  onClick,
  tone = "neutral",
}: {
  icon: IconName;
  title: string;
  onClick: () => void;
  tone?: "neutral" | "danger";
}) {
  const danger = tone === "danger";
  return (
    <button
      title={title}
      aria-label={title}
      onClick={(e) => {
        e.stopPropagation();
        onClick();
      }}
      style={{
        width: 30,
        height: 30,
        borderRadius: 6,
        border: `1px solid ${danger ? "var(--red-200)" : "var(--adaptive-200)"}`,
        background: danger ? "var(--red-50)" : "var(--card)",
        cursor: "pointer",
        display: "inline-flex",
        alignItems: "center",
        justifyContent: "center",
      }}
    >
      <Icon name={icon} size={14} color={danger ? "var(--red-600)" : "var(--adaptive-600)"} />
    </button>
  );
}

// ── Page header (title + subtitle + actions), shared across views ─────────
export function PageHeader({
  title,
  subtitle,
  actions,
}: {
  title: string;
  subtitle?: ReactNode;
  actions?: ReactNode;
}) {
  return (
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
          {title}
        </h1>
        {subtitle != null && (
          <div style={{ marginTop: 5, fontSize: 13, color: "var(--adaptive-500)" }}>{subtitle}</div>
        )}
      </div>
      {actions && <div style={{ display: "flex", alignItems: "center", gap: 10 }}>{actions}</div>}
    </div>
  );
}
