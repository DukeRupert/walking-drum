import type { PageLoad } from './$types';

// Price model definition
interface Price {
  id: string;
  product_id: string;
  name: string;
  amount: number;
  currency: string;
  interval: string;
  interval_count: number;
  active: boolean;
  created_at: string; // ISO date string format
  updated_at: string; // ISO date string format
}

// Product model definition
interface Product {
  id: string;            // UUID in string format
  name: string;
  description: string;
  image_url: string;
  active: boolean;
  stock_level: number;
  weight: number;        // Weight in grams
  origin: string;
  roast_level: string;
  flavor_notes: string;
  stripe_id: string;
  created_at: string;    // ISO date string format
  updated_at: string;    // ISO date string format
}

// Customer model definition
interface Customer {
  id: string;
  email: string;
  first_name: string;
  last_name: string;
  phone_number: string;
  active: boolean;
  created_at: string;    // ISO date string format
  updated_at: string;    // ISO date string format
}

// Paginated customer response
interface CustomerResponse {
  data: Customer[];
  meta: {
    page: number;
    per_page: number;
    total: number;
    total_pages: number;
    has_next: boolean;
    has_prev: boolean;
  }
}


export const load: PageLoad = async ({ params, fetch }) => {
    try {
      const { priceId } = params;
      
      // Fetch price details
      const priceResponse = await fetch(`/api/v1/prices/${priceId}`);
      if (!priceResponse.ok) {
        throw new Error('Failed to fetch price details');
      }
      const priceData: Price = await priceResponse.json();
      console.log(`priceData: ${priceData}`)
      
      // Fetch product details
      const productResponse = await fetch(`/api/v1/products/${priceData.product_id}`);
      if (!productResponse.ok) {
        throw new Error('Failed to fetch product details');
      }
      const productData: Product = await productResponse.json();
      
      // Fetch customer (in a real app, this would be from auth)
      // Here we're just getting the first customer for demonstration
      const customerResponse = await fetch('/api/v1/customers');
      if (!customerResponse.ok) {
        throw new Error('Failed to fetch customer details');
      }
      const customerData: CustomerResponse = await customerResponse.json();
      
      return {
        price: priceData,
        product: productData,
        customer: customerData.data[0]
      };
    } catch (error) {
      console.error('Error loading data:', error);
      return {
        price: null,
        product: null,
        customer: null,
        error: error?.message ?? 'missing error message'
      };
    }
  };