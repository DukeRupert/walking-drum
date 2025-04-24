<!-- src/routes/admin/products/[id]/+page.svelte -->
<script lang="ts">
	import { enhance } from '$app/forms';
	import { goto } from '$app/navigation';
	import type { PageProps } from './$types';

	let { data, form }: PageProps = $props();
	let { product, metadataArray } = $derived(data);

	// Form state
	let metadataFields = $derived(
		metadataArray.length > 0 ? [...metadataArray] : [{ key: '', value: '' }]
	);
	let isSubmitting = $state(false);

	function formatDate(dateString: string) {
		return new Date(dateString).toLocaleString();
	}

	function addMetadataField() {
		metadataFields = [...metadataFields, { key: '', value: '' }];
	}

	function removeMetadataField(index: number) {
		metadataFields = metadataFields.filter((_, i) => i !== index);
	}

	function handleCancel() {
		goto('/products');
	}

	function stringifyMetadataValue(value: any): string {
		if (typeof value === 'string') {
			return value;
		}
		return JSON.stringify(value);
	}
</script>

<div class="mx-auto max-w-4xl px-4 py-8 sm:px-6 lg:px-8">
	<div class="flex items-center justify-between">
		<div>
			<h1 class="text-2xl font-semibold text-gray-900">Edit Product</h1>
			<p class="mt-1 text-sm text-gray-500">
				Update the details for {product.name}
			</p>
		</div>
		<div>
			<button
				type="button"
				onclick={handleCancel}
				class="inline-flex items-center rounded-md bg-white px-3 py-2 text-sm font-semibold text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 hover:bg-gray-50"
			>
				Back to Products
			</button>
		</div>
	</div>

	<div class="mt-6 bg-white shadow-sm ring-1 ring-gray-900/5 sm:rounded-xl md:col-span-2">
		{#if form?.success}
			<div class="p-6">
				<div class="rounded-md bg-green-50 p-4">
					<div class="flex">
						<div class="flex-shrink-0">
							<svg
								class="h-5 w-5 text-green-400"
								viewBox="0 0 20 20"
								fill="currentColor"
								aria-hidden="true"
							>
								<path
									fill-rule="evenodd"
									d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.857-9.565a.75.75 0 00-1.214-.882l-3.483 4.79-1.88-1.88a.75.75 0 10-1.06 1.061l2.5 2.5a.75.75 0 001.137-.089l4-5.5z"
									clip-rule="evenodd"
								/>
							</svg>
						</div>
						<div class="ml-3">
							<p class="text-sm font-medium text-green-800">Product updated successfully!</p>
						</div>
					</div>
				</div>

				<div class="mt-4 flex justify-end">
					<button
						type="button"
						onclick={handleCancel}
						class="rounded-md bg-indigo-600 px-3 py-2 text-sm font-semibold text-white shadow-sm hover:bg-indigo-500 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-indigo-600"
					>
						Return to Products
					</button>
				</div>
			</div>
		{:else}
			<form
				method="POST"
				action="?/updateProduct"
				use:enhance={() => {
					isSubmitting = true;

					return async ({ update, result }) => {
						await update();
						isSubmitting = false;
						if (result.type == 'success') {
							setTimeout(() => {
								handleCancel();
							}, 750);
						}
					};
				}}
			>
				<div class="px-4 py-6 sm:p-8">
					{#if form?.error}
						<div class="mb-6 rounded-md bg-red-50 p-4">
							<div class="flex">
								<div class="flex-shrink-0">
									<svg
										class="h-5 w-5 text-red-400"
										viewBox="0 0 20 20"
										fill="currentColor"
										aria-hidden="true"
									>
										<path
											fill-rule="evenodd"
											d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.28 7.22a.75.75 0 00-1.06 1.06L8.94 10l-1.72 1.72a.75.75 0 101.06 1.06L10 11.06l1.72 1.72a.75.75 0 101.06-1.06L11.06 10l1.72-1.72a.75.75 0 00-1.06-1.06L10 8.94 8.28 7.22z"
											clip-rule="evenodd"
										/>
									</svg>
								</div>
								<div class="ml-3">
									<h3 class="text-sm font-medium text-red-800">{form.error}</h3>
								</div>
							</div>
						</div>
					{/if}

					<div class="space-y-6">
						<!-- Product ID (readonly) -->
						<div>
							<label for="productId" class="block text-sm font-medium leading-6 text-gray-900">
								Product ID
							</label>
							<div class="mt-2">
								<input
									type="text"
									id="productId"
									value={product.id}
									readonly
									class="block w-full rounded-md border-0 bg-gray-50 px-2 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm sm:leading-6"
								/>
							</div>
						</div>

						<!-- Created/Updated timestamps (readonly) -->
						<div class="grid grid-cols-1 gap-4 sm:grid-cols-2">
							<div>
								<label for="createdAt" class="block text-sm font-medium leading-6 text-gray-900">
									Created At
								</label>
								<div class="mt-2">
									<input
										type="text"
										id="createdAt"
										value={formatDate(product.created_at)}
										readonly
										class="block w-full rounded-md border-0 bg-gray-50 px-2 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm sm:leading-6"
									/>
								</div>
							</div>
							<div>
								<label for="updatedAt" class="block text-sm font-medium leading-6 text-gray-900">
									Updated At
								</label>
								<div class="mt-2">
									<input
										type="text"
										id="updatedAt"
										value={formatDate(product.updated_at)}
										readonly
										class="block w-full rounded-md border-0 bg-gray-50 px-2 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm sm:leading-6"
									/>
								</div>
							</div>
						</div>

						<!-- Name field -->
						<div>
							<label for="name" class="block text-sm font-medium leading-6 text-gray-900">
								Name <span class="text-red-500">*</span>
							</label>
							<div class="mt-2">
								<input
									type="text"
									name="name"
									id="name"
									value={form?.product?.name !== undefined ? form.product.name : product.name}
									required
									class="block w-full rounded-md border-0 px-2 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm sm:leading-6"
								/>
							</div>
						</div>

						<!-- Description field -->
						<div>
							<label for="description" class="block text-sm font-medium leading-6 text-gray-900">
								Description
							</label>
							<div class="mt-2">
								<textarea
									id="description"
									name="description"
									rows="4"
									class="block w-full rounded-md border-0 px-2 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm sm:leading-6"
									>{form?.product?.description !== undefined
										? form.product.description
										: product.description}</textarea
								>
							</div>
						</div>

						<!-- Is Active field -->
						<div class="relative flex items-start">
							<div class="flex h-6 items-center">
								<input
									id="is_active"
									name="is_active"
									type="checkbox"
									value="true"
									checked={form?.product?.is_active !== undefined
										? form.product.is_active
										: product.is_active}
									class="h-4 w-4 rounded border-gray-300 text-indigo-600 focus:ring-indigo-600"
								/>
							</div>
							<div class="ml-3 text-sm leading-6">
								<label for="is_active" class="font-medium text-gray-900">Active</label>
								<p class="text-gray-500">Make this product active and available</p>
							</div>
						</div>

						<!-- Stripe Product ID (readonly if present) -->
						{#if product.stripe_product_id}
							<div>
								<label
									for="stripeProductId"
									class="block text-sm font-medium leading-6 text-gray-900"
								>
									Stripe Product ID
								</label>
								<div class="mt-2">
									<input
										type="text"
										id="stripeProductId"
										value={product.stripe_product_id}
										readonly
										class="block w-full rounded-md border-0 bg-gray-50 px-2 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm sm:leading-6"
									/>
								</div>
							</div>
						{/if}

						<!-- Metadata fields -->
						<div>
							<div class="flex items-center justify-between">
								<label class="block text-sm font-medium leading-6 text-gray-900">
									Metadata (Optional)
								</label>
								<button
									type="button"
									onclick={addMetadataField}
									class="text-sm text-indigo-600 hover:text-indigo-500"
								>
									+ Add field
								</button>
							</div>

							<div class="mt-2 space-y-2">
								{#each metadataFields as field, index}
									<div class="flex space-x-2">
										<div class="w-1/3">
											<input
												type="text"
												name="metadata_key"
												placeholder="Key"
												value={field.key}
												class="block w-full rounded-md border-0 px-2 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm sm:leading-6"
											/>
										</div>
										<div class="flex-1">
											<input
												type="text"
												name="metadata_value"
												placeholder="Value"
												value={stringifyMetadataValue(field.value)}
												class="block w-full rounded-md border-0 px-2 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm sm:leading-6"
											/>
										</div>
										{#if index > 0 || metadataFields.length > 1}
											<button
												type="button"
												onclick={() => removeMetadataField(index)}
												class="inline-flex items-center rounded-md text-sm font-medium text-red-600 hover:text-red-500"
											>
												<svg class="h-5 w-5" viewBox="0 0 20 20" fill="currentColor">
													<path
														fill-rule="evenodd"
														d="M10 18a8 8 0 100-16 8 8 0 000 16zM7 9a1 1 0 000 2h6a1 1 0 100-2H7z"
														clip-rule="evenodd"
													/>
												</svg>
											</button>
										{/if}
									</div>
								{/each}
							</div>
							<p class="mt-1 text-xs text-gray-500">
								For JSON values like numbers or booleans, use valid JSON syntax (e.g., true, 42,
								"string")
							</p>
						</div>
					</div>
				</div>

				<div
					class="flex items-center justify-end gap-x-6 border-t border-gray-900/10 px-4 py-4 sm:px-8"
				>
					<button
						type="button"
						onclick={handleCancel}
						class="text-sm font-semibold leading-6 text-gray-900"
					>
						Cancel
					</button>
					<button
						type="submit"
						disabled={isSubmitting}
						class="rounded-md bg-indigo-600 px-3 py-2 text-sm font-semibold text-white shadow-sm hover:bg-indigo-500 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-indigo-600 disabled:cursor-not-allowed disabled:opacity-50"
					>
						{#if isSubmitting}
							<svg
								class="-ml-1 mr-2 inline-block h-4 w-4 animate-spin text-white"
								xmlns="http://www.w3.org/2000/svg"
								fill="none"
								viewBox="0 0 24 24"
							>
								<circle
									class="opacity-25"
									cx="12"
									cy="12"
									r="10"
									stroke="currentColor"
									stroke-width="4"
								></circle>
								<path
									class="opacity-75"
									fill="currentColor"
									d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
								></path>
							</svg>
							Saving...
						{:else}
							Save Changes
						{/if}
					</button>
				</div>
			</form>
		{/if}
	</div>
</div>
