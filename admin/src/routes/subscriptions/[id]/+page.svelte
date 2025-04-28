<!-- src/routes/subscriptions/[id]/+page.svelte -->
<script lang="ts">
	import { goto } from '$app/navigation';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();

	let { subscription, metadataArray } = $derived(data);

	// Format date
	function formatDate(dateString?: string): string {
		if (!dateString) return '—';
		return new Date(dateString).toLocaleString();
	}

	// Format date for display (date only)
	function formatDateOnly(dateString?: string): string {
		if (!dateString) return '—';
		return new Date(dateString).toLocaleDateString();
	}

	// Get status badge classes
	function getStatusClasses(status: string): string {
		switch (status.toLowerCase()) {
			case 'active':
				return 'bg-green-50 text-green-700';
			case 'trialing':
				return 'bg-blue-50 text-blue-700';
			case 'past_due':
				return 'bg-yellow-50 text-yellow-700';
			case 'canceled':
				return 'bg-gray-50 text-gray-700';
			case 'incomplete':
				return 'bg-orange-50 text-orange-700';
			case 'incomplete_expired':
				return 'bg-red-50 text-red-700';
			case 'unpaid':
				return 'bg-red-50 text-red-700';
			default:
				return 'bg-gray-50 text-gray-700';
		}
	}

	// Navigate back to the subscriptions list
	function goBackToSubscriptions() {
		goto('/subscriptions');
	}

	// Format JSON string
	function stringifyValue(value: any): string {
		if (typeof value === 'object' && value !== null) {
			return JSON.stringify(value, null, 2);
		}
		return String(value);
	}

	// Calculate remaining days in billing period
	function getRemainingDays(endDate: string): number {
		const end = new Date(endDate);
		const now = new Date();

		const diffTime = end.getTime() - now.getTime();
		const diffDays = Math.ceil(diffTime / (1000 * 60 * 60 * 24));

		return Math.max(0, diffDays);
	}

	// Functions for subscription actions
	function handleCancel() {
		// Implement cancel subscription API call
		alert('Cancel subscription functionality to be implemented');
	}

	function handleUpdate() {
		// Navigate to update page
		goto(`/subscriptions/${subscription.id}/edit`);
	}
</script>

<div class="mx-auto max-w-7xl px-4 py-8 sm:px-6 lg:px-8">
	<div class="flex items-center justify-between">
		<div>
			<h1 class="text-2xl font-semibold text-gray-900">Subscription Details</h1>
			<p class="mt-1 text-sm text-gray-500">
				{subscription.stripe_id}
			</p>
		</div>
		<div class="flex gap-3">
			<button
				type="button"
				onclick={goBackToSubscriptions}
				class="inline-flex items-center rounded-md bg-white px-3 py-2 text-sm font-semibold text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 hover:bg-gray-50"
			>
				Back to Subscriptions
			</button>

			{#if ['active', 'trialing'].includes(subscription.status) && !subscription.cancel_at_period_end}
				<button
					type="button"
					onclick={handleCancel}
					class="inline-flex items-center rounded-md bg-white px-3 py-2 text-sm font-semibold text-red-600 shadow-sm ring-1 ring-inset ring-red-300 hover:bg-red-50"
				>
					Cancel Subscription
				</button>
			{/if}

			<button
				type="button"
				onclick={handleUpdate}
				class="inline-flex items-center rounded-md bg-indigo-600 px-3 py-2 text-sm font-semibold text-white shadow-sm hover:bg-indigo-500 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-indigo-600"
			>
				Update
			</button>
		</div>
	</div>

	<div class="mt-6 grid grid-cols-1 gap-x-6 gap-y-8 md:grid-cols-2">
		<!-- Subscription Summary Card -->
		<div class="bg-white shadow-sm ring-1 ring-gray-900/5 sm:rounded-xl">
			<div class="px-4 py-6 sm:p-8">
				<div class="flex items-center justify-between">
					<div>
						<h2 class="text-base font-semibold leading-7 text-gray-900">Subscription Summary</h2>
						<p class="mt-1 max-w-2xl text-sm leading-6 text-gray-500">
							Created on {formatDate(subscription.created_at)}
						</p>
					</div>
					<span
						class="inline-flex items-center rounded-md px-2 py-1 text-xs font-medium {getStatusClasses(
							subscription.status
						)}"
					>
						{subscription.status.charAt(0).toUpperCase() +
							subscription.status.slice(1).replace(/_/g, ' ')}
					</span>
				</div>

				<div class="mt-6 border-t border-gray-100">
					<dl class="divide-y divide-gray-100">
						<div class="px-4 py-4 sm:grid sm:grid-cols-3 sm:gap-4 sm:px-0">
							<dt class="text-sm font-medium leading-6 text-gray-900">Stripe ID</dt>
							<dd class="mt-1 text-sm leading-6 text-gray-700 sm:col-span-2 sm:mt-0">
								{subscription.stripe_id}
							</dd>
						</div>
						<div class="px-4 py-4 sm:grid sm:grid-cols-3 sm:gap-4 sm:px-0">
							<dt class="text-sm font-medium leading-6 text-gray-900">Internal ID</dt>
							<dd class="mt-1 text-sm leading-6 text-gray-700 sm:col-span-2 sm:mt-0">
								{subscription.id}
							</dd>
						</div>
						<div class="px-4 py-4 sm:grid sm:grid-cols-3 sm:gap-4 sm:px-0">
							<dt class="text-sm font-medium leading-6 text-gray-900">User ID</dt>
							<dd class="mt-1 text-sm leading-6 text-gray-700 sm:col-span-2 sm:mt-0">
								<a
									href={`/users/${subscription.user_id}`}
									class="text-indigo-600 hover:text-indigo-900"
								>
									{subscription.user_id}
								</a>
							</dd>
						</div>
						<div class="px-4 py-4 sm:grid sm:grid-cols-3 sm:gap-4 sm:px-0">
							<dt class="text-sm font-medium leading-6 text-gray-900">Price ID</dt>
							<dd class="mt-1 text-sm leading-6 text-gray-700 sm:col-span-2 sm:mt-0">
								<a
									href={`/prices/${subscription.price_id}`}
									class="text-indigo-600 hover:text-indigo-900"
								>
									{subscription.price_id}
								</a>
							</dd>
						</div>
						<div class="px-4 py-4 sm:grid sm:grid-cols-3 sm:gap-4 sm:px-0">
							<dt class="text-sm font-medium leading-6 text-gray-900">Quantity</dt>
							<dd class="mt-1 text-sm leading-6 text-gray-700 sm:col-span-2 sm:mt-0">
								{subscription.quantity}
							</dd>
						</div>
						{#if subscription.latest_invoice_id}
							<div class="px-4 py-4 sm:grid sm:grid-cols-3 sm:gap-4 sm:px-0">
								<dt class="text-sm font-medium leading-6 text-gray-900">Latest Invoice</dt>
								<dd class="mt-1 text-sm leading-6 text-gray-700 sm:col-span-2 sm:mt-0">
									<a
										href={`/invoices?subscription_id=${subscription.id}`}
										class="text-indigo-600 hover:text-indigo-900"
									>
										{subscription.latest_invoice_id}
									</a>
								</dd>
							</div>
						{/if}
					</dl>
				</div>
			</div>
		</div>

		<!-- Billing Period Card -->
		<div class="bg-white shadow-sm ring-1 ring-gray-900/5 sm:rounded-xl">
			<div class="px-4 py-6 sm:p-8">
				<h2 class="text-base font-semibold leading-7 text-gray-900">Billing Period</h2>

				<div class="mt-6 border-t border-gray-100">
					<dl class="divide-y divide-gray-100">
						<div class="px-4 py-4 sm:grid sm:grid-cols-3 sm:gap-4 sm:px-0">
							<dt class="text-sm font-medium leading-6 text-gray-900">Current Period Start</dt>
							<dd class="mt-1 text-sm leading-6 text-gray-700 sm:col-span-2 sm:mt-0">
								{formatDate(subscription.current_period_start)}
							</dd>
						</div>
						<div class="px-4 py-4 sm:grid sm:grid-cols-3 sm:gap-4 sm:px-0">
							<dt class="text-sm font-medium leading-6 text-gray-900">Current Period End</dt>
							<dd class="mt-1 text-sm leading-6 text-gray-700 sm:col-span-2 sm:mt-0">
								{formatDate(subscription.current_period_end)}
								{#if ['active', 'trialing'].includes(subscription.status)}
									<span class="ml-2 text-xs text-gray-500">
										({getRemainingDays(subscription.current_period_end)} days remaining)
									</span>
								{/if}
							</dd>
						</div>
						{#if subscription.trial_start}
							<div class="px-4 py-4 sm:grid sm:grid-cols-3 sm:gap-4 sm:px-0">
								<dt class="text-sm font-medium leading-6 text-gray-900">Trial Start</dt>
								<dd class="mt-1 text-sm leading-6 text-gray-700 sm:col-span-2 sm:mt-0">
									{formatDate(subscription.trial_start)}
								</dd>
							</div>
						{/if}
						{#if subscription.trial_end}
							<div class="px-4 py-4 sm:grid sm:grid-cols-3 sm:gap-4 sm:px-0">
								<dt class="text-sm font-medium leading-6 text-gray-900">Trial End</dt>
								<dd class="mt-1 text-sm leading-6 text-gray-700 sm:col-span-2 sm:mt-0">
									{formatDate(subscription.trial_end)}
									{#if subscription.status === 'trialing'}
										<span class="ml-2 text-xs text-gray-500">
											({getRemainingDays(subscription.trial_end)} days remaining)
										</span>
									{/if}
								</dd>
							</div>
						{/if}
						<div class="px-4 py-4 sm:grid sm:grid-cols-3 sm:gap-4 sm:px-0">
							<dt class="text-sm font-medium leading-6 text-gray-900">Cancel At Period End</dt>
							<dd class="mt-1 text-sm leading-6 text-gray-700 sm:col-span-2 sm:mt-0">
								{subscription.cancel_at_period_end ? 'Yes' : 'No'}
							</dd>
						</div>
						{#if subscription.cancel_at}
							<div class="px-4 py-4 sm:grid sm:grid-cols-3 sm:gap-4 sm:px-0">
								<dt class="text-sm font-medium leading-6 text-gray-900">Cancel At</dt>
								<dd class="mt-1 text-sm leading-6 text-gray-700 sm:col-span-2 sm:mt-0">
									{formatDate(subscription.cancel_at)}
								</dd>
							</div>
						{/if}
						{#if subscription.canceled_at}
							<div class="px-4 py-4 sm:grid sm:grid-cols-3 sm:gap-4 sm:px-0">
								<dt class="text-sm font-medium leading-6 text-gray-900">Canceled At</dt>
								<dd class="mt-1 text-sm leading-6 text-gray-700 sm:col-span-2 sm:mt-0">
									{formatDate(subscription.canceled_at)}
								</dd>
							</div>
						{/if}
						{#if subscription.ended_at}
							<div class="px-4 py-4 sm:grid sm:grid-cols-3 sm:gap-4 sm:px-0">
								<dt class="text-sm font-medium leading-6 text-gray-900">Ended At</dt>
								<dd class="mt-1 text-sm leading-6 text-gray-700 sm:col-span-2 sm:mt-0">
									{formatDate(subscription.ended_at)}
								</dd>
							</div>
						{/if}
					</dl>
				</div>
			</div>
		</div>

		<!-- Additional Details Card -->
		<div class="bg-white shadow-sm ring-1 ring-gray-900/5 sm:rounded-xl md:col-span-2">
			<div class="px-4 py-6 sm:p-8">
				<h2 class="text-base font-semibold leading-7 text-gray-900">Additional Details</h2>

				<div class="mt-6 border-t border-gray-100">
					<dl class="divide-y divide-gray-100">
						<div class="px-4 py-4 sm:grid sm:grid-cols-3 sm:gap-4 sm:px-0">
							<dt class="text-sm font-medium leading-6 text-gray-900">Customer ID</dt>
							<dd class="mt-1 text-sm leading-6 text-gray-700 sm:col-span-2 sm:mt-0">
								{subscription.customer_id}
							</dd>
						</div>
						{#if subscription.payment_method_id}
							<div class="px-4 py-4 sm:grid sm:grid-cols-3 sm:gap-4 sm:px-0">
								<dt class="text-sm font-medium leading-6 text-gray-900">Payment Method ID</dt>
								<dd class="mt-1 text-sm leading-6 text-gray-700 sm:col-span-2 sm:mt-0">
									{subscription.payment_method_id}
								</dd>
							</div>
						{/if}
						<div class="px-4 py-4 sm:grid sm:grid-cols-3 sm:gap-4 sm:px-0">
							<dt class="text-sm font-medium leading-6 text-gray-900">Created At</dt>
							<dd class="mt-1 text-sm leading-6 text-gray-700 sm:col-span-2 sm:mt-0">
								{formatDate(subscription.created_at)}
							</dd>
						</div>
						<div class="px-4 py-4 sm:grid sm:grid-cols-3 sm:gap-4 sm:px-0">
							<dt class="text-sm font-medium leading-6 text-gray-900">Updated At</dt>
							<dd class="mt-1 text-sm leading-6 text-gray-700 sm:col-span-2 sm:mt-0">
								{formatDate(subscription.updated_at)}
							</dd>
						</div>

						<!-- Metadata -->
						{#if metadataArray.length > 0}
							<div class="px-4 py-4 sm:px-0">
								<dt class="text-sm font-medium leading-6 text-gray-900">Metadata</dt>
								<dd class="mt-2 text-sm text-gray-900">
									<div class="rounded-md border border-gray-200 bg-gray-50 p-4">
										<ul class="space-y-2">
											{#each metadataArray as item}
												<li>
													<span class="font-medium">{item.key}:</span>
													{stringifyValue(item.value)}
												</li>
											{/each}
										</ul>
									</div>
								</dd>
							</div>
						{/if}
					</dl>
				</div>
			</div>
		</div>
	</div>

	<!-- Related Info Card -->
	<div class="mt-6 grid grid-cols-1 gap-x-6 gap-y-8 md:grid-cols-2">
		<!-- Invoices Button -->
		<div class="bg-white p-6 shadow-sm ring-1 ring-gray-900/5 sm:rounded-xl">
			<h3 class="text-base font-semibold leading-7 text-gray-900">Invoices</h3>
			<p class="mt-1 text-sm text-gray-500">View all invoices related to this subscription.</p>
			<div class="mt-4">
				<a
					href={`/invoices?subscription_id=${subscription.id}`}
					class="text-sm font-semibold leading-6 text-indigo-600 hover:text-indigo-500"
				>
					View invoices <span aria-hidden="true">→</span>
				</a>
			</div>
		</div>

		<!-- User Details Button -->
		<div class="bg-white p-6 shadow-sm ring-1 ring-gray-900/5 sm:rounded-xl">
			<h3 class="text-base font-semibold leading-7 text-gray-900">Customer</h3>
			<p class="mt-1 text-sm text-gray-500">
				View details about the customer associated with this subscription.
			</p>
			<div class="mt-4">
				<a
					href={`/users/${subscription.user_id}`}
					class="text-sm font-semibold leading-6 text-indigo-600 hover:text-indigo-500"
				>
					View customer <span aria-hidden="true">→</span>
				</a>
			</div>
		</div>
	</div>
</div>
