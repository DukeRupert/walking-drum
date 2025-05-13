<script lang="ts">
	import type { PageProps } from './$types';
	import { cart } from '$lib/stores/cart';
	import type { Price, Product } from '$lib/types';
	import { onMount } from 'svelte';

	let { data }: PageProps = $props();
	let { products, prices, error } = $derived(data);

	// Track quantities locally
	let quantities: Record<string, number> = $state({});

	function addToCart(product: Product, price: Price) {
		const quantity = quantities[price.id] || 1;
		cart.addItem(product, price, quantity);

		// Show confirmation message
		showAddedToCart(product.name, price.name);
	}

	// Added to cart notification
	let notification = $state({ show: false, message: '' });

	function showAddedToCart(productName: string, priceName: string) {
		notification = {
			show: true,
			message: `Added ${productName} - ${priceName} to your cart`
		};

		// Hide notification after 3 seconds
		setTimeout(() => {
			notification = { ...notification, show: false };
		}, 3000);
	}

	// Handle quantity change
	function updateQuantity(priceId: string, event: Event) {
		const selectElement = event.target as HTMLSelectElement;
		quantities[priceId] = parseInt(selectElement.value, 10);
	}

	onMount(() => {
		// Initialize quantities for each price
		if (prices) {
			prices.forEach((price) => {
				if (quantities[price.id] === undefined) {
					quantities[price.id] = 1;
				}
			});
		}
	});
</script>

<svelte:head>
	<title>Coffee Subscriptions | Walking Drum Coffee</title>
</svelte:head>

<div class="page-header">
	<h1>Coffee Subscriptions</h1>
	<p>Select your favorite coffee and subscription plan</p>
</div>

{#if notification.show}
	<div class="notification">
		<p>{notification.message}</p>
	</div>
{/if}

{#if error}
	<div class="error-container">
		<h3>Error Loading Products</h3>
		<p>{error}</p>
		<button onclick={() => window.location.reload()}>Try Again</button>
	</div>
{:else if products.length === 0}
	<div class="loading-container">
		{#if !error}
			<div class="spinner"></div>
			<p>Loading products...</p>
		{:else}
			<p>No products available at the moment.</p>
		{/if}
	</div>
{:else}
	<div class="products-container">
		{#each products as product (product.id)}
			<div class="product-card">
				<div class="product-image">
					<img
						src={product.image_url ||
							'https://images.unsplash.com/photo-1611854779393-1b2da9d400fe?w=500&auto=format&fit=crop&q=60&ixlib=rb-4.0.3&ixid=M3wxMjA3fDB8MHxzZWFyY2h8Mnx8Y29mZmVlJTIwYmVhbnxlbnwwfHwwfHx8MA%3D%3D'}
						alt={product.name}
					/>
				</div>
				<div class="product-info">
					<h2>{product.name}</h2>
					<p class="product-description">{product.description}</p>

					<div class="product-details">
						<div class="detail">
							<span class="label">Origin</span>
							<span class="value">{product.origin}</span>
						</div>
						<div class="detail">
							<span class="label">Roast Level</span>
							<span class="value">{product.roast_level}</span>
						</div>
						<div class="detail">
							<span class="label">Flavor Notes</span>
							<span class="value">{product.flavor_notes}</span>
						</div>
					</div>

					<div class="subscription-options">
						<h3>Subscription Options</h3>
						{#if prices.filter((price) => price.product_id === product.id).length > 0}
							<div class="price-options">
								{#each prices.filter((price) => price.product_id === product.id) as price}
									<div class="price-option">
										<div class="price-details">
											<span class="price-name">{price.name}</span>
											<span class="price-amount">${(price.amount / 100).toFixed(2)}</span>
											<span class="price-interval">per {price.interval}</span>
										</div>

										<!-- Add quantity selector -->
										<div class="quantity-selector">
											<label for="quantity-{price.id}">Quantity:</label>
											<select
												id="quantity-{price.id}"
												value={quantities[price.id]}
												onchange={(e) => updateQuantity(price.id, e)}
											>
												<option value="1">1</option>
												<option value="2">2</option>
												<option value="3">3</option>
												<option value="4">4</option>
												<option value="5">5</option>
											</select>
										</div>

										<button onclick={() => addToCart(product, price)}> Add to Cart </button>
									</div>
								{/each}
							</div>
						{:else}
							<p class="no-prices">No subscription options available at the moment.</p>
						{/if}
					</div>
				</div>
			</div>
		{/each}
	</div>
{/if}

<style>
	.products-container {
		display: flex;
		flex-direction: column;
		gap: 3rem;
	}

	.product-card {
		display: flex;
		background: white;
		border-radius: var(--radius);
		box-shadow: var(--shadow);
		overflow: hidden;
	}

	.product-image {
		flex: 0 0 300px;
		overflow: hidden;
	}

	.product-image img {
		width: 100%;
		height: 100%;
		object-fit: cover;
	}

	.product-info {
		flex: 1;
		padding: 2rem;
	}

	.product-description {
		color: var(--color-muted);
		margin-bottom: 1.5rem;
		line-height: 1.6;
	}

	.product-details {
		display: grid;
		grid-template-columns: repeat(3, 1fr);
		gap: 1rem;
		margin-bottom: 2rem;
		padding-bottom: 1.5rem;
		border-bottom: 1px solid var(--color-border);
	}

	.detail {
		display: flex;
		flex-direction: column;
	}

	.label {
		font-size: 0.8rem;
		color: var(--color-muted);
		margin-bottom: 0.2rem;
	}

	.value {
		font-size: 1rem;
		color: var(--color-text);
	}

	.subscription-options h3 {
		font-size: 1.2rem;
		margin-bottom: 1rem;
	}

	.price-options {
		display: flex;
		flex-direction: column;
		gap: 1rem;
	}

	.price-option {
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding: 1rem;
		background-color: var(--color-light);
		border-radius: var(--radius);
	}

	.price-details {
		display: flex;
		flex-direction: column;
	}

	.price-name {
		font-size: 1rem;
		color: var(--color-text);
		margin-bottom: 0.2rem;
	}

	.price-amount {
		font-size: 1.4rem;
		font-weight: bold;
		color: var(--color-text);
	}

	.price-interval {
		font-size: 0.8rem;
		color: var(--color-muted);
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

	.no-prices {
		color: var(--color-muted);
		font-style: italic;
	}

	.notification {
		position: fixed;
		top: 20px;
		right: 20px;
		background-color: var(--color-success);
		color: white;
		padding: 1rem;
		border-radius: var(--radius);
		box-shadow: var(--shadow);
		z-index: 1000;
		animation: slideIn 0.3s ease-out;
	}

	@keyframes slideIn {
		from {
			transform: translateX(100%);
			opacity: 0;
		}
		to {
			transform: translateX(0);
			opacity: 1;
		}
	}

	@media (max-width: 768px) {
		.product-card {
			flex-direction: column;
		}

		.product-image {
			flex: 0 0 200px;
		}

		.product-details {
			grid-template-columns: 1fr;
		}

		.price-option {
			flex-direction: column;
			align-items: flex-start;
			gap: 1rem;
		}

		.quantity-selector {
			margin-bottom: 1rem;
		}

		button {
			width: 100%;
		}
	}
</style>
