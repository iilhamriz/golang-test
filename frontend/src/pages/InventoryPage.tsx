import { useEffect, useState } from 'react';
import { getItems, createItem, adjustStock } from '../api/inventoryApi';
import { useDebounce } from '../hooks/useDebounce';
import { formatNumber } from '../utils/formatters';
import type { Item } from '../types';
import toast from 'react-hot-toast';

export default function InventoryPage() {
  const [items, setItems] = useState<Item[]>([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [search, setSearch] = useState('');
  const [skuSearch, setSkuSearch] = useState('');
  const debouncedSearch = useDebounce(search, 300);
  const debouncedSku = useDebounce(skuSearch, 300);
  const [showCreate, setShowCreate] = useState(false);
  const [showAdjust, setShowAdjust] = useState<Item | null>(null);
  const [form, setForm] = useState({ sku: '', name: '', description: '' });
  const [adjForm, setAdjForm] = useState({ quantity: 0, reason: '' });

  const fetchItems = async () => {
    try {
      const res = await getItems({ name: debouncedSearch, sku: debouncedSku, page, limit: 20 });
      setItems(res.data.data || []);
      setTotal(res.data.meta?.total || 0);
    } catch { toast.error('Failed to load items'); }
  };

  useEffect(() => { fetchItems(); }, [debouncedSearch, debouncedSku, page]);

  const handleCreate = async () => {
    try {
      await createItem(form);
      toast.success('Item created');
      setShowCreate(false);
      setForm({ sku: '', name: '', description: '' });
      fetchItems();
    } catch (e: any) { toast.error(e.response?.data?.error || 'Failed'); }
  };

  const handleAdjust = async () => {
    if (!showAdjust) return;
    try {
      await adjustStock(showAdjust.id, adjForm);
      toast.success('Stock adjusted');
      setShowAdjust(null);
      setAdjForm({ quantity: 0, reason: '' });
      fetchItems();
    } catch (e: any) { toast.error(e.response?.data?.error || 'Failed'); }
  };

  return (
    <div>
      <div className="flex justify-between items-center mb-6">
        <h2 className="text-2xl font-bold">Inventory</h2>
        <button onClick={() => setShowCreate(true)} className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700">+ New Item</button>
      </div>

      <div className="flex gap-3 mb-4">
        <input placeholder="Search by name..." value={search} onChange={e => setSearch(e.target.value)} className="border rounded-lg px-3 py-2 text-sm flex-1" />
        <input placeholder="Search by SKU..." value={skuSearch} onChange={e => setSkuSearch(e.target.value)} className="border rounded-lg px-3 py-2 text-sm flex-1" />
      </div>

      <div className="bg-white rounded-lg shadow overflow-hidden">
        <table className="w-full text-sm">
          <thead className="bg-gray-50">
            <tr>
              <th className="px-4 py-3 text-left font-medium text-gray-600">SKU</th>
              <th className="px-4 py-3 text-left font-medium text-gray-600">Name</th>
              <th className="px-4 py-3 text-right font-medium text-gray-600">Physical Stock</th>
              <th className="px-4 py-3 text-right font-medium text-gray-600">Available Stock</th>
              <th className="px-4 py-3 text-right font-medium text-gray-600">Actions</th>
            </tr>
          </thead>
          <tbody className="divide-y">
            {items.map(item => (
              <tr key={item.id} className="hover:bg-gray-50">
                <td className="px-4 py-3 font-mono text-xs">{item.sku}</td>
                <td className="px-4 py-3">{item.name}</td>
                <td className="px-4 py-3 text-right font-medium">{formatNumber(item.physical_stock)}</td>
                <td className="px-4 py-3 text-right font-medium">{formatNumber(item.available_stock)}</td>
                <td className="px-4 py-3 text-right">
                  <button onClick={() => { setShowAdjust(item); setAdjForm({ quantity: 0, reason: '' }); }} className="text-blue-600 hover:underline text-xs">Adjust</button>
                </td>
              </tr>
            ))}
            {items.length === 0 && <tr><td colSpan={5} className="px-4 py-8 text-center text-gray-500">No items found</td></tr>}
          </tbody>
        </table>
      </div>

      <div className="flex justify-between items-center mt-4 text-sm text-gray-600">
        <span>Total: {total}</span>
        <div className="flex gap-2">
          <button onClick={() => setPage(p => Math.max(1, p - 1))} disabled={page <= 1} className="px-3 py-1 border rounded disabled:opacity-50">Prev</button>
          <span className="px-3 py-1">Page {page}</span>
          <button onClick={() => setPage(p => p + 1)} disabled={items.length < 20} className="px-3 py-1 border rounded disabled:opacity-50">Next</button>
        </div>
      </div>

      {/* Create Item Modal */}
      {showCreate && (
        <div className="fixed inset-0 bg-black/40 flex items-center justify-center z-50">
          <div className="bg-white rounded-lg p-6 max-w-md w-full mx-4 shadow-xl">
            <h3 className="text-lg font-semibold mb-4">Create Item</h3>
            <div className="space-y-3">
              <input placeholder="SKU" value={form.sku} onChange={e => setForm({ ...form, sku: e.target.value })} className="w-full border rounded-lg px-3 py-2" />
              <input placeholder="Name" value={form.name} onChange={e => setForm({ ...form, name: e.target.value })} className="w-full border rounded-lg px-3 py-2" />
              <input placeholder="Description" value={form.description} onChange={e => setForm({ ...form, description: e.target.value })} className="w-full border rounded-lg px-3 py-2" />
            </div>
            <div className="flex justify-end gap-2 mt-4">
              <button onClick={() => setShowCreate(false)} className="px-4 py-2 border rounded-lg">Cancel</button>
              <button onClick={handleCreate} className="px-4 py-2 bg-blue-600 text-white rounded-lg">Create</button>
            </div>
          </div>
        </div>
      )}

      {/* Adjust Stock Modal */}
      {showAdjust && (
        <div className="fixed inset-0 bg-black/40 flex items-center justify-center z-50">
          <div className="bg-white rounded-lg p-6 max-w-md w-full mx-4 shadow-xl">
            <h3 className="text-lg font-semibold mb-2">Adjust Stock: {showAdjust.name}</h3>
            <p className="text-sm text-gray-500 mb-4">Current: {showAdjust.physical_stock} physical / {showAdjust.available_stock} available</p>
            <div className="space-y-3">
              <input type="number" placeholder="Quantity (+/-)" value={adjForm.quantity || ''} onChange={e => setAdjForm({ ...adjForm, quantity: parseInt(e.target.value) || 0 })} className="w-full border rounded-lg px-3 py-2" />
              <input placeholder="Reason" value={adjForm.reason} onChange={e => setAdjForm({ ...adjForm, reason: e.target.value })} className="w-full border rounded-lg px-3 py-2" />
            </div>
            <div className="flex justify-end gap-2 mt-4">
              <button onClick={() => setShowAdjust(null)} className="px-4 py-2 border rounded-lg">Cancel</button>
              <button onClick={handleAdjust} className="px-4 py-2 bg-blue-600 text-white rounded-lg">Adjust</button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
