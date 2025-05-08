import type { PageLoad } from './$types';
import { getPrice, getProduct, getCustomers } from '$lib/services/api';


export const load: PageLoad = async ({ params }) => {
  try {
    const priceId = params.priceId;

    // Get the price
    const price = await getPrice(priceId);
    console.log('Price:')
    console.log(price)

    // Get the product
    const product = await getProduct(price.product_id);
    console.log('Product: ')
    console.log(product)

    // Get the first customer (in a real app, this would be the authenticated user)
    const customersResponse = await getCustomers();
    const customer = customersResponse.data[0];
    console.log('Customer: ')
    console.log(customer)

    return {
      price,
      product,
      customer
    };
  } catch (error) {
    console.error('Error loading data:', error);
    return {
      price: null,
      product: null,
      customer: null,
      error: error.message
    };
  }
};