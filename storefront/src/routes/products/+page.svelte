<script lang="ts">
	import type { PageProps } from './$types';
	import { goto } from '$app/navigation';

	let { data }: PageProps = $props();
	let { products, prices, error } = $derived(data);
</script>

<svelte:head>
	<title>Coffee Subscriptions | Walking Drum Coffee</title>
</svelte:head>

<div class="page-header">
	<h1>Coffee Subscriptions</h1>
	<p>Select your favorite coffee and subscription plan</p>
</div>

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
		{#each products as product}
			<div class="product-card">
				<div class="product-image">
					<img src={product.image_url || 'https://unsplash.com/photos/coffee-bean-lot-TD4DBagg2wE'} alt={product.name} />
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
										<button onclick={() => goto(`/checkout/${price.id}`)}> Subscribe </button>
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

	.no-prices {
		color: var(--color-muted);
		font-style: italic;
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

		button {
			width: 100%;
		}
	}
</style>
