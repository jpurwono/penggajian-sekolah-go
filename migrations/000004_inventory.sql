-- 000004: Inventory — kategori, lokasi, item, stok, mutasi, pengadaan

CREATE TABLE inventory_categories (
  id          BIGSERIAL PRIMARY KEY,
  code        VARCHAR(20),
  name        VARCHAR(50) NOT NULL,
  parent_id   BIGINT REFERENCES inventory_categories(id),
  description TEXT,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
  deleted_at  TIMESTAMPTZ
);
CREATE TRIGGER trg_inventory_categories_updated_at BEFORE UPDATE ON inventory_categories
  FOR EACH ROW EXECUTE FUNCTION set_updated_at();
CREATE INDEX idx_inventory_categories_parent_id ON inventory_categories(parent_id);

CREATE TABLE inventory_locations (
  id          BIGSERIAL PRIMARY KEY,
  code        VARCHAR(20),
  name        VARCHAR(100) NOT NULL,
  building    VARCHAR(50),
  floor       VARCHAR(10),
  description TEXT,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
  deleted_at  TIMESTAMPTZ
);
CREATE TRIGGER trg_inventory_locations_updated_at BEFORE UPDATE ON inventory_locations
  FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TABLE inventory_items (
  id          BIGSERIAL PRIMARY KEY,
  code        VARCHAR(30)  NOT NULL UNIQUE,
  name        VARCHAR(100) NOT NULL,
  category_id BIGINT REFERENCES inventory_categories(id),
  unit        VARCHAR(15)  NOT NULL DEFAULT 'buah',
  min_stock   INT          NOT NULL DEFAULT 0,
  description TEXT,
  is_active   BOOLEAN      NOT NULL DEFAULT true,
  created_at  TIMESTAMPTZ  NOT NULL DEFAULT now(),
  updated_at  TIMESTAMPTZ  NOT NULL DEFAULT now(),
  deleted_at  TIMESTAMPTZ
);
CREATE TRIGGER trg_inventory_items_updated_at BEFORE UPDATE ON inventory_items
  FOR EACH ROW EXECUTE FUNCTION set_updated_at();
CREATE INDEX idx_inventory_items_category_id ON inventory_items(category_id);

CREATE TABLE inventory_stocks (
  id          BIGSERIAL PRIMARY KEY,
  item_id     BIGINT      NOT NULL REFERENCES inventory_items(id),
  location_id BIGINT      NOT NULL REFERENCES inventory_locations(id),
  condition   VARCHAR(15) NOT NULL DEFAULT 'baik'
              CONSTRAINT chk_stocks_condition CHECK (condition IN ('baik', 'rusak_ringan', 'rusak_berat')),
  quantity    INT         NOT NULL DEFAULT 0
              CONSTRAINT chk_stocks_quantity CHECK (quantity >= 0),
  created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
  CONSTRAINT uq_stocks_item_location_condition UNIQUE (item_id, location_id, condition)
);
CREATE TRIGGER trg_inventory_stocks_updated_at BEFORE UPDATE ON inventory_stocks
  FOR EACH ROW EXECUTE FUNCTION set_updated_at();
CREATE INDEX idx_inventory_stocks_location_id ON inventory_stocks(location_id);

CREATE TABLE inventory_mutations (
  id               BIGSERIAL PRIMARY KEY,
  mutation_no      VARCHAR(30) NOT NULL UNIQUE,
  type             VARCHAR(15) NOT NULL
                   CONSTRAINT chk_mutations_type CHECK (type IN ('in', 'out', 'transfer', 'repair', 'discard')),
  mutation_date    DATE        NOT NULL,
  item_id          BIGINT      NOT NULL REFERENCES inventory_items(id),
  from_location_id BIGINT REFERENCES inventory_locations(id),
  to_location_id   BIGINT REFERENCES inventory_locations(id),
  from_condition   VARCHAR(15)
                   CONSTRAINT chk_mutations_from_condition CHECK (from_condition IN ('baik', 'rusak_ringan', 'rusak_berat')),
  to_condition     VARCHAR(15)
                   CONSTRAINT chk_mutations_to_condition CHECK (to_condition IN ('baik', 'rusak_ringan', 'rusak_berat')),
  quantity         INT         NOT NULL
                   CONSTRAINT chk_mutations_quantity CHECK (quantity > 0),
  reference_no     VARCHAR(50),
  note             TEXT,
  created_by       BIGINT REFERENCES users(id),
  created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at       TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE TRIGGER trg_inventory_mutations_updated_at BEFORE UPDATE ON inventory_mutations
  FOR EACH ROW EXECUTE FUNCTION set_updated_at();
CREATE INDEX idx_inventory_mutations_item_id ON inventory_mutations(item_id);
CREATE INDEX idx_inventory_mutations_mutation_date ON inventory_mutations(mutation_date);
CREATE INDEX idx_inventory_mutations_type ON inventory_mutations(type);

CREATE TABLE inventory_purchases (
  id           BIGSERIAL PRIMARY KEY,
  purchase_no  VARCHAR(30) NOT NULL UNIQUE,
  purchase_date DATE       NOT NULL,
  supplier_name VARCHAR(100),
  description  TEXT,
  total_cost   BIGINT      NOT NULL DEFAULT 0,
  status       VARCHAR(10) NOT NULL DEFAULT 'draft'
               CONSTRAINT chk_purchases_status CHECK (status IN ('draft', 'received')),
  created_by   BIGINT REFERENCES users(id),
  created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE TRIGGER trg_inventory_purchases_updated_at BEFORE UPDATE ON inventory_purchases
  FOR EACH ROW EXECUTE FUNCTION set_updated_at();
CREATE INDEX idx_inventory_purchases_status ON inventory_purchases(status);
CREATE INDEX idx_inventory_purchases_purchase_date ON inventory_purchases(purchase_date);

CREATE TABLE inventory_purchase_items (
  id          BIGSERIAL PRIMARY KEY,
  purchase_id BIGINT NOT NULL REFERENCES inventory_purchases(id) ON DELETE CASCADE,
  item_id     BIGINT NOT NULL REFERENCES inventory_items(id),
  quantity    INT    NOT NULL
              CONSTRAINT chk_purchase_items_quantity CHECK (quantity > 0),
  unit_price  BIGINT NOT NULL DEFAULT 0,
  subtotal    BIGINT NOT NULL DEFAULT 0,
  location_id BIGINT NOT NULL REFERENCES inventory_locations(id),
  created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE TRIGGER trg_inventory_purchase_items_updated_at BEFORE UPDATE ON inventory_purchase_items
  FOR EACH ROW EXECUTE FUNCTION set_updated_at();
CREATE INDEX idx_inventory_purchase_items_purchase_id ON inventory_purchase_items(purchase_id);
CREATE INDEX idx_inventory_purchase_items_item_id ON inventory_purchase_items(item_id);
