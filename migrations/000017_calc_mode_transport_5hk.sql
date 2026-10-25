-- 000017: Mode hitung transport guru — tarif × absen kelas 5 hari kerja pertama

ALTER TABLE employee_components DROP CONSTRAINT chk_employee_components_calc_mode;

ALTER TABLE employee_components
  ADD CONSTRAINT chk_employee_components_calc_mode
  CHECK (calc_mode IN ('fixed', 'per_jam_hadir', 'per_jam_mengajar', 'per_hari_hadir', 'per_jam_minggu1', 'transport_5hk'));
