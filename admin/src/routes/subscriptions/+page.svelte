<!-- src/routes/subscriptions/+page.svelte -->
<script lang="ts">
    import type { PageProps } from './$types';
    
    let { data }: PageProps = $props();
    
   let { subscriptions, pagination, statusList, error } = $derived(data);
    
    // Format date
    function formatDate(dateString?: string): string {
      if (!dateString) return '—';
      return new Date(dateString).toLocaleString();
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
    
    // Apply user filter
    function applyUserFilter(event: Event) {
      const input = event.target as HTMLInputElement;
      const userId = input.value.trim();
      
      if (event instanceof KeyboardEvent && event.key !== 'Enter') {
        return;
      }
      
      const url = new URL(window.location.href);
      if (userId) {
        url.searchParams.set('user_id', userId);
      } else {
        url.searchParams.delete('user_id');
      }
      
      url.searchParams.set('offset', '0'); // Reset to first page
      window.location.href = url.toString();
    }
    
    // Clear user filter
    function clearUserFilter() {
      const url = new URL(window.location.href);
      url.searchParams.delete('user_id');
      url.searchParams.set('offset', '0'); // Reset to first page
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
    
    // Calculate current active period
    function getActivePeriod(subscription: any): string {
      const start = new Date(subscription.current_period_start);
      const end = new Date(subscription.current_period_end);
      
      const startFormatted = start.toLocaleDateString();
      const endFormatted = end.toLocaleDateString();
      
      return `${startFormatted} to ${endFormatted}`;
    }
    
    // Calculate remaining days in billing period
    function getRemainingDays(endDate: string): number {
      const end = new Date(endDate);
      const now = new Date();
      
      const diffTime = end.getTime() - now.getTime();
      const diffDays = Math.ceil(diffTime / (1000 * 60 * 60 * 24));
      
      return Math.max(0, diffDays);
    }
    
    let currentPage = $derived(Math.floor(pagination.offset / pagination.limit) + 1);
    let hasNextPage = $derived(subscriptions.length >= pagination.limit);
    let hasPrevPage = $derived(pagination.offset > 0);
  </script>
  
  <div class="px-4 sm:px-6 lg:px-8">
    <div class="sm:flex sm:items-center">
      <div class="sm:flex-auto">
        <h1 class="text-base font-semibold text-gray-900">Subscriptions</h1>
        <p class="mt-2 text-sm text-gray-700">
          A list of all subscriptions including their status, billing period, and other details.
        </p>
        
        {#if error}
          <div class="mt-2 rounded-md bg-red-50 p-2 text-sm text-red-700">
            {error}
          </div>
        {/if}
      </div>
      <div class="mt-4 sm:mt-0 sm:ml-16 sm:flex-none">
        <a
          href="/subscriptions/new"
          class="block rounded-md bg-indigo-600 px-3 py-2 text-center text-sm font-semibold text-white shadow-xs hover:bg-indigo-500 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-indigo-600"
        >
          Create subscription
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
      
      <!-- User ID filter -->
      <div class="relative flex-1 max-w-xs">
        <label for="user-filter" class="sr-only">Filter by User ID</label>
        <input
          id="user-filter"
          type="text"
          placeholder="Filter by User ID"
          value={pagination.userId}
          onkeydown={applyUserFilter}
          onblur={applyUserFilter}
          class="block w-full rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500 sm:text-sm py-2 px-3 pr-10"
        />
        {#if pagination.userId}
          <button 
            onclick={clearUserFilter}
            class="absolute inset-y-0 right-0 flex items-center pr-3 text-gray-400 hover:text-gray-500"
          >
            <svg class="h-5 w-5" viewBox="0 0 20 20" fill="currentColor">
              <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z" clip-rule="evenodd" />
            </svg>
          </button>
        {/if}
      </div>
    </div>
    
    <div class="mt-8 flow-root">
      <div class="-mx-4 -my-2 overflow-x-auto sm:-mx-6 lg:-mx-8">
        <div class="inline-block min-w-full py-2 align-middle sm:px-6 lg:px-8">
          <table class="min-w-full divide-y divide-gray-300">
            <thead>
              <tr>
                <th scope="col" class="py-3.5 pl-4 pr-3 text-left text-sm font-semibold text-gray-900 sm:pl-0">ID</th>
                <th scope="col" class="px-3 py-3.5 text-left text-sm font-semibold text-gray-900">Status</th>
                <th scope="col" class="px-3 py-3.5 text-left text-sm font-semibold text-gray-900">Customer</th>
                <th scope="col" class="px-3 py-3.5 text-left text-sm font-semibold text-gray-900">Period</th>
                <th scope="col" class="px-3 py-3.5 text-left text-sm font-semibold text-gray-900">Plan</th>
                <th scope="col" class="px-3 py-3.5 text-left text-sm font-semibold text-gray-900">Created</th>
                <th scope="col" class="relative py-3.5 pl-3 pr-4 sm:pr-0">
                  <span class="sr-only">Actions</span>
                </th>
              </tr>
            </thead>
            <tbody class="divide-y divide-gray-200">
              {#if subscriptions.length === 0}
                <tr>
                  <td colspan="7" class="py-4 text-center text-sm text-gray-500">
                    No subscriptions found.
                  </td>
                </tr>
              {/if}
              
              {#each subscriptions as subscription}
                <tr>
                  <td class="whitespace-nowrap py-4 pl-4 pr-3 text-sm font-medium text-gray-900 sm:pl-0">
                    <div>
                      {subscription.stripe_id}
                    </div>
                    <div class="text-xs text-gray-500">
                      ID: {subscription.id.substring(0, 8)}...
                    </div>
                  </td>
                  <td class="whitespace-nowrap px-3 py-4 text-sm">
                    <span class="inline-flex items-center rounded-md px-2 py-1 text-xs font-medium {getStatusClasses(subscription.status)}">
                      {subscription.status.charAt(0).toUpperCase() + subscription.status.slice(1).replace(/_/g, ' ')}
                    </span>
                    {#if subscription.cancel_at_period_end}
                      <span class="block mt-1 text-xs text-gray-500">
                        Cancels at period end
                      </span>
                    {/if}
                  </td>
                  <td class="whitespace-nowrap px-3 py-4 text-sm text-gray-500">
                    <a 
                      href={`/users/${subscription.user_id}`}
                      class="text-indigo-600 hover:text-indigo-900"
                    >
                      {subscription.user_id.substring(0, 8)}...
                    </a>
                  </td>
                  <td class="whitespace-nowrap px-3 py-4 text-sm text-gray-500">
                    <div>
                      {getActivePeriod(subscription)}
                    </div>
                    {#if subscription.status === 'active' || subscription.status === 'trialing'}
                      <div class="text-xs text-gray-500">
                        {getRemainingDays(subscription.current_period_end)} days remaining
                      </div>
                    {/if}
                    {#if subscription.trial_end}
                      <div class="text-xs text-indigo-600">
                        Trial ends: {formatDate(subscription.trial_end).split(',')[0]}
                      </div>
                    {/if}
                  </td>
                  <td class="whitespace-nowrap px-3 py-4 text-sm text-gray-500">
                    <a 
                      href={`/prices/${subscription.price_id}`}
                      class="text-indigo-600 hover:text-indigo-900"
                    >
                      {subscription.price_id.substring(0, 8)}...
                    </a>
                    {#if subscription.quantity > 1}
                      <div class="text-xs text-gray-500">
                        Qty: {subscription.quantity}
                      </div>
                    {/if}
                  </td>
                  <td class="whitespace-nowrap px-3 py-4 text-sm text-gray-500">
                    {formatDate(subscription.created_at).split(',')[0]}
                  </td>
                  <td class="relative whitespace-nowrap py-4 pl-3 pr-4 text-right text-sm font-medium sm:pr-0">
                    <a href={`/subscriptions/${subscription.id}`} class="text-indigo-600 hover:text-indigo-900">
                      View<span class="sr-only">, {subscription.id}</span>
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
            Showing <span class="font-medium">{pagination.offset + 1}</span> to <span class="font-medium">{pagination.offset + subscriptions.length}</span> results
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