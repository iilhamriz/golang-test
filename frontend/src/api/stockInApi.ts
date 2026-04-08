import api from './client';
import type { APIResponse, StockInTransaction } from '../types';

export const getStockIns = (params?: { status?: string; page?: number; limit?: number }) =>
  api.get<APIResponse<StockInTransaction[]>>('/stock-in', { params });

export const getStockIn = (id: string) =>
  api.get<APIResponse<StockInTransaction>>(`/stock-in/${id}`);

export const createStockIn = (data: { reference_no: string; notes?: string; created_by?: string; items: { item_id: string; quantity: number }[] }) =>
  api.post<APIResponse<StockInTransaction>>('/stock-in', data);

export const updateStockInStatus = (id: string, status: string) =>
  api.patch<APIResponse<StockInTransaction>>(`/stock-in/${id}/status`, { status });

export const cancelStockIn = (id: string) =>
  api.post<APIResponse<StockInTransaction>>(`/stock-in/${id}/cancel`);
