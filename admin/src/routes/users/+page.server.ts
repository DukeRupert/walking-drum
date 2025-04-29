// src/routes/admin/users/+page.server.ts

import type { PageServerLoad } from './$types';

interface UserResponse {
  id: string;
  email: string;
  name: string;
  created_at: string;
  updated_at: string;
  stripe_customer_id?: string;
  is_active: boolean;
  metadata?: Record<string, string>;
}

export const load: PageServerLoad = async ({ url, fetch }) => {
  // Get query parameters from the URL
  const limit = url.searchParams.get('limit') || '10';
  const offset = url.searchParams.get('offset') || '0';
  
  // Construct the API URL with query parameters
  const apiUrl = new URL('http://localhost:8080/api/users');
  apiUrl.searchParams.set('limit', limit);
  apiUrl.searchParams.set('offset', offset);
  
  try {
    // Fetch users from the API
    const response = await fetch(apiUrl.toString());
    
    if (!response.ok) {
      throw new Error(`API request failed with status ${response.status}`);
    }
    
    // Parse the response and handle null case
    const users: UserResponse[] = await response.json() || [];
    
    return {
      users,
      pagination: {
        limit: parseInt(limit),
        offset: parseInt(offset)
      }
    };
  } catch (error) {
    console.error('Failed to fetch users:', error);
    return {
      users: [] as UserResponse[],
      pagination: {
        limit: parseInt(limit),
        offset: parseInt(offset)
      },
      error: 'Failed to load users. Please try again later.'
    };
  }
};