<script lang="ts">
  import { goto } from '$app/navigation';
</script>

<svelte:head>
  <title>Checkout Canceled | Walking Drum Coffee</title>
</svelte:head>

<div class="cancel-container">
  <div class="cancel-icon">
    <svg viewBox="0 0 24 24" width="64" height="64">
      <circle cx="12" cy="12" r="11" fill="#f44336" />
      <path
        d="M15 9l-6 6M9 9l6 6"
        fill="none"
        stroke="white"
        stroke-width="2"
        stroke-linecap="round"
      />
    </svg>
  </div>
  
  <h1>Checkout Canceled</h1>
  <p class="cancel-message">Your subscription checkout was canceled.</p>
  <p>No charges have been made to your account.</p>
  
  <div class="cancel-actions">
    <button on:click={() => goto('/products')}>
      Return to Subscriptions
    </button>
  </div>
</div>

<style>
  .cancel-container {
    max-width: 500px;
    margin: 3rem auto;
    padding: 2rem;
    background-color: white;
    border-radius: var(--radius);
    box-shadow: var(--shadow);
    text-align: center;
  }

  .cancel-icon {
    margin: 0 auto 2rem;
    width: 80px;
    height: 80px;
  }

  h1 {
    margin-bottom: 1rem;
    color: #f44336;
  }

  .cancel-message {
    font-size: 1.2rem;
    color: var(--color-muted);
    margin-bottom: 0.5rem;
  }

  p {
    color: var(--color-muted);
    margin-bottom: 2rem;
  }

  .cancel-actions {
    display: flex;
    justify-content: center;
  }
</style>