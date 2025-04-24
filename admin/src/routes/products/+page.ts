// src/routes/admin/products/+page.ts

import type { PageLoad } from './$types';

// Interface matching the Go ProductResponse struct
interface ProductResponse {
  id: string;
  name: string;
  description: string;
  is_active: boolean;
  created_at: string;
  updated_at: string;
  stripe_product_id?: string;
  metadata?: Record<string, any>;
}

export const load: PageLoad = async ({ url, fetch }) => {
  // Get query parameters from the URL
  const limit = url.searchParams.get('limit') || '10';
  const offset = url.searchParams.get('offset') || '0';
  const activeOnly = url.searchParams.get('active') === 'true';
  
  // Construct the API URL with query parameters
  const apiUrl = new URL('http://localhost:8080/api/products');
  apiUrl.searchParams.set('limit', limit);
  apiUrl.searchParams.set('offset', offset);
  if (activeOnly) {
    apiUrl.searchParams.set('active', 'true');
  }
  
  try {
    // Fetch products from the API
    const response = await fetch(apiUrl.toString());
    
    if (!response.ok) {
      throw new Error(`API request failed with status ${response.status}`);
    }
    
    const products: ProductResponse[] = await response.json();
    
    return {
      products,
      pagination: {
        limit: parseInt(limit),
        offset: parseInt(offset),
        activeOnly
      }
    };
  } catch (error) {
    console.error('Failed to fetch products:', error);
    return {
      products: [] as ProductResponse[],
      pagination: {
        limit: parseInt(limit),
        offset: parseInt(offset),
        activeOnly
      },
      error: 'Failed to load products. Please try again later.'
    };
  }
};