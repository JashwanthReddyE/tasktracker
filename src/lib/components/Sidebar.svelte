<script lang="ts">
  import type { Person } from '$lib/types';
  
  let { people, activePersonFilter = $bindable(), categoryId, onAddPerson } = $props<{
    people: Person[];
    activePersonFilter: string;
    categoryId: string;
    onAddPerson: () => void;
  }>();

  let categoryPeople = $derived(people.filter(p => p.category_id === categoryId));
</script>

<aside class="w-48 bg-white/40 dark:bg-black/20 backdrop-blur-sm border-r border-gray-200/50 dark:border-white/10 flex flex-col shrink-0 transition-colors duration-300">
  <div class="h-11 px-4 border-b border-gray-200/50 dark:border-white/10 flex flex-col justify-center shrink-0">
    <span class="text-[10px] font-bold tracking-widest uppercase text-gray-500 dark:text-gray-400">People</span>
  </div>

  <div class="flex-1 overflow-y-auto p-2 grid grid-cols-2 gap-1 content-start scrollbar-hide">
    <button 
      class="col-span-2 flex items-center justify-center py-2 text-xs text-gray-500 dark:text-gray-400 hover:text-gray-900 dark:hover:text-gray-100 transition-all rounded-lg {activePersonFilter === '' ? 'bg-gray-200/70 dark:bg-white/10 font-semibold shadow-sm' : ''}"
      onclick={() => activePersonFilter = ''}
    >
      Everyone
    </button>
    
    {#each categoryPeople as person}
      <button 
        class="flex flex-col items-center gap-1.5 p-2 rounded-lg border border-transparent transition-all {activePersonFilter === person.id ? 'bg-white dark:bg-[#1a1a24] shadow-md border-gray-200 dark:border-gray-700 scale-[1.02]' : 'hover:bg-white/60 dark:hover:bg-white/5'}"
        onclick={() => activePersonFilter = activePersonFilter === person.id ? '' : person.id}
      >
        <div class="w-8 h-8 rounded-full flex items-center justify-center text-[11px] font-bold text-white shadow-sm ring-2 ring-white/50 dark:ring-black/50" style="background-color: hsl({person.hue}, 65%, 55%)">
          {person.name.charAt(0).toUpperCase()}
        </div>
        <span class="text-[10px] truncate w-full text-center text-gray-700 dark:text-gray-300 {activePersonFilter === person.id ? 'font-semibold' : ''}">{person.name}</span>
      </button>
    {/each}
  </div>

  <div class="p-2 border-t border-gray-200/50 dark:border-white/10 shrink-0">
    <button 
      class="w-full py-1.5 rounded-lg border border-dashed border-gray-300 dark:border-gray-600 text-gray-400 text-[11px] font-medium hover:border-gray-500 hover:text-gray-600 dark:hover:border-gray-400 dark:hover:text-gray-300 transition-colors"
      onclick={onAddPerson}
    >
      + Add Person
    </button>
  </div>
</aside>
