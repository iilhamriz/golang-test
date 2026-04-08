import { BrowserRouter, Routes, Route } from 'react-router-dom';
import Layout from './components/common/Layout';
import InventoryPage from './pages/InventoryPage';
import StockInPage from './pages/StockInPage';
import StockOutPage from './pages/StockOutPage';
import ReportPage from './pages/ReportPage';
import CustomerPage from './pages/CustomerPage';

export default function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route element={<Layout />}>
          <Route path="/" element={<InventoryPage />} />
          <Route path="/stock-in" element={<StockInPage />} />
          <Route path="/stock-out" element={<StockOutPage />} />
          <Route path="/reports" element={<ReportPage />} />
          <Route path="/customers" element={<CustomerPage />} />
        </Route>
      </Routes>
    </BrowserRouter>
  );
}
