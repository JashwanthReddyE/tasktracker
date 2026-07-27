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

    const { error } = await supabase
      .from('profiles')
      .update({ status: 'approved', team_id: teamId })
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
