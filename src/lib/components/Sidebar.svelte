<script lang="ts">
  import type { Profile } from '$lib/types';
  
  let { people, activePersonFilter = $bindable(), isOpen = $bindable(false) } = $props<{
    people: Profile[];
    activePersonFilter: string;
    isOpen?: boolean;
  }>();
</script>

<!-- Mobile Overlay (only visible when isOpen on md:hidden) -->
{#if isOpen}
  <!-- svelte-ignore a11y_click_events_have_key_events -->
  <!-- svelte-ignore a11y_no_static_element_interactions -->
  <div class="fixed inset-0 bg-black/40 dark:bg-black/60 backdrop-blur-sm z-30 md:hidden" onclick={() => isOpen = false}></div>
{/if}

<aside class="
  fixed md:relative top-0 right-0 h-full z-40 
  w-64 md:w-48 
  bg-white/90 dark:bg-[#1c1c26]/90 backdrop-blur-xl md:bg-white/40 md:dark:bg-black/20 md:backdrop-blur-sm 
  border-l border-gray-200/50 dark:border-white/10 
  flex flex-col shrink-0 transition-transform duration-300
  {isOpen ? 'translate-x-0' : 'translate-x-full md:translate-x-0'}
">
  <div class="h-14 md:h-11 px-4 border-b border-gray-200/50 dark:border-white/10 flex items-center justify-between shrink-0">
    <span class="text-[10px] md:text-[10px] font-bold tracking-widest uppercase text-gray-500 dark:text-gray-400">Team Directory</span>
    <button class="md:hidden p-2 text-gray-400 hover:text-gray-600 dark:hover:text-gray-200" onclick={() => isOpen = false}>
      <svg class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"/></svg>
    </button>
  </div>

  <div class="flex-1 overflow-y-auto p-2 space-y-1 scrollbar-hide">
    <button 
      class="w-full text-left px-3 py-2 rounded-lg text-sm font-medium transition-colors {activePersonFilter === '' ? 'bg-white dark:bg-white/10 text-gray-900 dark:text-white shadow-sm' : 'text-gray-600 dark:text-gray-400 hover:bg-white/50 dark:hover:bg-white/5'}"
      onclick={() => { activePersonFilter = ''; isOpen = false; }}
    >
      Everyone
    </button>
    {#each people as person}
      <button 
        class="w-full flex items-center gap-2 px-3 py-2 rounded-lg text-sm font-medium transition-colors {activePersonFilter === person.id ? 'bg-white dark:bg-white/10 text-gray-900 dark:text-white shadow-sm' : 'text-gray-600 dark:text-gray-400 hover:bg-white/50 dark:hover:bg-white/5'}"
        onclick={() => { activePersonFilter = person.id; isOpen = false; }}
      >
        <div class="w-6 h-6 md:w-5 md:h-5 rounded-full flex items-center justify-center text-[10px] text-white shrink-0 shadow-sm" style="background-color: hsl({person.hue}, 65%, 55%)">
          {person.name.charAt(0).toUpperCase()}
        </div>
        <span class="truncate">{person.name}</span>
      </button>
    {/each}
  </div>
</aside>
