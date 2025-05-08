import type { PageLoad } from './$types';
import { getPrice, getProduct, getCustomers } from '$lib/services/api';


export const load: PageLoad = async ({ params, url }) => {
  try {
    const priceId = params.priceId;
    // Get quantity from URL query parameter, default to 1
    const quantity = parseInt(url.searchParams.get('quantity') || '1', 10);

    // Get the price
    const price = await getPrice(priceId);

    // Get the product
    const product = await getProduct(price.product_id);

    // Get the first customer (in a real app, this would be the authenticated user)
    const customersResponse = await getCustomers();
    const customer = customersResponse.data[0];

    return {
      price,
      product,
      customer,
      quantity
    };
  } catch (error) {
    console.error('Error loading data:', error);
    return {
      price: null,
      product: null,
      customer: null,
      quantity: 1,  // Default quantity
      error: error.message
    };
  }
};