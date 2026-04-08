export const StockInStatus = {
  CREATED: 'CREATED',
  IN_PROGRESS: 'IN_PROGRESS',
  DONE: 'DONE',
  CANCELLED: 'CANCELLED',
} as const;

export const StockOutStatus = {
  DRAFT: 'DRAFT',
  IN_PROGRESS: 'IN_PROGRESS',
  DONE: 'DONE',
  CANCELLED: 'CANCELLED',
} as const;

export const statusColors: Record<string, string> = {
  CREATED: 'bg-blue-100 text-blue-800',
  DRAFT: 'bg-gray-100 text-gray-800',
  IN_PROGRESS: 'bg-amber-100 text-amber-800',
  DONE: 'bg-green-100 text-green-800',
  CANCELLED: 'bg-red-100 text-red-800',
};

export const nextStockInStatus: Record<string, string | null> = {
  CREATED: 'IN_PROGRESS',
  IN_PROGRESS: 'DONE',
  DONE: null,
  CANCELLED: null,
};

export const nextStockOutStatus: Record<string, string | null> = {
  DRAFT: 'IN_PROGRESS',
  IN_PROGRESS: 'DONE',
  DONE: null,
  CANCELLED: null,
};
