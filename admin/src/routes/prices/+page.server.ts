// src/routes/admin/prices/+page.server.ts

import type { PageServerLoad } from './$types';

// Define interface for the ProductResponse (referenced in PriceResponse)
interface ProductResponse {
  id: string;
  name: string;
  description: string;
  is_active: boolean;
  created_at: string;
  updated_at: string;
  stripe_product_id?: string;
  metadata?: Record<string, string>;
}

// Define interface for the PriceResponse
interface PriceResponse {
  id: string;
  product_id: string;
  amount: number;
  currency: string;
  interval_type: string;
  interval_count: number;
  trial_period_days?: number;
  created_at: string;
  updated_at: string;
  stripe_price_id?: string;
  is_active: boolean;
  nickname?: string;
  metadata?: Record<string, string>;
  product?: ProductResponse;
}

export const load: PageServerLoad = async ({ url, fetch }) => {
  // Get query parameters from the URL
  const limit = url.searchParams.get('limit') || '10';
  const offset = url.searchParams.get('offset') || '0';
  const activeOnly = url.searchParams.get('active') === 'true';
  const includeProduct = url.searchParams.get('include_product') === 'true';
  const productId = url.searchParams.get('product_id') || '';
  
  // Construct the API URL with query parameters
  const apiUrl = new URL('http://localhost:8080/api/prices');
  apiUrl.searchParams.set('limit', limit);
  apiUrl.searchParams.set('offset', offset);
  
  if (activeOnly) {
    apiUrl.searchParams.set('active', 'true');
  }
  
  if (includeProduct) {
    apiUrl.searchParams.set('include_product', 'true');
  }
  
  if (productId) {
    apiUrl.searchParams.set('product_id', productId);
  }
  
  try {
    // Fetch prices from the API
    const response = await fetch(apiUrl.toString());
    
    if (!response.ok) {
      throw new Error(`API request failed with status ${response.status}`);
    }
    
    const prices: PriceResponse[] = await response.json();
    
    return {
      prices,
      pagination: {
        limit: parseInt(limit),
        offset: parseInt(offset),
        activeOnly,
        includeProduct,
        productId
      }
    };
  } catch (error) {
    console.error('Failed to fetch prices:', error);
    return {
      prices: [] as PriceResponse[],
      pagination: {
        limit: parseInt(limit),
        offset: parseInt(offset),
        activeOnly,
        includeProduct,
        productId
      },
      error: 'Failed to load prices. Please try again later.'
    };
  }
};