// src/routes/admin/products/+page.server.ts

import { fail } from '@sveltejs/kit';
import type { Actions, PageServerLoad } from './$types';

// Define the interface for the product API response
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

// Define the create product request structure
interface CreateProductRequest {
  name: string;
  description: string;
  is_active: boolean;
  metadata?: Record<string, string>;
}

export const load: PageServerLoad = async ({ url, fetch }) => {
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

export const actions: Actions = {
  createProduct: async ({ request, fetch }) => {
    // Parse the form data
    const formData = await request.formData();
    
    const name = formData.get('name')?.toString() || '';
    const description = formData.get('description')?.toString() || '';
    const isActive = formData.get('is_active') === 'true';
    
    // Basic validation
    if (!name.trim()) {
      return fail(400, { 
        error: 'Product name is required',
        values: { name, description, isActive }
      });
    }
    
    // Handle metadata
    const metadata: Record<string, string> = {};
    const metadataKeys = formData.getAll('metadata_key');
    const metadataValues = formData.getAll('metadata_value');
    
    for (let i = 0; i < metadataKeys.length; i++) {
      const key = metadataKeys[i]?.toString().trim();
      const value = metadataValues[i]?.toString().trim();
      
      if (key && value !== undefined) {
        // Try to parse the value as JSON (for numbers, booleans, etc.)
        try {
          metadata[key] = JSON.parse(value);
        } catch {
          // If parsing fails, use the value as a string
          metadata[key] = value;
        }
      }
    }
    
    // Create the request payload
    const payload: CreateProductRequest = {
      name: name.trim(),
      description: description.trim(),
      is_active: isActive
    };
    
    // Only include metadata if we have some
    if (Object.keys(metadata).length > 0) {
      payload.metadata = metadata;
    }
    
    try {
      // Send the request to the API
      const response = await fetch('http://localhost:8080/api/products', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json'
        },
        body: JSON.stringify(payload)
      });
      
      // Handle API errors
      if (!response.ok) {
        if (response.status === 409) {
          return fail(409, { 
            error: 'A product with this name already exists',
            values: { name, description, isActive }
          });
        }
        
        const errorText = await response.text();
        return fail(response.status, { 
          error: `API Error: ${errorText || response.statusText}`,
          values: { name, description, isActive }
        });
      }
      
      // Parse and return the successful response
      const product: ProductResponse = await response.json();
      
      return { success: true, product };
    } catch (error) {
      console.error('Failed to create product:', error);
      return fail(500, { 
        error: 'Network error. Please try again.',
        values: { name, description, isActive }
      });
    }
  }
};