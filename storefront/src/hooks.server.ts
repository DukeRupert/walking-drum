import type { Handle } from '@sveltejs/kit';
import { PUBLIC_BASE_URL } from '$env/static/public';

export const handle: Handle = async ({ event, resolve }) => {
    // Proxy API requests to your Go backend during development
    if (event.url.pathname.startsWith('/api/')) {
        const apiUrl = PUBLIC_BASE_URL; // Your Go backend URL
        const targetUrl = new URL(event.url.pathname + event.url.search, apiUrl);

        // Required for CORS to work
        if (event.request.method === 'OPTIONS') {
            return new Response(null, {
                headers: {
                    'Access-Control-Allow-Methods': 'GET, POST, PUT, DELETE, PATCH, OPTIONS',
                    'Access-Control-Allow-Origin': '*',
                    'Access-Control-Allow-Headers': '*',
                }
            });
        }

        try {
            const response = await fetch(targetUrl, {
                method: event.request.method,
                headers: event.request.headers,
                body: event.request.method !== 'GET' && event.request.method !== 'HEAD'
                    ? await event.request.arrayBuffer()
                    : undefined,
            });

            const headers = new Headers();
            response.headers.forEach((value, key) => {
                headers.set(key, value);
            });

            return new Response(response.body, {
                status: response.status,
                statusText: response.statusText,
                headers,
            });
        } catch (error) {
            console.error('API proxy error:', error);
            return new Response(JSON.stringify({ error: 'API request failed' }), {
                status: 500,
                headers: {
                    'Content-Type': 'application/json',
                },
            });
        }
    }

    // For non-API requests, proceed with normal SvelteKit routing
    return resolve(event);
}