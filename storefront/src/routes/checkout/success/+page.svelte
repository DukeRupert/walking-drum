<script lang="ts">
    import { onMount } from 'svelte';
    import { goto } from '$app/navigation';
    import type { PageProps } from './$types';


	let { data }: PageProps = $props();
    
    const { subscriptionId, subscription, error } = data;
    
    function formatDate(dateString) {
      const date = new Date(dateString);
      return date.toLocaleDateString('en-US', {
        weekday: 'long',
        year: 'numeric',
        month: 'long',
        day: 'numeric'
      });
    }
  </script>
  
  <svelte:head>
    <title>Subscription Confirmed | Walking Drum Coffee</title>
  </svelte:head>
  
  {#if error}
    <div class="error-container">
      <h2>Something went wrong</h2>
      <p>{error}</p>
      <button onclick={() => goto('/products')}>Browse Subscriptions</button>
    </div>
  {:else}
    <div class="success-container">
      <div class="success-icon">
        <svg viewBox="0 0 24 24" width="64" height="64">
          <circle cx="12" cy="12" r="11" fill="#4BB543" />
          <path
            d="M9 12l2 2 4-4"
            fill="none"
            stroke="white"
            stroke-width="2"
            stroke-linecap="round"
            stroke-linejoin="round"
          />
        </svg>
      </div>
      
      <h1>Subscription Confirmed!</h1>
      <p class="success-message">Thank you for subscribing to Walking Drum Coffee.</p>
      
      <div class="subscription-details">
        <h2>Your Subscription Details</h2>
        <div class="detail-row">
          <span>Subscription ID:</span>
          <span>{subscriptionId}</span>
        </div>
        <div class="detail-row">
          <span>Product:</span>
          <span>{subscription.product_name}</span>
        </div>
        <div class="detail-row">
          <span>Status:</span>
          <span class="status-badge">{subscription.status}</span>
        </div>
        <div class="detail-row">
          <span>Price:</span>
          <span>${(subscription.amount / 100).toFixed(2)}/{subscription.interval}</span>
        </div>
        <div class="detail-row">
          <span>Next Delivery Date:</span>
          <span>{formatDate(subscription.next_delivery_date)}</span>
        </div>
      </div>
      
      <div class="what-next">
        <h3>What's Next?</h3>
        <ul>
          <li>You'll receive a confirmation email shortly.</li>
          <li>Your first coffee package will be roasted and shipped soon.</li>
          <li>You can manage your subscription anytime from your account dashboard.</li>
        </ul>
      </div>
      
      <div class="success-actions">
        <button class="button-secondary" onclick={() => goto('/account/subscriptions')}>
          Manage Subscription
        </button>
        <button on:click={() => goto('/products')}>
          Browse More Coffees
        </button>
      </div>
    </div>
  {/if}
  
  <style>
    .success-container {
      max-width: 600px;
      margin: 3rem auto;
      padding: 2rem;
      background-color: white;
      border-radius: var(--radius);
      box-shadow: var(--shadow);
      text-align: center;
    }
  
    .success-icon {
      margin: 0 auto 2rem;
      width: 80px;
      height: 80px;
    }
  
    h1 {
      margin-bottom: 1rem;
      color: var(--color-primary);
    }
  
    .success-message {
      font-size: 1.2rem;
      color: var(--color-muted);
      margin-bottom: 2.5rem;
    }
  
    .subscription-details {
      text-align: left;
      margin-bottom: 2.5rem;
      padding: 1.5rem;
      background-color: var(--color-light);
      border-radius: var(--radius);
    }
  
    .subscription-details h2 {
      margin-bottom: 1.5rem;
      font-size: 1.3rem;
    }
  
    .detail-row {
      display: flex;
      justify-content: space-between;
      margin-bottom: 1rem;
      padding-bottom: 1rem;
      border-bottom: 1px solid var(--color-border);
    }
  
    .detail-row:last-child {
      border-bottom: none;
      margin-bottom: 0;
      padding-bottom: 0;
    }
  
    .status-badge {
      background-color: var(--color-success);
      color: white;
      padding: 0.2rem 0.5rem;
      border-radius: 20px;
      font-size: 0.8rem;
      text-transform: capitalize;
    }
  
    .what-next {
      margin-bottom: 2.5rem;
      text-align: left;
    }
  
    .what-next h3 {
      margin-bottom: 1rem;
      font-size: 1.2rem;
    }
  
    .what-next ul {
      padding-left: 1.5rem;
      color: var(--color-muted);
    }
  
    .what-next li {
      margin-bottom: 0.5rem;
    }
  
    .success-actions {
      display: flex;
      justify-content: center;
      gap: 1.5rem;
    }
  
    .button-secondary {
      background-color: white;
      color: var(--color-text);
      border: 1px solid var(--color-border);
    }
  
    .button-secondary:hover {
      background-color: var(--color-light);
    }
  </style>