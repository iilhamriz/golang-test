import { useEffect, useState } from 'react';
import { getReportTransactions, getReportDetail } from '../api/reportApi';
import StatusBadge from '../components/common/StatusBadge';
import { formatDate } from '../utils/formatters';
import toast from 'react-hot-toast';

export default function ReportPage() {
  const [txns, setTxns] = useState<any[]>([]);
  const [detail, setDetail] = useState<any>(null);
  const [typeFilter, setTypeFilter] = useState('');

  const fetchList = async () => {
    try {
      const res = await getReportTransactions({ type: typeFilter || undefined, limit: 50 });
      setTxns(res.data.data || []);
    } catch { toast.error('Failed to load'); }
  };

  useEffect(() => { fetchList(); }, [typeFilter]);

  const loadDetail = async (type: string, id: string) => {
    try {
      const t = type === 'STOCK_IN' ? 'stock-in' : 'stock-out';
      const res = await getReportDetail(t, id);
      setDetail(res.data.data);
    } catch { toast.error('Failed to load detail'); }
  };

  return (
    <div>
      <h2 className="text-2xl font-bold mb-6">Reports (Completed Transactions)</h2>

      <div className="flex gap-3 mb-4">
        <select value={typeFilter} onChange={e => setTypeFilter(e.target.value)} className="border rounded-lg px-3 py-2 text-sm">
          <option value="">All Types</option>
          <option value="stock-in">Stock In</option>
          <option value="stock-out">Stock Out</option>
        </select>
      </div>

      <div className="grid grid-cols-3 gap-6">
        <div className="col-span-2 bg-white rounded-lg shadow overflow-hidden">
          <table className="w-full text-sm">
            <thead className="bg-gray-50">
              <tr>
                <th className="px-4 py-3 text-left">Type</th>
                <th className="px-4 py-3 text-left">Reference</th>
                <th className="px-4 py-3 text-left">Status</th>
                <th className="px-4 py-3 text-left">Date</th>
              </tr>
            </thead>
            <tbody className="divide-y">
              {txns.map(t => (
                <tr key={t.id} onClick={() => loadDetail(t.type, t.id)} className="hover:bg-blue-50 cursor-pointer">
                  <td className="px-4 py-3"><span className={`px-2 py-0.5 rounded text-xs ${t.type === 'STOCK_IN' ? 'bg-blue-50 text-blue-700' : 'bg-orange-50 text-orange-700'}`}>{t.type.replace('_', ' ')}</span></td>
                  <td className="px-4 py-3 font-medium">{t.reference_no}</td>
                  <td className="px-4 py-3"><StatusBadge status={t.status} /></td>
                  <td className="px-4 py-3 text-gray-500">{formatDate(t.updated_at)}</td>
                </tr>
              ))}
              {txns.length === 0 && <tr><td colSpan={4} className="px-4 py-8 text-center text-gray-500">No completed transactions</td></tr>}
            </tbody>
          </table>
        </div>

        <div className="bg-white rounded-lg shadow p-4">
          {detail ? (
            <div>
              <h3 className="font-semibold text-lg mb-4">Transaction Detail</h3>

              <h4 className="font-medium text-sm mb-2">Items</h4>
              <div className="space-y-1 mb-4">
                {(detail.transaction?.items || []).map((it: any) => (
                  <div key={it.id} className="flex justify-between text-sm bg-gray-50 px-3 py-2 rounded">
                    <span>{it.item_name || it.item_sku || it.item_id}</span>
                    <span className="font-medium">x{it.quantity}</span>
                  </div>
                ))}
              </div>

              <h4 className="font-medium text-sm mb-2">Audit Log</h4>
              <div className="space-y-1">
                {(detail.logs || []).map((log: any) => (
                  <div key={log.id} className="text-xs text-gray-500 bg-gray-50 px-3 py-2 rounded">
                    <div>{log.from_status || '(new)'} → {log.to_status}</div>
                    <div>{log.notes}</div>
                    <div className="text-gray-400">{formatDate(log.created_at)}</div>
                  </div>
                ))}
              </div>
            </div>
          ) : (
            <p className="text-gray-500 text-sm">Select a transaction to view report detail</p>
          )}
        </div>
      </div>
    </div>
  );
}
