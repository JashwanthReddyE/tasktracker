<script lang="ts">
  import { enhance } from '$app/forms';
  import type { Task, Category, Profile } from '$lib/types';
  
  let { 
    task = null,
    isOpen = $bindable(), 
    status = 'todo', 
    categoryId = '',
    categories,
    people
  } = $props<{
    task?: Task | null;
    isOpen: boolean;
    status: string;
    categoryId: string;
    categories: Category[];
    people: Profile[];
  }>();

  let title = $state('');
  let notes = $state('');
  let priority = $state('medium');
  let due_date = $state('');
  let selectedPeople = $state<string[]>([]);
  let isConfirmingDelete = $state(false);
  
  $effect(() => {
    if (isOpen) {
      title = task?.title || '';
      notes = task?.notes || '';
      priority = task?.priority || 'medium';
      due_date = task?.due_date || '';
      selectedPeople = task?.task_assignments?.map(ta => ta.user_id) || [];
      isConfirmingDelete = false;
    }
  });

  function togglePerson(id: string) {
    if (selectedPeople.includes(id)) {
      selectedPeople = selectedPeople.filter(p => p !== id);
    } else {
      selectedPeople = [...selectedPeople, id];
    }
  }

  function close() {
    isOpen = false;
    selectedPeople = [];
    isConfirmingDelete = false;
  }
</script>

{#if isOpen}
  <div class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/50 backdrop-blur-sm">
    <div class="bg-white dark:bg-[#1c1c26] rounded-xl shadow-2xl w-full max-w-md overflow-hidden border border-gray-200 dark:border-white/10">
      <div class="px-6 py-4 border-b border-gray-100 dark:border-white/10 flex items-center justify-between">
        <h2 class="text-lg font-bold text-gray-900 dark:text-white">{task ? 'Edit Task' : 'Add New Task'}</h2>
        <button class="text-gray-400 hover:text-gray-600 dark:hover:text-gray-200 transition-colors" onclick={close}>✕</button>
      </div>
      
      <form method="POST" action={task ? "?/updateTask" : "?/createTask"} use:enhance={() => {
        return async ({ result, update }) => {
          if (result.type === 'success') {
            close();
          } else if (result.type === 'failure') {
            console.error('Failed to save task:', result.data);
            alert('Failed to save task: ' + (result.data?.error || 'Unknown error'));
          }
          await update();
        };
      }} class="p-6 flex flex-col gap-4">
        
        {#if task}
          <input type="hidden" name="id" value={task.id} />
        {/if}
        <input type="hidden" name="status" value={status} />
        <input type="hidden" name="category_id" value={categoryId} />
        <input type="hidden" name="people_ids" value={JSON.stringify(selectedPeople)} />
        
        <div>
          <label for="title" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Title</label>
          <input type="text" id="title" name="title" bind:value={title} required class="w-full rounded-lg border border-gray-300 dark:border-gray-600 bg-transparent px-3 py-2 text-sm focus:ring-2 focus:ring-blue-500 outline-none transition-all dark:text-white" placeholder="What needs to be done?" />
        </div>
        
        <div>
          <label for="notes" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Notes</label>
          <textarea id="notes" name="notes" bind:value={notes} rows="3" class="w-full rounded-lg border border-gray-300 dark:border-gray-600 bg-transparent px-3 py-2 text-sm focus:ring-2 focus:ring-blue-500 outline-none transition-all dark:text-white" placeholder="Additional details..."></textarea>
        </div>

        <div class="grid grid-cols-2 gap-4">
          <div>
            <label for="priority" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Priority</label>
            <select id="priority" name="priority" bind:value={priority} class="w-full rounded-lg border border-gray-300 dark:border-gray-600 bg-transparent px-3 py-2 text-sm focus:ring-2 focus:ring-blue-500 outline-none transition-all dark:text-white">
              <option value="low" class="text-black">Low</option>
              <option value="medium" class="text-black">Medium</option>
              <option value="high" class="text-black">High</option>
            </select>
          </div>
          
          <div>
            <label for="due_date" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Due Date</label>
            <input type="date" id="due_date" name="due_date" bind:value={due_date} class="w-full rounded-lg border border-gray-300 dark:border-gray-600 bg-transparent px-3 py-2 text-sm focus:ring-2 focus:ring-blue-500 outline-none transition-all dark:text-white [color-scheme:light] dark:[color-scheme:dark]" />
          </div>
        </div>

        {#if people.length > 0}
          <div>
            <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">Assign People</label>
            <div class="flex flex-wrap gap-2 max-h-32 overflow-y-auto p-1">
              {#each people as person}
                <button 
                  type="button"
                  class="flex items-center gap-2 px-2 py-1 rounded-full border text-xs font-medium transition-colors {selectedPeople.includes(person.id) ? 'border-blue-500 bg-blue-50 dark:bg-blue-900/30 text-blue-700 dark:text-blue-300' : 'border-gray-200 dark:border-gray-700 hover:border-gray-300 dark:hover:border-gray-500 text-gray-600 dark:text-gray-400'}"
                  onclick={() => togglePerson(person.id)}
                >
                  <div class="w-4 h-4 rounded-full flex items-center justify-center text-[8px] text-white" style="background-color: hsl({person.hue}, 65%, 55%)">
                    {person.name.charAt(0).toUpperCase()}
                  </div>
                  {person.name}
                </button>
              {/each}
            </div>
          </div>
        {/if}

        <div class="mt-4 flex items-center {task ? 'justify-between' : 'justify-end'} gap-3">
          {#if task}
            <div class="flex items-center gap-2">
              {#if isConfirmingDelete}
                <button type="button" class="px-3 py-1.5 text-xs font-bold text-white bg-red-600 hover:bg-red-700 rounded-lg shadow-sm transition-all animate-in fade-in slide-in-from-left-2" onclick={() => {
                  const form = document.createElement('form');
                  form.method = 'POST';
                  form.action = '?/deleteTask';
                  const input = document.createElement('input');
                  input.type = 'hidden';
                  input.name = 'id';
                  input.value = task!.id;
                  form.appendChild(input);
                  document.body.appendChild(form);
                  form.submit();
                }}>Confirm Delete</button>
                <button type="button" class="px-3 py-1.5 text-xs font-medium text-gray-500 hover:text-gray-700 dark:hover:text-gray-300 transition-colors animate-in fade-in" onclick={() => isConfirmingDelete = false}>Cancel</button>
              {#else}
                <button type="button" class="px-4 py-2 text-sm font-bold text-red-500 hover:text-white hover:bg-red-500 rounded-lg transition-all border border-red-500 hover:border-transparent shadow-sm" onclick={() => isConfirmingDelete = true}>Delete</button>
              {/if}
            </div>
          {/if}
          <div class="flex gap-3">
            <button type="button" class="px-4 py-2 text-sm font-medium text-gray-600 dark:text-gray-400 hover:text-gray-900 dark:hover:text-white transition-colors" onclick={close}>Cancel</button>
            <button type="submit" class="px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white text-sm font-medium rounded-lg shadow-sm transition-colors">{task ? 'Save Changes' : 'Add Task'}</button>
          </div>
        </div>
      </form>
    </div>
  </div>
{/if}
