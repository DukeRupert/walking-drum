import type { PageLoad } from './$types';
import { PUBLIC_API_BASE_URL } from '$env/static/public';

export const load: PageLoad = async ({ url }) => {
  const sessionId = url.searchParams.get('session_id');
  
  if (!sessionId) {
    return {
      error: 'Missing session ID'
    };
  }
  
  try {
    // In a production app, you'd verify the session with your backend
    // This is a simplified example
    const response = await fetch(`${PUBLIC_API_BASE_URL}/checkout/verify-session?session_id=${sessionId}`);
    
    if (!response.ok) {
      throw new Error('Failed to verify session');
    }
    
    const data = await response.json();
    
    return {
      sessionId,
      subscription: data.subscription
    };
  } catch (error) {
    console.error('Error verifying session:', error);
    
    // For the POC, we'll simulate a successful subscription
    return {
      sessionId,
      subscription: {
        id: sessionId,
        product_name: 'Cloud 9 Espresso',
        amount: 1620,
        currency: 'usd',
        interval: 'week',
        next_delivery_date: new Date(Date.now() + 7 * 24 * 60 * 60 * 1000).toISOString(),
      }
    };
  }
  };
  