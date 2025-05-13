import type { PageLoad } from './$types';
import { getCustomers } from '$lib/services/api';
import { createMultipleItemCheckoutSession } from '$lib/services/stripe';
import { get } from 'svelte/store';
import { cart } from '$lib/stores/cart';
import { redirect } from '@sveltejs/kit';

export const load: PageLoad = async ({ fetch }) => {
  try {
    // Check if cart is empty
    const cartStore = get(cart);
    if (cartStore.items.length === 0) {
      // Redirect to products page if cart is empty
      throw redirect(303, '/products');
    }

    // Get the first customer (in a real app, this would be the authenticated user)
    const customersResponse = await getCustomers(fetch);
    const customer = customersResponse.data[0];

    if (!customer) {
      return {
        customer: null,
        error: 'No customer found'
      };
    }

    // Prepare items for checkout
    const items = cartStore.items.map(item => ({
      price_id: item.priceId,
      quantity: item.quantity
    }));

    // Create checkout session
    const checkoutSession = await createMultipleItemCheckoutSession(items, customer.id);

    return {
      customer,
      clientSecret: checkoutSession.client_secret
    };
  } catch (error) {
    if (error instanceof Response) {
      // This is a redirect, just throw it
      throw error;
    }
    
    console.error('Error loading data:', error);
    return {
      customer: null,
      error: error.message || 'Failed to create checkout session'
    };
  }
};