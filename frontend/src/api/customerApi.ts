import api from './client';
import type { APIResponse, Customer } from '../types';

export const getCustomers = (params?: { page?: number; limit?: number }) =>
  api.get<APIResponse<Customer[]>>('/customers', { params });

export const getCustomer = (id: string) =>
  api.get<APIResponse<Customer>>(`/customers/${id}`);

export const createCustomer = (data: { name: string; email?: string; phone?: string; address?: string }) =>
  api.post<APIResponse<Customer>>('/customers', data);
