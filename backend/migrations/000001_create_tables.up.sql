-- customers
CREATE TABLE IF NOT EXISTS customers (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name        VARCHAR(255) NOT NULL,
    email       VARCHAR(255),
    phone       VARCHAR(50),
    address     TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- items
CREATE TABLE IF NOT EXISTS items (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    sku             VARCHAR(100) NOT NULL UNIQUE,
    name            VARCHAR(255) NOT NULL,
    description     TEXT,
    physical_stock  INTEGER NOT NULL DEFAULT 0 CHECK (physical_stock >= 0),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_items_sku ON items(sku);
CREATE INDEX IF NOT EXISTS idx_items_name ON items(name);

-- stock_in_transactions
CREATE TABLE IF NOT EXISTS stock_in_transactions (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    reference_no    VARCHAR(100) NOT NULL UNIQUE,
    status          VARCHAR(20) NOT NULL DEFAULT 'CREATED'
                    CHECK (status IN ('CREATED', 'IN_PROGRESS', 'DONE', 'CANCELLED')),
    notes           TEXT,
    created_by      VARCHAR(255),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- stock_in_items
CREATE TABLE IF NOT EXISTS stock_in_items (
    id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    stock_in_transaction_id UUID NOT NULL REFERENCES stock_in_transactions(id) ON DELETE CASCADE,
    item_id                 UUID NOT NULL REFERENCES items(id),
    quantity                INTEGER NOT NULL CHECK (quantity > 0),
    created_at              TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_stock_in_items_txn ON stock_in_items(stock_in_transaction_id);

-- stock_out_transactions
CREATE TABLE IF NOT EXISTS stock_out_transactions (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    reference_no    VARCHAR(100) NOT NULL UNIQUE,
    customer_id     UUID REFERENCES customers(id),
    status          VARCHAR(20) NOT NULL DEFAULT 'DRAFT'
                    CHECK (status IN ('DRAFT', 'IN_PROGRESS', 'DONE', 'CANCELLED')),
    notes           TEXT,
    created_by      VARCHAR(255),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- stock_out_items
CREATE TABLE IF NOT EXISTS stock_out_items (
    id                       UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    stock_out_transaction_id UUID NOT NULL REFERENCES stock_out_transactions(id) ON DELETE CASCADE,
    item_id                  UUID NOT NULL REFERENCES items(id),
    quantity                 INTEGER NOT NULL CHECK (quantity > 0),
    created_at               TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_stock_out_items_txn ON stock_out_items(stock_out_transaction_id);

-- transaction_logs (audit history)
CREATE TABLE IF NOT EXISTS transaction_logs (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    transaction_type  VARCHAR(10) NOT NULL CHECK (transaction_type IN ('STOCK_IN', 'STOCK_OUT')),
    transaction_id    UUID NOT NULL,
    from_status       VARCHAR(20),
    to_status         VARCHAR(20) NOT NULL,
    notes             TEXT,
    created_by        VARCHAR(255),
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_transaction_logs_txn ON transaction_logs(transaction_type, transaction_id);

-- stock_adjustments
CREATE TABLE IF NOT EXISTS stock_adjustments (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    item_id     UUID NOT NULL REFERENCES items(id),
    quantity    INTEGER NOT NULL,
    reason      TEXT NOT NULL,
    created_by  VARCHAR(255),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- view: available stock = physical - reserved (DRAFT/IN_PROGRESS stock out)
CREATE OR REPLACE VIEW v_item_stock AS
SELECT
    i.id,
    i.sku,
    i.name,
    i.description,
    i.physical_stock,
    i.physical_stock - COALESCE(reserved.total, 0) AS available_stock,
    i.created_at,
    i.updated_at
FROM items i
LEFT JOIN (
    SELECT soi.item_id, SUM(soi.quantity) AS total
    FROM stock_out_items soi
    JOIN stock_out_transactions sot ON sot.id = soi.stock_out_transaction_id
    WHERE sot.status IN ('DRAFT', 'IN_PROGRESS')
    GROUP BY soi.item_id
) reserved ON reserved.item_id = i.id;
