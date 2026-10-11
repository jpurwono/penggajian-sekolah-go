-- 000011: Mode hitung tambahan — tarif × jam mengajar minggu ke-1

ALTER TABLE employee_components DROP CONSTRAINT chk_employee_components_calc_mode;

ALTER TABLE employee_components
  ADD CONSTRAINT chk_employee_components_calc_mode
  CHECK (calc_mode IN ('fixed', 'per_jam_hadir', 'per_jam_mengajar', 'per_hari_hadir', 'per_jam_minggu1'));
