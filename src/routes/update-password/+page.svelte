<script lang="ts">
  import { enhance } from '$app/forms';
  let { form } = $props();
  let isLoading = $state(false);
</script>

<div class="min-h-screen flex items-center justify-center bg-gradient-to-br from-gray-50 to-gray-200 dark:from-[#0a0a0f] dark:to-[#13131a] p-4 transition-colors duration-300">
  <div class="w-full max-w-md bg-white/70 dark:bg-[#1c1c26]/80 backdrop-blur-xl border border-gray-200 dark:border-white/10 rounded-2xl shadow-2xl p-8 animate-in fade-in zoom-in-95 duration-300">
    <div class="text-center mb-8">
      <h1 class="text-2xl font-bold tracking-tight text-gray-900 dark:text-white mb-2">Update Password</h1>
      <p class="text-sm font-medium text-gray-500 dark:text-gray-400">
        Please enter your new password below.
      </p>
    </div>
    
    {#if form?.error}
      <div class="bg-red-50/50 dark:bg-red-900/20 text-red-600 dark:text-red-400 p-3 rounded-lg mb-6 text-sm font-medium border border-red-100 dark:border-red-900/50">{form.error}</div>
    {/if}

    <form method="POST" action="?/updatePassword" use:enhance={() => {
      isLoading = true;
      return async ({ update }) => {
        isLoading = false;
        await update();
      };
    }}>
      <div class="space-y-5">
        <div>
          <label for="password" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">New Password</label>
          <input type="password" id="password" name="password" required class="w-full rounded-lg border border-gray-300 dark:border-gray-600 bg-white/50 dark:bg-black/20 px-4 py-2.5 text-sm focus:ring-2 focus:ring-blue-500 outline-none transition-all dark:text-white" placeholder="••••••••" />
        </div>

        <button type="submit" disabled={isLoading} class="w-full bg-gradient-to-r from-blue-600 to-purple-600 hover:from-blue-700 hover:to-purple-700 text-white py-2.5 rounded-lg shadow-md text-sm font-semibold transition-all disabled:opacity-70">
          {#if isLoading}
            <span class="animate-pulse">Updating...</span>
          {:else}Update Password{/if}
        </button>
      </div>
    </form>
  </div>
</div>
