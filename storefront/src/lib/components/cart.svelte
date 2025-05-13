<script lang="ts">
	import { cart, type CartItem } from '$lib/stores/cart';
	import { goto } from '$app/navigation';
    
    let { isOpen = $bindable(false)} = $props();

	function formatCurrency(amount: number): string {
		return (amount / 100).toFixed(2);
	}

	function getTotalPrice(): number {
		return $cart.items.reduce((sum, item) => sum + item.price.amount * item.quantity, 0);
	}

	function getItemCount(): number {
		return $cart.items.reduce((sum, item) => sum + item.quantity, 0);
	}

	function toggleCart() {
		isOpen = !isOpen;
	}

	function handleCheckout() {
		if ($cart.items.length > 0) {
			goto('/checkout/cart');
		}
	}
</script>

<div class="cart-widget">
	<button
		class="cart-button"
		onclick={toggleCart}
		aria-label="Shopping cart"
		aria-expanded={isOpen}
	>
		<svg
			class="w-6 h-6"
			fill="none"
			stroke="currentColor"
			viewBox="0 0 24 24"
			xmlns="http://www.w3.org/2000/svg"
		>
			<path
				stroke-linecap="round"
				stroke-linejoin="round"
				stroke-width="2"
				d="M3 3h2l.4 2M7 13h10l4-8H5.4M7 13L5.4 5M7 13l-2.293 2.293c-.63.63-.184 1.707.707 1.707H17m0 0a2 2 0 100 4 2 2 0 000-4zm-8 2a2 2 0 11-4 0 2 2 0 014 0z"
			></path>
		</svg>
		<span class="cart-count">{getItemCount()}</span>
	</button>

	{#if isOpen}
		<div class="cart-dropdown">
			<div class="cart-header">
				<h3>Your Subscription Cart</h3>
				<button onclick={toggleCart} class="close-button" aria-label="Close cart">
					&times;
				</button>
			</div>

			{#if $cart.items.length === 0}
				<div class="empty-cart">
					<p>Your cart is empty</p>
					<button onclick={() => goto('/products')} class="continue-shopping">
						Browse Subscriptions
					</button>
				</div>
			{:else}
				<div class="cart-items">
					{#each $cart.items as item}
						<div class="cart-item">
							<div class="item-info">
								<h4>{item.product.name}</h4>
								<p>{item.price.name} - ${formatCurrency(item.price.amount)}/{item.price.interval}</p>
							</div>
							<div class="item-quantity">
								<label for="quantity-{item.priceId}">Qty:</label>
								<select
									id="quantity-{item.priceId}"
									bind:value={item.quantity}
									onchange={() => cart.updateQuantity(item.priceId, item.quantity)}
								>
									{#each Array(5) as _, i}
										<option value={i + 1}>{i + 1}</option>
									{/each}
								</select>
								<button
									onclick={() => cart.removeItem(item.priceId)}
									class="remove-button"
									aria-label="Remove item"
								>
									&times;
								</button>
							</div>
						</div>
					{/each}
				</div>

				<div class="cart-footer">
					<div class="cart-total">
						<span>Total per cycle:</span>
						<span>${formatCurrency(getTotalPrice())}</span>
					</div>
					<button onclick={handleCheckout} class="checkout-button">
						Proceed to Checkout
					</button>
					<button onclick={() => goto('/products')} class="continue-shopping">
						Continue Shopping
					</button>
				</div>
			{/if}
		</div>
	{/if}
</div>

<style>
	.cart-widget {
		position: relative;
	}

	.cart-button {
		position: relative;
		background: none;
		border: none;
		cursor: pointer;
		color: var(--color-text);
	}

	.cart-count {
		position: absolute;
		top: -8px;
		right: -8px;
		background-color: var(--color-secondary);
		color: white;
		border-radius: 50%;
		width: 20px;
		height: 20px;
		display: flex;
		align-items: center;
		justify-content: center;
		font-size: 0.75rem;
		font-weight: bold;
	}

	.cart-dropdown {
		position: absolute;
		top: 100%;
		right: 0;
		width: 320px;
		background-color: white;
		border-radius: var(--radius);
		box-shadow: var(--shadow);
		z-index: 100;
		margin-top: 0.5rem;
	}

	.cart-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding: 1rem;
		border-bottom: 1px solid var(--color-border);
	}

	.cart-header h3 {
		margin: 0;
		font-size: 1.1rem;
	}

	.close-button {
		background: none;
		border: none;
		font-size: 1.5rem;
		cursor: pointer;
		color: var(--color-muted);
	}

	.empty-cart {
		padding: 2rem 1rem;
		text-align: center;
		color: var(--color-muted);
	}

	.cart-items {
		max-height: 300px;
		overflow-y: auto;
		padding: 0.5rem;
	}

	.cart-item {
		display: flex;
		justify-content: space-between;
		padding: 0.75rem;
		border-bottom: 1px solid var(--color-border);
	}

	.item-info {
		flex: 1;
	}

	.item-info h4 {
		margin: 0 0 0.25rem;
		font-size: 0.9rem;
	}

	.item-info p {
		margin: 0;
		font-size: 0.8rem;
		color: var(--color-muted);
	}

	.item-quantity {
		display: flex;
		align-items: center;
		gap: 0.5rem;
	}

	.item-quantity label {
		font-size: 0.8rem;
		color: var(--color-muted);
	}

	.item-quantity select {
		width: 50px;
		padding: 0.25rem;
		border-radius: 4px;
		border: 1px solid var(--color-border);
	}

	.remove-button {
		background: none;
		border: none;
		color: var(--color-muted);
		font-size: 1.2rem;
		cursor: pointer;
		padding: 0 0.25rem;
	}

	.cart-footer {
		padding: 1rem;
		border-top: 1px solid var(--color-border);
	}

	.cart-total {
		display: flex;
		justify-content: space-between;
		margin-bottom: 1rem;
		font-weight: bold;
	}

	.checkout-button {
		display: block;
		width: 100%;
		padding: 0.75rem;
		background-color: var(--color-primary);
		color: white;
		border: none;
		border-radius: var(--radius);
		font-weight: bold;
		cursor: pointer;
		margin-bottom: 0.5rem;
	}

	.continue-shopping {
		display: block;
		width: 100%;
		padding: 0.75rem;
		background-color: white;
		color: var(--color-text);
		border: 1px solid var(--color-border);
		border-radius: var(--radius);
		cursor: pointer;
	}
</style>