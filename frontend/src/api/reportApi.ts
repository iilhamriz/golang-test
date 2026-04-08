import api from './client';
import type { APIResponse } from '../types';

export const getReportTransactions = (params?: { type?: string; page?: number; limit?: number }) =>
  api.get<APIResponse<any[]>>('/reports/transactions', { params });

export const getReportDetail = (type: string, id: string) =>
  api.get<APIResponse<any>>(`/reports/transactions/${type}/${id}`);
