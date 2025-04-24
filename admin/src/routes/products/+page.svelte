<script lang="ts">
    import type { PageData } from './$types';
    
    export let data: PageData;
    
    $: ({ products, pagination, error } = data);
  </script>
  
  <div class="admin-panel">
    <h1>Products</h1>
    
    {#if error}
      <div class="error-message">{error}</div>
    {/if}
    
    <div class="filters">
      <label>
        <input type="checkbox" bind:checked={pagination.activeOnly} on:change={() => {
          // Update URL params and refresh
          const url = new URL(window.location.href);
          url.searchParams.set('active', pagination.activeOnly.toString());
          window.location.href = url.toString();
        }} />
        Show active only
      </label>
    </div>
    
    <table>
      <thead>
        <tr>
          <th>ID</th>
          <th>Name</th>
          <th>Description</th>
          <th>Status</th>
          <th>Created</th>
          <th>Updated</th>
          <th>Actions</th>
        </tr>
      </thead>
      <tbody>
        {#each products as product}
          <tr>
            <td>{product.id}</td>
            <td>{product.name}</td>
            <td>{product.description}</td>
            <td>{product.is_active ? 'Active' : 'Inactive'}</td>
            <td>{new Date(product.created_at).toLocaleString()}</td>
            <td>{new Date(product.updated_at).toLocaleString()}</td>
            <td>
              <button>Edit</button>
              <button>Delete</button>
            </td>
          </tr>
        {/each}
      </tbody>
    </table>
    
    <div class="pagination">
      <button disabled={pagination.offset <= 0} on:click={() => {
        const url = new URL(window.location.href);
        url.searchParams.set('offset', String(Math.max(0, pagination.offset - pagination.limit)));
        window.location.href = url.toString();
      }}>Previous</button>
      
      <span>Page {Math.floor(pagination.offset / pagination.limit) + 1}</span>
      
      <button disabled={products.length < pagination.limit} on:click={() => {
        const url = new URL(window.location.href);
        url.searchParams.set('offset', String(pagination.offset + pagination.limit));
        window.location.href = url.toString();
      }}>Next</button>
    </div>
  </div>
  
  <style>
    .admin-panel {
      max-width: 1200px;
      margin: 0 auto;
      padding: 20px;
    }
    
    table {
      width: 100%;
      border-collapse: collapse;
      margin: 20px 0;
    }
    
    th, td {
      padding: 10px;
      text-align: left;
      border-bottom: 1px solid #ddd;
    }
    
    th {
      background-color: #f2f2f2;
    }
    
    .error-message {
      color: red;
      padding: 10px;
      background-color: #ffeeee;
      border: 1px solid #ffcccc;
      margin-bottom: 20px;
    }
    
    .pagination {
      display: flex;
      justify-content: center;
      gap: 20px;
      align-items: center;
      margin-top: 20px;
    }
    
    .filters {
      margin-bottom: 20px;
    }
  </style>