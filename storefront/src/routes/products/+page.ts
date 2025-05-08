import type { PageLoad } from './$types';

export const load: PageLoad = async ({ fetch }) => {
    try {
        // Fetch products
        const productsResponse = await fetch('/api/v1/products');
        if (!productsResponse.ok) {
            throw new Error('Failed to fetch products');
        }
        const productsData = await productsResponse.json();

        // Fetch prices
        const pricesResponse = await fetch('/api/v1/prices');
        if (!pricesResponse.ok) {
            throw new Error('Failed to fetch prices');
        }
        const pricesData = await pricesResponse.json();

        return {
            products: productsData.data,
            prices: pricesData.data
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