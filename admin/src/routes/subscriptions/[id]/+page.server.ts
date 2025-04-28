// src/routes/admin/subscriptions/[id]/+page.server.ts

import { error } from '@sveltejs/kit';
import type { PageServerLoad } from './$types';

interface SubscriptionResponse {
  id: string;
  user_id: string;
  price_id: string;
  quantity: number;
  status: string;
  customer_id: string;
  current_period_start: string;
  current_period_end: string;
  cancel_at?: string;
  cancel_at_period_end: boolean;
  canceled_at?: string;
  ended_at?: string;
  trial_start?: string;
  trial_end?: string;
  resume_at?: string;
  latest_invoice_id?: string;
  payment_method_id?: string;
  stripe_id: string;
  created_at: string;
  updated_at: string;
  metadata?: Record<string, string>;
}

export const load: PageServerLoad = async ({ params, fetch }) => {
  const { id } = params;
  console.log(`Fetching subscription id: ${id}`)
  
  try {
    // Build the API URL
    const apiUrl = `http://localhost:8080/api/subscriptions/${id}`;
    
    // Fetch the subscription data
    const response = await fetch(apiUrl);
    
    if (!response.ok) {
      if (response.status === 404) {
        throw error(404, 'Subscription not found');
      }
      throw error(response.status, `Error fetching subscription: ${response.statusText}`);
    }
    
    const subscription: SubscriptionResponse = await response.json();
    
    // Convert the metadata to an array of key-value pairs for easier display
    const metadataArray = subscription.metadata 
      ? Object.entries(subscription.metadata).map(([key, value]) => ({ key, value }))
      : [];
    
    return {
      subscription,
      metadataArray
    };
  } catch (err) {
    console.error('Failed to load subscription:', err);
    throw error(500, 'Failed to load subscription details');
  }
};