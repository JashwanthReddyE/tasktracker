<script lang="ts">
  import type { Task } from '$lib/types';

  let { task } = $props<{ task: Task }>();

  function handleDragStart(e: DragEvent) {
    if (e.dataTransfer) {
      e.dataTransfer.setData('text/plain', task.id);
      e.dataTransfer.effectAllowed = 'move';
      
      setTimeout(() => {
        if (e.target instanceof HTMLElement) {
          e.target.classList.add('opacity-40', 'scale-95');
        }
      }, 0);
    }
  }

  function handleDragEnd(e: DragEvent) {
    if (e.target instanceof HTMLElement) {
      e.target.classList.remove('opacity-40', 'scale-95');
    }
  }
</script>

<div
  draggable="true"
  ondragstart={handleDragStart}
  ondragend={handleDragEnd}
  class="group bg-white dark:bg-[#1c1c26] rounded-xl p-3.5 border border-gray-200/60 dark:border-white/5 shadow-sm hover:shadow-md hover:border-gray-300 dark:hover:border-white/10 transition-all duration-200 cursor-grab active:cursor-grabbing relative overflow-hidden flex flex-col gap-1.5"
>
  <!-- Priority Indicator -->
  <div class="absolute top-0 left-0 w-1 h-full {task.priority === 'high' ? 'bg-red-500' : task.priority === 'low' ? 'bg-green-500' : 'bg-orange-400'}"></div>
  
  <h3 class="text-[13px] font-semibold text-gray-800 dark:text-gray-100 leading-snug pr-4">{task.title}</h3>
  
  {#if task.notes}
    <p class="text-[11px] text-gray-500 dark:text-gray-400 line-clamp-2 leading-relaxed">{task.notes}</p>
  {/if}
  
  <div class="flex items-center gap-2 mt-1.5 pt-2 border-t border-gray-100 dark:border-white/5">
    {#if task.due_date}
      <span class="text-[9px] font-bold tracking-wide uppercase px-1.5 py-0.5 rounded bg-gray-100 dark:bg-white/5 text-gray-500 dark:text-gray-400">
        {task.due_date}
      </span>
    {/if}
    
    <div class="ml-auto flex -space-x-1.5">
      {#each task.task_people as tp}
        <div class="w-5 h-5 rounded-full bg-blue-500 border border-white dark:border-[#1c1c26] shadow-sm flex items-center justify-center text-[8px] font-bold text-white z-0">
          U
        </div>
      {/each}
    </div>
  </div>
</div>
