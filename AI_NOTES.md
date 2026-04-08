# AI Usage Report

## AI Tools yang Digunakan
- **Claude Code** (Anthropic Claude Opus 4.6) — digunakan untuk generate kode backend dan frontend.
- **Cursor** (Agent) — digunakan untuk membantu memahami struktur file dan kode yang ada dan membantu debugging.

## Prompt Paling Kompleks

```
Build the complete Go backend for a Smart Inventory Core System.
Business Rules:
- Stock In: CREATED -> IN_PROGRESS -> DONE. Cancel only before DONE. Physical stock increases ONLY when status becomes DONE.
- Stock Out: Two-Phase Commit. Phase 1 = Create DRAFT (check available stock, reserve via structural stock_out_items). Phase 2 = DRAFT -> IN_PROGRESS -> DONE. When DONE decrease physical_stock. Cancel at DRAFT/IN_PROGRESS = just change status (reservation released automatically). Use SELECT...FOR UPDATE for concurrency.
- Available Stock = Physical Stock - SUM(reserved in DRAFT/IN_PROGRESS stock_out_items). Computed via SQL VIEW, not stored column.
```

## Kode yang Dimodifikasi Manual (Best Practice)

### 1. Database Transaction Handling di Stock Out Use Case

AI awalnya generate `CreateDraft()` tanpa `SELECT ... FOR UPDATE` pada item rows saat mengecek available stock. Ini berbahaya karena dua request concurrent bisa membaca available stock yang sama dan keduanya lolos pengecekan, mengakibatkan over-allocation.

**Modifikasi manual**: Menambahkan `GetByIDForUpdate()` (row-level lock) pada setiap item sebelum mengecek available stock di dalam database transaction:

```go
// Di stock_out_usecase.go CreateDraft():
for _, it := range input.Items {
    // Lock item row to prevent concurrent allocation
    _, err := u.itemRepo.GetByIDForUpdate(ctx, tx, it.ItemID)
    if err != nil {
        return nil, fmt.Errorf("item %s not found", it.ItemID)
    }
    avail, err := u.itemRepo.GetAvailableStock(ctx, tx, it.ItemID)
    // ...
}
```

Tanpa row lock ini, race condition bisa terjadi saat dua stock-out draft dibuat bersamaan untuk item yang sama — keduanya bisa lolos pengecekan stock dan menyebabkan available stock menjadi negatif.

### 2. Available Stock sebagai SQL VIEW vs Stored Column

AI awalnya menyarankan `available_stock` sebagai kolom tersimpan di tabel `items` yang di-update setiap kali ada transaksi. Ini rentan terhadap inconsistency jika ada bug atau crash di tengah update.

**Modifikasi manual**: Mengubah menjadi SQL VIEW `v_item_stock` yang menghitung available stock secara real-time:

```sql
CREATE VIEW v_item_stock AS
SELECT i.*,
    i.physical_stock - COALESCE(reserved.total, 0) AS available_stock
FROM items i
LEFT JOIN (
    SELECT soi.item_id, SUM(soi.quantity) AS total
    FROM stock_out_items soi
    JOIN stock_out_transactions sot ON sot.id = soi.stock_out_transaction_id
    WHERE sot.status IN ('DRAFT', 'IN_PROGRESS')
    GROUP BY soi.item_id
) reserved ON reserved.item_id = i.id;
```

Dengan pendekatan ini, available stock selalu konsisten dan tidak bisa out-of-sync. Cancel stock-out otomatis "melepas" reservasi tanpa perlu mutasi tambahan.
