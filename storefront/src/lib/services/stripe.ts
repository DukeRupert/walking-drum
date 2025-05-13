import { PUBLIC_API_BASE_URL } from '$env/static/public';

// Single price subscription checkout
export async function createCheckoutSession(priceId: string, customerId: string, quantity = 1) {
  try {
    const response = await fetch(`${PUBLIC_API_BASE_URL}/checkout/create-session`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        price_id: priceId,
        customer_id: customerId,
        quantity: quantity,
        return_url: `${window.location.origin}/checkout/result?session_id={CHECKOUT_SESSION_ID}`,
      }),
    });

    if (!response.ok) {
      const errorData = await response.json();
      throw new Error(errorData.error || 'Failed to create checkout session');
    }

    return await response.json();
  } catch (error) {
    console.error('Error creating checkout session:', error);
    throw error;
  }
}

// Multiple items subscription checkout
export async function createMultipleItemCheckoutSession(
  items: Array<{ price_id: string; quantity: number }>,
  customerId: string
) {
  try {
    const response = await fetch(`${PUBLIC_API_BASE_URL}/checkout/create-multi-item-session`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        items,
        customer_id: customerId,
        return_url: `${window.location.origin}/checkout/result?session_id={CHECKOUT_SESSION_ID}`,
      }),
    });

    if (!response.ok) {
      const errorData = await response.json();
      throw new Error(errorData.error || 'Failed to create checkout session');
    }

    return await response.json();
  } catch (error) {
    console.error('Error creating multi-item checkout session:', error);
    throw error;
  }
}