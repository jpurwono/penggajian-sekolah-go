-- 000009: Backfill item hasil generate lama agar ikut disinkronkan otomatis
-- (item manual murni tanpa komponen pegawai tidak tersentuh)

UPDATE payslip_items pi
SET is_auto = true
WHERE is_auto = false
  AND (
    pi.component_id IN (SELECT component_id FROM employee_components)
    OR pi.code_snapshot = 'GP'
  );
