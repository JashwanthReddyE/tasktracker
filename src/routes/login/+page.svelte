<script lang="ts">
  let { form } = $props();
</script>

<div class="max-w-md mx-auto mt-20 p-6 bg-white shadow-lg rounded-xl dark:bg-gray-800 border border-gray-100 dark:border-gray-700">
  <h1 class="text-3xl font-bold mb-6 text-center text-gray-900 dark:text-white tracking-tight">Task Tracker</h1>
  
  {#if form?.error}
    <div class="bg-red-50 text-red-600 p-3 rounded-lg mb-6 text-sm font-medium border border-red-100">{form.error}</div>
  {/if}
  
  {#if form?.success}
    <div class="bg-green-50 text-green-700 p-3 rounded-lg mb-6 text-sm font-medium border border-green-100">{form.success}</div>
  {/if}

  <form method="POST">
    <div class="mb-5">
      <label for="email" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Email</label>
      <input type="email" id="email" name="email" value={form?.email ?? ''} required class="block w-full border border-gray-300 rounded-lg shadow-sm py-2.5 px-3 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500 sm:text-sm dark:bg-gray-900 dark:border-gray-600 dark:text-white transition-colors" placeholder="you@example.com" />
    </div>

    <div class="mb-8">
      <label for="password" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Password</label>
      <input type="password" id="password" name="password" required class="block w-full border border-gray-300 rounded-lg shadow-sm py-2.5 px-3 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500 sm:text-sm dark:bg-gray-900 dark:border-gray-600 dark:text-white transition-colors" placeholder="••••••••" />
    </div>

    <div class="flex space-x-4">
      <button formaction="?/login" class="flex-1 bg-blue-600 text-white py-2.5 px-4 rounded-lg shadow-sm text-sm font-semibold hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 transition-colors">
        Sign In
      </button>
      <button formaction="?/signup" class="flex-1 bg-gray-100 text-gray-900 py-2.5 px-4 rounded-lg shadow-sm text-sm font-semibold hover:bg-gray-200 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-gray-200 transition-colors dark:bg-gray-700 dark:text-gray-100 dark:hover:bg-gray-600">
        Sign Up
      </button>
    </div>
  </form>
</div>
