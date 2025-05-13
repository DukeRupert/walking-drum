<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { createCheckoutSession } from '$lib/services/stripe';
	import { PUBLIC_STRIPE_KEY } from '$env/static/public';
	import type { PageProps } from './$types';
	import type { Stripe, StripeEmbeddedCheckout } from '@stripe/stripe-js';

	let { data }: PageProps = $props();
	let { price, product, quantity, customer, error } = $derived(data);
	$inspect(data)

	let loading = $state(false);
	let checkoutError: string | null = $state(null);
	let clientSecret: string | null = $state(null);
	let stripe: Stripe | null;
	let checkout: StripeEmbeddedCheckout | null;

	onMount(async () => {
		// Load Stripe.js
		const { loadStripe } = await import('@stripe/stripe-js');
		stripe = await loadStripe(PUBLIC_STRIPE_KEY);

		if (!stripe) {
			checkoutError = 'Failed to load Stripe';
			return;
		}

		// Create a checkout session
		try {
			loading = true;
			const response = await createCheckoutSession(price.id, customer.id, quantity);
			clientSecret = response.client_secret;

			// Initialize the embedded checkout
			if (clientSecret) {
				checkout = await stripe.initEmbeddedCheckout({
					clientSecret
				});

				// Mount the checkout form
				checkout.mount('#checkout');
			}
		} catch (err) {
			checkoutError = err.message || 'An error occurred';
		} finally {
			loading = false;
		}
	});

	// Clean up checkout on component unmount
	function handleUnmount() {
		if (checkout) {
			checkout.destroy();
		}
	}
</script>

<svelte:head>
	<title>Checkout | Walking Drum Coffee</title>
	<script>
		window.addEventListener('beforeunload', () => {
			window.dispatchEvent(new Event('unmount-checkout'));
		});
	</script>
</svelte:head>

<svelte:window on:unmount-checkout={handleUnmount} />

<div class="checkout-page">
	<div class="checkout-header">
		<h1>Complete Your Subscription</h1>
	</div>

	{#if error}
		<div class="error-container">
			<h3>Error Loading Checkout</h3>
			<p>{error}</p>
			<button onclick={() => goto('/products')}>Return to Products</button>
		</div>
	{:else if loading}
		<div class="loading-container">
			<div class="spinner"></div>
			<p>Preparing your checkout...</p>
		</div>
	{:else if checkoutError}
		<div class="error-container">
			<h3>Checkout Error</h3>
			<p>{checkoutError}</p>
			<button onclick={() => window.location.reload()}>Try Again</button>
		</div>
	{/if}
	<div class="checkout-content">
		<div class="order-summary">
			<h3>Order Summary</h3>

			<div class="product-summary">
				<img src={product?.image_url || '/images/coffee-placeholder.jpg'} alt={product?.name} />
				<div class="product-summary-details">
					<h4>{product?.name}</h4>
					<p>{price?.name}</p>
				</div>
			</div>

			<div class="price-breakdown">
        <div class="price-row">
          <span>Subscription</span>
          <span>${((price?.amount || 0) / 100).toFixed(2)}/{price?.interval}</span>
        </div>
        
        <div class="price-row">
          <span>Quantity</span>
          <span>{quantity}</span>
        </div>
        
        <div class="price-total">
          <span>Total per {price?.interval}</span>
          <span>${((price?.amount || 0) / 100 * quantity).toFixed(2)}</span>
        </div>
      </div>

			<div class="subscription-details">
				<h4>Subscription Details</h4>
				<p>You will receive your first delivery within 3-5 business days.</p>
				<p>Your subscription will automatically renew every {price?.interval}.</p>
				<p>You can pause or cancel anytime from your account.</p>
			</div>
		</div>
		<div id="checkout"></div>
	</div>
</div>

<style>
	.checkout-page {
		max-width: 1000px;
		margin: 0 auto;
	}

	.checkout-header {
		text-align: center;
		margin-bottom: 2rem;
	}

	.checkout-content {
		display: grid;
		grid-template-columns: 1fr 2fr;
		gap: 2rem;
		align-items: start;
	}

	.order-summary {
		background-color: white;
		border-radius: var(--radius);
		box-shadow: var(--shadow);
		padding: 1.5rem;
		position: sticky;
		top: 2rem;
	}

	.order-summary h3 {
		margin-bottom: 1.5rem;
		padding-bottom: 1rem;
		border-bottom: 1px solid var(--color-border);
	}

	.product-summary {
		display: flex;
		gap: 1rem;
		margin-bottom: 1.5rem;
		padding-bottom: 1.5rem;
		border-bottom: 1px solid var(--color-border);
	}

	.product-summary img {
		width: 80px;
		height: 80px;
		object-fit: cover;
		border-radius: var(--radius);
	}

	.product-summary-details h4 {
		margin-bottom: 0.2rem;
		font-size: 1rem;
	}

	.product-summary-details p {
		color: var(--color-muted);
		font-size: 0.9rem;
		margin: 0;
	}

	.price-breakdown {
		margin-bottom: 1.5rem;
	}

	.price-row {
		display: flex;
		justify-content: space-between;
		margin-bottom: 0.8rem;
		color: var(--color-muted);
	}

	.price-total {
		display: flex;
		justify-content: space-between;
		font-weight: bold;
		font-size: 1.1rem;
		margin-top: 1rem;
		padding-top: 1rem;
		border-top: 1px solid var(--color-border);
	}

	.subscription-details {
		margin-top: 1.5rem;
		padding-top: 1.5rem;
		border-top: 1px solid var(--color-border);
	}

	.subscription-details h4 {
		margin-bottom: 0.8rem;
		font-size: 1rem;
	}

	.subscription-details p {
		font-size: 0.9rem;
		color: var(--color-muted);
		margin-bottom: 0.5rem;
	}

	.checkout-form {
		background-color: white;
		border-radius: var(--radius);
		box-shadow: var(--shadow);
		overflow: hidden;
		min-height: 400px;
	}

	#checkout-container {
		width: 100%;
		min-height: 400px;
	}

	@media (max-width: 768px) {
		.checkout-content {
			grid-template-columns: 1fr;
		}

		.order-summary {
			position: static;
			margin-bottom: 2rem;
		}
	}
</style>
