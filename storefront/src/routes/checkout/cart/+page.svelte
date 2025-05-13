<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { cart } from '$lib/stores/cart';
	import { createMultipleItemCheckoutSession } from '$lib/services/stripe';
	import { PUBLIC_STRIPE_KEY } from '$env/static/public';
	import type { PageProps } from './$types';
	import type { Stripe, StripeEmbeddedCheckout } from '@stripe/stripe-js';

	let { data }: PageProps = $props();
	let { customer, error } = $derived(data);

	let loading = $state(false);
	let checkoutError: string | null = $state(null);
	let clientSecret: string | null = $state(null);
	let stripe: Stripe | null;
	let checkout: StripeEmbeddedCheckout | null;

	// Format currency for display
	function formatCurrency(amount: number): string {
		return (amount / 100).toFixed(2);
	}

	// Calculate the total amount
	function getTotalAmount(): number {
		return $cart.items.reduce((sum, item) => sum + item.price.amount * item.quantity, 0);
	}

	onMount(async () => {
		// If cart is empty, redirect to products page
		if ($cart.items.length === 0) {
			goto('/products');
			return;
		}

		// Load Stripe.js
		const { loadStripe } = await import('@stripe/stripe-js');
		stripe = await loadStripe(PUBLIC_STRIPE_KEY);

		if (!stripe) {
			checkoutError = 'Failed to load Stripe';
			return;
		}

		// Create a checkout session for multiple items
		try {
			loading = true;
			const items = $cart.items.map(item => ({
				price_id: item.priceId,
				quantity: item.quantity
			}));

			const response = await createMultipleItemCheckoutSession(items, customer?.id ?? '');
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

	// Remove an item from cart
	function removeItem(priceId: string) {
		cart.removeItem(priceId);
		if ($cart.items.length === 0) {
			goto('/products');
		}
	}

	// Update quantity of an item
	function updateQuantity(priceId: string, quantity: number) {
		cart.updateQuantity(priceId, quantity);
	}
</script>

<svelte:head>
	<title>Cart Checkout | Walking Drum Coffee</title>
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

			{#if $cart.items.length > 0}
				<div class="cart-items">
					{#each $cart.items as item}
						<div class="cart-item">
							<div class="item-info">
								<h4>{item.product.name}</h4>
								<p class="item-description">
									{item.price.name} - ${formatCurrency(item.price.amount)}/{item.price.interval}
								</p>
							</div>
							<div class="item-actions">
								<div class="quantity-selector">
									<label for="quantity-{item.priceId}">Qty:</label>
									<select
										id="quantity-{item.priceId}"
										bind:value={item.quantity}
										onchange={() => updateQuantity(item.priceId, item.quantity)}
										disabled={loading || !!clientSecret}
									>
										{#each Array(5) as _, i}
											<option value={i + 1}>{i + 1}</option>
										{/each}
									</select>
								</div>
								<button
									class="remove-button"
									onclick={() => removeItem(item.priceId)}
									disabled={loading || !!clientSecret}
								>
									Remove
								</button>
							</div>
							<div class="item-total">
								${formatCurrency(item.price.amount * item.quantity)}
							</div>
						</div>
					{/each}
				</div>

				<div class="price-breakdown">
					<div class="price-total">
						<span>Total per Billing Cycle</span>
						<span>${formatCurrency(getTotalAmount())}</span>
					</div>
				</div>

				<div class="subscription-details">
					<h4>Subscription Details</h4>
					<p>You will receive your first delivery within 3-5 business days.</p>
					<p>Your subscription will automatically renew based on each product's billing interval.</p>
					<p>You can pause or cancel anytime from your account.</p>
				</div>
			{:else}
				<div class="empty-cart">
					<p>Your cart is empty</p>
					<button onclick={() => goto('/products')}>Browse Products</button>
				</div>
			{/if}
		</div>
		<div id="checkout" class="checkout-form"></div>
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

	.cart-items {
		margin-bottom: 1.5rem;
	}

	.cart-item {
		display: grid;
		grid-template-columns: 1fr auto auto;
		gap: 1rem;
		padding: 1rem 0;
		border-bottom: 1px solid var(--color-border);
	}

	.item-info h4 {
		margin: 0 0 0.5rem;
		font-size: 1rem;
	}

	.item-description {
		margin: 0;
		font-size: 0.9rem;
		color: var(--color-muted);
	}

	.item-actions {
		display: flex;
		flex-direction: column;
		gap: 0.5rem;
	}

	.quantity-selector {
		display: flex;
		align-items: center;
		gap: 0.5rem;
	}

	.quantity-selector label {
		font-size: 0.9rem;
		color: var(--color-muted);
	}

	.quantity-selector select {
		padding: 0.3rem 0.5rem;
		border-radius: var(--radius);
		border: 1px solid var(--color-border);
		background-color: white;
	}

	.remove-button {
		background: none;
		border: none;
		color: var(--color-secondary);
		font-size: 0.9rem;
		cursor: pointer;
		padding: 0;
		text-decoration: underline;
	}

	.item-total {
		font-weight: bold;
		text-align: right;
	}

	.price-breakdown {
		margin: 1.5rem 0;
		padding-top: 1rem;
	}

	.price-total {
		display: flex;
		justify-content: space-between;
		font-weight: bold;
		font-size: 1.1rem;
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

	.empty-cart {
		text-align: center;
		padding: 2rem 0;
		color: var(--color-muted);
	}

	.checkout-form {
		background-color: white;
		border-radius: var(--radius);
		box-shadow: var(--shadow);
		overflow: hidden;
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

		.cart-item {
			grid-template-columns: 1fr;
		}

		.item-actions {
			flex-direction: row;
			justify-content: space-between;
			margin: 0.5rem 0;
		}

		.item-total {
			text-align: left;
		}
	}
</style>