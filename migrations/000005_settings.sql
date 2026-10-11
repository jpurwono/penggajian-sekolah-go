-- 000005: Pendukung — profil sekolah (singleton) & audit log (immutable)

CREATE TABLE school_profiles (
  id             BIGSERIAL PRIMARY KEY,
  school_name    VARCHAR(100) NOT NULL,
  address        TEXT,
  phone          VARCHAR(20),
  email          VARCHAR(100),
  principal_name VARCHAR(100),
  logo_url       VARCHAR(255),
  created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE TRIGGER trg_school_profiles_updated_at BEFORE UPDATE ON school_profiles
  FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TABLE audit_logs (
  id         BIGSERIAL PRIMARY KEY,
  user_id    BIGINT REFERENCES users(id) ON DELETE SET NULL,
  action     VARCHAR(20) NOT NULL,
  entity     VARCHAR(50) NOT NULL,
  entity_id  BIGINT      NOT NULL,
  old_values JSONB,
  new_values JSONB,
  ip         VARCHAR(45),
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_audit_logs_entity ON audit_logs(entity, entity_id);
CREATE INDEX idx_audit_logs_created_at ON audit_logs(created_at);
