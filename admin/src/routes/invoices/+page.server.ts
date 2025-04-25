// src/routes/admin/invoices/+page.server.ts
import type { PageServerLoad } from './$types';

// Define interface for the UserResponse (referenced in InvoiceResponse)
interface UserResponse {
    id: string;
    email: string;
    name?: string;
    created_at: string;
    updated_at: string;
}

// Define interface for the SubscriptionResponse (referenced in InvoiceResponse)
interface SubscriptionResponse {
    id: string;
    user_id: string;
    status: string;
    current_period_start: string;
    current_period_end: string;
    created_at: string;
    updated_at: string;
    price_id: string;
    stripe_subscription_id: string;
    cancel_at_period_end: boolean;
    canceled_at?: string;
    trial_start?: string;
    trial_end?: string;
}

// Define interface for the InvoiceResponse
interface InvoiceResponse {
    id: string;
    user_id: string;
    subscription_id?: string;
    status: string;
    amount_due: number;
    amount_paid: number;
    currency: string;
    invoice_pdf?: string;
    stripe_invoice_id: string;
    payment_intent_id?: string;
    period_start?: string;
    period_end?: string;
    created_at: string;
    updated_at: string;
    metadata?: Record<string, string>;
    user?: UserResponse;
    subscription?: SubscriptionResponse;
}

export const load: PageServerLoad = async ({ url, fetch }) => {
    // Get query parameters from the URL
    const limit = url.searchParams.get('limit') || '10';
    const offset = url.searchParams.get('offset') || '0';
    const status = url.searchParams.get('status') || '';
    const userId = url.searchParams.get('user_id') || '';
    const subscriptionId = url.searchParams.get('subscription_id') || '';
    const includeRelations = url.searchParams.get('include_relations') === 'true';

    // Construct the API URL with query parameters
    const apiUrl = new URL('http://localhost:8080/api/invoices');
    apiUrl.searchParams.set('limit', limit);
    apiUrl.searchParams.set('offset', offset);

    if (status) {
        apiUrl.searchParams.set('status', status);
    }

    if (userId) {
        apiUrl.searchParams.set('user_id', userId);
    }

    if (subscriptionId) {
        apiUrl.searchParams.set('subscription_id', subscriptionId);
    }

    if (includeRelations) {
        apiUrl.searchParams.set('include_relations', 'true');
    }

    try {
        // Fetch invoices from the API
        const response = await fetch(apiUrl.toString());

        if (!response.ok) {
            throw new Error(`API request failed with status ${response.status}`);
        }

        const invoices: InvoiceResponse[] = await response.json() || [];
       console.log(invoices)

        // Get list of invoice statuses for filter dropdown
        const statusList = [
            { value: 'draft', label: 'Draft' },
            { value: 'open', label: 'Open' },
            { value: 'paid', label: 'Paid' },
            { value: 'uncollectible', label: 'Uncollectible' },
            { value: 'void', label: 'Void' }
        ];

        return {
            invoices,
            pagination: {
                limit: parseInt(limit),
                offset: parseInt(offset),
                status,
                userId,
                subscriptionId,
                includeRelations
            },
            statusList
        };
    } catch (error) {
        console.error('Failed to fetch invoices:', error);
        return {
            invoices: [] as InvoiceResponse[],
            pagination: {
                limit: parseInt(limit),
                offset: parseInt(offset),
                status,
                userId,
                subscriptionId,
                includeRelations
            },
            statusList: [],
            error: 'Failed to load invoices. Please try again later.'
        };
    }
};