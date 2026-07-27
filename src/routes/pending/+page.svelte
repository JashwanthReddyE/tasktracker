<div class="min-h-screen bg-gray-50 dark:bg-[#0a0a0f] flex items-center justify-center p-4">
  <div class="max-w-md w-full bg-white dark:bg-[#13131a] p-8 rounded-2xl shadow-sm border border-gray-100 dark:border-white/5 text-center">
    <div class="w-16 h-16 bg-blue-100 dark:bg-blue-900/20 text-blue-600 dark:text-blue-400 rounded-full flex items-center justify-center mx-auto mb-6">
      <svg class="w-8 h-8" fill="none" viewBox="0 0 24 24" stroke="currentColor">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
      </svg>
    </div>
    
    <h1 class="text-2xl font-bold text-gray-900 dark:text-white mb-2">Approval Pending</h1>
    
    <p class="text-gray-600 dark:text-gray-400 mb-8">
      Your account has been created successfully, but it requires administrator approval before you can access the workspace.
    </p>

    <form action="/logout" method="POST">
      <button class="w-full bg-gray-100 dark:bg-white/5 text-gray-900 dark:text-white font-semibold py-2.5 rounded-xl hover:bg-gray-200 dark:hover:bg-white/10 transition-colors">
        Logout
      </button>
    </form>
  </div>
</div>
