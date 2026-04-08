import { useEffect, useState } from 'react';
import { getCustomers, createCustomer } from '../api/customerApi';
import type { Customer } from '../types';
import toast from 'react-hot-toast';

export default function CustomerPage() {
  const [customers, setCustomers] = useState<Customer[]>([]);
  const [showCreate, setShowCreate] = useState(false);
  const [form, setForm] = useState({ name: '', email: '', phone: '', address: '' });

  const fetchList = async () => {
    try {
      const res = await getCustomers({ limit: 50 });
      setCustomers(res.data.data || []);
    } catch { toast.error('Failed to load'); }
  };

  useEffect(() => { fetchList(); }, []);

  const handleCreate = async () => {
    try {
      await createCustomer(form);
      toast.success('Customer created');
      setShowCreate(false);
      setForm({ name: '', email: '', phone: '', address: '' });
      fetchList();
    } catch (e: any) { toast.error(e.response?.data?.error || 'Failed'); }
  };

  return (
    <div>
      <div className="flex justify-between items-center mb-6">
        <h2 className="text-2xl font-bold">Customers</h2>
        <button onClick={() => setShowCreate(true)} className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700">+ New Customer</button>
      </div>

      <div className="bg-white rounded-lg shadow overflow-hidden">
        <table className="w-full text-sm">
          <thead className="bg-gray-50">
            <tr>
              <th className="px-4 py-3 text-left">Name</th>
              <th className="px-4 py-3 text-left">Email</th>
              <th className="px-4 py-3 text-left">Phone</th>
              <th className="px-4 py-3 text-left">Address</th>
            </tr>
          </thead>
          <tbody className="divide-y">
            {customers.map(c => (
              <tr key={c.id} className="hover:bg-gray-50">
                <td className="px-4 py-3 font-medium">{c.name}</td>
                <td className="px-4 py-3 text-gray-500">{c.email || '-'}</td>
                <td className="px-4 py-3 text-gray-500">{c.phone || '-'}</td>
                <td className="px-4 py-3 text-gray-500">{c.address || '-'}</td>
              </tr>
            ))}
            {customers.length === 0 && <tr><td colSpan={4} className="px-4 py-8 text-center text-gray-500">No customers</td></tr>}
          </tbody>
        </table>
      </div>

      {showCreate && (
        <div className="fixed inset-0 bg-black/40 flex items-center justify-center z-50">
          <div className="bg-white rounded-lg p-6 max-w-md w-full mx-4 shadow-xl">
            <h3 className="text-lg font-semibold mb-4">Create Customer</h3>
            <div className="space-y-3">
              <input placeholder="Name *" value={form.name} onChange={e => setForm({ ...form, name: e.target.value })} className="w-full border rounded-lg px-3 py-2" />
              <input placeholder="Email" value={form.email} onChange={e => setForm({ ...form, email: e.target.value })} className="w-full border rounded-lg px-3 py-2" />
              <input placeholder="Phone" value={form.phone} onChange={e => setForm({ ...form, phone: e.target.value })} className="w-full border rounded-lg px-3 py-2" />
              <input placeholder="Address" value={form.address} onChange={e => setForm({ ...form, address: e.target.value })} className="w-full border rounded-lg px-3 py-2" />
            </div>
            <div className="flex justify-end gap-2 mt-4">
              <button onClick={() => setShowCreate(false)} className="px-4 py-2 border rounded-lg">Cancel</button>
              <button onClick={handleCreate} className="px-4 py-2 bg-blue-600 text-white rounded-lg">Create</button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
