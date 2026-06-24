import { useState } from "react";
import type { CSSProperties } from "react";
import type {
  CreateTourRequest,
  Tour,
  TourCategory,
  UpdateTourRequest,
} from "../api/resources";
import {
  Btn,
  Card,
  Chip,
  Drawer,
  DrawerSectionLabel,
  Field,
  Icon,
  PageHeader,
  controlStyle,
} from "../ui";
import type { ChipTone } from "../ui";

const CATEGORIES: TourCategory[] = ["walking", "day_trip", "food", "driving", "evening"];

const CAT_LABEL: Record<TourCategory, string> = {
  walking: "Walking",
  day_trip: "Day Trip",
  food: "Food",
  driving: "Driving",
  evening: "Evening",
};

const CAT_TONE: Record<TourCategory, ChipTone> = {
  walking: "orange",
  day_trip: "blue",
  food: "neutral",
  driving: "neutral",
  evening: "neutral",
};

const TOUR_COLORS = ["#ea580c", "#2563eb", "#7c3aed", "#0d9488", "#db2777", "#d97706", "#15803d", "#9333ea"];

function euros(cents: number): string {
  return (cents / 100).toLocaleString(undefined, { minimumFractionDigits: 0, maximumFractionDigits: 2 });
}

function durLabel(min: number): string {
  const h = Math.floor(min / 60);
  const m = min % 60;
  if (h && m) return `${h}h ${m}m`;
  if (h) return `${h}h`;
  return `${m}m`;
}

function crewLabel(t: Tour): string {
  const parts: string[] = [];
  if (t.guides_needed) parts.push(`${t.guides_needed} guide${t.guides_needed > 1 ? "s" : ""}`);
  if (t.drivers_needed) parts.push(`${t.drivers_needed} driver${t.drivers_needed > 1 ? "s" : ""}`);
  return parts.join(" + ") || "—";
}

function Stars({ value }: { value: number }) {
  return (
    <span
      style={{
        display: "inline-flex",
        alignItems: "center",
        gap: 4,
        fontSize: 12.5,
        fontWeight: 600,
        color: "var(--adaptive-700)",
        fontFeatureSettings: "'tnum'",
      }}
    >
      <svg width="13" height="13" viewBox="0 0 24 24" fill="var(--amber-500)" stroke="none">
        <path d="M12 2l2.9 6.3 6.9.7-5.1 4.6 1.4 6.8L12 17.8 5.9 20.4l1.4-6.8L2.2 9l6.9-.7z" />
      </svg>
      {value.toFixed(1)}
    </span>
  );
}

function MetaItem({ icon, children }: { icon: "clock" | "users" | "route" | "pin"; children: React.ReactNode }) {
  return (
    <span style={{ display: "inline-flex", alignItems: "center", gap: 6, fontSize: 12.5, color: "var(--adaptive-600)" }}>
      <Icon name={icon} size={14} color="var(--adaptive-400)" />
      {children}
    </span>
  );
}

function TourCard({ tour, onSelect }: { tour: Tour; onSelect: () => void }) {
  const color = tour.color ?? "#ea580c";
  return (
    <Card hover onClick={onSelect} style={{ overflow: "hidden", opacity: tour.active ? 1 : 0.72, display: "flex", flexDirection: "column" }}>
      <div style={{ height: 5, background: color }} />
      <div style={{ padding: 16, display: "flex", flexDirection: "column", gap: 10, flex: 1 }}>
        <div style={{ display: "flex", alignItems: "flex-start", gap: 10 }}>
          <div style={{ flex: 1, minWidth: 0 }}>
            <div style={{ fontSize: 15.5, fontWeight: 700, color: "var(--adaptive-900)", letterSpacing: "-0.01em" }}>
              {tour.name}
            </div>
            {tour.meeting_point && (
              <div style={{ display: "flex", alignItems: "center", gap: 6, marginTop: 3, fontSize: 12, color: "var(--adaptive-500)" }}>
                <Icon name="pin" size={13} color="var(--adaptive-400)" />
                {tour.meeting_point}
              </div>
            )}
          </div>
          {!tour.active && <Chip>Paused</Chip>}
        </div>

        <div style={{ display: "flex", flexWrap: "wrap", gap: "6px 14px" }}>
          <MetaItem icon="clock">{durLabel(tour.duration_min)}</MetaItem>
          <MetaItem icon="users">Max {tour.max_guests}</MetaItem>
          <MetaItem icon="route">{crewLabel(tour)}</MetaItem>
        </div>

        <div
          style={{
            display: "flex",
            alignItems: "center",
            gap: 8,
            marginTop: "auto",
            paddingTop: 12,
            borderTop: "1px solid var(--adaptive-100)",
          }}
        >
          <Chip tone={CAT_TONE[tour.category]}>{CAT_LABEL[tour.category]}</Chip>
          {tour.rating != null && tour.rating > 0 && <Stars value={tour.rating} />}
          <span style={{ marginLeft: "auto", fontSize: 14, fontWeight: 700, color: "var(--adaptive-900)" }}>
            €{euros(tour.price_cents)}
          </span>
        </div>
        <div style={{ fontSize: 11.5, color: "var(--adaptive-500)" }}>
          {tour.departure_times.length} departure{tour.departure_times.length === 1 ? "" : "s"}/day
        </div>
      </div>
    </Card>
  );
}

function TourDrawer({
  tour,
  onClose,
  onEdit,
  onDelete,
}: {
  tour: Tour;
  onClose: () => void;
  onEdit: () => void;
  onDelete: () => void;
}) {
  const color = tour.color ?? "#ea580c";
  return (
    <Drawer
      onClose={onClose}
      width={440}
      header={
        <>
          <span style={{ width: 12, height: 12, borderRadius: 4, background: color, flexShrink: 0 }} />
          <div style={{ minWidth: 0 }}>
            <div style={{ fontSize: 16, fontWeight: 700, color: "var(--adaptive-900)" }}>{tour.name}</div>
            <div style={{ display: "flex", alignItems: "center", gap: 8, marginTop: 3 }}>
              <Chip tone={CAT_TONE[tour.category]}>{CAT_LABEL[tour.category]}</Chip>
              {tour.rating != null && tour.rating > 0 && <Stars value={tour.rating} />}
              {!tour.active && <Chip>Paused</Chip>}
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
            onClick={onDelete}
          >
            Delete
          </Btn>
          <Btn variant="primary" icon="briefcase" style={{ marginLeft: "auto" }} onClick={onEdit}>
            Edit Tour
          </Btn>
        </>
      }
    >
      {tour.description && (
        <p style={{ margin: 0, fontSize: 13.5, lineHeight: 1.55, color: "var(--adaptive-600)" }}>{tour.description}</p>
      )}

      <div style={{ display: "grid", gridTemplateColumns: "1fr 1fr", gap: 10 }}>
        {(
          [
            ["clock", "Duration", durLabel(tour.duration_min)],
            ["users", "Max guests", String(tour.max_guests)],
            ["route", "Crew needed", crewLabel(tour)],
            ["pin", "Meeting point", tour.meeting_point ?? "—"],
          ] as const
        ).map(([ic, k, v]) => (
          <div key={k} style={{ border: "1px solid var(--adaptive-200)", borderRadius: 8, padding: "11px 12px" }}>
            <div style={{ display: "flex", alignItems: "center", gap: 6, fontSize: 11.5, color: "var(--adaptive-500)", marginBottom: 5 }}>
              <Icon name={ic} size={13} color="var(--adaptive-400)" />
              {k}
            </div>
            <div style={{ fontSize: 14, fontWeight: 600, color: "var(--adaptive-900)" }}>{v}</div>
          </div>
        ))}
      </div>

      <div
        style={{
          display: "flex",
          alignItems: "center",
          gap: 8,
          padding: "11px 14px",
          borderRadius: 8,
          background: "var(--adaptive-50)",
          border: "1px solid var(--adaptive-200)",
        }}
      >
        <Icon name="briefcase" size={16} color="var(--adaptive-500)" />
        <span style={{ fontSize: 13, color: "var(--adaptive-600)" }}>Price per guest</span>
        <span style={{ marginLeft: "auto", fontSize: 15, fontWeight: 700, color: "var(--adaptive-900)" }}>
          €{euros(tour.price_cents)}
        </span>
      </div>

      <div>
        <DrawerSectionLabel>
          Daily departures · {tour.departure_times.length}
        </DrawerSectionLabel>
        {tour.departure_times.length === 0 ? (
          <div style={{ fontSize: 13, color: "var(--adaptive-500)" }}>No departure times configured.</div>
        ) : (
          <div style={{ display: "flex", flexWrap: "wrap", gap: 8 }}>
            {tour.departure_times.map((t, i) => (
              <span
                key={i}
                style={{
                  fontSize: 13,
                  fontWeight: 600,
                  color: "var(--adaptive-800)",
                  background: "var(--adaptive-50)",
                  border: "1px solid var(--adaptive-200)",
                  borderRadius: 7,
                  padding: "5px 11px",
                  fontFeatureSettings: "'tnum'",
                }}
              >
                {t}
              </span>
            ))}
          </div>
        )}
      </div>
    </Drawer>
  );
}

type FormState = {
  name: string;
  category: TourCategory;
  meeting_point: string;
  duration_min: string;
  max_guests: string;
  price_eu: string;
  guides_needed: string;
  drivers_needed: string;
  rating: string;
  color: string;
  description: string;
  active: boolean;
  times: string[];
};

function tourToForm(t: Tour): FormState {
  return {
    name: t.name,
    category: t.category,
    meeting_point: t.meeting_point ?? "",
    duration_min: String(t.duration_min),
    max_guests: String(t.max_guests),
    price_eu: String(t.price_cents / 100),
    guides_needed: String(t.guides_needed),
    drivers_needed: String(t.drivers_needed),
    rating: t.rating != null ? String(t.rating) : "",
    color: t.color ?? TOUR_COLORS[0],
    description: t.description ?? "",
    active: t.active,
    times: [...t.departure_times],
  };
}

const EMPTY_FORM: FormState = {
  name: "",
  category: "walking",
  meeting_point: "",
  duration_min: "120",
  max_guests: "10",
  price_eu: "30",
  guides_needed: "1",
  drivers_needed: "0",
  rating: "",
  color: TOUR_COLORS[0],
  description: "",
  active: true,
  times: ["10:00"],
};

function TourForm({
  tour,
  onClose,
  onCreate,
  onUpdate,
}: {
  tour: Tour | null;
  onClose: () => void;
  onCreate: (body: CreateTourRequest) => Promise<void>;
  onUpdate: (id: string, body: UpdateTourRequest) => Promise<void>;
}) {
  const isEdit = Boolean(tour);
  const [f, setF] = useState<FormState>(tour ? tourToForm(tour) : EMPTY_FORM);
  const [submitting, setSubmitting] = useState(false);
  const set = <K extends keyof FormState>(k: K, v: FormState[K]) => setF((p) => ({ ...p, [k]: v }));
  const canSubmit = f.name.trim().length > 0 && !submitting;

  function body(): CreateTourRequest {
    const ratingNum = f.rating.trim() ? Number(f.rating) : undefined;
    return {
      name: f.name.trim(),
      category: f.category,
      meeting_point: f.meeting_point.trim() || undefined,
      duration_min: Number(f.duration_min) || 0,
      max_guests: Number(f.max_guests) || 0,
      guides_needed: Number(f.guides_needed) || 0,
      drivers_needed: Number(f.drivers_needed) || 0,
      departure_times: f.times.map((t) => t.trim()).filter(Boolean),
      price_cents: Math.round((Number(f.price_eu) || 0) * 100),
      rating: ratingNum,
      active: f.active,
      color: f.color,
      description: f.description.trim() || undefined,
    };
  }

  async function submit() {
    if (!canSubmit) return;
    setSubmitting(true);
    try {
      if (tour) {
        await onUpdate(tour.id, body());
      } else {
        await onCreate(body());
      }
      onClose();
    } finally {
      setSubmitting(false);
    }
  }

  const numField: CSSProperties = { ...controlStyle };

  return (
    <Drawer
      onClose={onClose}
      width={440}
      header={<div style={{ fontSize: 16, fontWeight: 600 }}>{isEdit ? "Edit Tour" : "New Tour"}</div>}
      footer={
        <>
          <Btn variant="tertiary" onClick={onClose} style={{ marginRight: "auto" }}>
            Cancel
          </Btn>
          <Btn variant="primary" icon="check" disabled={!canSubmit} onClick={() => void submit()}>
            {isEdit ? "Save Changes" : "Create Tour"}
          </Btn>
        </>
      }
    >
      <div style={{ display: "flex", flexDirection: "column", gap: 16 }}>
        <Field label="Tour name">
          <input value={f.name} onChange={(e) => set("name", e.target.value)} placeholder="e.g. Alfama Walking Tour" style={controlStyle} />
        </Field>
        <Field label="Category">
          <select value={f.category} onChange={(e) => set("category", e.target.value as TourCategory)} style={controlStyle}>
            {CATEGORIES.map((c) => (
              <option key={c} value={c}>
                {CAT_LABEL[c]}
              </option>
            ))}
          </select>
        </Field>
        <Field label="Meeting point">
          <input value={f.meeting_point} onChange={(e) => set("meeting_point", e.target.value)} placeholder="e.g. Praça do Comércio" style={controlStyle} />
        </Field>
        <div style={{ display: "flex", gap: 12 }}>
          <div style={{ flex: 1 }}>
            <Field label="Duration (min)">
              <input type="number" value={f.duration_min} onChange={(e) => set("duration_min", e.target.value)} style={numField} />
            </Field>
          </div>
          <div style={{ flex: 1 }}>
            <Field label="Max guests">
              <input type="number" value={f.max_guests} onChange={(e) => set("max_guests", e.target.value)} style={numField} />
            </Field>
          </div>
          <div style={{ flex: 1 }}>
            <Field label="Price (€)">
              <input type="number" step="0.01" value={f.price_eu} onChange={(e) => set("price_eu", e.target.value)} style={numField} />
            </Field>
          </div>
        </div>
        <div style={{ display: "flex", gap: 12 }}>
          <div style={{ flex: 1 }}>
            <Field label="Guides needed">
              <input type="number" min={0} value={f.guides_needed} onChange={(e) => set("guides_needed", e.target.value)} style={numField} />
            </Field>
          </div>
          <div style={{ flex: 1 }}>
            <Field label="Drivers needed">
              <input type="number" min={0} value={f.drivers_needed} onChange={(e) => set("drivers_needed", e.target.value)} style={numField} />
            </Field>
          </div>
          <div style={{ flex: 1 }}>
            <Field label="Rating (0–5)">
              <input type="number" step="0.1" min={0} max={5} value={f.rating} onChange={(e) => set("rating", e.target.value)} style={numField} />
            </Field>
          </div>
        </div>
        <Field label="Daily departures">
          <div style={{ display: "flex", flexWrap: "wrap", gap: 8 }}>
            {f.times.map((t, i) => (
              <div
                key={i}
                style={{ display: "inline-flex", alignItems: "center", gap: 4, border: "1px solid var(--adaptive-200)", borderRadius: 6, paddingRight: 4 }}
              >
                <input
                  value={t}
                  onChange={(e) => set("times", f.times.map((x, j) => (j === i ? e.target.value : x)))}
                  style={{ ...controlStyle, width: 80, border: 0, minHeight: 34, padding: "0 8px" }}
                />
                <button
                  onClick={() => set("times", f.times.filter((_, j) => j !== i))}
                  style={{ width: 22, height: 22, borderRadius: 5, border: 0, background: "var(--adaptive-100)", cursor: "pointer", display: "flex", alignItems: "center", justifyContent: "center" }}
                >
                  <Icon name="x" size={12} color="var(--adaptive-500)" />
                </button>
              </div>
            ))}
            <button
              onClick={() => set("times", [...f.times, "12:00"])}
              style={{
                height: 36,
                padding: "0 12px",
                borderRadius: 6,
                border: "1px dashed var(--adaptive-300)",
                background: "var(--card)",
                cursor: "pointer",
                fontFamily: "inherit",
                fontSize: 12.5,
                fontWeight: 600,
                color: "var(--adaptive-600)",
                display: "inline-flex",
                alignItems: "center",
                gap: 5,
              }}
            >
              <Icon name="plus" size={13} />
              Add
            </button>
          </div>
        </Field>
        <Field label="Accent color">
          <div style={{ display: "flex", gap: 8, flexWrap: "wrap" }}>
            {TOUR_COLORS.map((c) => (
              <button
                key={c}
                onClick={() => set("color", c)}
                style={{
                  width: 30,
                  height: 30,
                  borderRadius: 7,
                  background: c,
                  border: "2px solid var(--card)",
                  cursor: "pointer",
                  boxShadow: f.color === c ? `0 0 0 2px ${c}` : "0 0 0 1px var(--adaptive-200)",
                }}
              />
            ))}
          </div>
        </Field>
        <Field label="Description">
          <textarea
            value={f.description}
            onChange={(e) => set("description", e.target.value)}
            style={{ ...controlStyle, minHeight: 80, resize: "vertical" }}
            placeholder="What the tour covers…"
          />
        </Field>
        <label style={{ display: "flex", alignItems: "center", gap: 10, cursor: "pointer", fontSize: 13.5, fontWeight: 500, color: "var(--adaptive-800)" }}>
          <button
            onClick={() => set("active", !f.active)}
            style={{
              width: 38,
              height: 22,
              borderRadius: 9999,
              border: 0,
              cursor: "pointer",
              background: f.active ? "var(--green-500)" : "var(--adaptive-300)",
              position: "relative",
              flexShrink: 0,
            }}
          >
            <span style={{ position: "absolute", top: 2, left: f.active ? 18 : 2, width: 18, height: 18, borderRadius: "50%", background: "#fff", transition: "left .15s" }} />
          </button>
          {f.active ? "Active — bookable & schedulable" : "Paused — hidden from new schedules"}
        </label>
      </div>
    </Drawer>
  );
}

export function Tours({
  tours,
  onCreate,
  onUpdate,
  onDelete,
}: {
  tours: Tour[];
  onCreate: (body: CreateTourRequest) => Promise<void>;
  onUpdate: (id: string, body: UpdateTourRequest) => Promise<void>;
  onDelete: (id: string) => void;
}) {
  const [layout, setLayout] = useState<"grid" | "list">("grid");
  const [cat, setCat] = useState<"all" | TourCategory>("all");
  const [sel, setSel] = useState<Tour | null>(null);
  const [editing, setEditing] = useState<Tour | "new" | null>(null);

  const shown = cat === "all" ? tours : tours.filter((t) => t.category === cat);
  const activeCount = tours.filter((t) => t.active).length;
  const cats: ("all" | TourCategory)[] = ["all", ...CATEGORIES.filter((c) => tours.some((t) => t.category === c))];

  return (
    <div style={{ padding: "20px 24px 32px", display: "flex", flexDirection: "column", gap: 18 }}>
      <PageHeader
        title="Tours"
        subtitle={`${activeCount} active tour${activeCount === 1 ? "" : "s"} · ${tours.length} total`}
        actions={
          <>
            <div style={{ display: "inline-flex", background: "var(--adaptive-100)", borderRadius: 7, padding: 3, gap: 2 }}>
              {(["grid", "list"] as const).map((o) => {
                const on = o === layout;
                return (
                  <button
                    key={o}
                    onClick={() => setLayout(o)}
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
            <Btn variant="primary" icon="plus" onClick={() => setEditing("new")}>
              New Tour
            </Btn>
          </>
        }
      />

      <div style={{ display: "flex", gap: 8, flexWrap: "wrap" }}>
        {cats.map((c) => {
          const on = c === cat;
          const n = c === "all" ? tours.length : tours.filter((t) => t.category === c).length;
          return (
            <button
              key={c}
              onClick={() => setCat(c)}
              style={{
                display: "inline-flex",
                alignItems: "center",
                gap: 7,
                padding: "6px 12px",
                borderRadius: 9999,
                cursor: "pointer",
                fontFamily: "inherit",
                fontSize: 13,
                fontWeight: 600,
                border: `1px solid ${on ? "var(--primary-300)" : "var(--adaptive-200)"}`,
                background: on ? "var(--primary-50)" : "var(--card)",
                color: on ? "var(--primary-700)" : "var(--adaptive-600)",
              }}
            >
              {c === "all" ? "All" : CAT_LABEL[c]}
              <span style={{ fontSize: 11, fontWeight: 700, opacity: 0.7 }}>{n}</span>
            </button>
          );
        })}
      </div>

      {tours.length === 0 ? (
        <Card style={{ padding: 28 }}>
          <div style={{ textAlign: "center", color: "var(--adaptive-500)" }}>
            No tours yet. Create one to build your catalog.
          </div>
        </Card>
      ) : layout === "grid" ? (
        <div style={{ display: "grid", gridTemplateColumns: "repeat(auto-fill, minmax(300px, 1fr))", gap: 16 }}>
          {shown.map((t) => (
            <TourCard key={t.id} tour={t} onSelect={() => setSel(t)} />
          ))}
        </div>
      ) : (
        <Card style={{ overflow: "hidden" }}>
          <div style={{ overflowX: "auto" }}>
            <table style={{ width: "100%", borderCollapse: "collapse", fontSize: 13, minWidth: 760 }}>
              <thead>
                <tr style={{ background: "var(--adaptive-50)", textAlign: "left" }}>
                  {["Tour", "Category", "Duration", "Capacity", "Crew", "Departures", "Price", ""].map((h, i) => (
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
                {shown.map((t, i) => (
                  <tr
                    key={t.id}
                    onClick={() => setSel(t)}
                    style={{ cursor: "pointer", borderBottom: i < shown.length - 1 ? "1px solid var(--adaptive-100)" : "none", opacity: t.active ? 1 : 0.7 }}
                    onMouseEnter={(e) => (e.currentTarget.style.background = "var(--adaptive-50)")}
                    onMouseLeave={(e) => (e.currentTarget.style.background = "transparent")}
                  >
                    <td style={{ padding: "11px 16px" }}>
                      <div style={{ display: "flex", alignItems: "center", gap: 11 }}>
                        <span style={{ width: 10, height: 10, borderRadius: 3, background: t.color ?? "#ea580c", flexShrink: 0 }} />
                        <div>
                          <div style={{ fontWeight: 600, color: "var(--adaptive-900)" }}>{t.name}</div>
                          {t.meeting_point && <div style={{ fontSize: 11.5, color: "var(--adaptive-500)" }}>{t.meeting_point}</div>}
                        </div>
                      </div>
                    </td>
                    <td style={{ padding: "11px 16px" }}>
                      <Chip tone={CAT_TONE[t.category]}>{CAT_LABEL[t.category]}</Chip>
                    </td>
                    <td style={{ padding: "11px 16px", color: "var(--adaptive-700)", fontFeatureSettings: "'tnum'" }}>{durLabel(t.duration_min)}</td>
                    <td style={{ padding: "11px 16px", color: "var(--adaptive-700)" }}>{t.max_guests} guests</td>
                    <td style={{ padding: "11px 16px", color: "var(--adaptive-700)" }}>{crewLabel(t)}</td>
                    <td style={{ padding: "11px 16px", color: "var(--adaptive-500)", fontFeatureSettings: "'tnum'" }}>{t.departure_times.join(", ") || "—"}</td>
                    <td style={{ padding: "11px 16px", fontWeight: 700, color: "var(--adaptive-900)", fontFeatureSettings: "'tnum'" }}>€{euros(t.price_cents)}</td>
                    <td style={{ padding: "11px 16px", textAlign: "right" }}>
                      <Icon name="chevron" size={15} color="var(--adaptive-300)" />
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </Card>
      )}

      {sel && (
        <TourDrawer
          tour={sel}
          onClose={() => setSel(null)}
          onEdit={() => {
            setEditing(sel);
            setSel(null);
          }}
          onDelete={() => {
            onDelete(sel.id);
            setSel(null);
          }}
        />
      )}
      {editing && (
        <TourForm
          tour={editing === "new" ? null : editing}
          onClose={() => setEditing(null)}
          onCreate={onCreate}
          onUpdate={onUpdate}
        />
      )}
    </div>
  );
}
