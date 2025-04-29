<!-- src/routes/admin/users/[id]/+page.svelte -->
<script lang="ts">
    import type { PageProps } from './$types';
    
    let { data }: PageProps = $props();
    
    let { user } = $derived(data);
    
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

    // Format metadata object for display
    function formatMetadataObject(metadata?: Record<string, any>): { key: string; value: string }[] {
      if (!metadata) return [];
      return Object.entries(metadata).map(([key, value]) => ({
        key,
        value: typeof value === 'object' ? JSON.stringify(value) : String(value)
      }));
    }
    
    let metadataItems = $derived(formatMetadataObject(user.metadata));
</script>

<div class="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
  <div class="lg:flex lg:items-center lg:justify-between">
    <div class="min-w-0 flex-1">
      <nav class="flex" aria-label="Breadcrumb">
        <ol class="flex items-center space-x-4">
          <li>
            <div>
              <a href="/users" class="text-sm font-medium text-gray-500 hover:text-gray-700">Users</a>
            </div>
          </li>
          <li>
            <div class="flex items-center">
              <svg class="h-5 w-5 flex-shrink-0 text-gray-400" viewBox="0 0 20 20" fill="currentColor" aria-hidden="true">
                <path fill-rule="evenodd" d="M7.21 14.77a.75.75 0 01.02-1.06L11.168 10 7.23 6.29a.75.75 0 111.04-1.08l4.5 4.25a.75.75 0 010 1.08l-4.5 4.25a.75.75 0 01-1.06-.02z" clip-rule="evenodd" />
              </svg>
              <span class="ml-4 text-sm font-medium text-gray-500">User Details</span>
            </div>
          </li>
        </ol>
      </nav>
      <h2 class="mt-2 text-2xl font-bold leading-7 text-gray-900 sm:truncate sm:text-3xl sm:tracking-tight">
        {user.name || 'Unnamed User'}
      </h2>
      <div class="mt-1 flex flex-col sm:mt-0 sm:flex-row sm:flex-wrap sm:space-x-6">
        <div class="mt-2 flex items-center text-sm text-gray-500">
          <svg class="mr-1.5 h-5 w-5 flex-shrink-0 text-gray-400" viewBox="0 0 20 20" fill="currentColor">
            <path d="M2.003 5.884L10 9.882l7.997-3.998A2 2 0 0016 4H4a2 2 0 00-1.997 1.884z" />
            <path d="M18 8.118l-8 4-8-4V14a2 2 0 002 2h12a2 2 0 002-2V8.118z" />
          </svg>
          {user.email}
        </div>
        <div class="mt-2 flex items-center text-sm text-gray-500">
          <svg class="mr-1.5 h-5 w-5 flex-shrink-0 text-gray-400" viewBox="0 0 20 20" fill="currentColor">
            <path fill-rule="evenodd" d="M10 2a8 8 0 100 16 8 8 0 000-16zm0 14a6 6 0 110-12 6 6 0 010 12z" clip-rule="evenodd" />
            <path fill-rule="evenodd" d="M10 5a1 1 0 011 1v3.586l2.707 2.707a1 1 0 01-1.414 1.414l-3-3A1 1 0 019 10V6a1 1 0 011-1z" clip-rule="evenodd" />
          </svg>
          User since {formatDate(user.created_at).split(',')[0]}
        </div>
        <div class="mt-2 flex items-center text-sm">
          <span class="inline-flex items-center rounded-md px-2 py-1 text-xs font-medium {getStatusClasses(user.is_active)}">
            {user.is_active ? 'Active' : 'Inactive'}
          </span>
        </div>
      </div>
    </div>
    <div class="mt-5 flex lg:ml-4 lg:mt-0">
      <span class="ml-3">
        <a
          href={`/users/${user.id}/edit`}
          class="inline-flex items-center rounded-md bg-white px-3 py-2 text-sm font-semibold text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 hover:bg-gray-50"
        >
          <svg class="-ml-0.5 mr-1.5 h-5 w-5 text-gray-400" viewBox="0 0 20 20" fill="currentColor">
            <path d="M13.586 3.586a2 2 0 112.828 2.828l-.793.793-2.828-2.828.793-.793zM11.379 5.793L3 14.172V17h2.828l8.38-8.379-2.83-2.828z" />
          </svg>
          Edit
        </a>
      </span>
      <span class="ml-3">
        <button
          type="button"
          class="inline-flex items-center rounded-md bg-white px-3 py-2 text-sm font-semibold text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 hover:bg-gray-50"
        >
          <svg class="-ml-0.5 mr-1.5 h-5 w-5 text-gray-400" viewBox="0 0 20 20" fill="currentColor">
            <path d="M8 5a1 1 0 100 2h5.586l-1.293 1.293a1 1 0 001.414 1.414l3-3a1 1 0 000-1.414l-3-3a1 1 0 10-1.414 1.414L13.586 5H8zM12 15a1 1 0 100-2H6.414l1.293-1.293a1 1 0 10-1.414-1.414l-3 3a1 1 0 000 1.414l3 3a1 1 0 001.414-1.414L6.414 15H12z" />
          </svg>
          Reset Password
        </button>
      </span>
      {#if user.is_active}
        <span class="ml-3">
          <button
            type="button"
            class="inline-flex items-center rounded-md bg-white px-3 py-2 text-sm font-semibold text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 hover:bg-gray-50"
          >
            <svg class="-ml-0.5 mr-1.5 h-5 w-5 text-gray-400" viewBox="0 0 20 20" fill="currentColor">
              <path fill-rule="evenodd" d="M4.293 4.293a1 1 0 011.414 0L10 8.586l4.293-4.293a1 1 0 111.414 1.414L11.414 10l4.293 4.293a1 1 0 01-1.414 1.414L10 11.414l-4.293 4.293a1 1 0 01-1.414-1.414L8.586 10 4.293 5.707a1 1 0 010-1.414z" clip-rule="evenodd" />
            </svg>
            Deactivate
          </button>
        </span>
      {:else}
        <span class="ml-3">
          <button
            type="button"
            class="inline-flex items-center rounded-md bg-white px-3 py-2 text-sm font-semibold text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 hover:bg-gray-50"
          >
            <svg class="-ml-0.5 mr-1.5 h-5 w-5 text-gray-400" viewBox="0 0 20 20" fill="currentColor">
              <path fill-rule="evenodd" d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z" clip-rule="evenodd" />
            </svg>
            Activate
          </button>
        </span>
      {/if}
    </div>
  </div>

  <div class="mt-8 grid grid-cols-1 gap-6 lg:grid-cols-2">
    <!-- User Basic Info Card -->
    <div class="overflow-hidden rounded-lg bg-white shadow">
      <div class="px-4 py-5 sm:px-6">
        <h3 class="text-lg font-medium leading-6 text-gray-900">User Information</h3>
        <p class="mt-1 max-w-2xl text-sm text-gray-500">Personal details and application settings.</p>
      </div>
      <div class="border-t border-gray-200 px-4 py-5 sm:p-0">
        <dl class="sm:divide-y sm:divide-gray-200">
          <div class="py-4 sm:grid sm:grid-cols-3 sm:gap-4 sm:px-6 sm:py-5">
            <dt class="text-sm font-medium text-gray-500">Full name</dt>
            <dd class="mt-1 text-sm text-gray-900 sm:col-span-2 sm:mt-0">{user.name || '—'}</dd>
          </div>
          <div class="py-4 sm:grid sm:grid-cols-3 sm:gap-4 sm:px-6 sm:py-5">
            <dt class="text-sm font-medium text-gray-500">Email address</dt>
            <dd class="mt-1 text-sm text-gray-900 sm:col-span-2 sm:mt-0">{user.email}</dd>
          </div>
          <div class="py-4 sm:grid sm:grid-cols-3 sm:gap-4 sm:px-6 sm:py-5">
            <dt class="text-sm font-medium text-gray-500">User ID</dt>
            <dd class="mt-1 text-sm text-gray-900 sm:col-span-2 sm:mt-0">{user.id}</dd>
          </div>
          <div class="py-4 sm:grid sm:grid-cols-3 sm:gap-4 sm:px-6 sm:py-5">
            <dt class="text-sm font-medium text-gray-500">Status</dt>
            <dd class="mt-1 text-sm text-gray-900 sm:col-span-2 sm:mt-0">
              <span class="inline-flex items-center rounded-md px-2 py-1 text-xs font-medium {getStatusClasses(user.is_active)}">
                {user.is_active ? 'Active' : 'Inactive'}
              </span>
            </dd>
          </div>
          <div class="py-4 sm:grid sm:grid-cols-3 sm:gap-4 sm:px-6 sm:py-5">
            <dt class="text-sm font-medium text-gray-500">Created at</dt>
            <dd class="mt-1 text-sm text-gray-900 sm:col-span-2 sm:mt-0">{formatDate(user.created_at)}</dd>
          </div>
          <div class="py-4 sm:grid sm:grid-cols-3 sm:gap-4 sm:px-6 sm:py-5">
            <dt class="text-sm font-medium text-gray-500">Last updated</dt>
            <dd class="mt-1 text-sm text-gray-900 sm:col-span-2 sm:mt-0">{formatDate(user.updated_at)}</dd>
          </div>
        </dl>
      </div>
    </div>

    <!-- Payment Information Card -->
    <div class="overflow-hidden rounded-lg bg-white shadow">
      <div class="px-4 py-5 sm:px-6">
        <h3 class="text-lg font-medium leading-6 text-gray-900">Payment Information</h3>
        <p class="mt-1 max-w-2xl text-sm text-gray-500">Billing and subscription details.</p>
      </div>
      <div class="border-t border-gray-200 px-4 py-5 sm:p-0">
        <dl class="sm:divide-y sm:divide-gray-200">
          <div class="py-4 sm:grid sm:grid-cols-3 sm:gap-4 sm:px-6 sm:py-5">
            <dt class="text-sm font-medium text-gray-500">Stripe Customer ID</dt>
            <dd class="mt-1 text-sm text-gray-900 sm:col-span-2 sm:mt-0">{user.stripe_customer_id || '—'}</dd>
          </div>
          <!-- Placeholder for subscription info - you could add actual subscription data here -->
          <div class="py-4 sm:grid sm:grid-cols-3 sm:gap-4 sm:px-6 sm:py-5">
            <dt class="text-sm font-medium text-gray-500">Current Plan</dt>
            <dd class="mt-1 text-sm text-gray-900 sm:col-span-2 sm:mt-0">
              {#if user.stripe_customer_id}
                <a href={`/subscriptions?user_id=${user.id}`} class="text-indigo-600 hover:text-indigo-900">
                  View Subscriptions
                </a>
              {:else}
                No active subscription
              {/if}
            </dd>
          </div>
          <div class="py-4 sm:grid sm:grid-cols-3 sm:gap-4 sm:px-6 sm:py-5">
            <dt class="text-sm font-medium text-gray-500">Payment Method</dt>
            <dd class="mt-1 text-sm text-gray-900 sm:col-span-2 sm:mt-0">
              {#if user.stripe_customer_id}
                <a href={`/payment-methods?user_id=${user.id}`} class="text-indigo-600 hover:text-indigo-900">
                  View Payment Methods
                </a>
              {:else}
                No payment methods
              {/if}
            </dd>
          </div>
          <div class="py-4 sm:grid sm:grid-cols-3 sm:gap-4 sm:px-6 sm:py-5">
            <dt class="text-sm font-medium text-gray-500">Billing History</dt>
            <dd class="mt-1 text-sm text-gray-900 sm:col-span-2 sm:mt-0">
              {#if user.stripe_customer_id}
                <a href={`/invoices?user_id=${user.id}`} class="text-indigo-600 hover:text-indigo-900">
                  View Invoices
                </a>
              {:else}
                No billing history
              {/if}
            </dd>
          </div>
        </dl>
      </div>
    </div>

    <!-- Metadata Card -->
    {#if metadataItems.length > 0}
      <div class="overflow-hidden rounded-lg bg-white shadow col-span-1 lg:col-span-2">
        <div class="px-4 py-5 sm:px-6">
          <h3 class="text-lg font-medium leading-6 text-gray-900">User Metadata</h3>
          <p class="mt-1 max-w-2xl text-sm text-gray-500">Additional properties and customizations.</p>
        </div>
        <div class="border-t border-gray-200 px-4 py-5">
          <div class="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
            {#each metadataItems as item}
              <div class="rounded-md bg-gray-50 p-4">
                <p class="text-sm font-medium text-gray-500">{item.key}</p>
                <p class="mt-1 text-sm text-gray-900 break-words">{item.value}</p>
              </div>
            {/each}
          </div>
        </div>
      </div>
    {/if}

    <!-- Orders Section -->
    <div class="overflow-hidden rounded-lg bg-white shadow col-span-1 lg:col-span-2">
      <div class="px-4 py-5 sm:px-6 flex justify-between items-center">
        <div>
          <h3 class="text-lg font-medium leading-6 text-gray-900">Orders</h3>
          <p class="mt-1 max-w-2xl text-sm text-gray-500">Recent purchase history.</p>
        </div>
        <a 
          href={`/orders?user_id=${user.id}`}
          class="inline-flex items-center rounded-md bg-white px-3 py-2 text-sm font-semibold text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 hover:bg-gray-50"
        >
          View All Orders
        </a>
      </div>
      <div class="border-t border-gray-200">
        <!-- You could add a table of recent orders here or a placeholder -->
        <div class="px-4 py-5 text-center text-sm text-gray-500">
          User order history will be displayed here.
        </div>
      </div>
    </div>
  </div>
</div>