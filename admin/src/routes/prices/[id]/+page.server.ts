// src/routes/admin/prices/[id]/+page.server.ts

import { error } from '@sveltejs/kit';
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

export const load: PageServerLoad = async ({ params, fetch, url }) => {
  const { id } = params;
  const includeProduct = url.searchParams.get('include_product') === 'true' || true; // Default to true
  
  try {
    // Build the API URL
    const apiUrl = new URL(`http://localhost:8080/api/prices/${id}`);
    
    // Add query parameters
    if (includeProduct) {
      apiUrl.searchParams.set('include_product', 'true');
    }
    
    // Fetch the price data
    const response = await fetch(apiUrl.toString());
    
    if (!response.ok) {
      if (response.status === 404) {
        throw error(404, 'Price not found');
      }
      throw error(response.status, `Error fetching price: ${response.statusText}`);
    }
    
    const price: PriceResponse = await response.json();
    
    // Convert the metadata to an array of key-value pairs for easier display
    const metadataArray = price.metadata 
      ? Object.entries(price.metadata).map(([key, value]) => ({ key, value }))
      : [];
    
    return {
      price,
      metadataArray,
      includeProduct
    };
  } catch (err) {
    console.error('Failed to load price:', err);
    throw error(500, 'Failed to load price details');
  }
};