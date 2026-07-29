<script lang="ts">
  import { onMount } from 'svelte';
  import { invalidateAll } from '$app/navigation';
  import type { PageData } from './$types';
  import type { Task, Category, Profile } from '$lib/types';
  import Board from '$lib/components/Board.svelte';
  import Topbar from '$lib/components/Topbar.svelte';
  import Sidebar from '$lib/components/Sidebar.svelte';
  import TaskModal from '$lib/components/TaskModal.svelte';
  import CategoryModal from '$lib/components/CategoryModal.svelte';

  let { data } = $props<{ data: PageData }>();

  let supabase = $derived(data.supabase);
  let profile = $derived(data.profile as Profile);

  let my_teams = $derived(data.my_teams || []);
  
  // Use Svelte 5 state
  let tasks = $state(data.tasks as Task[]);
  let categories = $state(data.categories as Category[]);
  let people = $state(data.people as Profile[]);

  $effect(() => {
    tasks = data.tasks as Task[];
    categories = data.categories as Category[];
    people = data.people as Profile[];
  });
  
  let activeCategoryId = $state(categories.length > 0 ? categories[0].id : '');
  let activePersonFilter = $state('');

  // Modal states
  let isTaskModalOpen = $state(false);
  let newTaskStatus = $state('todo');
  let isCategoryModalOpen = $state(false);
  let isMobileSidebarOpen = $state(false);
  let editingTask = $state<Task | null>(null);

  function openTaskModal(status: string) {
    editingTask = null;
    newTaskStatus = status;
    isTaskModalOpen = true;
  }

  function handleTaskClick(task: Task) {
    editingTask = task;
    newTaskStatus = task.status;
    isTaskModalOpen = true;
  }

  // Derived state for filtering
  let filteredTasks = $derived(
    tasks.filter(t => 
      (activeCategoryId === '' || t.category_id === activeCategoryId) &&
      (activePersonFilter === '' || t.task_assignments?.some(a => a.user_id === activePersonFilter))
    )
  );

  async function handleTaskMove(taskId: string, newStatus: string) {
    const task = tasks.find(t => t.id === taskId);
    if (!task || task.status === newStatus) return;

    // Optimistic update
    task.status = newStatus;
    
    // Server update
    const formData = new FormData();
    formData.append('id', taskId);
    formData.append('status', newStatus);
    formData.append('position', task.position.toString());
    
    const res = await fetch('?/updateTask', { method: 'POST', body: formData });
    if (!res.ok) {
      // Rollback on failure (simplified)
      console.error('Failed to move task');
    }
  }

  onMount(() => {
    // Listen to real-time changes on tasks, categories, and people tables
    const channel = supabase.channel('public-db-changes')
      .on('postgres_changes', { event: '*', schema: 'public', table: 'tasks' }, payload => {
        if (payload.eventType === 'INSERT') {
          // Add basic structure, relies on invalidation or next refresh for joins if any
          tasks = [...tasks, { ...payload.new as Task, task_assignments: [], events: [] }];
        } else if (payload.eventType === 'UPDATE') {
          const index = tasks.findIndex(t => t.id === payload.new.id);
          if (index !== -1) {
            tasks[index] = { ...tasks[index], ...payload.new };
          }
        } else if (payload.eventType === 'DELETE') {
          tasks = tasks.filter(t => t.id !== payload.old.id);
        }
      })
      .on('postgres_changes', { event: '*', schema: 'public', table: 'categories' }, () => {
        invalidateAll().then(() => { categories = data.categories as Category[]; });
      })
      .on('postgres_changes', { event: '*', schema: 'public', table: 'profiles' }, () => {
        invalidateAll().then(() => { people = data.people as Profile[]; });
      })
      .subscribe();

    return () => {
      supabase.removeChannel(channel);
    };
  });
</script>

<svelte:head>
  <title>Task Tracker</title>
</svelte:head>

<div class="flex flex-col h-screen overflow-hidden bg-gradient-to-br from-gray-50 to-gray-200 dark:from-[#0a0a0f] dark:to-[#13131a] text-gray-900 dark:text-gray-100 transition-colors duration-300">
  <Topbar bind:activeCategoryId {categories} {profile} {my_teams} onAddCategory={() => isCategoryModalOpen = true} onToggleTeam={() => isMobileSidebarOpen = true} />
  
  <div class="flex flex-col md:flex-row flex-1 overflow-hidden relative">
    <Board {filteredTasks} {people} onMove={handleTaskMove} onAddTask={openTaskModal} onTaskClick={handleTaskClick} />
    <Sidebar bind:activePersonFilter {people} bind:isOpen={isMobileSidebarOpen} />
  </div>
</div>

<TaskModal task={editingTask} bind:isOpen={isTaskModalOpen} status={newTaskStatus} categoryId={activeCategoryId} {categories} {people} />
<CategoryModal bind:isOpen={isCategoryModalOpen} {categories} />
