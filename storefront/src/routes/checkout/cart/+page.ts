import type { PageLoad } from './$types';
import { getCustomers } from '$lib/services/api';

export const load: PageLoad = async ({ fetch }) => {
    try {
        // Get the first customer (in a real app, this would be the authenticated user)
        const customersResponse = await getCustomers(fetch);
        const customer = customersResponse.data[0];

        return {
            customer
        };
    } catch (error) {
        console.error('Error loading data:', error);
        if (error instanceof Error) {
            return {
                customer: null,
                error: error.message
            };
        } else {
            throw error
        }

    }
};