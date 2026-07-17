#!/usr/bin/env bash
# Re-seed the demo tenant's Live View data, anchored to the current time.
#
# Replaces all shifts tagged notes='Seeded demo shift' (and their attendance)
# with a fresh set for every active employee, spread over three windows so the
# live view shows a realistic mix right now:
#   rows 1-3: started 6h ago, ends in +2h
#   rows 4-6: started 3h ago, ends in +5h
#   rows 7-9: starts in +2h (upcoming)
# Attendance: 1 checked out, 1 on break (started 25m ago), 3 checked in,
# 1 no-show (started, no record); upcoming rows get none. The first six
# shifts alternate across the demo tours so "Group by tour" has content.
#
# Usage: ./scripts/seed-demo-liveview.sh  (needs the opero-postgres container)
set -euo pipefail

docker exec -i opero-postgres psql -U opero -d opero_tenant_demo <<'SQL'
BEGIN;
DELETE FROM attendance_records
WHERE shift_id IN (SELECT id FROM shifts WHERE notes = 'Seeded demo shift');
DELETE FROM shifts WHERE notes = 'Seeded demo shift';

WITH emp AS (
  SELECT id, row_number() OVER (ORDER BY full_name) AS rn
  FROM employees WHERE status = 'active'
), loc AS (
  SELECT id, row_number() OVER (ORDER BY name) AS rn FROM locations
), tur AS (
  SELECT id, row_number() OVER (ORDER BY name) AS rn FROM tours
), base AS (
  SELECT date_trunc('hour', now()) AS h
), new_shifts AS (
  INSERT INTO shifts (employee_id, location_id, tour_id, starts_at, ends_at, notes, status)
  SELECT e.id, l.id,
         CASE WHEN e.rn <= 6 AND (SELECT count(*) FROM tur) > 0
              THEN (SELECT id FROM tur WHERE rn = ((e.rn - 1) % (SELECT count(*) FROM tur)) + 1)
         END,
         CASE WHEN e.rn <= 3 THEN b.h - interval '6 hours'
              WHEN e.rn <= 6 THEN b.h - interval '3 hours'
              ELSE                b.h + interval '2 hours' END,
         CASE WHEN e.rn <= 3 THEN b.h + interval '2 hours'
              WHEN e.rn <= 6 THEN b.h + interval '5 hours'
              ELSE                b.h + interval '10 hours' END,
         'Seeded demo shift',
         'published'
  FROM emp e CROSS JOIN base b
  JOIN loc l ON l.rn = ((e.rn - 1) % (SELECT count(*) FROM loc)) + 1
  RETURNING id, employee_id, location_id, starts_at
)
INSERT INTO attendance_records
  (employee_id, shift_id, client_id, check_in_at, check_in_lat, check_in_lng,
   check_out_at, check_out_lat, check_out_lng, break_started_at, status)
SELECT s.employee_id, s.id, gen_random_uuid(),
       s.starts_at + (interval '1 minute' * (random()*10)::int),
       l.lat + (random()-0.5)/1000, l.lng + (random()-0.5)/1000,
       CASE WHEN q.rn = 1 THEN now() - interval '20 minutes' END,
       CASE WHEN q.rn = 1 THEN l.lat END,
       CASE WHEN q.rn = 1 THEN l.lng END,
       CASE WHEN q.rn = 2 THEN now() - interval '25 minutes' END,
       CASE WHEN q.rn = 1 THEN 'checked_out'
            WHEN q.rn = 2 THEN 'on_break'
            ELSE 'checked_in' END
FROM (SELECT *, row_number() OVER (ORDER BY starts_at, employee_id) AS rn FROM new_shifts) q
JOIN new_shifts s ON s.id = q.id
LEFT JOIN locations l ON l.id = s.location_id
WHERE q.rn <= 5
  AND s.starts_at <= now();
COMMIT;

SELECT e.full_name,
       to_char(s.starts_at, 'HH24:MI') AS starts,
       to_char(s.ends_at, 'HH24:MI') AS ends,
       coalesce(t.name, '-') AS tour,
       coalesce(a.status, '(none)') AS attendance,
       to_char(a.break_started_at, 'HH24:MI') AS break_at
FROM shifts s
JOIN employees e ON e.id = s.employee_id
LEFT JOIN tours t ON t.id = s.tour_id
LEFT JOIN attendance_records a ON a.shift_id = s.id
WHERE s.notes = 'Seeded demo shift'
ORDER BY s.starts_at, e.full_name;
SQL
