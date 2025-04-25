<!-- src/routes/prices/new/+page.svelte -->
<script lang="ts">
    import { enhance } from '$app/forms';
    import { goto } from '$app/navigation';
    import type { PageProps } from './$types';
    
    let { data, form }: PageProps = $props();
    
    let { products, defaultValues, currencies, intervalTypes, error } = $derived(data);
    
    let isSubmitting = $state(false);
    let metadataFields = $state([{ key: '', value: '' }]);
    
    // Handle adding new metadata field
    function addMetadataField() {
      metadataFields = [...metadataFields, { key: '', value: '' }];
    }
    
    // Handle removing metadata field
    function removeMetadataField(index: number) {
      metadataFields = metadataFields.filter((_, i) => i !== index);
    }
    
    // Format amount for display
    function formatAmountDisplay(amount: number, currency: string) {
      // Convert cents to dollars
      const dollars = amount / 100;
      
      return new Intl.NumberFormat('en-US', {
        style: 'currency',
        currency: currency.toUpperCase()
      }).format(dollars);
    }
    
    // Handle cancel button click
    function handleCancel() {
      goto('/prices');
    }
    
    // Selected values for reactive display
    let selectedAmount = $derived(form?.values?.amount || defaultValues.amount);
    let selectedCurrency = $derived(form?.values?.currency || defaultValues.currency);
  </script>
  
  <div class="px-4 sm:px-6 lg:px-8 py-8 max-w-4xl mx-auto">
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-2xl font-semibold text-gray-900">Add New Price</h1>
        <p class="mt-1 text-sm text-gray-500">
          Create a new pricing configuration for your products
        </p>
      </div>
      <div>
        <button 
          type="button"
          onclick={handleCancel}
          class="inline-flex items-center rounded-md bg-white px-3 py-2 text-sm font-semibold text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 hover:bg-gray-50"
        >
          Cancel
        </button>
      </div>
    </div>
    
    {#if error || form?.error}
      <div class="mt-4 rounded-md bg-red-50 p-4">
        <div class="flex">
          <div class="flex-shrink-0">
            <svg class="h-5 w-5 text-red-400" viewBox="0 0 20 20" fill="currentColor" aria-hidden="true">
              <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.28 7.22a.75.75 0 00-1.06 1.06L8.94 10l-1.72 1.72a.75.75 0 101.06 1.06L10 11.06l1.72 1.72a.75.75 0 101.06-1.06L11.06 10l1.72-1.72a.75.75 0 00-1.06-1.06L10 8.94 8.28 7.22z" clip-rule="evenodd" />
            </svg>
          </div>
          <div class="ml-3">
            <h3 class="text-sm font-medium text-red-800">{error || form?.error}</h3>
          </div>
        </div>
      </div>
    {/if}
    
    <div class="mt-6 bg-white shadow-sm ring-1 ring-gray-900/5 sm:rounded-xl md:col-span-2">
      <form 
        method="POST" 
        action="?/createPrice" 
        class="px-4 py-6 sm:p-8"
        use:enhance={() => {
          isSubmitting = true;
          
          return async ({ update }) => {
            await update();
            isSubmitting = false;
          };
        }}
      >
        <div class="space-y-8">
          <!-- Product selection -->
          <div>
            <label for="product_id" class="block text-sm font-medium leading-6 text-gray-900">
              Product <span class="text-red-500">*</span>
            </label>
            <div class="mt-2">
              <select
                id="product_id"
                name="product_id"
                required
                class="block w-full rounded-md border-0 py-1.5 px-2 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm sm:leading-6"
              >
                <option value="" disabled selected={!form?.values?.product_id}>Select a product</option>
                {#each products as product}
                  <option 
                    value={product.id} 
                    selected={form?.values?.product_id === product.id}
                  >
                    {product.name}
                  </option>
                {/each}
              </select>
            </div>
          </div>
          
          <!-- Price amount and currency -->
          <div class="grid grid-cols-1 gap-4 sm:grid-cols-2">
            <div>
              <label for="amount" class="block text-sm font-medium leading-6 text-gray-900">
                Amount (in cents) <span class="text-red-500">*</span>
              </label>
              <div class="mt-2">
                <input
                  type="number"
                  id="amount"
                  name="amount"
                  min="1"
                  required
                  value={form?.values?.amount ?? defaultValues.amount}
                  oninput={(e) => selectedAmount = Number(e.currentTarget.value)}
                  class="block w-full rounded-md border-0 py-1.5 px-2 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm sm:leading-6"
                />
              </div>
              <p class="mt-1 text-xs text-gray-500">
                Displayed as {formatAmountDisplay(Number(selectedAmount) || 0, selectedCurrency.toString() || 'usd')}
              </p>
            </div>
            
            <div>
              <label for="currency" class="block text-sm font-medium leading-6 text-gray-900">
                Currency <span class="text-red-500">*</span>
              </label>
              <div class="mt-2">
                <select
                  id="currency"
                  name="currency"
                  required
                  onchange={(e) => selectedCurrency = e.currentTarget.value}
                  class="block w-full rounded-md border-0 py-1.5 px-2 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm sm:leading-6"
                >
                  {#each currencies as { code, name }}
                    <option 
                      value={code} 
                      selected={form?.values?.currency === code || (code === defaultValues.currency && !form?.values?.currency)}
                    >
                      {name}
                    </option>
                  {/each}
                </select>
              </div>
            </div>
          </div>
          
          <!-- Billing interval -->
          <div class="grid grid-cols-1 gap-4 sm:grid-cols-2">
            <div>
              <label for="interval_type" class="block text-sm font-medium leading-6 text-gray-900">
                Billing Interval <span class="text-red-500">*</span>
              </label>
              <div class="mt-2">
                <select
                  id="interval_type"
                  name="interval_type"
                  required
                  class="block w-full rounded-md border-0 py-1.5 px-2 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm sm:leading-6"
                >
                  {#each intervalTypes as { value, label }}
                    <option 
                      value={value} 
                      selected={form?.values?.interval_type === value || (value === defaultValues.interval_type && !form?.values?.interval_type)}
                    >
                      {label}
                    </option>
                  {/each}
                </select>
              </div>
            </div>
            
            <div>
              <label for="interval_count" class="block text-sm font-medium leading-6 text-gray-900">
                Interval Count <span class="text-red-500">*</span>
              </label>
              <div class="mt-2">
                <input
                  type="number"
                  id="interval_count"
                  name="interval_count"
                  min="1"
                  required
                  value={form?.values?.interval_count ?? defaultValues.interval_count}
                  class="block w-full rounded-md border-0 py-1.5 px-2 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm sm:leading-6"
                />
              </div>
              <p class="mt-1 text-xs text-gray-500">
                E.g., for "every 3 months", set Interval to "Monthly" and Count to "3"
              </p>
            </div>
          </div>
          
          <!-- Trial period -->
          <div>
            <label for="trial_period_days" class="block text-sm font-medium leading-6 text-gray-900">
              Trial Period (days)
            </label>
            <div class="mt-2">
              <input
                type="number"
                id="trial_period_days"
                name="trial_period_days"
                min="0"
                value={form?.values?.trial_period_days ?? ''}
                class="block w-full rounded-md border-0 py-1.5 px-2 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm sm:leading-6"
              />
            </div>
            <p class="mt-1 text-xs text-gray-500">
              Optional. Leave empty for no trial period.
            </p>
          </div>
          
          <!-- Nickname -->
          <div>
            <label for="nickname" class="block text-sm font-medium leading-6 text-gray-900">
              Nickname
            </label>
            <div class="mt-2">
              <input
                type="text"
                id="nickname"
                name="nickname"
                value={form?.values?.nickname ?? ''}
                class="block w-full rounded-md border-0 py-1.5 px-2 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm sm:leading-6"
              />
            </div>
            <p class="mt-1 text-xs text-gray-500">
              Optional. A friendly name to help identify this price.
            </p>
          </div>
          
          <!-- Is Active switch -->
          <div class="relative flex items-start">
            <div class="flex h-6 items-center">
              <input
                id="is_active"
                name="is_active"
                type="checkbox"
                checked={form?.values?.is_active === 'on' || (!form && defaultValues.is_active)}
                class="h-4 w-4 rounded border-gray-300 text-indigo-600 focus:ring-indigo-600"
              />
            </div>
            <div class="ml-3 text-sm leading-6">
              <label for="is_active" class="font-medium text-gray-900">Active</label>
              <p class="text-gray-500">Make this price active and available immediately</p>
            </div>
          </div>
          
          <!-- Metadata fields -->
          <div>
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
              {#each metadataFields as _, index}
                <div class="flex space-x-2">
                  <div class="w-1/3">
                    <input
                      type="text"
                      name="metadata_key"
                      placeholder="Key"
                      bind:value={metadataFields[index].key}
                      class="block w-full rounded-md border-0 py-1.5 px-2 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm sm:leading-6"
                    />
                  </div>
                  <div class="flex-1">
                    <input
                      type="text"
                      name="metadata_value"
                      placeholder="Value"
                      bind:value={metadataFields[index].value}
                      class="block w-full rounded-md border-0 py-1.5 px-2 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm sm:leading-6"
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
        </div>
        
        <div class="mt-8 flex justify-end">
          <button
            type="button"
            onclick={handleCancel}
            class="mr-3 rounded-md bg-white px-3 py-2 text-sm font-semibold text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 hover:bg-gray-50"
          >
            Cancel
          </button>
          <button
            type="submit"
            disabled={isSubmitting}
            class="rounded-md bg-indigo-600 px-3 py-2 text-sm font-semibold text-white shadow-sm hover:bg-indigo-500 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-indigo-600 disabled:opacity-50 disabled:cursor-not-allowed"
          >
            {#if isSubmitting}
              <svg class="animate-spin -ml-1 mr-2 h-4 w-4 text-white inline-block" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
              </svg>
              Creating...
            {:else}
              Create Price
            {/if}
          </button>
        </div>
      </form>
    </div>
  </div>