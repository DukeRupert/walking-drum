import React, { useEffect, useState } from 'react';
import { useSearchParams } from 'react-router-dom';

const SuccessPage = () => {
  const [searchParams] = useSearchParams();
  const sessionId = searchParams.get('session_id');
  const [status, setStatus] = useState('loading');
  const [subscription, setSubscription] = useState(null);

  useEffect(() => {
    const verifySubscription = async () => {
      if (!sessionId) {
        setStatus('error');
        return;
      }

      try {
        // Verify the session with your backend
        const response = await fetch(`/api/v1/checkout/verify-session?session_id=${sessionId}`);
        
        if (!response.ok) {
          throw new Error('Failed to verify subscription');
        }
        
        const data = await response.json();
        setSubscription(data.subscription);
        setStatus('success');
      } catch (error) {
        console.error('Error verifying subscription:', error);
        setStatus('error');
      }
    };

    verifySubscription();
  }, [sessionId]);

  if (status === 'loading') {
    return (
      <div className="success-loading">
        <div className="spinner"></div>
        <p>Verifying your subscription...</p>
      </div>
    );
  }

  if (status === 'error') {
    return (
      <div className="success-error">
        <h2>Verification Failed</h2>
        <p>We couldn't verify your subscription. Please contact customer support.</p>
        <button onClick={() => window.location.href = '/'}>Return Home</button>
      </div>
    );
  }

  return (
    <div className="success-page">
      <div className="success-icon">✓</div>
      <h1>Subscription Confirmed!</h1>
      <p>Thank you for subscribing to our coffee delivery service.</p>
      
      {subscription && (
        <div className="subscription-details">
          <h3>Subscription Details</h3>
          <p>Your first delivery will be shipped soon.</p>
          <p>You can manage your subscription in your account dashboard.</p>
        </div>
      )}
      
      <div className="success-actions">
        <button onClick={() => window.location.href = '/account'}>Go to My Account</button>
        <button onClick={() => window.location.href = '/'}>Continue Shopping</button>
      </div>
    </div>
  );
};

export default SuccessPage;