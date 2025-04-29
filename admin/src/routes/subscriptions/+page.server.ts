// src/routes/admin/subscriptions/+page.server.ts

import type { PageServerLoad } from './$types';

interface SubscriptionResponse {
  id: string;
  user_id: string;
  price_id: string;
  quantity: number;
  status: string;
  collection_method: string;
  currency: string;
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

export const load: PageServerLoad = async ({ url, fetch }) => {
  // Get query parameters from the URL
  const limit = url.searchParams.get('limit') || '10';
  const offset = url.searchParams.get('offset') || '0';
  const status = url.searchParams.get('status') || '';
  const userId = url.searchParams.get('user_id') || '';
  
  // Construct the API URL with query parameters
  const apiUrl = new URL('http://localhost:8080/api/subscriptions');
  apiUrl.searchParams.set('limit', limit);
  apiUrl.searchParams.set('offset', offset);
  
  if (status) {
    apiUrl.searchParams.set('status', status);
  }
  
  if (userId) {
    apiUrl.searchParams.set('user_id', userId);
  }
  
  try {
    // Fetch subscriptions from the API
    const response = await fetch(apiUrl.toString());
    
    if (!response.ok) {
      throw new Error(`API request failed with status ${response.status}`);
    }
    
    // Parse the response and handle null case
    const subscriptions: SubscriptionResponse[] = await response.json() || [];
    
    // Get list of subscription statuses for filter dropdown
    const statusList = [
      { value: 'active', label: 'Active' },
      { value: 'past_due', label: 'Past Due' },
      { value: 'canceled', label: 'Canceled' },
      { value: 'incomplete', label: 'Incomplete' },
      { value: 'incomplete_expired', label: 'Incomplete Expired' },
      { value: 'trialing', label: 'Trialing' },
      { value: 'unpaid', label: 'Unpaid' }
    ];
    
    return {
      subscriptions,
      pagination: {
        limit: parseInt(limit),
        offset: parseInt(offset),
        status,
        userId
      },
      statusList
    };
  } catch (error) {
    console.error('Failed to fetch subscriptions:', error);
    return {
      subscriptions: [] as SubscriptionResponse[],
      pagination: {
        limit: parseInt(limit),
        offset: parseInt(offset),
        status,
        userId
      },
      statusList: [
        { value: 'active', label: 'Active' },
        { value: 'past_due', label: 'Past Due' },
        { value: 'canceled', label: 'Canceled' },
        { value: 'incomplete', label: 'Incomplete' },
        { value: 'incomplete_expired', label: 'Incomplete Expired' },
        { value: 'trialing', label: 'Trialing' },
        { value: 'unpaid', label: 'Unpaid' }
      ],
      error: 'Failed to load subscriptions. Please try again later.'
    };
  }
};