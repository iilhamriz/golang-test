import api from './client';
import type { APIResponse, Item } from '../types';

export const getItems = (params?: { name?: string; sku?: string; customer_id?: string; page?: number; limit?: number }) =>
  api.get<APIResponse<Item[]>>('/items', { params });

export const getItem = (id: string) =>
  api.get<APIResponse<Item>>(`/items/${id}`);

export const createItem = (data: { sku: string; name: string; description?: string }) =>
  api.post<APIResponse<Item>>('/items', data);

export const updateItem = (id: string, data: { sku?: string; name?: string; description?: string }) =>
  api.put<APIResponse<Item>>(`/items/${id}`, data);

export const adjustStock = (id: string, data: { quantity: number; reason: string }) =>
  api.post<APIResponse<Item>>(`/items/${id}/adjust`, data);
