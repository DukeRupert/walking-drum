<script lang="ts">
    import { onMount } from 'svelte';
    import { goto } from '$app/navigation';
    import type { PageProps } from './$types';

	let { data }: PageProps = $props();
  let { price, product, customer, error } = $derived(data);
    
    let loading = $state(false);
    let checkoutError: string | null = $state(null);
    let formData = $state({
      name: customer?.first_name + ' ' + customer?.last_name || '',
      email: customer?.email || '',
      address: {
        line1: '',
        line2: '',
        city: '',
        state: '',
        postal_code: '',
        country: 'US'
      },
      shippingMethod: 'standard'
    });
    
    let step = $state(1); // 1: Shipping, 2: Payment, 3: Review

    function goToStep(newStep: number) {
      if (newStep === 2 && step === 1) {
        // Validate shipping info before proceeding
        if (!formData.name || !formData.email || !formData.address.line1 || 
            !formData.address.city || !formData.address.state || 
            !formData.address.postal_code) {
          checkoutError = 'Please fill out all required fields';
          return;
        }
      }
      
      checkoutError = null;
      step = newStep;
    }
    
    async function handleSubmit() {
      try {
        loading = true;
        checkoutError = null;
        
        // In a real implementation, we would:
        // 1. Create a subscription in your backend
        // 2. Process payment via your chosen payment processor
        // For this POC, we'll simulate a successful checkout
        
        await new Promise(resolve => setTimeout(resolve, 1500)); // Simulate network request
        
        // Redirect to success page
        goto(`/checkout/success?subscription_id=sim_123456789`);
        
      } catch (err) {
        console.error('Checkout error:', err);
        checkoutError = err.message || 'An error occurred during checkout';
      } finally {
        loading = false;
      }
    }
  </script>
  
  <svelte:head>
    <title>Checkout | Walking Drum Coffee</title>
  </svelte:head>
  
  {#if error}
    <div class="error-container">
      <h3>Error Loading Checkout</h3>
      <p>{error}</p>
      <button onclick={() => goto('/products')}>Return to Products</button>
    </div>
  {:else if !price || !product || !customer}
    <div class="loading-container">
      <div class="spinner"></div>
      <p>Loading checkout...</p>
    </div>
  {:else}
    <div class="checkout-container">
      <div class="checkout-header">
        <h1>Complete Your Subscription</h1>
      </div>
      
      <div class="checkout-steps">
        <div class="step {step >= 1 ? 'active' : ''}" onclick={() => goToStep(1)}>
          <div class="step-number">1</div>
          <div class="step-text">Shipping</div>
        </div>
        <div class="step-line"></div>
        <div class="step {step >= 2 ? 'active' : ''}" onclick={() => step > 1 && goToStep(2)}>
          <div class="step-number">2</div>
          <div class="step-text">Payment</div>
        </div>
        <div class="step-line"></div>
        <div class="step {step >= 3 ? 'active' : ''}" onclick={() => step > 2 && goToStep(3)}>
          <div class="step-number">3</div>
          <div class="step-text">Review</div>
        </div>
      </div>
      
      <div class="checkout-content">
        <div class="checkout-main">
          {#if checkoutError}
            <div class="checkout-error">
              <p>{checkoutError}</p>
            </div>
          {/if}
          
          {#if step === 1}
            <!-- Shipping Information -->
            <div class="form-section">
              <h2>Shipping Information</h2>
              <div class="form-group">
                <label for="name">Full Name</label>
                <input 
                  type="text" 
                  id="name" 
                  bind:value={formData.name} 
                  required
                />
              </div>
              
              <div class="form-group">
                <label for="email">Email Address</label>
                <input 
                  type="email" 
                  id="email" 
                  bind:value={formData.email} 
                  required
                />
              </div>
              
              <div class="form-group">
                <label for="line1">Address Line 1</label>
                <input 
                  type="text" 
                  id="line1" 
                  bind:value={formData.address.line1} 
                  required
                />
              </div>
              
              <div class="form-group">
                <label for="line2">Address Line 2 (Optional)</label>
                <input 
                  type="text" 
                  id="line2" 
                  bind:value={formData.address.line2}
                />
              </div>
              
              <div class="form-row">
                <div class="form-group">
                  <label for="city">City</label>
                  <input 
                    type="text" 
                    id="city" 
                    bind:value={formData.address.city} 
                    required
                  />
                </div>
                
                <div class="form-group">
                  <label for="state">State</label>
                  <input 
                    type="text" 
                    id="state" 
                    bind:value={formData.address.state} 
                    required
                  />
                </div>
              </div>
              
              <div class="form-row">
                <div class="form-group">
                  <label for="postal_code">Postal Code</label>
                  <input 
                    type="text" 
                    id="postal_code" 
                    bind:value={formData.address.postal_code} 
                    required
                  />
                </div>
                
                <div class="form-group">
                  <label for="country">Country</label>
                  <select id="country" bind:value={formData.address.country}>
                    <option value="US">United States</option>
                    <option value="CA">Canada</option>
                    <option value="MX">Mexico</option>
                  </select>
                </div>
              </div>
              
              <div class="form-group">
                <label>Shipping Method</label>
                <div class="radio-group">
                  <label class="radio-label">
                    <input 
                      type="radio" 
                      name="shippingMethod" 
                      value="standard" 
                      bind:group={formData.shippingMethod}
                    />
                    <span>Standard Shipping (Free)</span>
                  </label>
                  <label class="radio-label">
                    <input 
                      type="radio" 
                      name="shippingMethod" 
                      value="express" 
                      bind:group={formData.shippingMethod}
                    />
                    <span>Express Shipping (+$5.99)</span>
                  </label>
                </div>
              </div>
              
              <div class="form-actions">
                <button class="button-secondary" onclick={() => goto('/products')}>
                  Back to Products
                </button>
                <button onclick={() => goToStep(2)}>
                  Continue to Payment
                </button>
              </div>
            </div>
          {:else if step === 2}
            <!-- Payment Information -->
            <div class="form-section">
              <h2>Payment Information</h2>
              <p>For this proof of concept, we'll simulate payment processing.</p>
              <p>In a real implementation, you would integrate with your payment processor of choice.</p>
              
              <div class="payment-demo">
                <div class="card-input">
                  <label>Card Number</label>
                  <input type="text" placeholder="4242 4242 4242 4242" />
                </div>
                
                <div class="form-row">
                  <div class="card-input">
                    <label>Expiry Date</label>
                    <input type="text" placeholder="MM/YY" />
                  </div>
                  
                  <div class="card-input">
                    <label>CVC</label>
                    <input type="text" placeholder="123" />
                  </div>
                </div>
              </div>
              
              <div class="form-actions">
                <button class="button-secondary" onclick={() => goToStep(1)}>
                  Back to Shipping
                </button>
                <button onclick={() => goToStep(3)}>
                  Continue to Review
                </button>
              </div>
            </div>
          {:else if step === 3}
            <!-- Review Order -->
            <div class="form-section">
              <h2>Review Your Subscription</h2>
              
              <div class="review-section">
                <h3>Shipping Address</h3>
                <div class="review-details">
                  <p>{formData.name}</p>
                  <p>{formData.address.line1}</p>
                  {#if formData.address.line2}
                    <p>{formData.address.line2}</p>
                  {/if}
                  <p>{formData.address.city}, {formData.address.state} {formData.address.postal_code}</p>
                  <p>{formData.address.country}</p>
                </div>
              </div>
              
              <div class="review-section">
                <h3>Shipping Method</h3>
                <div class="review-details">
                  <p>
                    {formData.shippingMethod === 'standard' ? 'Standard Shipping (Free)' : 'Express Shipping (+$5.99)'}
                  </p>
                </div>
              </div>
              
              <div class="review-section">
                <h3>Payment Method</h3>
                <div class="review-details">
                  <p>Credit Card ending in 4242</p>
                </div>
              </div>
              
              <div class="form-actions">
                <button class="button-secondary" onclick={() => goToStep(2)}>
                  Back to Payment
                </button>
                <button onclick={handleSubmit} disabled={loading}>
                  {loading ? 'Processing...' : 'Complete Subscription'}
                </button>
              </div>
            </div>
          {/if}
        </div>
        
        <div class="order-summary">
          <h3>Order Summary</h3>
          
          <div class="product-summary">
            <img src={product.image_url || '/images/coffee-placeholder.jpg'} alt={product.name} />
            <div class="product-summary-details">
              <h4>{product.name}</h4>
              <p>{price.name}</p>
            </div>
          </div>
          
          <div class="price-breakdown">
            <div class="price-row">
              <span>Subscription</span>
              <span>${(price.amount / 100).toFixed(2)}/{price.interval}</span>
            </div>
            
            {#if formData.shippingMethod === 'express'}
              <div class="price-row">
                <span>Express Shipping</span>
                <span>$5.99</span>
              </div>
            {:else}
              <div class="price-row">
                <span>Standard Shipping</span>
                <span>Free</span>
              </div>
            {/if}
            
            <div class="price-total">
              <span>Total per {price.interval}</span>
              <span>
                ${formData.shippingMethod === 'express' 
                  ? ((price.amount / 100) + 5.99).toFixed(2) 
                  : (price.amount / 100).toFixed(2)}
              </span>
            </div>
          </div>
          
          <div class="subscription-details">
            <h4>Subscription Details</h4>
            <p>You will receive your first delivery within 3-5 business days.</p>
            <p>Your subscription will automatically renew every {price.interval}.</p>
            <p>You can pause or cancel anytime from your account.</p>
          </div>
        </div>
      </div>
    </div>
  {/if}
  
  <style>
    .checkout-container {
      max-width: 1000px;
      margin: 0 auto;
    }
  
    .checkout-header {
      text-align: center;
      margin-bottom: 2rem;
    }
  
    .checkout-steps {
      display: flex;
      align-items: center;
      justify-content: center;
      margin-bottom: 2.5rem;
    }
  
    .step {
      display: flex;
      flex-direction: column;
      align-items: center;
      position: relative;
      cursor: pointer;
    }
  
    .step-number {
      width: 30px;
      height: 30px;
      border-radius: 50%;
      background-color: var(--color-light);
      display: flex;
      align-items: center;
      justify-content: center;
      margin-bottom: 0.5rem;
      border: 2px solid var(--color-border);
      font-weight: bold;
      transition: all 0.3s;
    }
  
    .step.active .step-number {
      background-color: var(--color-secondary);
      border-color: var(--color-secondary);
      color: white;
    }
  
    .step-text {
      font-size: 0.9rem;
      color: var(--color-muted);
      transition: color 0.3s;
    }
  
    .step.active .step-text {
      color: var(--color-text);
      font-weight: 500;
    }
  
    .step-line {
      height: 2px;
      width: 100px;
      background-color: var(--color-border);
      margin: 0 0.5rem;
    }
  
    .checkout-content {
      display: grid;
      grid-template-columns: 1fr 350px;
      gap: 2rem;
      align-items: start;
    }
  
    .checkout-main {
      background-color: white;
      border-radius: var(--radius);
      box-shadow: var(--shadow);
      padding: 2rem;
    }
  
    .checkout-error {
      background-color: #ffeeee;
      color: var(--color-error);
      padding: 1rem;
      border-radius: var(--radius);
      margin-bottom: 1.5rem;
      border: 1px solid #ffcccc;
    }
  
    .form-section h2 {
      margin-bottom: 1.5rem;
      color: var(--color-primary);
    }
  
    .form-group {
      margin-bottom: 1.5rem;
    }
  
    .form-row {
      display: grid;
      grid-template-columns: 1fr 1fr;
      gap: 1rem;
      margin-bottom: 1.5rem;
    }
  
    label {
      display: block;
      margin-bottom: 0.5rem;
      font-weight: 500;
      color: var(--color-text);
    }
  
    input, select {
      width: 100%;
      padding: 0.8rem;
      border: 1px solid var(--color-border);
      border-radius: var(--radius);
      font-size: 1rem;
      font-family: inherit;
    }
  
    input:focus, select:focus {
      outline: none;
      border-color: var(--color-secondary);
      box-shadow: 0 0 0 2px rgba(184, 92, 56, 0.2);
    }
  
    .radio-group {
      display: flex;
      flex-direction: column;
      gap: 0.8rem;
    }
  
    .radio-label {
      display: flex;
      align-items: center;
      gap: 0.5rem;
      cursor: pointer;
    }
  
    .radio-label input {
      width: auto;
    }
  
    .form-actions {
      display: flex;
      justify-content: space-between;
      margin-top: 2rem;
    }
  
    .button-secondary {
      background-color: white;
      color: var(--color-text);
      border: 1px solid var(--color-border);
    }
  
    .button-secondary:hover {
      background-color: var(--color-light);
    }
  
    .order-summary {
      background-color: white;
      border-radius: var(--radius);
      box-shadow: var(--shadow);
      padding: 1.5rem;
      position: sticky;
      top: 2rem;
    }
  
    .order-summary h3 {
      margin-bottom: 1.5rem;
      padding-bottom: 1rem;
      border-bottom: 1px solid var(--color-border);
    }
  
    .product-summary {
      display: flex;
      gap: 1rem;
      margin-bottom: 1.5rem;
      padding-bottom: 1.5rem;
      border-bottom: 1px solid var(--color-border);
    }
  
    .product-summary img {
      width: 80px;
      height: 80px;
      object-fit: cover;
      border-radius: var(--radius);
    }
  
    .product-summary-details h4 {
      margin-bottom: 0.2rem;
      font-size: 1rem;
    }
  
    .product-summary-details p {
      color: var(--color-muted);
      font-size: 0.9rem;
      margin: 0;
    }
  
    .price-breakdown {
      margin-bottom: 1.5rem;
    }
  
    .price-row {
      display: flex;
      justify-content: space-between;
      margin-bottom: 0.8rem;
      color: var(--color-muted);
    }
  
    .price-total {
      display: flex;
      justify-content: space-between;
      font-weight: bold;
      font-size: 1.1rem;
      margin-top: 1rem;
      padding-top: 1rem;
      border-top: 1px solid var(--color-border);
    }
  
    .subscription-details {
      margin-top: 1.5rem;
      padding-top: 1.5rem;
      border-top: 1px solid var(--color-border);
    }
  
    .subscription-details h4 {
      margin-bottom: 0.8rem;
      font-size: 1rem;
    }
  
    .subscription-details p {
      font-size: 0.9rem;
      color: var(--color-muted);
      margin-bottom: 0.5rem;
    }
  
    .payment-demo {
      background-color: var(--color-light);
      padding: 1.5rem;
      border-radius: var(--radius);
      margin: 1.5rem 0;
    }
  
    .card-input {
      margin-bottom: 1rem;
    }
  
    .review-section {
      margin-bottom: 1.5rem;
      padding-bottom: 1.5rem;
      border-bottom: 1px solid var(--color-border);
    }
  
    .review-section h3 {
      font-size: 1.1rem;
      margin-bottom: 0.8rem;
    }
  
    .review-details p {
      margin-bottom: 0.3rem;
      color: var(--color-muted);
    }
  
    @media (max-width: 768px) {
      .checkout-content {
        grid-template-columns: 1fr;
      }
  
      .order-summary {
        position: static;
      }
  
      .step-line {
        width: 50px;
      }
    }
  </style>