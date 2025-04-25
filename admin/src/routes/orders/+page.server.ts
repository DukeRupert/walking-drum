// src/routes/admin/orders/+page.server.ts

import type { PageServerLoad } from './$types';

interface Address {
  line1: string;
  line2?: string;
  city: string;
  state: string;
  postal_code: string;
  country: string;
}

interface OrderItemResponse {
  id: string;
  order_id: string;
  product_id: string;
  price_id: string;
  quantity: number;
  amount: number;
  currency: string;
  name: string;
  description?: string;
  metadata?: Record<string, string>;
  product?: {
    id: string;
    name: string;
    description: string;
    is_active: boolean;
  };
  price?: {
    id: string;
    amount: number;
    currency: string;
    interval_type: string;
    interval_count: number;
  };
}

interface UserResponse {
  id: string;
  email: string;
  name?: string;
  created_at: string;
  updated_at: string;
}

interface OrderResponse {
  id: string;
  user_id?: string;
  status: string;
  total_amount: number;
  currency: string;
  created_at: string;
  updated_at: string;
  completed_at?: string;
  shipping_address?: Address;
  billing_address?: Address;
  payment_intent_id?: string;
  stripe_customer_id?: string;
  metadata?: Record<string, string>;
  items?: OrderItemResponse[];
  user?: UserResponse;
}

export const load: PageServerLoad = async ({ url, fetch }) => {
  // Get query parameters from the URL
  const limit = url.searchParams.get('limit') || '10';
  const offset = url.searchParams.get('offset') || '0';
  const status = url.searchParams.get('status') || '';
  const includeItems = url.searchParams.get('include_items') !== 'false'; // Default to true
  
  // Construct the API URL with query parameters
  const apiUrl = new URL('http://localhost:8080/api/orders');
  apiUrl.searchParams.set('limit', limit);
  apiUrl.searchParams.set('offset', offset);
  
  if (status) {
    apiUrl.searchParams.set('status', status);
  }
  
  if (!includeItems) {
    apiUrl.searchParams.set('include_items', 'false');
  }
  
  try {
    // Fetch orders from the API
    const response = await fetch(apiUrl.toString());
    
    if (!response.ok) {
      throw new Error(`API request failed with status ${response.status}`);
    }
    
    // Parse the response and handle null case
    const orders: OrderResponse[] = await response.json() || [];
    
    // Get list of order statuses for filter dropdown
    const statusList = [
      { value: 'pending', label: 'Pending' },
      { value: 'processing', label: 'Processing' },
      { value: 'completed', label: 'Completed' },
      { value: 'cancelled', label: 'Cancelled' },
      { value: 'failed', label: 'Failed' }
    ];
    
    return {
      orders,
      pagination: {
        limit: parseInt(limit),
        offset: parseInt(offset),
        status,
        includeItems
      },
      statusList
    };
  } catch (error) {
    console.error('Failed to fetch orders:', error);
    return {
      orders: [] as OrderResponse[],
      pagination: {
        limit: parseInt(limit),
        offset: parseInt(offset),
        status,
        includeItems
      },
      statusList: [
        { value: 'pending', label: 'Pending' },
        { value: 'processing', label: 'Processing' },
        { value: 'completed', label: 'Completed' },
        { value: 'cancelled', label: 'Cancelled' },
        { value: 'failed', label: 'Failed' }
      ],
      error: 'Failed to load orders. Please try again later.'
    };
  }
};