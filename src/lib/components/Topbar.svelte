<script lang="ts">
  import type { Category, Profile } from '$lib/types';
  
  let { categories, activeCategoryId = $bindable(), profile, onAddCategory, onToggleTeam } = $props<{
    categories: Category[];
    activeCategoryId: string;
    profile: Profile;
    onAddCategory: () => void;
    onToggleTeam: () => void;
  }>();

  function toggleTheme() {
    document.documentElement.classList.toggle('dark');
  }
</script>

<header class="h-14 bg-white dark:bg-[#0f0f13] border-b border-gray-200 dark:border-white/10 flex items-center px-4 md:px-6 gap-3 md:gap-4 z-10 sticky top-0 shadow-sm">
  <div class="flex items-baseline shrink-0">
    <h1 class="text-sm md:text-base font-black tracking-widest uppercase text-transparent bg-clip-text bg-gradient-to-r from-blue-500 to-purple-600">Tasks</h1>
    <span class="hidden sm:inline ml-2 text-xs font-medium text-gray-500 dark:text-gray-400">Tracker</span>
  </div>

  <div class="w-px h-5 bg-gray-300 dark:bg-gray-700 mx-1 md:mx-2 shrink-0"></div>

  <div class="flex items-center gap-2 overflow-x-auto flex-1 scrollbar-hide">
    {#each categories as cat}
      <button 
        class="px-3 py-1.5 rounded-lg text-xs font-semibold tracking-wide transition-all duration-200 {activeCategoryId === cat.id ? 'bg-blue-100 text-blue-700 dark:bg-blue-500/20 dark:text-blue-300 shadow-sm' : 'text-gray-500 hover:text-gray-900 hover:bg-gray-100 dark:text-gray-400 dark:hover:text-gray-100 dark:hover:bg-white/5'}"
        onclick={() => activeCategoryId = cat.id}
      >
        {cat.label}
      </button>
    {/each}
    <button 
      class="w-7 h-7 rounded-lg border border-dashed border-gray-300 dark:border-gray-600 text-gray-400 flex items-center justify-center hover:border-gray-500 hover:text-gray-600 dark:hover:border-gray-400 dark:hover:text-gray-300 transition-colors" 
      aria-label="Add Category"
      onclick={onAddCategory}
    >
      +
    </button>
  </div>

  <div class="flex items-center gap-2 md:gap-3 shrink-0 ml-auto">
    <button onclick={onToggleTeam} class="md:hidden p-2 rounded-lg bg-blue-50 dark:bg-blue-900/20 text-blue-600 dark:text-blue-400 hover:bg-blue-100 dark:hover:bg-blue-900/40 transition-colors" title="Team Directory">
      <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197M13 7a4 4 0 11-8 0 4 4 0 018 0z" />
      </svg>
    </button>
    {#if profile?.role === 'admin'}
      <a href="/admin" class="flex px-3 py-1.5 rounded-lg border border-purple-200 dark:border-purple-900/50 text-purple-600 dark:text-purple-400 text-xs font-semibold hover:bg-purple-50 dark:hover:bg-purple-900/20 transition-colors">
        Manage Users
      </a>
    {/if}
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
