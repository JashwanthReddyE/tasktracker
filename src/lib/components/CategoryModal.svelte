<script lang="ts">
  import { enhance } from '$app/forms';
  import type { Category } from '$lib/types';
  
  let { isOpen = $bindable(), categories } = $props<{
    isOpen: boolean;
    categories: Category[];
  }>();

  let tempCategories = $state<Category[]>([]);
  
  $effect(() => {
    if (isOpen) {
      tempCategories = JSON.parse(JSON.stringify(categories));
    }
  });

  function addCategory() {
    tempCategories = [...tempCategories, { id: crypto.randomUUID(), label: 'New Category', position: tempCategories.length }];
  }

  function removeCategory(index: number) {
    tempCategories = tempCategories.filter((_, i) => i !== index);
  }

  function close() {
    isOpen = false;
  }
</script>

{#if isOpen}
  <div class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/50 backdrop-blur-sm">
    <div class="bg-white dark:bg-[#1c1c26] rounded-xl shadow-2xl w-full max-w-md overflow-hidden border border-gray-200 dark:border-white/10">
      <div class="px-6 py-4 border-b border-gray-100 dark:border-white/10 flex items-center justify-between">
        <h2 class="text-lg font-bold text-gray-900 dark:text-white">Manage Categories</h2>
        <button class="text-gray-400 hover:text-gray-600 dark:hover:text-gray-200 transition-colors" onclick={close}>✕</button>
      </div>
      
      <form method="POST" action="?/replaceCategories" use:enhance={({ formData, cancel }) => {
        const labels = tempCategories.map((c: any) => c.label.toLowerCase().trim());
        const uniqueLabels = new Set(labels);
        if (labels.length !== uniqueLabels.size) {
          alert('Category names must be unique.');
          cancel();
          return;
        }
        
        formData.set('categories', JSON.stringify(tempCategories));
        return async ({ result, update }) => {
          if (result.type === 'success') {
            close();
          } else if (result.type === 'failure') {
            alert('Failed to save categories: ' + (result.data?.error || 'Unknown error'));
          }
          await update();
        };
      }} class="p-6 flex flex-col gap-4">
        
        <input type="hidden" name="categories" value={JSON.stringify(tempCategories)} />
        
        <div class="flex flex-col gap-2 max-h-60 overflow-y-auto pr-2">
          {#each tempCategories as cat, i}
            <div class="flex items-center gap-2">
              <input type="text" bind:value={cat.label} required class="flex-1 rounded-lg border border-gray-300 dark:border-gray-600 bg-transparent px-3 py-1.5 text-sm focus:ring-2 focus:ring-blue-500 outline-none transition-all dark:text-white" />
              <button type="button" class="p-1.5 text-red-500 hover:bg-red-50 dark:hover:bg-red-900/20 rounded-md transition-colors" onclick={() => removeCategory(i)}>
                <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"></path></svg>
              </button>
            </div>
          {/each}
          {#if tempCategories.length === 0}
            <p class="text-sm text-gray-500 text-center py-4">No categories left. Add one!</p>
          {/if}
        </div>
        
        <button type="button" class="w-full py-2 border border-dashed border-gray-300 dark:border-gray-600 rounded-lg text-sm font-medium text-gray-600 dark:text-gray-400 hover:bg-gray-50 dark:hover:bg-white/5 transition-colors" onclick={addCategory}>
          + Add Category
        </button>

        <div class="mt-4 flex justify-end gap-3 pt-4 border-t border-gray-100 dark:border-white/10">
          <button type="button" class="px-4 py-2 text-sm font-medium text-gray-600 dark:text-gray-400 hover:text-gray-900 dark:hover:text-white transition-colors" onclick={close}>Cancel</button>
          <button type="submit" class="px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white text-sm font-medium rounded-lg shadow-sm transition-colors">Save Changes</button>
        </div>
      </form>
    </div>
  </div>
{/if}
