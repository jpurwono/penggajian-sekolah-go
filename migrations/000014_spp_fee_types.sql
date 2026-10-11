-- 000014: SPP — master jenis biaya & generate tahunan multi-komponen

CREATE TABLE spp_fee_types (
    id             BIGSERIAL PRIMARY KEY,
    code           VARCHAR(20)  NOT NULL UNIQUE,
    name           VARCHAR(100) NOT NULL,
    frequency      VARCHAR(10)  NOT NULL DEFAULT 'monthly'
                   CONSTRAINT chk_spp_fee_types_freq CHECK (frequency IN ('monthly', 'annual')),
    default_amount BIGINT       NOT NULL DEFAULT 0,
    description    TEXT,
    is_active      BOOLEAN      NOT NULL DEFAULT true,
    created_at     TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at     TIMESTAMPTZ  NOT NULL DEFAULT now(),
    deleted_at     TIMESTAMPTZ
);
CREATE TRIGGER trg_spp_fee_types_updated_at BEFORE UPDATE ON spp_fee_types
  FOR EACH ROW EXECUTE FUNCTION set_updated_at();

ALTER TABLE spp_bills ADD COLUMN fee_type_id BIGINT REFERENCES spp_fee_types(id);
CREATE INDEX idx_spp_bills_fee_type_id ON spp_bills(fee_type_id);

-- Backfill: tagihan SPP bulanan yang sudah ada → fee_type SPP
INSERT INTO spp_fee_types (code, name, frequency, default_amount, description)
VALUES
  ('SPP', 'SPP Bulanan', 'monthly', 0, 'Sumbangan Pembinaan Pendidikan bulanan'),
  ('BUKU', 'Uang Buku', 'annual', 0, 'Pembelian buku pelajaran'),
  ('SERAGAM', 'Uang Seragam', 'annual', 0, 'Pembelian seragam sekolah'),
  ('KEGIATAN', 'Uang Kegiatan', 'annual', 0, 'Kegiatan siswa / ekstrakurikuler'),
  ('DAFTAR', 'Uang Pendaftaran', 'annual', 0, 'Biaya pendaftaran siswa baru'),
  ('LAIN', 'Biaya Lain-lain', 'annual', 0, 'Biaya lainnya')
ON CONFLICT (code) DO NOTHING;

-- Tagihan existing yang namanya diawali "SPP " → link ke fee_type SPP
UPDATE spp_bills SET fee_type_id = (SELECT id FROM spp_fee_types WHERE code = 'SPP')
WHERE name ILIKE 'SPP %' AND fee_type_id IS NULL;
