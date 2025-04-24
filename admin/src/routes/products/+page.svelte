<!-- src/routes/admin/products/+page.svelte -->
<script lang="ts">
	import { onMount } from 'svelte';
	import type { PageProps } from './$types';
	import { enhance } from '$app/forms';

	let { data, form }: PageProps = $props();
	let { products, pagination, error } = data;

	// Form state
	let showAddModal = $state(false);
	let metadataFields = $state([{ key: '', value: '' }]);

	function openAddModal() {
		// Reset form state
		metadataFields = [{ key: '', value: '' }];

		// Show modal
		showAddModal = true;
	}

	function closeAddModal() {
		showAddModal = false;
	}

	function addMetadataField() {
		metadataFields = [...metadataFields, { key: '', value: '' }];
	}

	function removeMetadataField(index: number) {
		metadataFields = metadataFields.filter((_, i) => i !== index);
	}

	// Triggered after a successful form submission
	function handleSuccess() {
		// Close the modal after a short delay to show the success message
		setTimeout(() => {
			closeAddModal();
			// Reload the page to show the updated product list
			window.location.reload();
		}, 1500);
	}

	function formatDate(dateString: string) {
		return new Date(dateString).toLocaleString();
	}

	function toggleActiveFilter() {
		const url = new URL(window.location.href);
		url.searchParams.set('active', (!pagination.activeOnly).toString());
		url.searchParams.set('offset', '0'); // Reset to first page
		window.location.href = url.toString();
	}

	function goToPage(direction: 'prev' | 'next') {
		const url = new URL(window.location.href);
		const newOffset =
			direction === 'next'
				? pagination.offset + pagination.limit
				: Math.max(0, pagination.offset - pagination.limit);

		url.searchParams.set('offset', newOffset.toString());
		window.location.href = url.toString();
	}

	const currentPage = Math.floor(pagination.offset / pagination.limit) + 1;
	const hasNextPage = products.length >= pagination.limit;
	const hasPrevPage = pagination.offset > 0;

	// Close modal on escape key
	onMount(() => {
		const handleKeydown = (e: KeyboardEvent) => {
			if (e.key === 'Escape' && showAddModal) {
				closeAddModal();
			}
		};

		window.addEventListener('keydown', handleKeydown);

		return () => {
			window.removeEventListener('keydown', handleKeydown);
		};
	});

	// Check if form submission was successful
	$effect(() => {
		if (form?.success && showAddModal) {
			handleSuccess();
		}
	});
</script>

<div class="px-4 sm:px-6 lg:px-8">
	<div class="sm:flex sm:items-center">
		<div class="sm:flex-auto">
			<h1 class="text-base font-semibold text-gray-900">Products</h1>
			<p class="mt-2 text-sm text-gray-700">
				A list of all the products in your system including their name, description, status, and
				creation date.
			</p>

			{#if error}
				<div class="mt-2 rounded-md bg-red-50 p-2 text-sm text-red-700">
					{error}
				</div>
			{/if}
		</div>
		<div class="mt-4 sm:ml-16 sm:mt-0 sm:flex-none">
			<button
				type="button"
				onclick={openAddModal}
				class="shadow-xs block rounded-md bg-indigo-600 px-3 py-2 text-center text-sm font-semibold text-white hover:bg-indigo-500 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-indigo-600"
			>
				Add product
			</button>
		</div>
	</div>

	<div class="mt-4 flex items-center">
		<label class="flex items-center text-sm text-gray-700">
			<input
				type="checkbox"
				checked={pagination.activeOnly}
				onchange={toggleActiveFilter}
				class="h-4 w-4 rounded border-gray-300 text-indigo-600 focus:ring-indigo-500"
			/>
			<span class="ml-2">Show active products only</span>
		</label>
	</div>

	<div class="mt-8 flow-root">
		<div class="-mx-4 -my-2 overflow-x-auto sm:-mx-6 lg:-mx-8">
			<div class="inline-block min-w-full py-2 align-middle sm:px-6 lg:px-8">
				<table class="min-w-full divide-y divide-gray-300">
					<thead>
						<tr>
							<th
								scope="col"
								class="py-3.5 pl-4 pr-3 text-left text-sm font-semibold text-gray-900 sm:pl-0"
								>Name</th
							>
							<th scope="col" class="px-3 py-3.5 text-left text-sm font-semibold text-gray-900"
								>Description</th
							>
							<th scope="col" class="px-3 py-3.5 text-left text-sm font-semibold text-gray-900"
								>Status</th
							>
							<th scope="col" class="px-3 py-3.5 text-left text-sm font-semibold text-gray-900"
								>Created</th
							>
							<th scope="col" class="px-3 py-3.5 text-left text-sm font-semibold text-gray-900"
								>Updated</th
							>
							<th scope="col" class="relative py-3.5 pl-3 pr-4 sm:pr-0">
								<span class="sr-only">Actions</span>
							</th>
						</tr>
					</thead>
					<tbody class="divide-y divide-gray-200">
						{#if products.length === 0}
							<tr>
								<td colspan="6" class="py-4 text-center text-sm text-gray-500">
									No products found.
								</td>
							</tr>
						{/if}

						{#each products as product}
							<tr>
								<td
									class="whitespace-nowrap py-4 pl-4 pr-3 text-sm font-medium text-gray-900 sm:pl-0"
								>
									{product.name}
								</td>
								<td class="max-w-xs truncate px-3 py-4 text-sm text-gray-500">
									{product.description}
								</td>
								<td class="whitespace-nowrap px-3 py-4 text-sm">
									{#if product.is_active}
										<span
											class="inline-flex items-center rounded-md bg-green-50 px-2 py-1 text-xs font-medium text-green-700"
										>
											Active
										</span>
									{:else}
										<span
											class="inline-flex items-center rounded-md bg-gray-50 px-2 py-1 text-xs font-medium text-gray-600"
										>
											Inactive
										</span>
									{/if}
								</td>
								<td class="whitespace-nowrap px-3 py-4 text-sm text-gray-500">
									{formatDate(product.created_at)}
								</td>
								<td class="whitespace-nowrap px-3 py-4 text-sm text-gray-500">
									{formatDate(product.updated_at)}
								</td>
								<td
									class="relative whitespace-nowrap py-4 pl-3 pr-4 text-right text-sm font-medium sm:pr-0"
								>
									<a
										href="/admin/products/{product.id}"
										class="mr-4 text-indigo-600 hover:text-indigo-900"
									>
										Edit<span class="sr-only">, {product.name}</span>
									</a>
									<button class="text-red-600 hover:text-red-900">
										Delete<span class="sr-only">, {product.name}</span>
									</button>
								</td>
							</tr>
						{/each}
					</tbody>
				</table>
			</div>
		</div>
	</div>

	<!-- Pagination controls -->
	<div class="mt-5 flex items-center justify-between">
		<div class="flex flex-1 justify-between sm:hidden">
			<button
				onclick={() => hasPrevPage && goToPage('prev')}
				disabled={!hasPrevPage}
				class="relative inline-flex items-center rounded-md border border-gray-300 bg-white px-4 py-2 text-sm font-medium text-gray-700 hover:bg-gray-50 disabled:cursor-not-allowed disabled:opacity-50"
			>
				Previous
			</button>
			<button
				onclick={() => hasNextPage && goToPage('next')}
				disabled={!hasNextPage}
				class="relative ml-3 inline-flex items-center rounded-md border border-gray-300 bg-white px-4 py-2 text-sm font-medium text-gray-700 hover:bg-gray-50 disabled:cursor-not-allowed disabled:opacity-50"
			>
				Next
			</button>
		</div>
		<div class="hidden sm:flex sm:flex-1 sm:items-center sm:justify-between">
			<div>
				<p class="text-sm text-gray-700">
					Showing <span class="font-medium">{pagination.offset + 1}</span> to
					<span class="font-medium">{pagination.offset + products.length}</span> results
				</p>
			</div>
			<div>
				<nav class="isolate inline-flex -space-x-px rounded-md shadow-sm" aria-label="Pagination">
					<button
						onclick={() => hasPrevPage && goToPage('prev')}
						disabled={!hasPrevPage}
						class="relative inline-flex items-center rounded-l-md px-2 py-2 text-gray-400 ring-1 ring-inset ring-gray-300 hover:bg-gray-50 focus:z-20 focus:outline-offset-0 disabled:cursor-not-allowed disabled:opacity-50"
					>
						<span class="sr-only">Previous</span>
						<svg class="h-5 w-5" viewBox="0 0 20 20" fill="currentColor" aria-hidden="true">
							<path
								fill-rule="evenodd"
								d="M12.79 5.23a.75.75 0 01-.02 1.06L8.832 10l3.938 3.71a.75.75 0 11-1.04 1.08l-4.5-4.25a.75.75 0 010-1.08l4.5-4.25a.75.75 0 011.06.02z"
								clip-rule="evenodd"
							/>
						</svg>
					</button>
					<span
						class="relative inline-flex items-center px-4 py-2 text-sm font-semibold text-gray-900 ring-1 ring-inset ring-gray-300 focus:outline-offset-0"
					>
						Page {currentPage}
					</span>
					<button
						onclick={() => hasNextPage && goToPage('next')}
						disabled={!hasNextPage}
						class="relative inline-flex items-center rounded-r-md px-2 py-2 text-gray-400 ring-1 ring-inset ring-gray-300 hover:bg-gray-50 focus:z-20 focus:outline-offset-0 disabled:cursor-not-allowed disabled:opacity-50"
					>
						<span class="sr-only">Next</span>
						<svg class="h-5 w-5" viewBox="0 0 20 20" fill="currentColor" aria-hidden="true">
							<path
								fill-rule="evenodd"
								d="M7.21 14.77a.75.75 0 01.02-1.06L11.168 10 7.23 6.29a.75.75 0 111.04-1.08l4.5 4.25a.75.75 0 010 1.08l-4.5 4.25a.75.75 0 01-1.06-.02z"
								clip-rule="evenodd"
							/>
						</svg>
					</button>
				</nav>
			</div>
		</div>
	</div>
</div>


<!-- Add Product Modal -->
{#if showAddModal}
  <div class="fixed inset-0 bg-gray-500 bg-opacity-75 transition-opacity z-10"></div>
  
  <div class="fixed inset-0 z-10 overflow-y-auto">
    <div class="flex min-h-full items-end justify-center p-4 text-center sm:items-center sm:p-0">
      <div class="relative transform overflow-hidden rounded-lg bg-white px-4 pb-4 pt-5 text-left shadow-xl transition-all sm:my-8 sm:w-full sm:max-w-lg sm:p-6">
        <div class="absolute right-0 top-0 hidden pr-4 pt-4 sm:block">
          <button 
            type="button" 
            onclick={closeAddModal}
            class="rounded-md bg-white text-gray-400 hover:text-gray-500 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:ring-offset-2"
          >
            <span class="sr-only">Close</span>
            <svg class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" aria-hidden="true">
              <path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>
        
        <div class="sm:flex sm:items-start">
          <div class="mt-3 text-center sm:ml-4 sm:mt-0 sm:text-left w-full">
            <h3 class="text-base font-semibold leading-6 text-gray-900">Add New Product</h3>
            
            {#if form?.success}
              <div class="mt-2 rounded-md bg-green-50 p-4">
                <div class="flex">
                  <div class="flex-shrink-0">
                    <svg class="h-5 w-5 text-green-400" viewBox="0 0 20 20" fill="currentColor" aria-hidden="true">
                      <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.857-9.565a.75.75 0 00-1.214-.882l-3.483 4.79-1.88-1.88a.75.75 0 10-1.06 1.061l2.5 2.5a.75.75 0 001.137-.089l4-5.5z" clip-rule="evenodd" />
                    </svg>
                  </div>
                  <div class="ml-3">
                    <p class="text-sm font-medium text-green-800">Product created successfully!</p>
                  </div>
                </div>
              </div>
            {:else}
              <div class="mt-2">
                <form method="POST" action="?/createProduct" use:enhance>
                  {#if form?.error}
                    <div class="rounded-md bg-red-50 p-2 text-sm text-red-700 mb-4">
                      {form.error}
                    </div>
                  {/if}
                  
                  <!-- Name field -->
                  <div class="mb-4">
                    <label for="name" class="block text-sm font-medium leading-6 text-gray-900">
                      Name <span class="text-red-500">*</span>
                    </label>
                    <div class="mt-2">
                      <input
                        type="text"
                        name="name"
                        id="name"
                        value={form?.values?.name || ''}
                        required
                        class="block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm sm:leading-6"
                      />
                    </div>
                  </div>
                  
                  <!-- Description field -->
                  <div class="mb-4">
                    <label for="description" class="block text-sm font-medium leading-6 text-gray-900">
                      Description
                    </label>
                    <div class="mt-2">
                      <textarea
                        id="description"
                        name="description"
                        rows="3"
                        value={form?.values?.description || ''}
                        class="block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm sm:leading-6"
                      ></textarea>
                    </div>
                  </div>
                  
                  <!-- Is Active field -->
                  <div class="relative flex items-start mb-4">
                    <div class="flex h-6 items-center">
                      <input
                        id="is_active"
                        name="is_active"
                        type="checkbox"
                        value="true"
                        checked={form?.values?.isActive !== false}
                        class="h-4 w-4 rounded border-gray-300 text-indigo-600 focus:ring-indigo-600"
                      />
                    </div>
                    <div class="ml-3 text-sm leading-6">
                      <label for="is_active" class="font-medium text-gray-900">Active</label>
                      <p class="text-gray-500">Make this product active and available immediately</p>
                    </div>
                  </div>
                  
                  <!-- Metadata fields -->
                  <div class="mb-4">
                    <div class="flex justify-between items-center">
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
                              bind:value={field.key}
                              class="block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm sm:leading-6"
                            />
                          </div>
                          <div class="flex-1">
                            <input
                              type="text"
                              name="metadata_value"
                              placeholder="Value"
                              bind:value={field.value}
                              class="block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm sm:leading-6"
                            />
                          </div>
                          {#if index > 0 || metadataFields.length > 1}
                            <button 
                              type="button" 
                              onclick={() => removeMetadataField(index)}
                              class="inline-flex items-center rounded-md text-sm font-medium text-red-600 hover:text-red-500"
                            >
                              <svg class="h-5 w-5" viewBox="0 0 20 20" fill="currentColor">
                                <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM7 9a1 1 0 000 2h6a1 1 0 100-2H7z" clip-rule="evenodd" />
                              </svg>
                            </button>
                          {/if}
                        </div>
                      {/each}
                    </div>
                    <p class="mt-1 text-xs text-gray-500">
                      For JSON values like numbers or booleans, use valid JSON syntax (e.g., true, 42, "string")
                    </p>
                  </div>
                  
                  <div class="mt-5 sm:mt-4 sm:flex sm:flex-row-reverse">
                    <button
                      type="submit"
                      class="inline-flex w-full justify-center rounded-md bg-indigo-600 px-3 py-2 text-sm font-semibold text-white shadow-sm hover:bg-indigo-500 sm:ml-3 sm:w-auto"
                    >
                      Create Product
                    </button>
                    <button
                      type="button"
                      onclick={closeAddModal}
                      class="mt-3 inline-flex w-full justify-center rounded-md bg-white px-3 py-2 text-sm font-semibold text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 hover:bg-gray-50 sm:mt-0 sm:w-auto"
                    >
                      Cancel
                    </button>
                  </div>
                </form>
              </div>
            {/if}
          </div>
        </div>
      </div>
    </div>
  </div>
{/if}