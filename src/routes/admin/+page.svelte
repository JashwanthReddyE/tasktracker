<script lang="ts">
  import type { PageData } from './$types';
  import { enhance } from '$app/forms';

  let { data } = $props<{ data: PageData }>();
  let pendingUsers = $derived(data.profiles.filter((p: any) => p.status === 'pending'));
  let approvedUsers = $derived(data.profiles.filter((p: any) => p.status === 'approved'));
</script>

<svelte:head>
  <title>Admin Dashboard</title>
</svelte:head>

<div class="min-h-screen bg-gray-50 dark:bg-[#0a0a0f] p-6 md:p-12 transition-colors duration-300">
  <div class="max-w-5xl mx-auto space-y-8">
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-3xl font-black tracking-tight text-gray-900 dark:text-white">Admin Dashboard</h1>
        <p class="text-gray-500 dark:text-gray-400 mt-1">Manage users, approvals, and teams.</p>
      </div>
      <a href="/" class="px-4 py-2 bg-white dark:bg-white/10 text-gray-700 dark:text-gray-200 rounded-lg text-sm font-semibold hover:bg-gray-100 dark:hover:bg-white/20 transition-colors border border-gray-200 dark:border-white/5">
        Back to App
      </a>
    </div>

    <div class="grid grid-cols-1 lg:grid-cols-3 gap-8">
      
      <!-- Pending Approvals -->
      <div class="lg:col-span-2 space-y-6">
        <div class="bg-white dark:bg-[#13131a] rounded-2xl p-6 shadow-sm border border-gray-200 dark:border-white/5">
          <h2 class="text-lg font-bold text-gray-900 dark:text-white mb-4 flex items-center gap-2">
            Pending Approvals
            <span class="bg-orange-100 text-orange-700 dark:bg-orange-500/20 dark:text-orange-400 text-xs px-2 py-0.5 rounded-full">{pendingUsers.length}</span>
          </h2>
          
          {#if pendingUsers.length === 0}
            <div class="py-8 text-center text-gray-500 dark:text-gray-400 border-2 border-dashed border-gray-200 dark:border-white/5 rounded-xl">
              No pending users.
            </div>
          {:else}
            <div class="space-y-4">
              {#each pendingUsers as user}
                <div class="p-4 bg-white dark:bg-[#13131a] rounded-xl border border-gray-100 dark:border-white/5 flex items-center justify-between shadow-sm">
                  <div>
                    <div class="font-medium text-gray-900 dark:text-gray-100">{user.name}</div>
                    <div class="text-sm text-gray-500 dark:text-gray-400">{user.email}</div>
                    {#if user.requested_team_id}
                      {@const team = data.teams.find(t => t.id === user.requested_team_id)}
                      <div class="text-xs text-blue-500 dark:text-blue-400 mt-1">Requested to join: {team?.name || 'Unknown Team'}</div>
                    {/if}
                  </div>
                  <div class="flex items-center gap-3">
                    <form method="POST" action="?/approveUser" use:enhance class="flex items-center gap-2">
                      <input type="hidden" name="user_id" value={user.id} />
                      <select name="team_id" class="text-sm rounded-lg border border-gray-200 dark:border-white/10 bg-gray-50 dark:bg-white/5 px-2 py-1 outline-none text-gray-700 dark:text-gray-300">
                        {#each data.teams as team}
                          <option value={team.id} selected={team.id === user.requested_team_id}>{team.name}</option>
                        {/each}
                      </select>
                      <button class="px-4 py-2 rounded-lg bg-green-500 hover:bg-green-600 text-white text-sm font-semibold transition-colors shadow-lg shadow-green-500/20">
                        Approve
                      </button>
                    </form>
                    <form method="POST" action="?/denyUser" use:enhance>
                      <input type="hidden" name="user_id" value={user.id} />
                      <button class="px-4 py-2 rounded-lg bg-red-500 hover:bg-red-600 text-white text-sm font-semibold transition-colors shadow-lg shadow-red-500/20">
                        Deny
                      </button>
                    </form>
                  </div>
                </div>
              {/each}
            </div>
          {/if}
        </div>

        <!-- Approved Users -->
        <div class="bg-white dark:bg-[#13131a] rounded-2xl p-6 shadow-sm border border-gray-200 dark:border-white/5">
          <h2 class="text-lg font-bold text-gray-900 dark:text-white mb-4">Approved Users</h2>
          <div class="space-y-3">
            {#each approvedUsers as user}
              <div class="flex items-center justify-between p-3 rounded-lg hover:bg-gray-50 dark:hover:bg-white/5 transition-colors">
                <div class="flex items-center gap-3">
                  <div class="w-8 h-8 rounded-full flex items-center justify-center text-xs text-white font-bold shrink-0" style="background-color: hsl({user.hue}, 65%, 55%)">
                    {user.name.charAt(0).toUpperCase()}
                  </div>
                  <div>
                    <p class="font-medium text-gray-900 dark:text-white text-sm">
                      {user.name}
                      {#if user.role === 'admin'}
                        <span class="ml-2 text-[10px] uppercase tracking-wider bg-purple-100 text-purple-700 dark:bg-purple-500/20 dark:text-purple-300 px-1.5 py-0.5 rounded">Admin</span>
                      {/if}
                    </p>
                    <p class="text-xs text-gray-500">{user.email} • Team: {data.teams.find((t: any) => t.id === user.team_id)?.name || 'None'}</p>
                  </div>
                </div>
                {#if user.role !== 'admin'}
                  <div class="flex items-center gap-4">
                    <form method="POST" action="?/promoteToAdmin" use:enhance>
                      <input type="hidden" name="user_id" value={user.id} />
                      <button class="text-xs font-semibold text-gray-500 hover:text-purple-600 dark:text-gray-400 dark:hover:text-purple-400">
                        Make Admin
                      </button>
                    </form>
                    <form method="POST" action="?/removeUser" use:enhance>
                      <input type="hidden" name="user_id" value={user.id} />
                      <button class="text-xs font-semibold text-red-500 hover:text-red-700 dark:text-red-400 dark:hover:text-red-300">
                        Revoke Access
                      </button>
                    </form>
                  </div>
                {/if}
              </div>
            {/each}
          </div>
        </div>
      </div>

      <!-- Teams Sidebar -->
      <div class="space-y-6">
        <div class="bg-white dark:bg-[#13131a] rounded-2xl p-6 shadow-sm border border-gray-200 dark:border-white/5">
          <h2 class="text-lg font-bold text-gray-900 dark:text-white mb-4">Teams</h2>
          <ul class="space-y-2 mb-6">
            {#each data.teams as team}
              <li class="flex items-center justify-between px-3 py-2 bg-gray-50 dark:bg-white/5 rounded-lg text-sm text-gray-700 dark:text-gray-300 font-medium border border-gray-100 dark:border-white/5">
                {team.name}
              </li>
            {/each}
          </ul>
          
          <form method="POST" action="?/createTeam" use:enhance class="space-y-3">
            <div>
              <label class="block text-xs font-medium text-gray-700 dark:text-gray-400 mb-1">Create New Team</label>
              <input type="text" name="name" required placeholder="e.g. Marketing" class="w-full rounded-lg border-gray-300 dark:border-gray-600 bg-white dark:bg-black/40 px-3 py-2 text-sm focus:ring-blue-500 focus:border-blue-500 dark:text-white outline-none">
            </div>
            <button class="w-full px-4 py-2 bg-gray-900 dark:bg-white text-white dark:text-gray-900 rounded-lg text-sm font-semibold hover:bg-gray-800 dark:hover:bg-gray-100 transition-colors">
              Add Team
            </button>
          </form>
        </div>
      </div>
      
    </div>
  </div>
</div>
