<script lang="ts">
  import type { Category } from '$lib/types';
  
  let { categories, activeCategoryId = $bindable() } = $props<{
    categories: Category[];
    activeCategoryId: string;
  }>();

  function toggleTheme() {
    document.documentElement.classList.toggle('dark');
  }
</script>

<header class="h-14 bg-white/60 dark:bg-black/40 backdrop-blur-md border-b border-gray-200/50 dark:border-white/10 flex items-center px-6 gap-4 z-10 sticky top-0 shadow-sm">
  <div class="flex items-baseline shrink-0">
    <h1 class="text-sm font-black tracking-widest uppercase text-transparent bg-clip-text bg-gradient-to-r from-blue-500 to-purple-600">Tasks</h1>
    <span class="ml-2 text-xs font-medium text-gray-500 dark:text-gray-400">Tracker</span>
  </div>

  <div class="w-px h-5 bg-gray-300 dark:bg-gray-700 mx-2"></div>

  <div class="flex items-center gap-2 overflow-x-auto flex-1 scrollbar-hide">
    {#each categories as cat}
      <button 
        class="px-3 py-1.5 rounded-lg text-xs font-semibold tracking-wide transition-all duration-200 {activeCategoryId === cat.id ? 'bg-blue-100 text-blue-700 dark:bg-blue-500/20 dark:text-blue-300 shadow-sm' : 'text-gray-500 hover:text-gray-900 hover:bg-gray-100 dark:text-gray-400 dark:hover:text-gray-100 dark:hover:bg-white/5'}"
        onclick={() => activeCategoryId = cat.id}
      >
        {cat.label}
      </button>
    {/each}
    <button class="w-7 h-7 rounded-lg border border-dashed border-gray-300 dark:border-gray-600 text-gray-400 flex items-center justify-center hover:border-gray-500 hover:text-gray-600 dark:hover:border-gray-400 dark:hover:text-gray-300 transition-colors" aria-label="Add Category">
      +
    </button>
  </div>

  <div class="flex items-center gap-3 shrink-0 ml-auto">
    <button onclick={toggleTheme} class="p-2 rounded-lg bg-gray-100 dark:bg-white/5 text-gray-600 dark:text-gray-300 hover:bg-gray-200 dark:hover:bg-white/10 transition-colors" title="Toggle Theme">
      <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M20.354 15.354A9 9 0 018.646 3.646 9.003 9.003 0 0012 21a9.003 9.003 0 008.354-5.646z" />
      </svg>
    </button>
    <form action="/logout" method="POST">
      <button class="px-3 py-1.5 rounded-lg border border-red-200 dark:border-red-900/50 text-red-600 dark:text-red-400 text-xs font-semibold hover:bg-red-50 dark:hover:bg-red-900/20 transition-colors">
        Logout
      </button>
    </form>
  </div>
</header>
