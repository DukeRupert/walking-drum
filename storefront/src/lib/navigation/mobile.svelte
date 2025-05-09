<script lang="ts">
    let { isMobileOpen = $bindable() } = $props();
</script>

<!-- Off-canvas menu for mobile, show/hide based on off-canvas menu state. -->
<div class="relative z-40 lg:hidden" role="dialog" aria-modal="true" style="visibility: {isMobileOpen ? 'visible' : 'hidden'}">
    <div 
        class="fixed inset-0 bg-black/25 transition-opacity ease-linear duration-300" 
        style="opacity: {isMobileOpen ? '1' : '0'}"
        aria-hidden="true"
    ></div>

    <div class="fixed inset-0 z-40 flex">
        <!-- Off-canvas menu, show/hide based on off-canvas menu state. -->
        <div 
            class="relative flex w-full max-w-xs flex-col overflow-y-auto bg-white pb-12 shadow-xl transition ease-in-out duration-300 transform"
            style="transform: {isMobileOpen ? 'translateX(0)' : 'translateX(-100%)'};"
        >
            <div class="flex px-4 pb-2 pt-5">
                <button
                    type="button"
                    onclick={() => isMobileOpen = false}
                    class="-m-2 inline-flex items-center justify-center rounded-md p-2 text-gray-400"
                >
                    <span class="sr-only">Close menu</span>
                    <svg
                        class="size-6"
                        fill="none"
                        viewBox="0 0 24 24"
                        stroke-width="1.5"
                        stroke="currentColor"
                        aria-hidden="true"
                        data-slot="icon"
                    >
                        <path stroke-linecap="round" stroke-linejoin="round" d="M6 18 18 6M6 6l12 12" />
                    </svg>
                </button>
            </div>

            <div class="space-y-6 border-t border-gray-200 px-4 py-6">
                <div class="flow-root">
                    <a href="#" class="-m-2 block p-2 font-medium text-gray-900">Coffee</a>
                </div>
                <div class="flow-root">
                    <a href="#" class="-m-2 block p-2 font-medium text-gray-900">About us</a>
                </div>
            </div>

            <div class="space-y-6 border-t border-gray-200 px-4 py-6">
                <div class="flow-root">
                    <a href="#" class="-m-2 block p-2 font-medium text-gray-900">Create an account</a>
                </div>
                <div class="flow-root">
                    <a href="#" class="-m-2 block p-2 font-medium text-gray-900">Sign in</a>
                </div>
            </div>
        </div>
    </div>
</div>