<!-- src/routes/invoices/+page.svelte -->
<script lang="ts">
    import type { PageProps } from './$types';
    
    let { data }: PageProps = $props();
    
    let { invoices, pagination, statusList, error } = $derived(data);
    
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
        case 'paid':
          return 'bg-green-50 text-green-700';
        case 'open':
          return 'bg-blue-50 text-blue-700';
        case 'draft':
          return 'bg-gray-50 text-gray-700';
        case 'uncollectible':
          return 'bg-red-50 text-red-700';
        case 'void':
          return 'bg-purple-50 text-purple-700';
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
    
    // Toggle include relations
    function toggleIncludeRelations() {
      const url = new URL(window.location.href);
      url.searchParams.set('include_relations', (!pagination.includeRelations).toString());
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
    const hasNextPage = false
    const hasPrevPage = pagination.offset > 0;
  </script>
  
  <div class="px-4 sm:px-6 lg:px-8">
    <div class="sm:flex sm:items-center">
      <div class="sm:flex-auto">
        <h1 class="text-base font-semibold text-gray-900">Invoices</h1>
        <p class="mt-2 text-sm text-gray-700">
          A list of all invoices including their status, amount, and related information.
        </p>
        
        {#if error}
          <div class="mt-2 rounded-md bg-red-50 p-2 text-sm text-red-700">
            {error}
          </div>
        {/if}
      </div>
      <div class="mt-4 sm:mt-0 sm:ml-16 sm:flex-none">
        <a
          href="/invoices/new"
          class="block rounded-md bg-indigo-600 px-3 py-2 text-center text-sm font-semibold text-white shadow-xs hover:bg-indigo-500 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-indigo-600"
        >
          Create invoice
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
      
      <!-- Include relations checkbox -->
      <label class="flex items-center text-sm text-gray-700">
        <input type="checkbox" 
          checked={pagination.includeRelations} 
          onchange={toggleIncludeRelations}
          class="h-4 w-4 rounded border-gray-300 text-indigo-600 focus:ring-indigo-500"
        />
        <span class="ml-2">Include user and subscription details</span>
      </label>
    </div>
    
    <div class="mt-8 flow-root">
      <div class="-mx-4 -my-2 overflow-x-auto sm:-mx-6 lg:-mx-8">
        <div class="inline-block min-w-full py-2 align-middle sm:px-6 lg:px-8">
          <table class="min-w-full divide-y divide-gray-300">
            <thead>
              <tr>
                <th scope="col" class="py-3.5 pl-4 pr-3 text-left text-sm font-semibold text-gray-900 sm:pl-0">Invoice</th>
                <th scope="col" class="px-3 py-3.5 text-left text-sm font-semibold text-gray-900">Status</th>
                <th scope="col" class="px-3 py-3.5 text-left text-sm font-semibold text-gray-900">Amount</th>
                <th scope="col" class="px-3 py-3.5 text-left text-sm font-semibold text-gray-900">User</th>
                <th scope="col" class="px-3 py-3.5 text-left text-sm font-semibold text-gray-900">Period</th>
                <th scope="col" class="px-3 py-3.5 text-left text-sm font-semibold text-gray-900">Created</th>
                <th scope="col" class="relative py-3.5 pl-3 pr-4 sm:pr-0">
                  <span class="sr-only">Actions</span>
                </th>
              </tr>
            </thead>
            <tbody class="divide-y divide-gray-200">
              {#if invoices.length === 0}
                <tr>
                  <td colspan="7" class="py-4 text-center text-sm text-gray-500">
                    No invoices found.
                  </td>
                </tr>
              {/if}
              
              {#each invoices as invoice}
                <tr>
                  <td class="whitespace-nowrap py-4 pl-4 pr-3 text-sm font-medium text-gray-900 sm:pl-0">
                    <div>
                      {invoice.stripe_invoice_id}
                    </div>
                    <div class="text-xs text-gray-500">
                      ID: {invoice.id}
                    </div>
                  </td>
                  <td class="whitespace-nowrap px-3 py-4 text-sm">
                    <span class="inline-flex items-center rounded-md px-2 py-1 text-xs font-medium {getStatusClasses(invoice.status)}">
                      {invoice.status.charAt(0).toUpperCase() + invoice.status.slice(1)}
                    </span>
                  </td>
                  <td class="whitespace-nowrap px-3 py-4 text-sm text-gray-900">
                    <div>
                      {formatAmount(invoice.amount_due, invoice.currency)}
                    </div>
                    {#if invoice.amount_paid > 0 && invoice.amount_paid !== invoice.amount_due}
                      <div class="text-xs text-gray-500">
                        Paid: {formatAmount(invoice.amount_paid, invoice.currency)}
                      </div>
                    {/if}
                  </td>
                  <td class="whitespace-nowrap px-3 py-4 text-sm text-gray-500">
                    {#if pagination.includeRelations && invoice.user}
                      <div class="font-medium text-gray-900">{invoice.user.name || invoice.user.email}</div>
                      {#if invoice.user.name}
                        <div class="text-xs">{invoice.user.email}</div>
                      {/if}
                    {:else}
                      <a 
                        href={`/users/${invoice.user_id}`}
                        class="text-indigo-600 hover:text-indigo-900"
                      >
                        {invoice.user_id}
                      </a>
                    {/if}
                  </td>
                  <td class="whitespace-nowrap px-3 py-4 text-sm text-gray-500">
                    {#if invoice.period_start && invoice.period_end}
                      <div>
                        {formatDate(invoice.period_start).split(',')[0]}
                      </div>
                      <div>
                        to {formatDate(invoice.period_end).split(',')[0]}
                      </div>
                    {:else}
                      <span class="text-gray-400">—</span>
                    {/if}
                  </td>
                  <td class="whitespace-nowrap px-3 py-4 text-sm text-gray-500">
                    {formatDate(invoice.created_at)}
                  </td>
                  <td class="relative whitespace-nowrap py-4 pl-3 pr-4 text-right text-sm font-medium sm:pr-0">
                    <a href={`/invoices/${invoice.id}`} class="text-indigo-600 hover:text-indigo-900 mr-4">
                      View<span class="sr-only">, {invoice.id}</span>
                    </a>
                    
                    {#if invoice.invoice_pdf}
                      <a href={invoice.invoice_pdf} target="_blank" rel="noopener noreferrer" class="text-indigo-600 hover:text-indigo-900">
                        PDF<span class="sr-only">, {invoice.id}</span>
                      </a>
                    {/if}
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
            Showing <span class="font-medium">{pagination.offset + 1}</span> to <span class="font-medium">{pagination.offset + invoices.length}</span> results
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