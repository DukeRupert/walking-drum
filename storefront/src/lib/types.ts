export interface Product {
    id: string;            
    name: string;
    description: string;
    image_url: string;
    active: boolean;
    stock_level: number;   
    weight: number;        
    origin: string;
    roast_level: string;
    flavor_notes: string;
    stripe_id: string;
    created_at: string;    
    updated_at: string;    
  }
  
  export interface Price {
    id: string;
    product_id: string;
    name: string;
    amount: number;
    currency: string;
    interval: string;
    interval_count: number;
    active: boolean;
    created_at: string;
    updated_at: string;
  }
  
  export interface Customer {
    id: string;
    email: string;
    first_name: string;
    last_name: string;
    phone_number: string;
    active: boolean;
    created_at: string;
    updated_at: string;
  }
  
  export interface ProductResponse {
    data: Product[];
    meta: {
      page: number;
      per_page: number;
      total: number;
      total_pages: number;
      has_next: boolean;
      has_prev: boolean;
    }
  }
  
  export interface PriceResponse {
    data: Price[];
    meta: {
      page: number;
      per_page: number;
      total: number;
      total_pages: number;
      has_next: boolean;
      has_prev: boolean;
    }
  }
  
  export interface CustomerResponse {
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