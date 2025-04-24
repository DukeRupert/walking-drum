<!-- src/routes/admin/prices/+page.svelte -->
<script lang="ts">
    import type { PageProps } from './$types';
    
    let { data }: PageProps = $props();
    
    let { prices, pagination, error } = $derived(data);
    
    // Format amount as currency
    function formatAmount(amount: number, currency: string): string {
      // Convert cents to dollars
      const dollars = amount / 100;
      
      return new Intl.NumberFormat('en-US', {
        style: 'currency',
        currency: currency.toUpperCase()
      }).format(dollars);
    }
    
    // Format date
    function formatDate(dateString: string): string {
      return new Date(dateString).toLocaleString();
    }
    
    // Format billing interval
    function formatInterval(type: string, count: number): string {
      if (type === 'one_time') return 'One-time payment';
      
      const typeDisplay = type === 'day' ? 'day' :
                           type === 'week' ? 'week' :
                           type === 'month' ? 'month' :
                           type === 'year' ? 'year' : type;
      
      return `Every ${count === 1 ? 'one' : count} ${typeDisplay}${count === 1 ? '' : 's'}`;
    }
    
    // Toggle active filter
    function toggleActiveFilter() {
      const url = new URL(window.location.href);
      url.searchParams.set('active', (!pagination.activeOnly).toString());
      url.searchParams.set('offset', '0'); // Reset to first page
      window.location.href = url.toString();
    }
    
    // Toggle include product
    function toggleIncludeProduct() {
      const url = new URL(window.location.href);
      url.searchParams.set('include_product', (!pagination.includeProduct).toString());
      window.location.href = url.toString();
    }
    
    // Pagination functions
    function goToPage(direction: 'prev' | 'next') {
      const url = new URL(window.location.href);
      const newOffset = direction === 'next' 
        ? pagination.offset + pagination.limit
        : Math.max(0, pagination.offset - pagination.limit);
      
      url.searchParams.set('offset', newOffset.toString());
      window.location.href = url.toString();
    }
    
    const currentPage = Math.floor(pagination.offset / pagination.limit) + 1;
    const hasNextPage = prices.length >= pagination.limit;
    const hasPrevPage = pagination.offset > 0;
  </script>
  
  <div class="px-4 sm:px-6 lg:px-8">
    <div class="sm:flex sm:items-center">
      <div class="sm:flex-auto">
        <h1 class="text-base font-semibold text-gray-900">Prices</h1>
        <p class="mt-2 text-sm text-gray-700">
          A list of all your pricing configurations including amounts, billing intervals, and associated products.
        </p>
        
        {#if error}
          <div class="mt-2 rounded-md bg-red-50 p-2 text-sm text-red-700">
            {error}
          </div>
        {/if}
      </div>
      <div class="mt-4 sm:mt-0 sm:ml-16 sm:flex-none">
        <a
          href="/prices/new"
          class="block rounded-md bg-indigo-600 px-3 py-2 text-center text-sm font-semibold text-white shadow-xs hover:bg-indigo-500 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-indigo-600"
        >
          Add price
        </a>
      </div>
    </div>
    
    <div class="mt-4 flex flex-col sm:flex-row gap-4 items-start sm:items-center">
      <label class="flex items-center text-sm text-gray-700">
        <input type="checkbox" 
          checked={pagination.activeOnly} 
          onchange={toggleActiveFilter}
          class="h-4 w-4 rounded border-gray-300 text-indigo-600 focus:ring-indigo-500"
        />
        <span class="ml-2">Show active prices only</span>
      </label>
      
      <label class="flex items-center text-sm text-gray-700">
        <input type="checkbox" 
          checked={pagination.includeProduct} 
          onchange={toggleIncludeProduct}
          class="h-4 w-4 rounded border-gray-300 text-indigo-600 focus:ring-indigo-500"
        />
        <span class="ml-2">Include product details</span>
      </label>
    </div>
    
    <div class="mt-8 flow-root">
      <div class="-mx-4 -my-2 overflow-x-auto sm:-mx-6 lg:-mx-8">
        <div class="inline-block min-w-full py-2 align-middle sm:px-6 lg:px-8">
          <table class="min-w-full divide-y divide-gray-300">
            <thead>
              <tr>
                <th scope="col" class="py-3.5 pr-3 pl-4 text-left text-sm font-semibold text-gray-900 sm:pl-0">Price</th>
                <th scope="col" class="px-3 py-3.5 text-left text-sm font-semibold text-gray-900">Billing</th>
                <th scope="col" class="px-3 py-3.5 text-left text-sm font-semibold text-gray-900">Product</th>
                <th scope="col" class="px-3 py-3.5 text-left text-sm font-semibold text-gray-900">Status</th>
                <th scope="col" class="px-3 py-3.5 text-left text-sm font-semibold text-gray-900">Created</th>
                <th scope="col" class="relative py-3.5 pr-4 pl-3 sm:pr-0">
                  <span class="sr-only">Actions</span>
                </th>
              </tr>
            </thead>
            <tbody class="divide-y divide-gray-200">
              {#if prices.length === 0}
                <tr>
                  <td colspan="6" class="py-4 text-center text-sm text-gray-500">
                    No prices found.
                  </td>
                </tr>
              {/if}
              
              {#each prices as price}
                <tr>
                  <td class="py-4 pr-3 pl-4 text-sm font-medium whitespace-nowrap text-gray-900 sm:pl-0">
                    <div>
                      {formatAmount(price.amount, price.currency)}
                      {#if price.nickname}
                        <span class="ml-1 text-gray-500">({price.nickname})</span>
                      {/if}
                    </div>
                    <div class="text-xs text-gray-500 font-normal">
                      ID: {price.id}
                    </div>
                  </td>
                  <td class="px-3 py-4 text-sm text-gray-500">
                    {formatInterval(price.interval_type, price.interval_count)}
                    {#if price.trial_period_days}
                      <div class="text-xs">
                        {price.trial_period_days} day trial
                      </div>
                    {/if}
                  </td>
                  <td class="px-3 py-4 text-sm text-gray-500">
                    {#if pagination.includeProduct && price.product}
                      <div class="font-medium">{price.product.name}</div>
                      <div class="text-xs truncate max-w-xs">{price.product.description}</div>
                    {:else}
                      <a 
                        href={`/products/${price.product_id}`}
                        class="text-indigo-600 hover:text-indigo-900"
                      >
                        {price.product_id}
                      </a>
                    {/if}
                  </td>
                  <td class="px-3 py-4 text-sm whitespace-nowrap">
                    {#if price.is_active}
                      <span class="inline-flex items-center rounded-md bg-green-50 px-2 py-1 text-xs font-medium text-green-700">
                        Active
                      </span>
                    {:else}
                      <span class="inline-flex items-center rounded-md bg-gray-50 px-2 py-1 text-xs font-medium text-gray-600">
                        Inactive
                      </span>
                    {/if}
                  </td>
                  <td class="px-3 py-4 text-sm whitespace-nowrap text-gray-500">
                    {formatDate(price.created_at)}
                  </td>
                  <td class="relative py-4 pr-4 pl-3 text-right text-sm font-medium whitespace-nowrap sm:pr-0">
                    <a href={`/prices/${price.id}`} class="text-indigo-600 hover:text-indigo-900 mr-4">
                      Edit<span class="sr-only">, {price.id}</span>
                    </a>
                    <button class="text-red-600 hover:text-red-900">
                      Delete<span class="sr-only">, {price.id}</span>
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
          class="relative inline-flex items-center rounded-md border border-gray-300 bg-white px-4 py-2 text-sm font-medium text-gray-700 hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed"
        >
          Previous
        </button>
        <button 
          onclick={() => hasNextPage && goToPage('next')}
          disabled={!hasNextPage}
          class="relative ml-3 inline-flex items-center rounded-md border border-gray-300 bg-white px-4 py-2 text-sm font-medium text-gray-700 hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed"
        >
          Next
        </button>
      </div>
      <div class="hidden sm:flex sm:flex-1 sm:items-center sm:justify-between">
        <div>
          <p class="text-sm text-gray-700">
            Showing <span class="font-medium">{pagination.offset + 1}</span> to <span class="font-medium">{pagination.offset + prices.length}</span> results
          </p>
        </div>
        <div>
          <nav class="isolate inline-flex -space-x-px rounded-md shadow-sm" aria-label="Pagination">
            <button
              onclick={() => hasPrevPage && goToPage('prev')}
              disabled={!hasPrevPage}
              class="relative inline-flex items-center rounded-l-md px-2 py-2 text-gray-400 ring-1 ring-inset ring-gray-300 hover:bg-gray-50 focus:z-20 focus:outline-offset-0 disabled:opacity-50 disabled:cursor-not-allowed"
            >
              <span class="sr-only">Previous</span>
              <svg class="h-5 w-5" viewBox="0 0 20 20" fill="currentColor" aria-hidden="true">
                <path fill-rule="evenodd" d="M12.79 5.23a.75.75 0 01-.02 1.06L8.832 10l3.938 3.71a.75.75 0 11-1.04 1.08l-4.5-4.25a.75.75 0 010-1.08l4.5-4.25a.75.75 0 011.06.02z" clip-rule="evenodd" />
              </svg>
            </button>
            <span class="relative inline-flex items-center px-4 py-2 text-sm font-semibold text-gray-900 ring-1 ring-inset ring-gray-300 focus:outline-offset-0">
              Page {currentPage}
            </span>
            <button
              onclick={() => hasNextPage && goToPage('next')}
              disabled={!hasNextPage}
              class="relative inline-flex items-center rounded-r-md px-2 py-2 text-gray-400 ring-1 ring-inset ring-gray-300 hover:bg-gray-50 focus:z-20 focus:outline-offset-0 disabled:opacity-50 disabled:cursor-not-allowed"
            >
              <span class="sr-only">Next</span>
              <svg class="h-5 w-5" viewBox="0 0 20 20" fill="currentColor" aria-hidden="true">
                <path fill-rule="evenodd" d="M7.21 14.77a.75.75 0 01.02-1.06L11.168 10 7.23 6.29a.75.75 0 111.04-1.08l4.5 4.25a.75.75 0 010 1.08l-4.5 4.25a.75.75 0 01-1.06-.02z" clip-rule="evenodd" />
              </svg>
            </button>
          </nav>
        </div>
      </div>
    </div>
  </div>