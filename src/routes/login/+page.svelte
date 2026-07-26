<script lang="ts">
  import { enhance } from '$app/forms';
  let { form } = $props();
  
  let mode = $state<'login' | 'signup' | 'forgot'>('login');
  let isLoading = $state(false);

  function switchMode(newMode: 'login' | 'signup' | 'forgot') {
    mode = newMode;
    // Clear form errors/success when switching
    if (form) {
      form.error = undefined;
      form.success = undefined;
    }
  }
</script>

<div class="min-h-screen flex items-center justify-center bg-gradient-to-br from-gray-50 to-gray-200 dark:from-[#0a0a0f] dark:to-[#13131a] p-4 transition-colors duration-300">
  <div class="w-full max-w-md bg-white/70 dark:bg-[#1c1c26]/80 backdrop-blur-xl border border-gray-200 dark:border-white/10 rounded-2xl shadow-2xl p-8">
    <div class="text-center mb-8">
      <h1 class="text-3xl font-black tracking-widest uppercase text-transparent bg-clip-text bg-gradient-to-r from-blue-500 to-purple-600 mb-2">Tasks</h1>
      <p class="text-sm font-medium text-gray-500 dark:text-gray-400">
        {#if mode === 'login'}Welcome back! Please enter your details.
        {:else if mode === 'signup'}Create your account to get started.
        {:else}Enter your email to reset your password.{/if}
      </p>
    </div>
    
    {#if form?.error}
      <div class="bg-red-50/50 dark:bg-red-900/20 text-red-600 dark:text-red-400 p-3 rounded-lg mb-6 text-sm font-medium border border-red-100 dark:border-red-900/50">{form.error}</div>
    {/if}
    
    {#if form?.success}
      <div class="bg-green-50/50 dark:bg-green-900/20 text-green-700 dark:text-green-400 p-3 rounded-lg mb-6 text-sm font-medium border border-green-100 dark:border-green-900/50">{form.success}</div>
    {/if}

    <form method="POST" action="?/{mode === 'forgot' ? 'resetPassword' : mode}" use:enhance={() => {
      isLoading = true;
      return async ({ update }) => {
        isLoading = false;
        await update();
      };
    }}>
      <div class="space-y-5">
        {#if mode === 'signup'}
          <div>
            <label for="name" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Full Name</label>
            <input type="text" id="name" name="name" value={form?.name ?? ''} required class="w-full rounded-lg border border-gray-300 dark:border-gray-600 bg-white/50 dark:bg-black/20 px-4 py-2.5 text-sm focus:ring-2 focus:ring-blue-500 outline-none transition-all dark:text-white" placeholder="John Doe" />
          </div>
        {/if}

        <div>
          <label for="email" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Email</label>
          <input type="email" id="email" name="email" value={form?.email ?? ''} required class="w-full rounded-lg border border-gray-300 dark:border-gray-600 bg-white/50 dark:bg-black/20 px-4 py-2.5 text-sm focus:ring-2 focus:ring-blue-500 outline-none transition-all dark:text-white" placeholder="you@example.com" />
        </div>

        {#if mode !== 'forgot'}
          <div>
            <div class="flex items-center justify-between mb-1">
              <label for="password" class="block text-sm font-medium text-gray-700 dark:text-gray-300">Password</label>
              {#if mode === 'login'}
                <button type="button" class="text-xs font-medium text-blue-600 dark:text-blue-400 hover:underline" onclick={() => switchMode('forgot')}>Forgot password?</button>
              {/if}
            </div>
            <input type="password" id="password" name="password" required class="w-full rounded-lg border border-gray-300 dark:border-gray-600 bg-white/50 dark:bg-black/20 px-4 py-2.5 text-sm focus:ring-2 focus:ring-blue-500 outline-none transition-all dark:text-white" placeholder="••••••••" />
          </div>
        {/if}

        <button type="submit" disabled={isLoading} class="w-full bg-gradient-to-r from-blue-600 to-purple-600 hover:from-blue-700 hover:to-purple-700 text-white py-2.5 rounded-lg shadow-md text-sm font-semibold transition-all disabled:opacity-70">
          {#if isLoading}
            <span class="animate-pulse">Please wait...</span>
          {:else if mode === 'login'}Sign In
          {:else if mode === 'signup'}Create Account
          {:else}Send Reset Link{/if}
        </button>
      </div>
    </form>

    <div class="mt-8 text-center text-sm text-gray-600 dark:text-gray-400">
      {#if mode === 'login'}
        Don't have an account? <button type="button" class="font-semibold text-blue-600 dark:text-blue-400 hover:underline" onclick={() => switchMode('signup')}>Sign up</button>
      {:else}
        Back to <button type="button" class="font-semibold text-blue-600 dark:text-blue-400 hover:underline" onclick={() => switchMode('login')}>Sign in</button>
      {/if}
    </div>
  </div>
</div>
