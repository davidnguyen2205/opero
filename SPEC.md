# Opero — Product Specification

**Opero** is a people-operations (HR) platform for service companies under ~100 employees, starting with travel/tourism agencies. This spec documents the high-fidelity prototype built on the **Blazeup UI design system**.

- **Demo tenant:** Tagus Trails — a Lisbon day-tours agency (18 staff)
- **Demo "now":** Saturday 21 Jun 2026, 11:42 (peak morning-tour time)
- **Manager persona:** Helena Bastos, Operations Lead
- **Field persona:** Inês Carvalho, Tour Guide
- **Two surfaces:** a manager **web app** and a field-staff **mobile app** (shown in an iPhone frame)

---

## 1. Product scope

The prototype delivers the daily field-ops loop end-to-end:

1. **Org / people core** — employees, departments, roles, employment types, access levels.
2. **Roster scheduling** — managers build and publish weekly shifts (web).
3. **Mobile attendance** — field staff check in/out with geolocation + photo, offline-tolerant (mobile).
4. **Manager live view** — who is working right now (web).

Plus the supporting modules that emerged: **Tours** catalog, **Time Off** approvals, and the full **field app** (schedule, inbox, notifications, profile).

---

## 2. Design system & foundations

- **Base:** Blazeup UI — orange brand (`--primary-600 #ea580c`), adaptive neutral scale that flips for dark mode, mostly-flat surfaces with 1px borders, a 3px hover ring instead of drop shadows, Inter type, Lucide-style 1.75px stroke icons.
- **Tokens:** all color/type/spacing via `var(--*)`. Semantic colors: green (success/on-shift), amber (warning/break), red (error/late), blue (info/done), violet (assist/leave).
- **Structure:** sectioned left sidebar (Operations / People groups) + breadcrumb top bar with search + notifications.
- **Status system** (shared): `working` (On shift), `break` (On break), `late` (Running late), `upcoming` (Upcoming), `done` (Checked out), `off` (Off today), `leave` (On leave).

### Tweaks (live controls)
- **Accent theme** — Orange / Steel Blue / Polar Green / Assist (violet)
- **Dark mode** toggle
- **Live view default layout** — board / list / map
- **Density** — compact / regular / comfy

---

## 3. Data model (demo)

- **Staff (18):** id, name, role, department, employment type, phone, email, location, languages, employee ID, join date, manager, emergency contact, and derived stats (on-time %, hours/week, tours/month, shifts/month).
- **Departments (3):** Field Ops, Operations, Office — each with a lead, color, icon, description, and member roles.
- **Roles (4):** Tour Guide, Driver, Ops, Office — each mapped to a department and an **access level**.
- **Access levels (3):** Mobile, Web · Manager, Web · Admin — each with an explicit permission set.
- **Employment types:** Full-time, Part-time, Seasonal, Contract.
- **Tours (8):** label, location, category, duration, capacity, required crew (guides/drivers), daily departure times, price, rating, active/paused, description.
- **Live assignments:** per-staff current status, tour, check-in time, location, guest count, lateness, photo flag.
- **Weekly roster:** per-staff shift per day (Mon–Sun), each published or draft.
- **Time-off requests (12):** staff, type, date range, days, status (pending/approved/declined), filed time, note, reviewer.
- **Leave allowance:** 22 days/year; approved leave reflected on the roster and live view.

---

## 4. Manager web app

### 4.1 Live View — "Who's working now" (hero)
- **Header:** live pulse, current timestamp, "updated just now," layout switch, refresh.
- **KPI strip (clickable filters):** On shift, Running late, On break, Upcoming, Checked out, plus a non-interactive **On leave** count (fed by approved time off).
- **Three layouts:**
  - **Board** — kanban columns by status (On shift / Running late / On break / Upcoming / Checked out), each card showing staff, tour, location, guests, time-on-shift / lateness / start time, and a check-in photo thumbnail.
  - **List** — sortable table: staff, status, tour, location, check-in, time on shift.
  - **Map** — stylized Lisbon map with color-coded staff pins (initials + status dot) and a side list of in-the-field crew.
- **Detail drawer** (any person): status, tour window, check-in time, location, guests, check-in photo, today's activity timeline, and Message / Call / Reassign actions.

### 4.2 Roster
- **Weekly grid** (Mon–Sun), staff grouped by role (Tour Guides, Drivers), today's column highlighted.
- **Shift chips** colored per tour, showing time; **draft** shifts dashed with a DRAFT tag.
- **Add Shift drawer** — staff, day, start/end, tour picker; new shifts created as drafts.
- **Publish** — publishes all drafts at once with a "field staff notified" toast; unpublished count shown in the header.
- **Approved leave** renders as a hatched **Leave** cell instead of an empty slot.
- Tour color legend below the grid.

### 4.3 Tours
- **Catalog** with **Grid / List** toggle and category filter (Walking, Day Trip, Food, Driving, Evening).
- **Grid card:** color bar, name, location, duration, capacity, required crew, category, rating, price, assigned-crew avatars, departures/day, and a live "N live" badge when crew are out.
- **List view:** sortable table (tour, category, duration, capacity, crew, departures, live, price) with a per-row pencil edit button.
- **Tour detail drawer:** full spec, live "crew on this tour now" banner, today's departures with per-departure status (Running / Completed / Scheduled / Unstaffed), assigned crew, and this week's staff.
- **New / Edit Tour form:** name, category, meeting point, duration, capacity, price, crew (guides/drivers), add/remove departure times, accent color, description, active/paused toggle. Saves update the catalog live.

### 4.4 Directory (people core)
- Department tabs with counts; sortable table (name, role, department, employment, phone, status).
- Live status per person; row click opens the **Employee profile**.
- **Add Member** opens the employee form.

### 4.5 Employee profile
- **Header:** avatar, live status, role · department · employee ID, employment/tenure/language chips, and Message / **Edit** / Assign Shift actions.
- **Live banner** when on shift (tour, check-in, location, guests).
- **Stats:** on-time rate, hours this week, tours/month, shifts/month.
- **This week's schedule** strip (from the roster), today highlighted.
- **Recent activity** timeline (check-in/out, photos, approvals).
- **Contact**, **Employment**, and **Role & Access** cards — Role & Department cross-link to those screens; **Change** opens the edit form.
- Breadcrumb shows People › Directory › {name}.

### 4.6 Employee form (add / edit)
- **Identity:** name, email, phone, location, languages.
- **Role & access:** department + role selectors (changing department auto-filters roles and shows the resulting access level).
- **Employment:** type, reports-to.
- **Emergency contact.**
- Saves propagate everywhere instantly (e.g. moving someone to Operations switches their role to Ops with Web·Manager access). New members get a generated avatar, employee ID, and join date.

### 4.7 Departments
- **Grid / List** toggle.
- **Grid card:** icon, lead, headcount, live on-shift count, roles breakdown, member avatar stack.
- **List view:** sortable (department, lead, people, roles, on-shift) with per-row pencil edit.
- **Detail drawer:** description, lead, roles (cross-link to role), employment mix bar, full member list.
- **New / Edit Department form:** name, description, lead, icon, color.

### 4.8 Roles
- Sortable table (role, department, people, access, permissions) with per-row pencil edit.
- **Detail drawer:** description, access level, full permission list, people with this role.
- **New / Edit Role form:** name, department, access level (with live permissions preview), color, description.

### 4.9 Time Off (approvals)
- **Summary stats:** pending / approved / declined counts, days off booked.
- **Tabs:** Pending, Approved, Declined, All (with counts).
- **Sortable table:** employee, type (Holiday / Sick / Personal), dates, days, filed, status — with inline **approve / decline** on pending rows.
- **Review drawer:** note, filed date, reviewer, employee leave-balance bar, and Approve / Decline (or move back to pending).
- Actions update counts/tabs live; approvals flow into the roster (Leave cells) and Live View (On leave).

---

## 5. Mobile field app

Shown in an iPhone frame with a persistent status bar, app header (logo, notifications bell, avatar), bottom tab bar, and stacked screens. A left-side **screen navigator** jumps to any screen; an **offline-mode** toggle simulates no-signal behavior.

### Tabs
- **Today** — greeting, next shift card (tour, time, location, guests), and the **check-in flow**.
- **My Schedule** — week ahead with day-off rows, NEXT / DRAFT flags, tour colors; row opens shift detail.
- **Inbox** — conversations (Dispatch, manager, guides group) with unread badges.
- **Me (Profile)** — avatar, stats, time-off balance bar + requests, contact info, sign out.

### Check-in flow (Today)
Location confirm (geo within 25 m) → photo capture (geo-tagged) → confirm summary → **on-shift timer** (with Break / Check Out) → shift complete summary. Offline mode shows "saved offline · will sync" states throughout.

### Stacked screens
- **Shift detail** — full brief: window, meeting point, mini-map, guest notes, check-in CTA.
- **Message thread** — chat with a working composer.
- **Notifications** — roster published, time-off approved, meeting-point updates, shift reminders.
- **Request Time Off** — type (Holiday / Sick / Personal), date range, note; routed to the manager for approval.

### Login & onboarding
- **Sign In** — work email + password, forgot password, "sign in with phone" alternative.
- **Onboarding carousel** — 4 slides (shifts in your pocket → check in from the field → works offline → stay in the loop) with progress dots, Skip, and inline location/notification permission toggles. Flow: Sign In → Onboarding → app.

---

## 6. Cross-cutting behaviors

- **Sortable tables** everywhere (Live list, Directory, Departments, Roles, Tours, Time Off) via a shared sort utility with asc/desc carets; numeric columns sort numerically, status by severity.
- **Edit forms** follow one pattern (right-side drawer, local state, color/select controls, confirmation toast); session-persistent (in-memory).
- **Connected data:** one approved time-off record drives the Time Off status, the roster Leave cell, and the Live View "On leave" count + Directory status simultaneously.
- **Responsive & accessible:** flex/grid layouts, reduced-motion support, tidy scrollbars, keyboard-visible focus per Blazeup.

---

## 7. Deliverables

- `Opero.html` — the full prototype (loads modular JS/JSX).
- `Opero (standalone).html` — single self-contained offline file.
- Module files under `opero/` — data, shared UI, app shell, and one file per module (live, roster, tours, people, employee, employee-form, org, timeoff, field, field-screens, field-data).

---

## 8. Not yet built / future

- Real map tiles (Live View map is a stylized abstraction).
- Wiring **Assign Shift** (employee profile) and **delete** actions on table rows.
- Global ⌘K search behavior.
- Shift-swap requests, payroll/export, and reporting.
