<script lang="ts">
  import type { Profile } from '$lib/types';
  
  let { people, activePersonFilter = $bindable() } = $props<{
    people: Profile[];
    activePersonFilter: string;
  }>();
</script>

<aside class="w-48 bg-white/40 dark:bg-black/20 backdrop-blur-sm border-l border-gray-200/50 dark:border-white/10 flex flex-col shrink-0 transition-colors duration-300">
  <div class="h-11 px-4 border-b border-gray-200/50 dark:border-white/10 flex flex-col justify-center shrink-0">
    <span class="text-[10px] font-bold tracking-widest uppercase text-gray-500 dark:text-gray-400">Team Directory</span>
  </div>

  <div class="flex-1 overflow-y-auto p-2 space-y-1 scrollbar-hide">
    <button 
      class="w-full text-left px-3 py-2 rounded-lg text-sm font-medium transition-colors {activePersonFilter === '' ? 'bg-white dark:bg-white/10 text-gray-900 dark:text-white shadow-sm' : 'text-gray-600 dark:text-gray-400 hover:bg-white/50 dark:hover:bg-white/5'}"
      onclick={() => activePersonFilter = ''}
    >
      Everyone
    </button>
    {#each people as person}
      <button 
        class="w-full flex items-center gap-2 px-3 py-2 rounded-lg text-sm font-medium transition-colors {activePersonFilter === person.id ? 'bg-white dark:bg-white/10 text-gray-900 dark:text-white shadow-sm' : 'text-gray-600 dark:text-gray-400 hover:bg-white/50 dark:hover:bg-white/5'}"
        onclick={() => activePersonFilter = person.id}
      >
        <div class="w-5 h-5 rounded-full flex items-center justify-center text-[10px] text-white shrink-0 shadow-sm" style="background-color: hsl({person.hue}, 65%, 55%)">
          {person.name.charAt(0).toUpperCase()}
        </div>
        <span class="truncate">{person.name}</span>
      </button>
    {/each}
  </div>
</aside>
