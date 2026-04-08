export interface Item {
  id: string;
  sku: string;
  name: string;
  description: string;
  physical_stock: number;
  available_stock: number;
  created_at: string;
  updated_at: string;
}

export interface Customer {
  id: string;
  name: string;
  email: string;
  phone: string;
  address: string;
  created_at: string;
  updated_at: string;
}

export interface StockInTransaction {
  id: string;
  reference_no: string;
  status: string;
  notes: string;
  created_by: string;
  items?: StockInItem[];
  logs?: TransactionLog[];
  created_at: string;
  updated_at: string;
}

export interface StockInItem {
  id: string;
  stock_in_transaction_id: string;
  item_id: string;
  quantity: number;
  item_name?: string;
  item_sku?: string;
  created_at: string;
}

export interface StockOutTransaction {
  id: string;
  reference_no: string;
  customer_id: string | null;
  customer_name?: string;
  status: string;
  notes: string;
  created_by: string;
  items?: StockOutItem[];
  logs?: TransactionLog[];
  created_at: string;
  updated_at: string;
}

export interface StockOutItem {
  id: string;
  stock_out_transaction_id: string;
  item_id: string;
  quantity: number;
  item_name?: string;
  item_sku?: string;
  created_at: string;
}

export interface TransactionLog {
  id: string;
  transaction_type: string;
  transaction_id: string;
  from_status: string;
  to_status: string;
  notes: string;
  created_by: string;
  created_at: string;
}

export interface APIResponse<T> {
  success: boolean;
  data: T;
  meta?: { page: number; limit: number; total: number };
  error: string | null;
}
