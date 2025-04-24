// src/routes/admin/products/[id]/+page.server.ts
import { error, fail } from '@sveltejs/kit';
import type { Actions, PageServerLoad } from './$types';

// Interface for product response from API
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

// Interface for product update request
interface UpdateProductRequest {
  name?: string;
  description?: string;
  is_active?: boolean;
  metadata?: Record<string, string>;
}

export const load: PageServerLoad = async ({ params, fetch }) => {
  const { id } = params;
  console.log('Hit endpoint: /products/[id]')
  console.log(`ID: ${id}`)
  
  try {
    // Fetch the product data from the API
    const response = await fetch(`http://localhost:8080/api/products/${id}`);
    
    if (!response.ok) {
      if (response.status === 404) {
        throw error(404, 'Product not found');
      }
      throw error(response.status, `Error fetching product: ${response.statusText}`);
    }
    
    const product: ProductResponse = await response.json();
    
    // Convert the metadata to an array of key-value pairs for easier form handling
    const metadataArray = product.metadata 
      ? Object.entries(product.metadata).map(([key, value]) => ({ key, value }))
      : [];
    
    return {
      product,
      metadataArray
    };
  } catch (err) {
    console.error('Failed to load product:', err);
    throw error(500, 'Failed to load product details');
  }
};

export const actions: Actions = {
  updateProduct: async ({ request, params, fetch }) => {
    const { id } = params;
    const formData = await request.formData();
    
    const name = formData.get('name')?.toString().trim();
    const description = formData.get('description')?.toString().trim();
    const isActive = formData.get('is_active') === 'true';
    
    // Validation
    if (!name) {
      return fail(400, {
        error: 'Product name is required',
        values: { name, description, isActive }
      });
    }
    
    // Process metadata from form
    const metadataKeys = formData.getAll('metadata_key');
    const metadataValues = formData.getAll('metadata_value');
    const metadata: Record<string, string> = {};
    
    for (let i = 0; i < metadataKeys.length; i++) {
      const key = metadataKeys[i]?.toString().trim();
      const value = metadataValues[i]?.toString().trim();
      
      if (key && value !== undefined) {
        // Try to parse as JSON if possible
        try {
          metadata[key] = JSON.parse(value);
        } catch {
          metadata[key] = value;
        }
      }
    }
    
    // Build the update request payload
    const updatePayload: UpdateProductRequest = {
      name,
      description,
      is_active: isActive
    };
    
    // Only include metadata if there are entries
    if (Object.keys(metadata).length > 0) {
      updatePayload.metadata = metadata;
    }
    
    try {
      // Send PUT request to update the product
      const response = await fetch(`http://localhost:8080/api/products/${id}`, {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json'
        },
        body: JSON.stringify(updatePayload)
      });
      
      if (!response.ok) {
        if (response.status === 404) {
          return fail(404, { error: 'Product not found' });
        }
        
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
      
      const updatedProduct: ProductResponse = await response.json();
      
      return { 
        success: true, 
        product: updatedProduct 
      };
    } catch (err) {
      console.error('Failed to update product:', err);
      return fail(500, { 
        error: 'Network error. Please try again.',
        values: { name, description, isActive }
      });
    }
  }
};