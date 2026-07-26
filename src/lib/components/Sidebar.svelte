<script lang="ts">
  import type { Profile } from '$lib/types';
  
  let { people, activePersonFilter = $bindable() } = $props<{
    people: Profile[];
    activePersonFilter: string;
  }>();
</script>

<aside class="w-full md:w-48 h-auto md:h-full bg-white/40 dark:bg-black/20 backdrop-blur-sm border-t md:border-t-0 md:border-l border-gray-200/50 dark:border-white/10 flex flex-col shrink-0 transition-colors duration-300 z-20">
  <div class="hidden md:flex h-11 px-4 border-b border-gray-200/50 dark:border-white/10 flex-col justify-center shrink-0">
    <span class="text-[10px] font-bold tracking-widest uppercase text-gray-500 dark:text-gray-400">Team Directory</span>
  </div>

  <div class="flex flex-row md:flex-col overflow-x-auto md:overflow-y-auto p-2 gap-2 md:gap-0 md:space-y-1 scrollbar-hide shrink-0">
    <button 
      class="shrink-0 whitespace-nowrap md:w-full text-center md:text-left px-4 md:px-3 py-1.5 md:py-2 rounded-full md:rounded-lg text-xs md:text-sm font-medium transition-colors {activePersonFilter === '' ? 'bg-white dark:bg-white/10 text-gray-900 dark:text-white shadow-sm border border-gray-200 dark:border-white/10 md:border-transparent' : 'text-gray-600 dark:text-gray-400 hover:bg-white/50 dark:hover:bg-white/5 border border-transparent'}"
      onclick={() => activePersonFilter = ''}
    >
      Everyone
    </button>
    {#each people as person}
      <button 
        class="shrink-0 whitespace-nowrap md:w-full flex items-center gap-2 px-3 py-1.5 md:py-2 rounded-full md:rounded-lg text-xs md:text-sm font-medium transition-colors {activePersonFilter === person.id ? 'bg-white dark:bg-white/10 text-gray-900 dark:text-white shadow-sm border border-gray-200 dark:border-white/10 md:border-transparent' : 'text-gray-600 dark:text-gray-400 hover:bg-white/50 dark:hover:bg-white/5 border border-transparent'}"
        onclick={() => activePersonFilter = person.id}
      >
        <div class="w-4 h-4 md:w-5 md:h-5 rounded-full flex items-center justify-center text-[9px] md:text-[10px] text-white shrink-0 shadow-sm" style="background-color: hsl({person.hue}, 65%, 55%)">
          {person.name.charAt(0).toUpperCase()}
        </div>
        <span class="truncate hidden md:block">{person.name}</span>
        <span class="truncate md:hidden">{person.name.split(' ')[0]}</span>
      </button>
    {/each}
  </div>
</aside>
