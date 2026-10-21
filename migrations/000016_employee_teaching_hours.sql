-- Jumlah jam mengajar pegawai (per bulan): dipakai sebagai kuantitas
-- komponen honor mengajar per jam (calc mode per_jam_mengajar) saat
-- data jam pada slip belum tersedia.
ALTER TABLE employees ADD COLUMN IF NOT EXISTS teaching_hours INT;

COMMENT ON COLUMN employees.teaching_hours IS 'Jumlah jam mengajar per bulan; qty default honor mengajar per jam';
