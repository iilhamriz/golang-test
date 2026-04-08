# Smart Inventory Core System

Sistem manajemen stok dengan Backend Go dan Frontend React.

## Arsitektur

### Backend (Go)
- **Clean Architecture**: Domain → Use Case → Adapter → Infrastructure
- **Database**: PostgreSQL dengan `pgx/v5` driver
- **Router**: `go-chi/chi` (lightweight, stdlib-compatible)
- **SOLID & DRY**: Repository interfaces di domain layer, implementasi di adapter layer
- **Concurrency-safe**: `SELECT ... FOR UPDATE` untuk stock allocation dan state transitions
- **Two-Phase Commit**: Stock Out menggunakan structural reservation (DRAFT/IN_PROGRESS items implicitly reserve stock)

### Frontend (React)
- **React 18** + TypeScript + Vite
- **Zustand** untuk state management (lightweight)
- **Tailwind CSS** untuk styling
- **React Router v6** untuk routing
- **Axios** untuk HTTP client

### Key Design Decisions
1. **Available Stock = Physical Stock - Reserved**: Dihitung via SQL VIEW (`v_item_stock`), bukan kolom terpisah. Menghindari dual-write inconsistency.
2. **Structural Reservation**: Stock Out items di status DRAFT/IN_PROGRESS otomatis "reserve" stock. Cancel = hapus reservasi tanpa mutasi tambahan.
3. **State Machine**: Transisi status divalidasi di use case layer dengan allowed transitions map.
4. **Audit Trail**: Setiap perubahan status tercatat di `transaction_logs` table.

## Cara Menjalankan

### Prerequisites
- Go 1.22+
- PostgreSQL 14+
- Node.js 18+
- npm

### Database
```bash
createdb smart_inventory
```

### Backend
```bash
cd backend
cp .env.example .env
# Edit .env sesuai konfigurasi PostgreSQL anda
go mod tidy
make run
# Server berjalan di http://localhost:8080
```

### Frontend
```bash
cd frontend
npm install
npm run dev
# UI berjalan di http://localhost:5173
```

## API Endpoints

| Method | Path | Deskripsi |
|--------|------|-----------|
| POST | /api/v1/stock-in | Create stock in (CREATED) |
| GET | /api/v1/stock-in | List stock in transactions |
| GET | /api/v1/stock-in/:id | Detail stock in |
| PATCH | /api/v1/stock-in/:id/status | Update status |
| POST | /api/v1/stock-in/:id/cancel | Cancel |
| POST | /api/v1/stock-out | Create stock out draft (DRAFT) |
| GET | /api/v1/stock-out | List stock out transactions |
| GET | /api/v1/stock-out/:id | Detail stock out |
| PATCH | /api/v1/stock-out/:id/status | Update status |
| POST | /api/v1/stock-out/:id/cancel | Cancel (rollback reservation) |
| POST | /api/v1/items | Create item |
| GET | /api/v1/items | List items (filter: name, sku, customer_id) |
| GET | /api/v1/items/:id | Detail item |
| PUT | /api/v1/items/:id | Update item |
| POST | /api/v1/items/:id/adjust | Stock adjustment |
| POST | /api/v1/customers | Create customer |
| GET | /api/v1/customers | List customers |
| GET | /api/v1/reports/transactions | List DONE transactions |
| GET | /api/v1/reports/transactions/:type/:id | Detail report |
