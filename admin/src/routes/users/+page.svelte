<!-- src/routes/admin/users/+page.svelte -->
<script lang="ts">
    import type { PageProps } from './$types';
    
    let { data }: PageProps = $props();
    
    let { users, pagination, error } = $derived(data);
    
    // Format date
    function formatDate(dateString?: string): string {
      if (!dateString) return '—';
      return new Date(dateString).toLocaleString();
    }
    
    // Get status badge classes based on active status
    function getStatusClasses(isActive: boolean): string {
      return isActive 
        ? 'bg-green-50 text-green-700' 
        : 'bg-gray-50 text-gray-700';
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
    let hasNextPage = $derived(users.length >= pagination.limit);
    let hasPrevPage = $derived(pagination.offset > 0);
    
    // Format metadata as string
    function formatMetadata(metadata?: Record<string, any>): string {
      if (!metadata || Object.keys(metadata).length === 0) return '—';
      
      // Return first 2 key-value pairs if there are many
      const entries = Object.entries(metadata);
      if (entries.length > 2) {
        return entries.slice(0, 2).map(([key, value]) => `${key}: ${value}`).join(', ') + '...';
      }
      
      return entries.map(([key, value]) => `${key}: ${value}`).join(', ');
    }
</script>

<div class="px-4 sm:px-6 lg:px-8">
  <div class="sm:flex sm:items-center">
    <div class="sm:flex-auto">
      <h1 class="text-base font-semibold text-gray-900">Users</h1>
      <p class="mt-2 text-sm text-gray-700">
        A list of all users in your coffee subscription service including their details and status.
      </p>
      
      {#if error}
        <div class="mt-2 rounded-md bg-red-50 p-2 text-sm text-red-700">
          {error}
        </div>
      {/if}
    </div>
    <div class="mt-4 sm:mt-0 sm:ml-16 sm:flex-none">
      <a
        href="/users/new"
        class="block rounded-md bg-indigo-600 px-3 py-2 text-center text-sm font-semibold text-white shadow-xs hover:bg-indigo-500 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-indigo-600"
      >
        Add user
      </a>
    </div>
  </div>
  
  <div class="mt-8 flow-root">
    <div class="-mx-4 -my-2 overflow-x-auto sm:-mx-6 lg:-mx-8">
      <div class="inline-block min-w-full py-2 align-middle sm:px-6 lg:px-8">
        <table class="min-w-full divide-y divide-gray-300">
          <thead>
            <tr>
              <th scope="col" class="py-3.5 pl-4 pr-3 text-left text-sm font-semibold text-gray-900 sm:pl-0">User ID</th>
              <th scope="col" class="px-3 py-3.5 text-left text-sm font-semibold text-gray-900">Name</th>
              <th scope="col" class="px-3 py-3.5 text-left text-sm font-semibold text-gray-900">Email</th>
              <th scope="col" class="px-3 py-3.5 text-left text-sm font-semibold text-gray-900">Status</th>
              <th scope="col" class="px-3 py-3.5 text-left text-sm font-semibold text-gray-900">Customer ID</th>
              <th scope="col" class="px-3 py-3.5 text-left text-sm font-semibold text-gray-900">Metadata</th>
              <th scope="col" class="px-3 py-3.5 text-left text-sm font-semibold text-gray-900">Created</th>
              <th scope="col" class="relative py-3.5 pl-3 pr-4 sm:pr-0">
                <span class="sr-only">Actions</span>
              </th>
            </tr>
          </thead>
          <tbody class="divide-y divide-gray-200">
            {#if users.length === 0}
              <tr>
                <td colspan="8" class="py-4 text-center text-sm text-gray-500">
                  No users found.
                </td>
              </tr>
            {/if}
            
            {#each users as user}
              <tr>
                <td class="whitespace-nowrap py-4 pl-4 pr-3 text-sm font-medium text-gray-900 sm:pl-0">
                  {user.id.substring(0, 8)}...
                </td>
                <td class="whitespace-nowrap px-3 py-4 text-sm text-gray-900">
                  {user.name || '—'}
                </td>
                <td class="whitespace-nowrap px-3 py-4 text-sm text-gray-500">
                  {user.email}
                </td>
                <td class="whitespace-nowrap px-3 py-4 text-sm">
                  <span class="inline-flex items-center rounded-md px-2 py-1 text-xs font-medium {getStatusClasses(user.is_active)}">
                    {user.is_active ? 'Active' : 'Inactive'}
                  </span>
                </td>
                <td class="whitespace-nowrap px-3 py-4 text-sm text-gray-500">
                  {user.stripe_customer_id ? user.stripe_customer_id.substring(0, 8) + '...' : '—'}
                </td>
                <td class="whitespace-nowrap px-3 py-4 text-sm text-gray-500">
                  {formatMetadata(user.metadata)}
                </td>
                <td class="whitespace-nowrap px-3 py-4 text-sm text-gray-500">
                  {formatDate(user.created_at)}
                </td>
                <td class="relative whitespace-nowrap py-4 pl-3 pr-4 text-right text-sm font-medium sm:pr-0">
                  <a href={`/users/${user.id}`} class="text-indigo-600 hover:text-indigo-900">
                    View<span class="sr-only">, {user.id}</span>
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
          Showing <span class="font-medium">{pagination.offset + 1}</span> to <span class="font-medium">{pagination.offset + users.length}</span> results
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