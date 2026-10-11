-- 000008: Komponen berbasis tarif (potongan otomatis mengikuti jam hadir)

ALTER TABLE employee_components
  ADD COLUMN calc_mode VARCHAR(20) NOT NULL DEFAULT 'fixed'
  CONSTRAINT chk_employee_components_calc_mode
  CHECK (calc_mode IN ('fixed', 'per_jam_hadir', 'per_jam_mengajar', 'per_hari_hadir'));

ALTER TABLE payslip_items
  ADD COLUMN is_auto BOOLEAN NOT NULL DEFAULT false;
