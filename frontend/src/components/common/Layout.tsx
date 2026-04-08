import { NavLink, Outlet } from 'react-router-dom';
import { Toaster } from 'react-hot-toast';

const navItems = [
  { to: '/', label: 'Inventory', icon: '📦' },
  { to: '/stock-in', label: 'Stock In', icon: '📥' },
  { to: '/stock-out', label: 'Stock Out', icon: '📤' },
  { to: '/reports', label: 'Reports', icon: '📊' },
  { to: '/customers', label: 'Customers', icon: '👥' },
];

export default function Layout() {
  return (
    <div className="flex h-screen">
      <Toaster position="top-right" />
      <aside className="w-56 bg-gray-900 text-white flex flex-col">
        <div className="p-4 border-b border-gray-700">
          <h1 className="text-lg font-bold">Smart Inventory</h1>
        </div>
        <nav className="flex-1 p-2">
          {navItems.map(({ to, label, icon }) => (
            <NavLink
              key={to}
              to={to}
              end={to === '/'}
              className={({ isActive }) =>
                `flex items-center gap-3 px-3 py-2.5 rounded-lg mb-1 text-sm transition-colors ${
                  isActive ? 'bg-blue-600 text-white' : 'text-gray-300 hover:bg-gray-800'
                }`
              }
            >
              <span>{icon}</span>
              <span>{label}</span>
            </NavLink>
          ))}
        </nav>
      </aside>
      <main className="flex-1 overflow-auto p-6">
        <Outlet />
      </main>
    </div>
  );
}
