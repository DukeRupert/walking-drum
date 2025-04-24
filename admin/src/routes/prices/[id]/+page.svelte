<!-- src/routes/prices/[id]/+page.svelte -->
<script lang="ts">
    import { goto } from '$app/navigation';
    import type { PageProps } from './$types';
    
    let { data }: PageProps = $props();
    
    let { price, metadataArray } = $derived(data);
    
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
    
    // Navigate back to the prices list
    function goBackToPrices() {
      goto('/prices');
    }
    
    // Navigate to edit price page (to be implemented)
    function goToEditPrice() {
      goto(`/prices/${price.id}/edit`);
    }
    
    // Format JSON string
    function stringifyValue(value: any): string {
      if (typeof value === 'object' && value !== null) {
        return JSON.stringify(value, null, 2);
      }
      return String(value);
    }
  </script>
  
  <div class="px-4 sm:px-6 lg:px-8 py-8 max-w-4xl mx-auto">
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-2xl font-semibold text-gray-900">Price Details</h1>
        <p class="mt-1 text-sm text-gray-500">
          Viewing details for price {price.nickname ? price.nickname : formatAmount(price.amount, price.currency)}
        </p>
      </div>
      <div class="flex gap-3">
        <button 
          type="button"
          onclick={goBackToPrices}
          class="inline-flex items-center rounded-md bg-white px-3 py-2 text-sm font-semibold text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 hover:bg-gray-50"
        >
          Back to Prices
        </button>
        <button
          type="button"
          onclick={goToEditPrice}
          class="inline-flex items-center rounded-md bg-indigo-600 px-3 py-2 text-sm font-semibold text-white shadow-sm hover:bg-indigo-500 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-indigo-600"
        >
          Edit Price
        </button>
      </div>
    </div>
    
    <div class="mt-6 bg-white shadow-sm ring-1 ring-gray-900/5 sm:rounded-xl md:col-span-2">
      <div class="px-4 py-6 sm:p-8">
        <div class="grid grid-cols-1 gap-x-6 gap-y-8 sm:grid-cols-6">
          <!-- Price ID -->
          <div class="sm:col-span-3">
            <h3 class="text-sm font-medium leading-6 text-gray-900">Price ID</h3>
            <div class="mt-2 text-sm text-gray-500">
              {price.id}
            </div>
          </div>
          
          <!-- Stripe Price ID (if present) -->
          {#if price.stripe_price_id}
            <div class="sm:col-span-3">
              <h3 class="text-sm font-medium leading-6 text-gray-900">Stripe Price ID</h3>
              <div class="mt-2 text-sm text-gray-500">
                {price.stripe_price_id}
              </div>
            </div>
          {/if}
          
          <!-- Amount -->
          <div class="sm:col-span-3">
            <h3 class="text-sm font-medium leading-6 text-gray-900">Amount</h3>
            <div class="mt-2 text-sm text-gray-900 font-medium text-lg">
              {formatAmount(price.amount, price.currency)}
              <span class="text-xs font-normal text-gray-500 ml-1">
                ({price.amount} cents {price.currency.toUpperCase()})
              </span>
            </div>
          </div>
          
          <!-- Nickname (if present) -->
          {#if price.nickname}
            <div class="sm:col-span-3">
              <h3 class="text-sm font-medium leading-6 text-gray-900">Nickname</h3>
              <div class="mt-2 text-sm text-gray-900">
                {price.nickname}
              </div>
            </div>
          {/if}
          
          <!-- Billing Interval -->
          <div class="sm:col-span-3">
            <h3 class="text-sm font-medium leading-6 text-gray-900">Billing Interval</h3>
            <div class="mt-2 text-sm text-gray-900">
              {formatInterval(price.interval_type, price.interval_count)}
            </div>
          </div>
          
          <!-- Trial Period (if present) -->
          {#if price.trial_period_days !== undefined && price.trial_period_days !== null}
            <div class="sm:col-span-3">
              <h3 class="text-sm font-medium leading-6 text-gray-900">Trial Period</h3>
              <div class="mt-2 text-sm text-gray-900">
                {price.trial_period_days} {price.trial_period_days === 1 ? 'day' : 'days'}
              </div>
            </div>
          {/if}
          
          <!-- Status -->
          <div class="sm:col-span-3">
            <h3 class="text-sm font-medium leading-6 text-gray-900">Status</h3>
            <div class="mt-2">
              {#if price.is_active}
                <span class="inline-flex items-center rounded-md bg-green-50 px-2 py-1 text-xs font-medium text-green-700">
                  Active
                </span>
              {:else}
                <span class="inline-flex items-center rounded-md bg-gray-50 px-2 py-1 text-xs font-medium text-gray-600">
                  Inactive
                </span>
              {/if}
            </div>
          </div>
          
          <!-- Created/Updated timestamps -->
          <div class="sm:col-span-3">
            <h3 class="text-sm font-medium leading-6 text-gray-900">Created At</h3>
            <div class="mt-2 text-sm text-gray-500">
              {formatDate(price.created_at)}
            </div>
          </div>
          
          <div class="sm:col-span-3">
            <h3 class="text-sm font-medium leading-6 text-gray-900">Updated At</h3>
            <div class="mt-2 text-sm text-gray-500">
              {formatDate(price.updated_at)}
            </div>
          </div>
          
          <!-- Product Information (if included) -->
          <div class="sm:col-span-6">
            <h3 class="text-sm font-medium leading-6 text-gray-900">Product</h3>
            <div class="mt-2">
              {#if price.product}
                <div class="bg-gray-50 border border-gray-200 rounded-md p-4">
                  <div class="flex justify-between items-start">
                    <div>
                      <h4 class="text-sm font-medium text-gray-900">{price.product.name}</h4>
                      <p class="text-sm text-gray-500 mt-1">{price.product.description}</p>
                    </div>
                    <div>
                      {#if price.product.is_active}
                        <span class="inline-flex items-center rounded-md bg-green-50 px-2 py-1 text-xs font-medium text-green-700">
                          Active
                        </span>
                      {:else}
                        <span class="inline-flex items-center rounded-md bg-gray-50 px-2 py-1 text-xs font-medium text-gray-600">
                          Inactive
                        </span>
                      {/if}
                    </div>
                  </div>
                  <div class="mt-3 text-xs text-gray-500">
                    Product ID: {price.product.id}
                  </div>
                  <div class="mt-3">
                    <a 
                      href={`/products/${price.product.id}`}
                      class="text-indigo-600 hover:text-indigo-900 text-sm font-medium"
                    >
                      View Product Details
                    </a>
                  </div>
                </div>
              {:else}
                <div class="text-sm text-gray-500">
                  Product ID: 
                  <a 
                    href={`/products/${price.product_id}`}
                    class="text-indigo-600 hover:text-indigo-900"
                  >
                    {price.product_id}
                  </a>
                </div>
                <div class="mt-2">
                  <a 
                    href={`/prices/${price.id}?include_product=true`}
                    class="text-indigo-600 hover:text-indigo-900 text-sm font-medium"
                  >
                    Load Product Details
                  </a>
                </div>
              {/if}
            </div>
          </div>
          
          <!-- Metadata (if present) -->
          {#if metadataArray.length > 0}
            <div class="sm:col-span-6">
              <h3 class="text-sm font-medium leading-6 text-gray-900">Metadata</h3>
              <div class="mt-2 overflow-hidden">
                <table class="min-w-full divide-y divide-gray-300">
                  <thead>
                    <tr>
                      <th scope="col" class="py-2 pr-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Key</th>
                      <th scope="col" class="py-2 px-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Value</th>
                    </tr>
                  </thead>
                  <tbody class="divide-y divide-gray-200 bg-white">
                    {#each metadataArray as { key, value }}
                      <tr>
                        <td class="py-2 pr-3 text-xs text-gray-900 font-medium">{key}</td>
                        <td class="py-2 px-3 text-xs text-gray-500 font-mono whitespace-pre-wrap">{stringifyValue(value)}</td>
                      </tr>
                    {/each}
                  </tbody>
                </table>
              </div>
            </div>
          {/if}
        </div>
      </div>
    </div>
  </div>