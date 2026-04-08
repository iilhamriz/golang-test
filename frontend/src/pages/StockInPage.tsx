import { useEffect, useState } from 'react';
import { getStockIns, getStockIn, createStockIn, updateStockInStatus, cancelStockIn } from '../api/stockInApi';
import { getItems } from '../api/inventoryApi';
import StatusBadge from '../components/common/StatusBadge';
import ConfirmDialog from '../components/common/ConfirmDialog';
import { formatDate } from '../utils/formatters';
import { nextStockInStatus } from '../types/enums';
import type { StockInTransaction, Item } from '../types';
import toast from 'react-hot-toast';

export default function StockInPage() {
  const [txns, setTxns] = useState<StockInTransaction[]>([]);
  const [selected, setSelected] = useState<StockInTransaction | null>(null);
  const [showCreate, setShowCreate] = useState(false);
  const [showCancel, setShowCancel] = useState(false);
  const [items, setItems] = useState<Item[]>([]);
  const [form, setForm] = useState({ reference_no: '', notes: '', items: [{ item_id: '', quantity: 1 }] });

  const fetchList = async () => {
    try {
      const res = await getStockIns({ limit: 50 });
      setTxns(res.data.data || []);
    } catch { toast.error('Failed to load'); }
  };

  const fetchDetail = async (id: string) => {
    try {
      const res = await getStockIn(id);
      setSelected(res.data.data);
    } catch { toast.error('Failed to load detail'); }
  };

  const loadItems = async () => {
    const res = await getItems({ limit: 100 });
    setItems(res.data.data || []);
  };

  useEffect(() => { fetchList(); loadItems(); }, []);

  const handleCreate = async () => {
    try {
      await createStockIn({ ...form, items: form.items.filter(i => i.item_id) });
      toast.success('Stock In created');
      setShowCreate(false);
      setForm({ reference_no: '', notes: '', items: [{ item_id: '', quantity: 1 }] });
      fetchList();
    } catch (e: any) { toast.error(e.response?.data?.error || 'Failed'); }
  };

  const handleAdvance = async () => {
    if (!selected) return;
    const next = nextStockInStatus[selected.status];
    if (!next) return;
    try {
      await updateStockInStatus(selected.id, next);
      toast.success(`Status updated to ${next}`);
      fetchDetail(selected.id);
      fetchList();
    } catch (e: any) { toast.error(e.response?.data?.error || 'Failed'); }
  };

  const handleCancel = async () => {
    if (!selected) return;
    try {
      await cancelStockIn(selected.id);
      toast.success('Cancelled');
      setShowCancel(false);
      fetchDetail(selected.id);
      fetchList();
    } catch (e: any) { toast.error(e.response?.data?.error || 'Failed'); }
  };

  const addRow = () => setForm({ ...form, items: [...form.items, { item_id: '', quantity: 1 }] });
  const removeRow = (i: number) => setForm({ ...form, items: form.items.filter((_, idx) => idx !== i) });

  return (
    <div>
      <div className="flex justify-between items-center mb-6">
        <h2 className="text-2xl font-bold">Stock In</h2>
        <button onClick={() => setShowCreate(true)} className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700">+ New Stock In</button>
      </div>

      <div className="grid grid-cols-3 gap-6">
        {/* List */}
        <div className="col-span-2 bg-white rounded-lg shadow overflow-hidden">
          <table className="w-full text-sm">
            <thead className="bg-gray-50">
              <tr>
                <th className="px-4 py-3 text-left">Reference</th>
                <th className="px-4 py-3 text-left">Status</th>
                <th className="px-4 py-3 text-left">Date</th>
              </tr>
            </thead>
            <tbody className="divide-y">
              {txns.map(t => (
                <tr key={t.id} onClick={() => fetchDetail(t.id)} className="hover:bg-blue-50 cursor-pointer">
                  <td className="px-4 py-3 font-medium">{t.reference_no}</td>
                  <td className="px-4 py-3"><StatusBadge status={t.status} /></td>
                  <td className="px-4 py-3 text-gray-500">{formatDate(t.created_at)}</td>
                </tr>
              ))}
              {txns.length === 0 && <tr><td colSpan={3} className="px-4 py-8 text-center text-gray-500">No transactions</td></tr>}
            </tbody>
          </table>
        </div>

        {/* Detail */}
        <div className="bg-white rounded-lg shadow p-4">
          {selected ? (
            <div>
              <h3 className="font-semibold text-lg mb-2">{selected.reference_no}</h3>
              <div className="mb-3"><StatusBadge status={selected.status} /></div>
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
                {nextStockInStatus[selected.status] && (
                  <button onClick={handleAdvance} className="px-3 py-1.5 bg-blue-600 text-white rounded-lg text-sm hover:bg-blue-700">
                    → {nextStockInStatus[selected.status]}
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

      {/* Create Modal */}
      {showCreate && (
        <div className="fixed inset-0 bg-black/40 flex items-center justify-center z-50">
          <div className="bg-white rounded-lg p-6 max-w-lg w-full mx-4 shadow-xl max-h-[80vh] overflow-auto">
            <h3 className="text-lg font-semibold mb-4">Create Stock In</h3>
            <div className="space-y-3">
              <input placeholder="Reference No" value={form.reference_no} onChange={e => setForm({ ...form, reference_no: e.target.value })} className="w-full border rounded-lg px-3 py-2" />
              <input placeholder="Notes" value={form.notes} onChange={e => setForm({ ...form, notes: e.target.value })} className="w-full border rounded-lg px-3 py-2" />
              <div>
                <label className="text-sm font-medium">Items</label>
                {form.items.map((item, i) => (
                  <div key={i} className="flex gap-2 mt-2">
                    <select value={item.item_id} onChange={e => { const newItems = [...form.items]; newItems[i].item_id = e.target.value; setForm({ ...form, items: newItems }); }} className="flex-1 border rounded-lg px-3 py-2 text-sm">
                      <option value="">Select item...</option>
                      {items.map(it => <option key={it.id} value={it.id}>{it.sku} - {it.name}</option>)}
                    </select>
                    <input type="number" min={1} value={item.quantity} onChange={e => { const newItems = [...form.items]; newItems[i].quantity = parseInt(e.target.value) || 1; setForm({ ...form, items: newItems }); }} className="w-20 border rounded-lg px-3 py-2 text-sm" />
                    {form.items.length > 1 && <button onClick={() => removeRow(i)} className="text-red-500 text-sm">X</button>}
                  </div>
                ))}
                <button onClick={addRow} className="text-blue-600 text-sm mt-2">+ Add item</button>
              </div>
            </div>
            <div className="flex justify-end gap-2 mt-4">
              <button onClick={() => setShowCreate(false)} className="px-4 py-2 border rounded-lg">Cancel</button>
              <button onClick={handleCreate} className="px-4 py-2 bg-blue-600 text-white rounded-lg">Create</button>
            </div>
          </div>
        </div>
      )}

      <ConfirmDialog open={showCancel} title="Cancel Transaction" message="Are you sure you want to cancel this stock in?" onConfirm={handleCancel} onCancel={() => setShowCancel(false)} />
    </div>
  );
}
