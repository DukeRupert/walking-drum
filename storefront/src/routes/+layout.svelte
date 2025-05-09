<script lang="ts">
	import '../app.css';
	import Navigation from '$lib/navigation/index.svelte'

	let { children } = $props();

	const announcement = 'Get free delivery on orders over $100'
</script>

<div class="app">
	<Navigation {announcement} />
  
	<main class="container">
	  {@render children()}
	</main>
  
	<footer>
	  <div class="container">
		<p>&copy; {new Date().getFullYear()} Walking Drum Coffee. All rights reserved.</p>
	  </div>
	</footer>
  </div>
  
  <style>
	.app {
	  display: flex;
	  flex-direction: column;
	  min-height: 100vh;
	}
  
	.container {
	  max-width: 1200px;
	  margin: 0 auto;
	  padding: 0 1rem;
	}
  
	header {
	  background-color: #2c1810;
	  color: white;
	  padding: 1rem 0;
	}
  
	header .container {
	  display: flex;
	  justify-content: space-between;
	  align-items: center;
	}
  
	.logo a {
	  font-size: 1.5rem;
	  font-weight: bold;
	  color: white;
	  text-decoration: none;
	}
  
	nav ul {
	  display: flex;
	  list-style: none;
	  gap: 2rem;
	  margin: 0;
	  padding: 0;
	}
  
	nav a {
	  color: white;
	  text-decoration: none;
	  font-weight: 500;
	  transition: color 0.2s;
	}
  
	nav a:hover {
	  color: #e6c9a8;
	}
  
	main {
	  flex: 1;
	  padding: 2rem 0;
	}
  
	footer {
	  background-color: #f5f5f5;
	  padding: 2rem 0;
	  margin-top: auto;
	}
  </style>
