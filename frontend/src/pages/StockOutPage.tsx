import { useEffect, useState } from 'react';
import { getStockOuts, getStockOut, createStockOut, updateStockOutStatus, cancelStockOut } from '../api/stockOutApi';
import { getItems } from '../api/inventoryApi';
import { getCustomers } from '../api/customerApi';
import StatusBadge from '../components/common/StatusBadge';
import ConfirmDialog from '../components/common/ConfirmDialog';
import { formatDate, formatNumber } from '../utils/formatters';
import { nextStockOutStatus } from '../types/enums';
import type { StockOutTransaction, Item, Customer } from '../types';
import toast from 'react-hot-toast';

export default function StockOutPage() {
  const [txns, setTxns] = useState<StockOutTransaction[]>([]);
  const [selected, setSelected] = useState<StockOutTransaction | null>(null);
  const [showCreate, setShowCreate] = useState(false);
  const [showCancel, setShowCancel] = useState(false);
  const [items, setItems] = useState<Item[]>([]);
  const [customers, setCustomers] = useState<Customer[]>([]);
  const [form, setForm] = useState({ reference_no: '', customer_id: '', notes: '', items: [{ item_id: '', quantity: 1 }] });

  const fetchList = async () => {
    try {
      const res = await getStockOuts({ limit: 50 });
      setTxns(res.data.data || []);
    } catch { toast.error('Failed to load'); }
  };

  const fetchDetail = async (id: string) => {
    try {
      const res = await getStockOut(id);
      setSelected(res.data.data);
    } catch { toast.error('Failed to load detail'); }
  };

  useEffect(() => {
    fetchList();
    getItems({ limit: 100 }).then(res => setItems(res.data.data || []));
    getCustomers({ limit: 100 }).then(res => setCustomers(res.data.data || []));
  }, []);

  const handleCreate = async () => {
    try {
      await createStockOut({
        ...form,
        customer_id: form.customer_id || undefined,
        items: form.items.filter(i => i.item_id),
      });
      toast.success('Stock Out draft created');
      setShowCreate(false);
      setForm({ reference_no: '', customer_id: '', notes: '', items: [{ item_id: '', quantity: 1 }] });
      fetchList();
    } catch (e: any) { toast.error(e.response?.data?.error || 'Failed'); }
  };

  const handleAdvance = async () => {
    if (!selected) return;
    const next = nextStockOutStatus[selected.status];
    if (!next) return;
    try {
      await updateStockOutStatus(selected.id, next);
      toast.success(`Status updated to ${next}`);
      fetchDetail(selected.id);
      fetchList();
    } catch (e: any) { toast.error(e.response?.data?.error || 'Failed'); }
  };

  const handleCancel = async () => {
    if (!selected) return;
    try {
      await cancelStockOut(selected.id);
      toast.success('Cancelled — stock reservation released');
      setShowCancel(false);
      fetchDetail(selected.id);
      fetchList();
    } catch (e: any) { toast.error(e.response?.data?.error || 'Failed'); }
  };

  const addRow = () => setForm({ ...form, items: [...form.items, { item_id: '', quantity: 1 }] });
  const removeRow = (i: number) => setForm({ ...form, items: form.items.filter((_, idx) => idx !== i) });

  const getAvailableStock = (itemId: string) => {
    const item = items.find(i => i.id === itemId);
    return item ? item.available_stock : null;
  };

  return (
    <div>
      <div className="flex justify-between items-center mb-6">
        <h2 className="text-2xl font-bold">Stock Out</h2>
        <button onClick={() => setShowCreate(true)} className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700">+ New Stock Out</button>
      </div>

      <div className="grid grid-cols-3 gap-6">
        <div className="col-span-2 bg-white rounded-lg shadow overflow-hidden">
          <table className="w-full text-sm">
            <thead className="bg-gray-50">
              <tr>
                <th className="px-4 py-3 text-left">Reference</th>
                <th className="px-4 py-3 text-left">Customer</th>
                <th className="px-4 py-3 text-left">Status</th>
                <th className="px-4 py-3 text-left">Date</th>
              </tr>
            </thead>
            <tbody className="divide-y">
              {txns.map(t => (
                <tr key={t.id} onClick={() => fetchDetail(t.id)} className="hover:bg-blue-50 cursor-pointer">
                  <td className="px-4 py-3 font-medium">{t.reference_no}</td>
                  <td className="px-4 py-3 text-gray-500">{t.customer_name || '-'}</td>
                  <td className="px-4 py-3"><StatusBadge status={t.status} /></td>
                  <td className="px-4 py-3 text-gray-500">{formatDate(t.created_at)}</td>
                </tr>
              ))}
              {txns.length === 0 && <tr><td colSpan={4} className="px-4 py-8 text-center text-gray-500">No transactions</td></tr>}
            </tbody>
          </table>
        </div>

        <div className="bg-white rounded-lg shadow p-4">
          {selected ? (
            <div>
              <h3 className="font-semibold text-lg mb-2">{selected.reference_no}</h3>
              <div className="mb-2"><StatusBadge status={selected.status} /></div>
              {selected.customer_name && <p className="text-sm text-gray-500 mb-1">Customer: {selected.customer_name}</p>}
              <p className="text-sm text-gray-500 mb-4">{selected.notes || 'No notes'}</p>

              <h4 className="font-medium text-sm mb-2">Items</h4>
              <div className="space-y-1 mb-4">
                {selected.items?.map(it => (
                  <div key={it.id} className="flex justify-between text-sm bg-gray-50 px-3 py-2 rounded">
                    <span>{it.item_name || it.item_sku || it.item_id}</span>
                    <span className="font-medium">x{it.quantity}</span>
                  </div>
                ))}
              </div>

              <h4 className="font-medium text-sm mb-2">History</h4>
              <div className="space-y-1 mb-4">
                {selected.logs?.map(log => (
                  <div key={log.id} className="text-xs text-gray-500 bg-gray-50 px-3 py-2 rounded">
                    {log.from_status || '(new)'} → {log.to_status} — {formatDate(log.created_at)}
                  </div>
                ))}
              </div>

              <div className="flex gap-2">
                {nextStockOutStatus[selected.status] && (
                  <button onClick={handleAdvance} className="px-3 py-1.5 bg-blue-600 text-white rounded-lg text-sm hover:bg-blue-700">
                    → {nextStockOutStatus[selected.status]}
                  </button>
                )}
                {selected.status !== 'DONE' && selected.status !== 'CANCELLED' && (
                  <button onClick={() => setShowCancel(true)} className="px-3 py-1.5 bg-red-600 text-white rounded-lg text-sm hover:bg-red-700">Cancel</button>
                )}
              </div>
            </div>
          ) : (
            <p className="text-gray-500 text-sm">Select a transaction to view details</p>
          )}
        </div>
      </div>

      {showCreate && (
        <div className="fixed inset-0 bg-black/40 flex items-center justify-center z-50">
          <div className="bg-white rounded-lg p-6 max-w-lg w-full mx-4 shadow-xl max-h-[80vh] overflow-auto">
            <h3 className="text-lg font-semibold mb-4">Create Stock Out (Draft)</h3>
            <div className="space-y-3">
              <input placeholder="Reference No" value={form.reference_no} onChange={e => setForm({ ...form, reference_no: e.target.value })} className="w-full border rounded-lg px-3 py-2" />
              <select value={form.customer_id} onChange={e => setForm({ ...form, customer_id: e.target.value })} className="w-full border rounded-lg px-3 py-2">
                <option value="">Select customer (optional)...</option>
                {customers.map(c => <option key={c.id} value={c.id}>{c.name}</option>)}
              </select>
              <input placeholder="Notes" value={form.notes} onChange={e => setForm({ ...form, notes: e.target.value })} className="w-full border rounded-lg px-3 py-2" />
              <div>
                <label className="text-sm font-medium">Items</label>
                {form.items.map((item, i) => (
                  <div key={i} className="flex gap-2 mt-2 items-center">
                    <select value={item.item_id} onChange={e => { const newItems = [...form.items]; newItems[i].item_id = e.target.value; setForm({ ...form, items: newItems }); }} className="flex-1 border rounded-lg px-3 py-2 text-sm">
                      <option value="">Select item...</option>
                      {items.map(it => <option key={it.id} value={it.id}>{it.sku} - {it.name} (avail: {formatNumber(it.available_stock)})</option>)}
                    </select>
                    <input type="number" min={1} value={item.quantity} onChange={e => { const newItems = [...form.items]; newItems[i].quantity = parseInt(e.target.value) || 1; setForm({ ...form, items: newItems }); }} className="w-20 border rounded-lg px-3 py-2 text-sm" />
                    {item.item_id && (
                      <span className={`text-xs ${(getAvailableStock(item.item_id) || 0) < item.quantity ? 'text-red-500' : 'text-green-600'}`}>
                        {getAvailableStock(item.item_id)} avail
                      </span>
                    )}
                    {form.items.length > 1 && <button onClick={() => removeRow(i)} className="text-red-500 text-sm">X</button>}
                  </div>
                ))}
                <button onClick={addRow} className="text-blue-600 text-sm mt-2">+ Add item</button>
              </div>
            </div>
            <div className="flex justify-end gap-2 mt-4">
              <button onClick={() => setShowCreate(false)} className="px-4 py-2 border rounded-lg">Cancel</button>
              <button onClick={handleCreate} className="px-4 py-2 bg-blue-600 text-white rounded-lg">Create Draft</button>
            </div>
          </div>
        </div>
      )}

      <ConfirmDialog open={showCancel} title="Cancel Transaction" message="Cancel this stock out? Reserved stock will be released." onConfirm={handleCancel} onCancel={() => setShowCancel(false)} />
    </div>
  );
}
