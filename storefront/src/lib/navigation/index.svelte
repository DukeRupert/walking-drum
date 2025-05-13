<script lang="ts">
	import { PUBLIC_NAME, PUBLIC_LOGO } from '$env/static/public';
	import Mobile from './mobile.svelte';
	import Cart from '$lib/components/cart.svelte';

	let { announcement = '' } = $props();

	let isMobileOpen = $state(false);
	let isCartOpen = $state(false);
</script>

<div class="bg-white">
	<Mobile bind:isMobileOpen={isMobileOpen} />

	<header class="relative">
		<nav aria-label="Top">
			<!-- Top navigation -->
			<div class="bg-gray-900">
				<div class="mx-auto flex h-10 max-w-7xl items-center justify-between px-4 sm:px-6 lg:px-8">
					<!-- Currency selector -->
					<div class="hidden lg:block lg:flex-1"></div>

					{#if announcement}
						<p class="flex-1 text-center text-sm font-medium text-white lg:flex-none">
							{announcement}
						</p>
					{/if}

					<div class="hidden lg:flex lg:flex-1 lg:items-center lg:justify-end lg:space-x-6">
						<a href="#" class="text-sm font-medium text-white hover:text-gray-100"
							>Create an account</a
						>
						<span class="h-6 w-px bg-gray-600" aria-hidden="true"></span>
						<a href="#" class="text-sm font-medium text-white hover:text-gray-100">Sign in</a>
					</div>
				</div>
			</div>

			<!-- Secondary navigation -->
			<div class="bg-white">
				<div class="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
					<div class="border-b border-gray-200">
						<div class="flex h-16 items-center justify-between">
							<!-- Logo (lg+) -->
							<div class="hidden lg:flex lg:items-center">
								<a href="/">
									<span class="sr-only">{PUBLIC_NAME}</span>
									<img class="h-8 w-auto" src={PUBLIC_LOGO} alt={`${PUBLIC_NAME} logo`} />
								</a>
							</div>

							<div class="hidden h-full lg:flex">
								<!-- Mega menus -->
								<div class="ml-8">
									<div class="flex h-full justify-center space-x-8">
										<a
											href="/products"
											class="flex items-center text-sm font-medium text-gray-700 hover:text-gray-800"
											>Coffee</a
										>
                    <a
                    href="#"
                    class="flex items-center text-sm font-medium text-gray-700 hover:text-gray-800"
                    >About us</a
                  >
									</div>
								</div>
							</div>

							<!-- Mobile menu and search (lg-) -->
							<div class="flex flex-1 items-center lg:hidden">
								<!-- Mobile menu toggle, controls the 'mobileMenuOpen' state. -->
								<button type="button" onclick={() => isMobileOpen = true} class="-ml-2 rounded-md bg-white p-2 text-gray-400">
									<span class="sr-only">Open menu</span>
									<svg
										class="size-6"
										fill="none"
										viewBox="0 0 24 24"
										stroke-width="1.5"
										stroke="currentColor"
										aria-hidden="true"
										data-slot="icon"
									>
										<path
											stroke-linecap="round"
											stroke-linejoin="round"
											d="M3.75 6.75h16.5M3.75 12h16.5m-16.5 5.25h16.5"
										/>
									</svg>
								</button>

								<!-- Search -->
								<a href="#" class="ml-2 p-2 text-gray-400 hover:text-gray-500">
									<span class="sr-only">Search</span>
									<svg
										class="size-6"
										fill="none"
										viewBox="0 0 24 24"
										stroke-width="1.5"
										stroke="currentColor"
										aria-hidden="true"
										data-slot="icon"
									>
										<path
											stroke-linecap="round"
											stroke-linejoin="round"
											d="m21 21-5.197-5.197m0 0A7.5 7.5 0 1 0 5.196 5.196a7.5 7.5 0 0 0 10.607 10.607Z"
										/>
									</svg>
								</a>
							</div>

							<!-- Logo (lg-) -->
							<a href="/" class="lg:hidden">
								<span class="sr-only">{PUBLIC_NAME}</span>
								<img src={PUBLIC_LOGO} alt={`${PUBLIC_NAME} logo`} class="h-8 w-auto" />
							</a>

							<div class="flex flex-1 items-center justify-end">
								<div class="flex items-center lg:ml-8">
									<div class="flex space-x-8">
										<div class="hidden lg:flex">
											<a href="#" class="-m-2 p-2 text-gray-400 hover:text-gray-500">
												<span class="sr-only">Search</span>
												<svg
													class="size-6"
													fill="none"
													viewBox="0 0 24 24"
													stroke-width="1.5"
													stroke="currentColor"
													aria-hidden="true"
													data-slot="icon"
												>
													<path
														stroke-linecap="round"
														stroke-linejoin="round"
														d="m21 21-5.197-5.197m0 0A7.5 7.5 0 1 0 5.196 5.196a7.5 7.5 0 0 0 10.607 10.607Z"
													/>
												</svg>
											</a>
										</div>

										<div class="flex">
											<a href="#" class="-m-2 p-2 text-gray-400 hover:text-gray-500">
												<span class="sr-only">Account</span>
												<svg
													class="size-6"
													fill="none"
													viewBox="0 0 24 24"
													stroke-width="1.5"
													stroke="currentColor"
													aria-hidden="true"
													data-slot="icon"
												>
													<path
														stroke-linecap="round"
														stroke-linejoin="round"
														d="M15.75 6a3.75 3.75 0 1 1-7.5 0 3.75 3.75 0 0 1 7.5 0ZM4.501 20.118a7.5 7.5 0 0 1 14.998 0A17.933 17.933 0 0 1 12 21.75c-2.676 0-5.216-.584-7.499-1.632Z"
													/>
												</svg>
											</a>
										</div>
									</div>

									<span class="mx-4 h-6 w-px bg-gray-200 lg:mx-6" aria-hidden="true"></span>

									<!-- Cart Component -->
									<div class="flow-root">
										<Cart bind:isOpen={isCartOpen} />
									</div>
								</div>
							</div>
						</div>
					</div>
				</div>
			</div>
		</nav>
	</header>
</div>