<script lang="ts">
  import type { Task } from '$lib/types';
  import Card from './Card.svelte';

  let { filteredTasks, onMove, onAddTask } = $props<{
    filteredTasks: Task[];
    onMove: (taskId: string, newStatus: string) => void;
    onAddTask: (status: string) => void;
  }>();

  const columns = [
    { id: 'todo', label: 'To Do' },
    { id: 'working', label: 'In Progress' },
    { id: 'done', label: 'Completed' }
  ];

  function handleDragOver(e: DragEvent) {
    e.preventDefault(); // necessary to allow dropping
    if (e.dataTransfer) {
      e.dataTransfer.dropEffect = 'move';
    }
    if (e.currentTarget instanceof HTMLElement) {
      e.currentTarget.classList.add('bg-black/5', 'dark:bg-white/5');
    }
  }

  function handleDragLeave(e: DragEvent) {
    if (e.currentTarget instanceof HTMLElement) {
      e.currentTarget.classList.remove('bg-black/5', 'dark:bg-white/5');
    }
  }

  function handleDrop(e: DragEvent, status: string) {
    e.preventDefault();
    if (e.currentTarget instanceof HTMLElement) {
      e.currentTarget.classList.remove('bg-black/5', 'dark:bg-white/5');
    }
    const taskId = e.dataTransfer?.getData('text/plain');
    if (taskId) {
      onMove(taskId, status);
    }
  }
</script>

<main class="flex-1 overflow-x-auto flex gap-6 p-6 scrollbar-hide">
  {#each columns as col, i}
    <div class="flex-1 min-w-[280px] flex flex-col gap-4 relative">
      <!-- Column Header -->
      <div class="flex items-center justify-between sticky top-0 z-10 px-1">
        <div class="flex items-center gap-2">
          <div class="w-2 h-2 rounded-full {col.id === 'todo' ? 'bg-orange-400' : col.id === 'working' ? 'bg-blue-500' : 'bg-green-500'} shadow-sm"></div>
          <h2 class="text-xs font-bold uppercase tracking-widest text-gray-700 dark:text-gray-300">{col.label}</h2>
          <span class="ml-1 text-[10px] font-bold px-1.5 py-0.5 rounded-md bg-white/60 dark:bg-white/10 text-gray-500 dark:text-gray-400 border border-gray-200/50 dark:border-white/5 shadow-sm">
            {filteredTasks.filter(t => t.status === col.id).length}
          </span>
        </div>
        <button 
          class="w-6 h-6 rounded-md flex items-center justify-center text-gray-400 hover:bg-white hover:shadow-sm dark:hover:bg-white/10 hover:text-gray-700 dark:hover:text-gray-200 transition-all"
          onclick={() => onAddTask(col.id)}
        >
          +
        </button>
      </div>

      <!-- Column Body (Droppable Area) -->
      <div
        class="flex-1 flex flex-col gap-3 rounded-2xl transition-colors duration-200 pb-10"
        ondragover={handleDragOver}
        ondragleave={handleDragLeave}
        ondrop={(e) => handleDrop(e, col.id)}
        role="region"
        aria-label="{col.label} column"
      >
        {#each filteredTasks.filter(t => t.status === col.id) as task (task.id)}
          <Card {task} />
        {/each}
        
        {#if filteredTasks.filter(t => t.status === col.id).length === 0}
          <div class="h-24 rounded-xl border-2 border-dashed border-gray-300/60 dark:border-white/10 flex items-center justify-center bg-white/20 dark:bg-black/10">
            <span class="text-xs text-gray-400 dark:text-gray-500 font-medium">Drop tasks here</span>
          </div>
        {/if}
      </div>
    </div>
    
    {#if i < columns.length - 1}
      <div class="w-0 mt-8 mb-4 shrink-0 border-r-2 border-dashed border-gray-200 dark:border-gray-700 hidden md:block"></div>
    {/if}
  {/each}
</main>
