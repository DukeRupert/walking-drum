<!-- src/routes/checkout/result/+page.svelte -->
<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { PUBLIC_API_BASE_URL, PUBLIC_STRIPE_KEY } from '$env/static/public';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();

	let status = $state('loading');

	onMount(async () => {
		try {
			// Get session ID from URL
			const url = new URL(window.location.href);
			const sessionId = url.searchParams.get('session_id');

			if (!sessionId) {
				status = 'error';
				return;
			}

			// Retrieve the session status from your backend
			const response = await fetch(
				`${PUBLIC_API_BASE_URL}/checkout/verify-session?session_id=${sessionId}`
			);

			if (!response.ok) {
				console.error('Failed to verify session');
				status = 'error';
				return;
			}

			const data = await response.json();

			// Check the payment status
			if (data.status === 'complete' || data.status === 'paid') {
				// Success!
				status = 'success';
				// Redirect to success page
				goto('/checkout/success?session_id=' + sessionId);
			} else {
				// Not completed
				status = 'canceled';
				goto('/checkout/cancel');
			}
		} catch (error) {
			console.error(error);
			status = 'error';
		}
	});
</script>

<div class="loading-container">
	{#if status === 'loading'}
		<div class="spinner"></div>
		<p>Processing your payment...</p>
	{:else if status === 'error'}
		<h2>Error</h2>
		<p>There was a problem processing your payment.</p>
		<button onclick={() => goto('/products')}>Return to Products</button>
	{/if}
</div>
