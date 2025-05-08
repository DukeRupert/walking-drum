import React, { useState, useEffect } from 'react';
import { loadStripe } from '@stripe/stripe-js';
import { EmbeddedCheckoutProvider, EmbeddedCheckout } from '@stripe/react-stripe-js';

// Initialize Stripe outside component to avoid recreating on each render
const stripePromise = loadStripe('pk_test_51RJKxwQglLwb4wZwJ24cQS5he6FQlahumJMF4VJBNOyJ7KrxfPcBJILtBpw9xZogP7HBKeafQR0c7mnY1TKzBCfg00dcTx2GwS');

const SubscriptionCheckout = ({ priceId, customerId }) => {
  const [clientSecret, setClientSecret] = useState('');
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  useEffect(() => {
    // Create a checkout session when the component mounts
    const createCheckoutSession = async () => {
      try {
        setLoading(true);
        
        const response = await fetch('/api/v1/checkout/create-session', {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
          },
          body: JSON.stringify({
            price_id: priceId,
            customer_id: customerId,
            success_url: `${window.location.origin}/subscription/success?session_id={CHECKOUT_SESSION_ID}`,
            cancel_url: `${window.location.origin}/subscription/cancel`,
          }),
        });

        if (!response.ok) {
          const errorData = await response.json();
          throw new Error(errorData.error || 'Failed to create checkout session');
        }

        const { client_secret } = await response.json();
        setClientSecret(client_secret);
        setError(null);
      } catch (err) {
        console.error('Error creating checkout session:', err);
        setError(err.message);
      } finally {
        setLoading(false);
      }
    };

    if (priceId && customerId) {
      createCheckoutSession();
    } else {
      setError('Missing required information: price ID or customer ID');
      setLoading(false);
    }
  }, [priceId, customerId]);

  if (loading) {
    return (
      <div className="checkout-loading">
        <div className="spinner"></div>
        <p>Loading checkout...</p>
      </div>
    );
  }

  if (error) {
    return (
      <div className="checkout-error">
        <h3>Error</h3>
        <p>{error}</p>
        <button onClick={() => window.location.reload()}>Try Again</button>
      </div>
    );
  }

  return (
    <div className="checkout-container">
      {clientSecret && (
        <EmbeddedCheckoutProvider
          stripe={stripePromise}
          options={{ clientSecret }}
        >
          <EmbeddedCheckout />
        </EmbeddedCheckoutProvider>
      )}
    </div>
  );
};

export default SubscriptionCheckout;