import type { PageLoad } from './$types';
import { getProducts, getPrices } from '$lib/services/api';

export const load: PageLoad = async () => {
    try {
        const [productsResponse, pricesResponse] = await Promise.all([
          getProducts(),
          getPrices()
        ]);
    
        return {
          products: productsResponse.data,
          prices: pricesResponse.data
        };
      } catch (error) {
        console.error('Error loading data:', error);
        return {
          products: [],
          prices: [],
          error: error.message
        };
      }
};