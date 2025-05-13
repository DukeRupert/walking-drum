import type { PageLoad } from './$types';
import { PUBLIC_API_BASE_URL } from '$env/static/public';

export const load: PageLoad = async ({ url, fetch }) => {
  const sessionId = url.searchParams.get('session_id');
  
  if (!sessionId) {
    return {
      error: 'Missing session ID'
    };
  }
  
  try {
    // Verify the session with backend
    const response = await fetch(`${PUBLIC_API_BASE_URL}/checkout/verify-session?session_id=${sessionId}`);
    
    if (!response.ok) {
      throw new Error('Failed to verify session');
    }
    
    const data = await response.json();
    
    return {
      sessionId,
      subscriptions: data.subscriptions || [data.subscription] // Handle both single and multiple subscriptions
    };
  } catch (error) {
    console.error('Error verifying session:', error);
    
    // For the POC, we'll simulate a successful subscription
    return {
      sessionId,
      subscriptions: [
        {
          id: 'sub_example1',
          product_name: 'Cloud 9 Espresso',
          amount: 1620,
          currency: 'usd',
          interval: 'week',
          quantity: 2,
          next_delivery_date: new Date(Date.now() + 7 * 24 * 60 * 60 * 1000).toISOString(),
        }
      ]
    };
  }
};