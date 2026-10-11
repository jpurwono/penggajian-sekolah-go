-- 000010: Hari kerja sekolah & minggu efektif kalender untuk perhitungan jam

ALTER TABLE school_profiles
  ADD COLUMN working_days SMALLINT NOT NULL DEFAULT 5
  CONSTRAINT chk_school_profiles_working_days CHECK (working_days IN (5, 6));

ALTER TABLE payroll_periods
  ADD COLUMN effective_weeks NUMERIC(5,2);
