import React, { useState, useEffect } from 'react';
import { useParams } from 'react-router-dom';
import SubscriptionCheckout from '../components/SubscriptionCheckout';

const CheckoutPage = () => {
  const { priceId } = useParams();
  const [product, setProduct] = useState(null);
  const [price, setPrice] = useState(null);
  const [customer, setCustomer] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  useEffect(() => {
    // Fetch the data needed for checkout
    const fetchData = async () => {
      try {
        setLoading(true);
        
        // Fetch price details
        const priceResponse = await fetch(`https://literate-space-meme-wg644q6vqv535x4v-8080.app.github.dev/api/v1/prices/${priceId}`);
        if (!priceResponse.ok) {
          throw new Error('Failed to fetch price details');
        }
        const priceData = await priceResponse.json();
        setPrice(priceData.data);
        
        // Fetch product details
        const productResponse = await fetch(`https://literate-space-meme-wg644q6vqv535x4v-8080.app.github.dev/api/v1/products/${priceData.data.product_id}`);
        if (!productResponse.ok) {
          throw new Error('Failed to fetch product details');
        }
        const productData = await productResponse.json();
        setProduct(productData.data);
        
        // In a real application, you would get the current authenticated customer
        // For this POC, we'll just use the first customer from the list
        const customerResponse = await fetch('https://literate-space-meme-wg644q6vqv535x4v-8080.app.github.dev/api/v1/customers');
        if (!customerResponse.ok) {
          throw new Error('Failed to fetch customer details');
        }
        const customerData = await customerResponse.json();
        setCustomer(customerData.data[0]);
        
        setError(null);
      } catch (err) {
        console.error('Error fetching data:', err);
        setError(err.message);
      } finally {
        setLoading(false);
      }
    };

    if (priceId) {
      fetchData();
    }
  }, [priceId]);

  if (loading) {
    return (
      <div className="page-loading">
        <div className="spinner"></div>
        <p>Loading subscription details...</p>
      </div>
    );
  }

  if (error) {
    return (
      <div className="page-error">
        <h2>Error</h2>
        <p>{error}</p>
        <button onClick={() => window.location.href = '/'}>Return Home</button>
      </div>
    );
  }

  if (!product || !price || !customer) {
    return (
      <div className="page-error">
        <h2>Missing Information</h2>
        <p>Could not load all required information for checkout.</p>
        <button onClick={() => window.location.href = '/'}>Return Home</button>
      </div>
    );
  }

  return (
    <div className="checkout-page">
      <div className="checkout-header">
        <h1>Subscribe to {product.name}</h1>
        <div className="subscription-details">
          <p>{product.description}</p>
          <div className="price-info">
            <span className="price">
              ${(price.amount / 100).toFixed(2)}
            </span>
            <span className="interval">
              per {price.interval}
            </span>
          </div>
        </div>
      </div>
      
      <SubscriptionCheckout 
        productId={product.id}
        priceId={price.id}
        customerId={customer.id}
      />
    </div>
  );
};

export default CheckoutPage;