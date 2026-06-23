import { useState } from "react";
import type { CreateLocationRequest, Location } from "../api/resources";
import { Btn, Card, Drawer, Field, Icon, PageHeader, colorForId, controlStyle } from "../ui";

function AddLocationDrawer({
  onClose,
  onCreate,
}: {
  onClose: () => void;
  onCreate: (body: CreateLocationRequest) => Promise<void>;
}) {
  const [name, setName] = useState("");
  const [address, setAddress] = useState("");
  const [lat, setLat] = useState("");
  const [lng, setLng] = useState("");
  const [submitting, setSubmitting] = useState(false);
  const canSubmit = name.trim().length > 0 && !submitting;

  async function submit() {
    if (!canSubmit) return;
    setSubmitting(true);
    try {
      await onCreate({
        name: name.trim(),
        address: address.trim() || undefined,
        lat: lat.trim() ? Number(lat) : undefined,
        lng: lng.trim() ? Number(lng) : undefined,
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
      header={<div style={{ fontSize: 16, fontWeight: 600 }}>Add location</div>}
      footer={
        <>
          <Btn variant="tertiary" onClick={onClose} style={{ marginRight: "auto" }}>
            Cancel
          </Btn>
          <Btn variant="primary" icon="plus" disabled={!canSubmit} onClick={() => void submit()}>
            Create location
          </Btn>
        </>
      }
    >
      <div style={{ display: "flex", flexDirection: "column", gap: 16 }}>
        <Field label="Name">
          <input value={name} onChange={(e) => setName(e.target.value)} style={controlStyle} required />
        </Field>
        <Field label="Address">
          <textarea
            value={address}
            onChange={(e) => setAddress(e.target.value)}
            style={{ ...controlStyle, minHeight: 70, resize: "vertical" }}
          />
        </Field>
        <div style={{ display: "flex", gap: 12 }}>
          <div style={{ flex: 1 }}>
            <Field label="Latitude">
              <input type="number" step="any" value={lat} onChange={(e) => setLat(e.target.value)} style={controlStyle} />
            </Field>
          </div>
          <div style={{ flex: 1 }}>
            <Field label="Longitude">
              <input type="number" step="any" value={lng} onChange={(e) => setLng(e.target.value)} style={controlStyle} />
            </Field>
          </div>
        </div>
      </div>
    </Drawer>
  );
}

export function Locations({
  locations,
  onCreate,
  onDelete,
}: {
  locations: Location[];
  onCreate: (body: CreateLocationRequest) => Promise<void>;
  onDelete: (id: string) => void;
}) {
  const [adding, setAdding] = useState(false);

  return (
    <div style={{ padding: "20px 24px 32px", display: "flex", flexDirection: "column", gap: 18 }}>
      <PageHeader
        title="Locations"
        subtitle={`${locations.length} location${locations.length === 1 ? "" : "s"} for shift assignment`}
        actions={
          <Btn variant="primary" icon="plus" onClick={() => setAdding(true)}>
            Add location
          </Btn>
        }
      />

      {locations.length === 0 ? (
        <Card style={{ padding: 28 }}>
          <div style={{ textAlign: "center", color: "var(--adaptive-500)" }}>
            No locations yet. Create one to assign it to shifts.
          </div>
        </Card>
      ) : (
        <Card style={{ overflow: "hidden" }}>
          <div style={{ overflowX: "auto" }}>
            <table style={{ width: "100%", borderCollapse: "collapse", fontSize: 13, minWidth: 680 }}>
              <thead>
                <tr style={{ background: "var(--adaptive-50)", textAlign: "left" }}>
                  {["Name", "Address", "Coordinates", ""].map((h, i) => (
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
                {locations.map((loc, i) => (
                  <tr
                    key={loc.id}
                    style={{ borderBottom: i < locations.length - 1 ? "1px solid var(--adaptive-100)" : "none" }}
                    onMouseEnter={(e) => (e.currentTarget.style.background = "var(--adaptive-50)")}
                    onMouseLeave={(e) => (e.currentTarget.style.background = "transparent")}
                  >
                    <td style={{ padding: "10px 16px" }}>
                      <div style={{ display: "flex", alignItems: "center", gap: 10 }}>
                        <span
                          style={{
                            width: 28,
                            height: 28,
                            borderRadius: 7,
                            background: `color-mix(in srgb, ${colorForId(loc.id)} 12%, transparent)`,
                            display: "flex",
                            alignItems: "center",
                            justifyContent: "center",
                            flexShrink: 0,
                          }}
                        >
                          <Icon name="pin" size={15} color={colorForId(loc.id)} />
                        </span>
                        <span style={{ fontWeight: 600, color: "var(--adaptive-900)" }}>{loc.name}</span>
                      </div>
                    </td>
                    <td style={{ padding: "10px 16px", color: "var(--adaptive-700)" }}>{loc.address ?? "—"}</td>
                    <td style={{ padding: "10px 16px", color: "var(--adaptive-500)", fontFeatureSettings: "'tnum'" }}>
                      {loc.lat != null && loc.lng != null ? `${loc.lat}, ${loc.lng}` : "—"}
                    </td>
                    <td style={{ padding: "10px 16px", textAlign: "right" }}>
                      <button
                        onClick={() => onDelete(loc.id)}
                        style={{
                          border: "1px solid var(--red-200)",
                          color: "var(--red-700)",
                          background: "var(--red-50)",
                          borderRadius: 6,
                          padding: "5px 10px",
                          fontSize: 12,
                          fontWeight: 600,
                          cursor: "pointer",
                          fontFamily: "inherit",
                        }}
                      >
                        Delete
                      </button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </Card>
      )}

      {adding && <AddLocationDrawer onClose={() => setAdding(false)} onCreate={onCreate} />}
    </div>
  );
}
