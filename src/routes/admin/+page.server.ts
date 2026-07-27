import { fail } from '@sveltejs/kit';
import type { PageServerLoad, Actions } from './$types';

export const load: PageServerLoad = async ({ locals: { supabase, safeGetSession } }) => {
  const { session, user } = await safeGetSession();

  // Load all profiles
  const { data: profiles, error: pError } = await supabase
    .from('profiles')
    .select('*')
    .order('created_at', { ascending: false });

  // Load all teams
  const { data: teams, error: tError } = await supabase
    .from('teams')
    .select('*')
    .order('name', { ascending: true });

  return {
    profiles: profiles || [],
    teams: teams || [],
  };
};

export const actions: Actions = {
  approveUser: async ({ request, locals: { supabase, user } }) => {
    if (!user) return fail(401, { error: 'Unauthorized' });
    
    const formData = await request.formData();
    const targetUserId = formData.get('user_id') as string;
    const teamId = formData.get('team_id') as string;
    
    if (!targetUserId || !teamId) return fail(400, { error: 'Missing fields' });


    // Admin can override the team during approval
    const teamIdOverride = formData.get('team_id') as string;
    
    if (!targetUserId) return fail(400, { error: 'Missing user ID' });

    // First fetch the profile to get their requested team if no override
    const { data: profile } = await supabase.from('profiles').select('requested_team_id').eq('id', targetUserId).single();
    
    const finalTeamId = teamIdOverride || profile?.requested_team_id || '00000000-0000-0000-0000-000000000000';

    // 1. Approve them and set active team
    const { error: profileError } = await supabase
      .from('profiles')
      .update({ status: 'approved', team_id: finalTeamId })
      .eq('id', targetUserId);
      
    if (profileError) return fail(500, { error: profileError.message });

    // 2. Add them to team_members
    const { error: teamError } = await supabase
      .from('team_members')
      .insert({ team_id: finalTeamId, user_id: targetUserId })
      .select()
      .single();

    // Ignore conflict error if they are already in the team
    if (teamError && teamError.code !== '23505') {
       return fail(500, { error: teamError.message });
    }

    return { success: true };
  },

  denyUser: async ({ request, locals: { supabase, user } }) => {
    if (!user) return fail(401, { error: 'Unauthorized' });
    
    const formData = await request.formData();
    const targetUserId = formData.get('user_id') as string;
    
    if (!targetUserId) return fail(400, { error: 'Missing user ID' });

    const { error } = await supabase
      .from('profiles')
      .update({ status: 'rejected' })
      .eq('id', targetUserId);
      
    if (error) return fail(500, { error: error.message });
    return { success: true };
  },
  
  promoteToAdmin: async ({ request, locals: { supabase, user } }) => {
    if (!user) return fail(401, { error: 'Unauthorized' });
    
    const formData = await request.formData();
    const targetUserId = formData.get('user_id') as string;
    
    if (!targetUserId) return fail(400, { error: 'Missing user ID' });

    const { error } = await supabase
      .from('profiles')
      .update({ role: 'admin' })
      .eq('id', targetUserId);
      
    if (error) return fail(500, { error: error.message });
    return { success: true };
  },

  removeUser: async ({ request, locals: { supabase, user } }) => {
    if (!user) return fail(401, { error: 'Unauthorized' });
    
    const formData = await request.formData();
    const targetUserId = formData.get('user_id') as string;
    
    if (!targetUserId) return fail(400, { error: 'Missing user ID' });

    // We 'remove' them by revoking their approval and team, locking them out
    const { error } = await supabase
      .from('profiles')
      .update({ status: 'pending', role: 'user', team_id: null })
      .eq('id', targetUserId);
      
    if (error) return fail(500, { error: error.message });
    return { success: true };
  },

  createTeam: async ({ request, locals: { supabase, user } }) => {
    if (!user) return fail(401, { error: 'Unauthorized' });
    
    const formData = await request.formData();
    const name = formData.get('name') as string;
    
    if (!name) return fail(400, { error: 'Missing team name' });

    const { error } = await supabase
      .from('teams')
      .insert({ name, created_by: user.id });
      
    if (error) return fail(500, { error: error.message });
    return { success: true };
  }
};
