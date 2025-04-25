<!-- src/routes/admin/orders/+page.svelte -->
<script lang="ts">
    import type { PageProps } from './$types';
    
    let { data }: PageProps = $props();
    
    let {orders, pagination, statusList, error } = $derived(data);
    
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
    function formatDate(dateString?: string): string {
      if (!dateString) return '—';
      return new Date(dateString).toLocaleString();
    }
    
    // Get status badge classes
    function getStatusClasses(status: string): string {
      switch (status.toLowerCase()) {
        case 'completed':
          return 'bg-green-50 text-green-700';
        case 'processing':
          return 'bg-blue-50 text-blue-700';
        case 'pending':
          return 'bg-yellow-50 text-yellow-700';
        case 'cancelled':
          return 'bg-gray-50 text-gray-700';
        case 'failed':
          return 'bg-red-50 text-red-700';
        default:
          return 'bg-gray-50 text-gray-700';
      }
    }
    
    // Apply status filter
    function applyStatusFilter(event: Event) {
      const select = event.target as HTMLSelectElement;
      const status = select.value;
      
      const url = new URL(window.location.href);
      if (status) {
        url.searchParams.set('status', status);
      } else {
        url.searchParams.delete('status');
      }
      
      url.searchParams.set('offset', '0'); // Reset to first page
      window.location.href = url.toString();
    }
    
    // Toggle include items
    function toggleIncludeItems() {
      const url = new URL(window.location.href);
      url.searchParams.set('include_items', (!pagination.includeItems).toString());
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
    
    let currentPage = $derived(Math.floor(pagination.offset / pagination.limit) + 1);
    let hasNextPage = $derived(orders.length >= pagination.limit);
    let hasPrevPage = $derived(pagination.offset > 0);
    
    // Get truncated address display
    function formatAddress(address?: {
      line1: string;
      line2?: string;
      city: string;
      state: string;
      postal_code: string;
      country: string;
    }): string {
      if (!address) return '—';
      return `${address.city}, ${address.state}, ${address.country}`;
    }
    
    // Count total items in an order
    function countItems(items?: any[]): number {
      if (!items) return 0;
      return items.reduce((sum, item) => sum + (item.quantity || 1), 0);
    }
  </script>
  
  <div class="px-4 sm:px-6 lg:px-8">
    <div class="sm:flex sm:items-center">
      <div class="sm:flex-auto">
        <h1 class="text-base font-semibold text-gray-900">Orders</h1>
        <p class="mt-2 text-sm text-gray-700">
          A list of all orders including their status, amount, and other details.
        </p>
        
        {#if error}
          <div class="mt-2 rounded-md bg-red-50 p-2 text-sm text-red-700">
            {error}
          </div>
        {/if}
      </div>
      <div class="mt-4 sm:mt-0 sm:ml-16 sm:flex-none">
        <a
          href="/orders/new"
          class="block rounded-md bg-indigo-600 px-3 py-2 text-center text-sm font-semibold text-white shadow-xs hover:bg-indigo-500 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-indigo-600"
        >
          Create order
        </a>
      </div>
    </div>
    
    <div class="mt-4 flex flex-col sm:flex-row gap-4 items-start sm:items-center">
      <!-- Status filter -->
      <div>
        <label for="status-filter" class="sr-only">Filter by status</label>
        <select
          id="status-filter"
          onchange={applyStatusFilter}
          class="block w-full rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500 sm:text-sm py-2 px-3"
        >
          <option value="">All statuses</option>
          {#each statusList as status}
            <option 
              value={status.value} 
              selected={pagination.status === status.value}
            >
              {status.label}
            </option>
          {/each}
        </select>
      </div>
      
      <!-- Include items checkbox -->
      <label class="flex items-center text-sm text-gray-700">
        <input type="checkbox" 
          checked={pagination.includeItems} 
          onchange={toggleIncludeItems}
          class="h-4 w-4 rounded border-gray-300 text-indigo-600 focus:ring-indigo-500"
        />
        <span class="ml-2">Include order items</span>
      </label>
    </div>
    
    <div class="mt-8 flow-root">
      <div class="-mx-4 -my-2 overflow-x-auto sm:-mx-6 lg:-mx-8">
        <div class="inline-block min-w-full py-2 align-middle sm:px-6 lg:px-8">
          <table class="min-w-full divide-y divide-gray-300">
            <thead>
              <tr>
                <th scope="col" class="py-3.5 pl-4 pr-3 text-left text-sm font-semibold text-gray-900 sm:pl-0">Order ID</th>
                <th scope="col" class="px-3 py-3.5 text-left text-sm font-semibold text-gray-900">Status</th>
                <th scope="col" class="px-3 py-3.5 text-left text-sm font-semibold text-gray-900">Total</th>
                <th scope="col" class="px-3 py-3.5 text-left text-sm font-semibold text-gray-900">Customer</th>
                <th scope="col" class="px-3 py-3.5 text-left text-sm font-semibold text-gray-900">Shipping</th>
                <th scope="col" class="px-3 py-3.5 text-left text-sm font-semibold text-gray-900">Date</th>
                <th scope="col" class="relative py-3.5 pl-3 pr-4 sm:pr-0">
                  <span class="sr-only">Actions</span>
                </th>
              </tr>
            </thead>
            <tbody class="divide-y divide-gray-200">
              {#if orders.length === 0}
                <tr>
                  <td colspan="7" class="py-4 text-center text-sm text-gray-500">
                    No orders found.
                  </td>
                </tr>
              {/if}
              
              {#each orders as order}
                <tr>
                  <td class="whitespace-nowrap py-4 pl-4 pr-3 text-sm font-medium text-gray-900 sm:pl-0">
                    {order.id.substring(0, 8)}...
                  </td>
                  <td class="whitespace-nowrap px-3 py-4 text-sm">
                    <span class="inline-flex items-center rounded-md px-2 py-1 text-xs font-medium {getStatusClasses(order.status)}">
                      {order.status.charAt(0).toUpperCase() + order.status.slice(1)}
                    </span>
                  </td>
                  <td class="whitespace-nowrap px-3 py-4 text-sm text-gray-900">
                    <div>
                      {formatAmount(order.total_amount, order.currency)}
                    </div>
                    {#if order.items}
                      <div class="text-xs text-gray-500">
                        {countItems(order.items)} item{countItems(order.items) !== 1 ? 's' : ''}
                      </div>
                    {/if}
                  </td>
                  <td class="whitespace-nowrap px-3 py-4 text-sm text-gray-500">
                    {#if order.user}
                      <div class="font-medium text-gray-900">{order.user.name || 'N/A'}</div>
                      <div class="text-xs">{order.user.email || 'No email'}</div>
                    {:else if order.user_id}
                      <a 
                        href={`/users/${order.user_id}`}
                        class="text-indigo-600 hover:text-indigo-900"
                      >
                        {order.user_id.substring(0, 8)}...
                      </a>
                    {:else}
                      <span class="text-gray-400">Guest order</span>
                    {/if}
                  </td>
                  <td class="whitespace-nowrap px-3 py-4 text-sm text-gray-500">
                    {formatAddress(order.shipping_address)}
                  </td>
                  <td class="whitespace-nowrap px-3 py-4 text-sm text-gray-500">
                    {formatDate(order.created_at)}
                  </td>
                  <td class="relative whitespace-nowrap py-4 pl-3 pr-4 text-right text-sm font-medium sm:pr-0">
                    <a href={`/orders/${order.id}`} class="text-indigo-600 hover:text-indigo-900">
                      View<span class="sr-only">, {order.id}</span>
                    </a>
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
            Showing <span class="font-medium">{pagination.offset + 1}</span> to <span class="font-medium">{pagination.offset + orders.length}</span> results
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