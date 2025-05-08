import React, { useState, useEffect } from 'react';
import { Link } from 'react-router-dom';

const ProductPage = () => {
  const [products, setProducts] = useState([]);
  const [prices, setPrices] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  useEffect(() => {
    const fetchData = async () => {
      try {
        setLoading(true);
        
        // Fetch products
        const productsResponse = await fetch('https://literate-space-meme-wg644q6vqv535x4v-8080.app.github.dev/api/v1/products');
        if (!productsResponse.ok) {
          throw new Error('Failed to fetch products');
        }
        const productsData = await productsResponse.json();
        setProducts(productsData.data);
        
        // Fetch prices
        const pricesResponse = await fetch('https://literate-space-meme-wg644q6vqv535x4v-8080.app.github.dev/api/v1/prices');
        if (!pricesResponse.ok) {
          throw new Error('Failed to fetch prices');
        }
        const pricesData = await pricesResponse.json();
        setPrices(pricesData.data);
        
      } catch (err) {
        console.error('Error fetching data:', err);
        setError(err.message);
      } finally {
        setLoading(false);
      }
    };

    fetchData();
  }, []);

  if (loading) {
    return (
      <div className="page-loading">
        <div className="spinner"></div>
        <p>Loading products...</p>
      </div>
    );
  }

  if (error) {
    return (
      <div className="page-error">
        <h2>Error</h2>
        <p>{error}</p>
        <button onClick={() => window.location.reload()}>Try Again</button>
      </div>
    );
  }

  return (
    <div className="product-page">
      <h1>Coffee Subscriptions</h1>
      <p className="page-description">
        Discover our selection of premium coffees delivered straight to your door.
      </p>
      
      <div className="products-container">
        {products.map(product => {
          const productPrices = prices.filter(price => price.product_id === product.id);
          
          return (
            <div key={product.id} className="product-card">
              <div className="product-image">
                <img src={product.image_url || '/placeholder-coffee.jpg'} alt={product.name} />
              </div>
              <div className="product-info">
                <h2>{product.name}</h2>
                <p className="product-description">{product.description}</p>
                <div className="product-details">
                  <div className="detail">
                    <span className="label">Origin:</span>
                    <span className="value">{product.origin}</span>
                  </div>
                  <div className="detail">
                    <span className="label">Roast Level:</span>
                    <span className="value">{product.roast_level}</span>
                  </div>
                  <div className="detail">
                    <span className="label">Flavor Notes:</span>
                    <span className="value">{product.flavor_notes}</span>
                  </div>
                </div>
                
                <div className="subscription-options">
                  <h3>Subscription Options</h3>
                  {productPrices.length > 0 ? (
                    <div className="price-options">
                      {productPrices.map(price => (
                        <div key={price.id} className="price-option">
                          <div className="price-details">
                            <span className="price-name">{price.name}</span>
                            <span className="price-amount">${(price.amount / 100).toFixed(2)}</span>
                            <span className="price-interval">per {price.interval}</span>
                          </div>
                          <Link to={`/checkout/${price.id}`} className="subscribe-button">
                            Subscribe
                          </Link>
                        </div>
                      ))}
                    </div>
                  ) : (
                    <p className="no-prices">No subscription options available at the moment.</p>
                  )}
                </div>
              </div>
            </div>
          );
        })}
      </div>
    </div>
  );
};

export default ProductPage;