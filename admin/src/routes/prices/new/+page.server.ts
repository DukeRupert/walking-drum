// src/routes/admin/prices/new/+page.server.ts

import { fail, redirect } from '@sveltejs/kit';
import type { Actions, PageServerLoad } from './$types';

// Interface for product data
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

// Interface for price creation request
interface CreatePriceRequest {
    product_id: string;
    amount: number;
    currency: string;
    interval_type: string;
    interval_count: number;
    trial_period_days?: number;
    is_active: boolean;
    nickname?: string;
    metadata?: Record<string, string>;
}

// Default values for price
interface DefaultValues {
    amount: number, // $10.00
    currency: string,
    interval_type: string,
    interval_count: number,
    is_active: boolean
};

export const load: PageServerLoad = async ({ fetch }) => {
    // Define default values for the form
    const defaultValues: DefaultValues = {
        amount: 1000, // $10.00
        currency: 'usd',
        interval_type: 'month',
        interval_count: 1,
        is_active: true
    };

    try {
        // Fetch the list of products to populate the product dropdown
        const productsResponse = await fetch('http://localhost:8080/api/products?limit=100');

        if (!productsResponse.ok) {
            throw new Error(`Failed to fetch products: ${productsResponse.statusText}`);
        }

        const products: ProductResponse[] = await productsResponse.json();

        return {
            products: products.filter(p => p.is_active), // Only show active products
            defaultValues,
            currencies: [
                { code: 'usd', name: 'US Dollar ($)' },
                { code: 'eur', name: 'Euro (€)' },
                { code: 'gbp', name: 'British Pound (£)' },
                { code: 'jpy', name: 'Japanese Yen (¥)' },
                { code: 'cad', name: 'Canadian Dollar (C$)' },
                { code: 'aud', name: 'Australian Dollar (A$)' }
            ],
            intervalTypes: [
                { value: 'one_time', label: 'One-time' },
                { value: 'day', label: 'Daily' },
                { value: 'week', label: 'Weekly' },
                { value: 'month', label: 'Monthly' },
                { value: 'year', label: 'Yearly' }
            ]
        };
    } catch (error) {
        console.error('Error loading data for price creation form:', error);
        return {
            products: [],
            defaultValues: defaultValues,
            currencies: [],
            intervalTypes: [],
            error: 'Failed to load required data. Please try again later.'
        };
    }
};

export const actions: Actions = {
    createPrice: async ({ request, fetch }) => {
        const formData = await request.formData();

        const productId = formData.get('product_id')?.toString();
        const amountStr = formData.get('amount')?.toString();
        const currency = formData.get('currency')?.toString()?.toLowerCase();
        const intervalType = formData.get('interval_type')?.toString();
        const intervalCountStr = formData.get('interval_count')?.toString();
        const trialPeriodDaysStr = formData.get('trial_period_days')?.toString();
        const isActive = formData.get('is_active') === 'on';
        const nickname = formData.get('nickname')?.toString();

        // Validate required fields
        if (!productId) {
            return fail(400, { error: 'Product is required', values: Object.fromEntries(formData) });
        }

        if (!amountStr || isNaN(Number(amountStr))) {
            return fail(400, { error: 'Amount must be a valid number', values: Object.fromEntries(formData) });
        }

        const amount = Number(amountStr);
        if (amount <= 0) {
            return fail(400, { error: 'Amount must be greater than zero', values: Object.fromEntries(formData) });
        }

        if (!currency) {
            return fail(400, { error: 'Currency is required', values: Object.fromEntries(formData) });
        }

        if (!intervalType) {
            return fail(400, { error: 'Interval type is required', values: Object.fromEntries(formData) });
        }

        if (!intervalCountStr || isNaN(Number(intervalCountStr))) {
            return fail(400, { error: 'Interval count must be a valid number', values: Object.fromEntries(formData) });
        }

        const intervalCount = Number(intervalCountStr);
        if (intervalCount <= 0) {
            return fail(400, { error: 'Interval count must be greater than zero', values: Object.fromEntries(formData) });
        }

        // Handle optional fields
        let trialPeriodDays: number | undefined;
        if (trialPeriodDaysStr && trialPeriodDaysStr.trim() !== '') {
            trialPeriodDays = Number(trialPeriodDaysStr);
            if (isNaN(trialPeriodDays) || trialPeriodDays < 0) {
                return fail(400, { error: 'Trial period days must be a valid positive number', values: Object.fromEntries(formData) });
            }
        }

        // Handle metadata
        const metadata: Record<string, string> = {};
        const metadataKeys = formData.getAll('metadata_key');
        const metadataValues = formData.getAll('metadata_value');

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

        // Build the request payload
        const payload: CreatePriceRequest = {
            product_id: productId,
            amount: amount,
            currency: currency,
            interval_type: intervalType,
            interval_count: intervalCount,
            is_active: isActive
        };

        if (trialPeriodDays !== undefined) {
            payload.trial_period_days = trialPeriodDays;
        }

        if (nickname && nickname.trim() !== '') {
            payload.nickname = nickname.trim();
        }

        if (Object.keys(metadata).length > 0) {
            payload.metadata = metadata;
        }

        try {
            // Send the request to create the price
            const response = await fetch('http://localhost:8080/api/prices', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify(payload)
            });

            if (!response.ok) {
                const errorText = await response.text();

                if (response.status === 404) {
                    return fail(404, { error: 'Product not found', values: Object.fromEntries(formData) });
                }

                if (response.status === 409) {
                    return fail(409, { error: 'Price with this configuration already exists', values: Object.fromEntries(formData) });
                }

                return fail(response.status, {
                    error: `Error creating price: ${errorText || response.statusText}`,
                    values: Object.fromEntries(formData)
                });
            }

            const createdPrice = await response.json();

            // Redirect to the price details page
            throw redirect(303, `/admin/prices/${createdPrice.id}`);
        } catch (error) {
            if (error instanceof Response) {
                throw error; // This is the redirect
            }

            console.error('Failed to create price:', error);
            return fail(500, {
                error: 'Failed to create price. Please try again.',
                values: Object.fromEntries(formData)
            });
        }
    }
};