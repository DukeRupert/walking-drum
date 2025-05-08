import React from 'react';
import { BrowserRouter, Routes, Route } from 'react-router-dom';
import ProductPage from './pages/ProductPage';
import CheckoutPage from './pages/CheckoutPage';
import SuccessPage from './pages/SuccessPage';
import CancelPage from './pages/CancelPage';
import './styles/checkout.css';
import './styles/products.css';
import './App.css'

const App = () => {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/" element={<ProductPage />} />
        <Route path="/checkout/:priceId" element={<CheckoutPage />} />
        <Route path="/subscription/success" element={<SuccessPage />} />
        <Route path="/subscription/cancel" element={<CancelPage />} />
      </Routes>
    </BrowserRouter>
  );
};

export default App
