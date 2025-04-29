<!-- src/routes/admin/+layout.svelte -->
<script lang="ts">
	import '../app.css';
	import { page } from '$app/state';
	import { onMount } from 'svelte';

	let { children } = $props();

	// State for mobile menu
	let isMobileMenuOpen = $state(false);

	// Toggle mobile menu
	function toggleMobileMenu() {
		isMobileMenuOpen = !isMobileMenuOpen;
	}

	// Close mobile menu when clicking outside or pressing escape
	function closeMobileMenu() {
		isMobileMenuOpen = false;
	}

	// Handle keyboard events for accessibility
	function handleKeydown(event: KeyboardEvent) {
		if (event.key === 'Escape' && isMobileMenuOpen) {
			closeMobileMenu();
		}
	}

	// Set up event listeners
	onMount(() => {
		// Add event listener for keyboard events
		window.addEventListener('keydown', handleKeydown);

		// Clean up on component unmount
		return () => {
			window.removeEventListener('keydown', handleKeydown);
		};
	});

	// Helper to check if a navigation item is active
	function isActive(href: string) {
		return page.url.pathname.startsWith(href);
	}

	// Icons for navigation items (simplified SVG paths)
	const icons = {
		box: 'M2.25 12.75V12A2.25 2.25 0 014.5 9.75h15A2.25 2.25 0 0121.75 12v.75m-8.69-6.44l-2.12-2.12a1.5 1.5 0 00-1.061-.44H4.5A2.25 2.25 0 002.25 6v12a2.25 2.25 0 002.25 2.25h15A2.25 2.25 0 0021.75 18V9a2.25 2.25 0 00-2.25-2.25h-5.379a1.5 1.5 0 01-1.06-.44Z',
		tag: 'M9.568 3H5.25A2.25 2.25 0 003 5.25v4.318c0 .597.237 1.17.659 1.591l9.581 9.581c.699.699 1.78.872 2.607.33a18.095 18.095 0 005.223-5.223c.542-.827.369-1.908-.33-2.607L11.16 3.66A2.25 2.25 0 009.568 3z M6 6h.008v.008H6V6z',
		receipt:
			'M9 14.25l6-6m4.5-3.493V21.75l-3.75-1.5-3.75 1.5-3.75-1.5-3.75 1.5V4.757c0-1.108.806-2.057 1.907-2.185a48.507 48.507 0 0111.186 0c1.1.128 1.907 1.077 1.907 2.185z',
		shoppingCart:
			'M2.25 3h1.386c.51 0 .955.343 1.087.835l.383 1.437M7.5 14.25a3 3 0 00-3 3h15.75m-12.75-3h11.218c1.121-2.3 2.1-4.684 2.924-7.138a60.114 60.114 0 00-16.536-1.84M7.5 14.25L5.106 5.272M6 20.25a.75.75 0 11-1.5 0 .75.75 0 011.5 0zm12.75 0a.75.75 0 11-1.5 0 .75.75 0 011.5 0z',
		package:
			'M21 7.5l-9-5.25L3 7.5m18 0l-9 5.25m9-5.25v9l-9 5.25M3 7.5l9 5.25M3 7.5v9l9 5.25m0-9v9',
		repeat:
			'M16.023 9.348h4.002v-.001M2.25 9.75v-4.5m0 4.5h4.5m-4.5 0L9 15M3 20.25v-8.25m18 0v8.25m-18 0h18M3 12h18',
		users:
			'M15 19.128a9.38 9.38 0 002.625.372 9.337 9.337 0 004.121-.952 4.125 4.125 0 00-7.533-2.493M15 19.128v-.003c0-1.113-.285-2.16-.786-3.07M15 19.128v.106A12.318 12.318 0 018.624 21c-2.331 0-4.512-.645-6.374-1.766l-.001-.109a6.375 6.375 0 0111.964-3.07M12 6.375a3.375 3.375 0 11-6.75 0 3.375 3.375 0 016.75 0zm8.25 2.25a2.625 2.625 0 11-5.25 0 2.625 2.625 0 015.25 0z',
		user: 'M15.75 6a3.75 3.75 0 11-7.5 0 3.75 3.75 0 017.5 0zM4.501 20.118a7.5 7.5 0 0114.998 0A17.933 17.933 0 0112 21.75c-2.676 0-5.216-.584-7.499-1.632z',
		settings:
			'M9.594 3.94c.09-.542.56-.94 1.11-.94h2.593c.55 0 1.02.398 1.11.94l.213 1.281c.063.374.313.686.645.87.074.04.147.083.22.127.325.196.72.257 1.075.124l1.217-.456a1.125 1.125 0 011.37.49l1.296 2.247a1.125 1.125 0 01-.26 1.431l-1.003.827c-.293.241-.438.613-.43.992a7.723 7.723 0 010 .255c-.008.378.137.75.43.991l1.004.827c.424.35.534.955.26 1.43l-1.298 2.247a1.125 1.125 0 01-1.369.491l-1.217-.456c-.355-.133-.75-.072-1.076.124a6.47 6.47 0 01-.22.128c-.331.183-.581.495-.644.869l-.213 1.281c-.09.543-.56.94-1.11.94h-2.594c-.55 0-1.019-.398-1.11-.94l-.213-1.281c-.062-.374-.312-.686-.644-.87a6.52 6.52 0 01-.22-.127c-.325-.196-.72-.257-1.076-.124l-1.217.456a1.125 1.125 0 01-1.369-.49l-1.297-2.247a1.125 1.125 0 01.26-1.431l1.004-.827c.292-.24.437-.613.43-.991a6.932 6.932 0 010-.255c.007-.38-.138-.751-.43-.992l-1.004-.827a1.125 1.125 0 01-.26-1.43l1.297-2.247a1.125 1.125 0 011.37-.491l1.216.456c.356.133.751.072 1.076-.124.072-.044.146-.086.22-.128.332-.183.582-.495.644-.869l.214-1.28z M15 12a3 3 0 11-6 0 3 3 0 016 0z'
	};

	const navigationItems = [
		{ href: '/products', label: 'Products', icon: 'box' as const },
		{ href: '/prices', label: 'Prices', icon: 'tag' as const },
		{ href: '/invoices', label: 'Invoices', icon: 'receipt' as const },
		{ href: '/orders', label: 'Orders', icon: 'package' as const },
		{ href: '/subscriptions', label: 'Subscriptions', icon: 'repeat' as const },
		{ href: '/users', label: 'Users', icon: 'user' as const }
	];
</script>

<div class="h-full bg-gray-50">
	<!-- Off-canvas menu for mobile -->
	<div
		class="relative z-50 lg:hidden"
		role="dialog"
		aria-modal="true"
		class:hidden={!isMobileMenuOpen}
	>
		<!-- Backdrop -->
		<div
			class="fixed inset-0 bg-gray-900/80 transition-opacity duration-300 ease-linear"
			class:opacity-100={isMobileMenuOpen}
			class:opacity-0={!isMobileMenuOpen}
			aria-hidden="true"
			onclick={closeMobileMenu}
		></div>

		<!-- Mobile menu panel -->
		<div class="fixed inset-0 flex">
			<div
				class="relative mr-16 flex w-full max-w-xs flex-1 transform transition duration-300 ease-in-out"
				class:translate-x-0={isMobileMenuOpen}
				class:translate-x-[-100%]={!isMobileMenuOpen}
			>
				<!-- Close button -->
				<div
					class="absolute left-full top-0 flex w-16 justify-center pt-5 duration-300 ease-in-out"
					class:opacity-100={isMobileMenuOpen}
					class:opacity-0={!isMobileMenuOpen}
				>
					<button type="button" class="-m-2.5 p-2.5" onclick={closeMobileMenu}>
						<span class="sr-only">Close sidebar</span>
						<svg
							class="size-6 text-white"
							fill="none"
							viewBox="0 0 24 24"
							stroke-width="1.5"
							stroke="currentColor"
							aria-hidden="true"
						>
							<path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12" />
						</svg>
					</button>
				</div>

				<!-- Sidebar component for mobile -->
				<div class="flex grow flex-col gap-y-5 overflow-y-auto bg-indigo-600 px-6 pb-4">
					<div class="flex h-16 shrink-0 items-center">
						<img
							class="h-8 w-auto"
							src="https://tailwindcss.com/plus-assets/img/logos/mark.svg?color=white"
							alt="Your Company"
						/>
					</div>
					<nav class="flex flex-1 flex-col">
						<ul role="list" class="flex flex-1 flex-col gap-y-7">
							<li>
								<ul role="list" class="-mx-2 space-y-1">
									{#each navigationItems as { href, label, icon }}
										<li>
											<a
												{href}
												class="group flex gap-x-3 rounded-md p-2 text-sm/6 font-semibold {isActive(
													href
												)
													? 'bg-indigo-700 text-white'
													: 'text-indigo-200 hover:bg-indigo-700 hover:text-white'}"
												onclick={() => isMobileMenuOpen && closeMobileMenu()}
											>
												<svg
													class="size-6 shrink-0 {isActive(href)
														? 'text-white'
														: 'text-indigo-200 group-hover:text-white'}"
													fill="none"
													viewBox="0 0 24 24"
													stroke-width="1.5"
													stroke="currentColor"
													aria-hidden="true"
												>
													<path stroke-linecap="round" stroke-linejoin="round" d={icons[icon]} />
												</svg>
												{label}
											</a>
										</li>
									{/each}
								</ul>
							</li>
							<li class="mt-auto">
								<a
									href="/admin/settings"
									class="group -mx-2 flex gap-x-3 rounded-md p-2 text-sm/6 font-semibold text-indigo-200 hover:bg-indigo-700 hover:text-white"
								>
									<svg
										class="size-6 shrink-0 text-indigo-200 group-hover:text-white"
										fill="none"
										viewBox="0 0 24 24"
										stroke-width="1.5"
										stroke="currentColor"
										aria-hidden="true"
									>
										<path
											stroke-linecap="round"
											stroke-linejoin="round"
											d="M9.594 3.94c.09-.542.56-.94 1.11-.94h2.593c.55 0 1.02.398 1.11.94l.213 1.281c.063.374.313.686.645.87.074.04.147.083.22.127.325.196.72.257 1.075.124l1.217-.456a1.125 1.125 0 011.37.49l1.296 2.247a1.125 1.125 0 01-.26 1.431l-1.003.827c-.293.241-.438.613-.43.992a7.723 7.723 0 010 .255c-.008.378.137.75.43.991l1.004.827c.424.35.534.955.26 1.43l-1.298 2.247a1.125 1.125 0 01-1.369.491l-1.217-.456c-.355-.133-.75-.072-1.076.124a6.47 6.47 0 01-.22.128c-.331.183-.581.495-.644.869l-.213 1.281c-.09.543-.56.94-1.11.94h-2.594c-.55 0-1.019-.398-1.11-.94l-.213-1.281c-.062-.374-.312-.686-.644-.87a6.52 6.52 0 01-.22-.127c-.325-.196-.72-.257-1.076-.124l-1.217.456a1.125 1.125 0 01-1.369-.49l-1.297-2.247a1.125 1.125 0 01.26-1.431l1.004-.827c.292-.24.437-.613.43-.991a6.932 6.932 0 010-.255c.007-.38-.138-.751-.43-.992l-1.004-.827a1.125 1.125 0 01-.26-1.43l1.297-2.247a1.125 1.125 0 011.37-.491l1.216.456c.356.133.751.072 1.076-.124.072-.044.146-.086.22-.128.332-.183.582-.495.644-.869l.214-1.28z"
										/>
										<path
											stroke-linecap="round"
											stroke-linejoin="round"
											d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"
										/>
									</svg>
									Settings
								</a>
							</li>
						</ul>
					</nav>
				</div>
			</div>
		</div>
	</div>

	<!-- Static sidebar for desktop -->
	<div class="hidden lg:fixed lg:inset-y-0 lg:z-50 lg:flex lg:w-72 lg:flex-col">
		<!-- Sidebar component for desktop -->
		<div class="flex grow flex-col gap-y-5 overflow-y-auto bg-indigo-600 px-6 pb-4">
			<div class="flex h-16 shrink-0 items-center">
				<img
					class="h-8 w-auto"
					src="https://tailwindcss.com/plus-assets/img/logos/mark.svg?color=white"
					alt="Your Company"
				/>
			</div>
			<nav class="flex flex-1 flex-col">
				<ul role="list" class="flex flex-1 flex-col gap-y-7">
					<li>
						<ul role="list" class="-mx-2 space-y-1">
							{#each navigationItems as { href, label, icon }}
								<li>
									<a
										{href}
										class="group flex gap-x-3 rounded-md p-2 text-sm/6 font-semibold {isActive(href)
											? 'bg-indigo-700 text-white'
											: 'text-indigo-200 hover:bg-indigo-700 hover:text-white'}"
									>
										<svg
											class="size-6 shrink-0 {isActive(href)
												? 'text-white'
												: 'text-indigo-200 group-hover:text-white'}"
											fill="none"
											viewBox="0 0 24 24"
											stroke-width="1.5"
											stroke="currentColor"
											aria-hidden="true"
										>
											<path stroke-linecap="round" stroke-linejoin="round" d={icons[icon]} />
										</svg>
										{label}
									</a>
								</li>
							{/each}
						</ul>
					</li>
					<li class="mt-auto">
						<a
							href="/admin/settings"
							class="group -mx-2 flex gap-x-3 rounded-md p-2 text-sm/6 font-semibold text-indigo-200 hover:bg-indigo-700 hover:text-white"
						>
							<svg
								class="size-6 shrink-0 text-indigo-200 group-hover:text-white"
								fill="none"
								viewBox="0 0 24 24"
								stroke-width="1.5"
								stroke="currentColor"
								aria-hidden="true"
							>
								<path
									stroke-linecap="round"
									stroke-linejoin="round"
									d="M9.594 3.94c.09-.542.56-.94 1.11-.94h2.593c.55 0 1.02.398 1.11.94l.213 1.281c.063.374.313.686.645.87.074.04.147.083.22.127.325.196.72.257 1.075.124l1.217-.456a1.125 1.125 0 011.37.49l1.296 2.247a1.125 1.125 0 01-.26 1.431l-1.003.827c-.293.241-.438.613-.43.992a7.723 7.723 0 010 .255c-.008.378.137.75.43.991l1.004.827c.424.35.534.955.26 1.43l-1.298 2.247a1.125 1.125 0 01-1.369.491l-1.217-.456c-.355-.133-.75-.072-1.076.124a6.47 6.47 0 01-.22.128c-.331.183-.581.495-.644.869l-.213 1.281c-.09.543-.56.94-1.11.94h-2.594c-.55 0-1.019-.398-1.11-.94l-.213-1.281c-.062-.374-.312-.686-.644-.87a6.52 6.52 0 01-.22-.127c-.325-.196-.72-.257-1.076-.124l-1.217.456a1.125 1.125 0 01-1.369-.49l-1.297-2.247a1.125 1.125 0 01.26-1.431l1.004-.827c.292-.24.437-.613.43-.991a6.932 6.932 0 010-.255c.007-.38-.138-.751-.43-.992l-1.004-.827a1.125 1.125 0 01-.26-1.43l1.297-2.247a1.125 1.125 0 011.37-.491l1.216.456c.356.133.751.072 1.076-.124.072-.044.146-.086.22-.128.332-.183.582-.495.644-.869l.214-1.28z"
								/>
								<path
									stroke-linecap="round"
									stroke-linejoin="round"
									d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"
								/>
							</svg>
							Settings
						</a>
					</li>
				</ul>
			</nav>
		</div>
	</div>

	<!-- Main content -->
	<div class="lg:pl-72">
		<!-- Header -->
		<div
			class="shadow-xs sticky top-0 z-40 flex h-16 shrink-0 items-center gap-x-4 border-b border-gray-200 bg-white px-4 sm:gap-x-6 sm:px-6 lg:px-8"
		>
			<!-- Mobile menu button -->
			<button type="button" class="-m-2.5 p-2.5 text-gray-700 lg:hidden" onclick={toggleMobileMenu}>
				<span class="sr-only">Open sidebar</span>
				<svg
					class="size-6"
					fill="none"
					viewBox="0 0 24 24"
					stroke-width="1.5"
					stroke="currentColor"
					aria-hidden="true"
				>
					<path
						stroke-linecap="round"
						stroke-linejoin="round"
						d="M3.75 6.75h16.5M3.75 12h16.5m-16.5 5.25h16.5"
					/>
				</svg>
			</button>

			<!-- Separator -->
			<div class="h-6 w-px bg-gray-900/10 lg:hidden" aria-hidden="true"></div>

			<!-- Search and user profile section -->
			<div class="flex flex-1 gap-x-4 self-stretch lg:gap-x-6">
				<form class="grid flex-1 grid-cols-1" action="#" method="GET">
					<input
						type="search"
						name="search"
						aria-label="Search"
						class="outline-hidden col-start-1 row-start-1 block size-full bg-white pl-8 text-base text-gray-900 placeholder:text-gray-400 sm:text-sm/6"
						placeholder="Search"
					/>
					<svg
						class="pointer-events-none col-start-1 row-start-1 size-5 self-center text-gray-400"
						viewBox="0 0 20 20"
						fill="currentColor"
						aria-hidden="true"
					>
						<path
							fill-rule="evenodd"
							d="M9 3.5a5.5 5.5 0 100 11 5.5 5.5 0 000-11zM2 9a7 7 0 1112.452 4.391l3.328 3.329a.75.75 0 11-1.06 1.06l-3.329-3.328A7 7 0 012 9z"
							clip-rule="evenodd"
						/>
					</svg>
				</form>
				<div class="flex items-center gap-x-4 lg:gap-x-6">
					<!-- Notifications -->
					<button type="button" class="-m-2.5 p-2.5 text-gray-400 hover:text-gray-500">
						<span class="sr-only">View notifications</span>
						<svg
							class="size-6"
							fill="none"
							viewBox="0 0 24 24"
							stroke-width="1.5"
							stroke="currentColor"
							aria-hidden="true"
						>
							<path
								stroke-linecap="round"
								stroke-linejoin="round"
								d="M14.857 17.082a23.848 23.848 0 005.454-1.31A8.967 8.967 0 0118 9.75V9A6 6 0 006 9v.75a8.967 8.967 0 01-2.312 6.022c1.733.64 3.56 1.085 5.455 1.31m5.714 0a24.255 24.255 0 01-5.714 0m5.714 0a3 3 0 11-5.714 0"
							/>
						</svg>
					</button>

					<!-- Separator -->
					<div class="hidden lg:block lg:h-6 lg:w-px lg:bg-gray-900/10" aria-hidden="true"></div>

					<!-- User profile - simplified version without dropdown -->
					<div class="relative">
						<button type="button" class="-m-1.5 flex items-center p-1.5" id="user-menu-button">
							<span class="sr-only">Open user menu</span>
							<img
								class="size-8 rounded-full bg-gray-50"
								src="https://images.unsplash.com/photo-1472099645785-5658abf4ff4e?ixlib=rb-1.2.1&ixid=eyJhcHBfaWQiOjEyMDd9&auto=format&fit=facearea&facepad=2&w=256&h=256&q=80"
								alt=""
							/>
							<span class="hidden lg:flex lg:items-center">
								<span class="ml-4 text-sm/6 font-semibold text-gray-900" aria-hidden="true"
									>Admin User</span
								>
								<svg
									class="ml-2 size-5 text-gray-400"
									viewBox="0 0 20 20"
									fill="currentColor"
									aria-hidden="true"
								>
									<path
										fill-rule="evenodd"
										d="M5.22 8.22a.75.75 0 011.06 0L10 11.94l3.72-3.72a.75.75 0 111.06 1.06l-4.25 4.25a.75.75 0 01-1.06 0L5.22 9.28a.75.75 0 010-1.06z"
										clip-rule="evenodd"
									/>
								</svg>
							</span>
						</button>

						<!-- User dropdown menu would go here -->
					</div>
				</div>
			</div>
		</div>

		<!-- Page content -->
		<main class="py-10">
			<div class="px-4 sm:px-6 lg:px-8">
				{@render children()}
			</div>
		</main>
	</div>
</div>
