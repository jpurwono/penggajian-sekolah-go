-- 000015: Payroll — template komponen gaji, otomatis diterapkan ke pegawai baru

CREATE TABLE payroll_templates (
    id            BIGSERIAL PRIMARY KEY,
    name          VARCHAR(100) NOT NULL,
    description   TEXT,
    employee_type VARCHAR(20)  NULL
                  CONSTRAINT chk_payroll_templates_etype CHECK (
                      employee_type IS NULL OR employee_type IN ('tetap', 'pppk', 'honorer', 'staff_tu')
                  ),
    is_default    BOOLEAN      NOT NULL DEFAULT false,
    is_active     BOOLEAN      NOT NULL DEFAULT true,
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ  NOT NULL DEFAULT now(),
    deleted_at    TIMESTAMPTZ
);
CREATE TRIGGER trg_payroll_templates_updated_at BEFORE UPDATE ON payroll_templates
  FOR EACH ROW EXECUTE FUNCTION set_updated_at();
CREATE INDEX idx_payroll_templates_deleted_at ON payroll_templates(deleted_at);
-- Satu default aktif per tipe pegawai (NULL = semua tipe)
CREATE UNIQUE INDEX uq_payroll_templates_default
  ON payroll_templates(COALESCE(employee_type, '*'))
  WHERE is_default AND is_active AND deleted_at IS NULL;

CREATE TABLE payroll_template_components (
    id           BIGSERIAL PRIMARY KEY,
    template_id  BIGINT       NOT NULL REFERENCES payroll_templates(id) ON DELETE CASCADE,
    component_id BIGINT       NOT NULL REFERENCES pay_components(id),
    calc_mode    VARCHAR(20)  NOT NULL DEFAULT 'fixed'
                 CONSTRAINT chk_payroll_tpl_comp_mode CHECK (
                     calc_mode IN ('fixed', 'per_jam_hadir', 'per_jam_mengajar', 'per_hari_hadir', 'per_jam_minggu1')
                 ),
    amount       BIGINT       NOT NULL DEFAULT 0,
    note         TEXT,
    created_at   TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ  NOT NULL DEFAULT now()
);
CREATE TRIGGER trg_payroll_template_components_updated_at BEFORE UPDATE ON payroll_template_components
  FOR EACH ROW EXECUTE FUNCTION set_updated_at();
CREATE INDEX idx_payroll_tpl_comp_template ON payroll_template_components(template_id);
CREATE UNIQUE INDEX uq_payroll_tpl_comp ON payroll_template_components(template_id, component_id);
