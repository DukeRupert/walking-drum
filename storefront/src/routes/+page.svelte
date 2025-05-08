<script lang="ts">
	import { goto } from '$app/navigation';
</script>

<svelte:head>
	<title>Walking Drum Coffee Subscriptions</title>
</svelte:head>

<section class="hero">
	<div class="hero-content">
		<h1>Premium Coffee, Delivered Fresh</h1>
		<p>Subscribe to our artisan coffee and never run out of your morning brew again.</p>
		<button onclick={() => goto('/products')}>View Subscriptions</button>
	</div>
	<img src="https://unsplash.com/photos/a-large-pile-of-coffee-beans-c7du_McsRKo" alt="Coffee beans and cup" class="hero-image" />
</section>

<section class="features">
	<div class="feature">
		<div class="feature-icon">📦</div>
		<h3>Free Shipping</h3>
		<p>Get your coffee delivered with free shipping on all subscription plans.</p>
	</div>
	<div class="feature">
		<div class="feature-icon">🔄</div>
		<h3>Flexible Schedules</h3>
		<p>Choose weekly, bi-weekly, or monthly deliveries to suit your coffee consumption.</p>
	</div>
	<div class="feature">
		<div class="feature-icon">☕</div>
		<h3>Freshly Roasted</h3>
		<p>All beans are roasted to order, ensuring maximum freshness and flavor.</p>
	</div>
</section>

<style>
	.hero {
		display: flex;
		align-items: center;
		gap: 2rem;
		margin-bottom: 4rem;
	}

	.hero-content {
		flex: 1;
	}

	.hero h1 {
		font-size: 2.8rem;
		margin-bottom: 1rem;
	}

	.hero p {
		font-size: 1.2rem;
		color: var(--color-muted);
		margin-bottom: 2rem;
	}

	.hero-image {
		flex: 1;
		max-width: 50%;
		border-radius: var(--radius);
		object-fit: cover;
	}

	.features {
		display: grid;
		grid-template-columns: repeat(3, 1fr);
		gap: 2rem;
		margin-bottom: 4rem;
	}

	.feature {
		text-align: center;
		padding: 2rem;
		background-color: var(--color-light);
		border-radius: var(--radius);
	}

	.feature-icon {
		font-size: 3rem;
		margin-bottom: 1rem;
	}

	@media (max-width: 768px) {
		.hero {
			flex-direction: column-reverse;
		}

		.hero-image {
			max-width: 100%;
		}

		.features {
			grid-template-columns: 1fr;
		}
	}
</style>
