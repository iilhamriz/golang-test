import api from './client';
import type { APIResponse, StockOutTransaction } from '../types';

export const getStockOuts = (params?: { status?: string; page?: number; limit?: number }) =>
  api.get<APIResponse<StockOutTransaction[]>>('/stock-out', { params });

export const getStockOut = (id: string) =>
  api.get<APIResponse<StockOutTransaction>>(`/stock-out/${id}`);

export const createStockOut = (data: { reference_no: string; customer_id?: string; notes?: string; created_by?: string; items: { item_id: string; quantity: number }[] }) =>
  api.post<APIResponse<StockOutTransaction>>('/stock-out', data);

export const updateStockOutStatus = (id: string, status: string) =>
  api.patch<APIResponse<StockOutTransaction>>(`/stock-out/${id}/status`, { status });

export const cancelStockOut = (id: string) =>
  api.post<APIResponse<StockOutTransaction>>(`/stock-out/${id}/cancel`);
