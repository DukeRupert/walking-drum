// src/routes/admin/users/[id]/+page.server.ts

import type { PageServerLoad } from './$types';
import { error } from '@sveltejs/kit';

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

export const load: PageServerLoad = async ({ params, fetch }) => {
  const userId = params.id;
  
  try {
    // Fetch user from the API
    const response = await fetch(`http://localhost:8080/api/users/${userId}`);
    
    if (response.status === 404) {
      throw error(404, 'User not found');
    }
    
    if (!response.ok) {
      throw error(response.status, `Failed to load user: ${response.statusText}`);
    }
    
    // Parse the response
    const user: UserResponse = await response.json();
    
    return {
      user
    };
  } catch (err) {   
    console.error('Failed to fetch user:', err);
    throw error(500, 'Failed to load user. Please try again later.');
  }
};