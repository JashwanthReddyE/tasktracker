import { redirect } from '@sveltejs/kit'
import type { LayoutServerLoad } from './$types'

export const load: LayoutServerLoad = async ({ url, locals: { supabase, safeGetSession }, cookies }) => {
  const { session, user } = await safeGetSession()
  let profile = null
  let my_teams = []

  if (user) {
    const { data } = await supabase.from('profiles').select('*').eq('id', user.id).single()
    profile = data
    
    // Fetch all teams the user belongs to for the team switcher
    const { data: teamMembers } = await supabase.from('team_members').select('team_id, teams(*)').eq('user_id', user.id)
    if (teamMembers) {
      my_teams = teamMembers.map(tm => tm.teams)
    }
  }

  // Define public routes
  const publicRoutes = ['/login', '/logout', '/update-password']
  const isPublicRoute = publicRoutes.includes(url.pathname)

  if (user && profile) {
    // If pending and not on pending page or public route, redirect to pending
    if (profile.status === 'pending' && url.pathname !== '/pending' && !isPublicRoute) {
      throw redirect(303, '/pending')
    }
    // If admin route and not admin
    if (url.pathname.startsWith('/admin') && profile.role !== 'admin') {
      throw redirect(303, '/')
    }
  }

  return {
    session,
    profile,
    my_teams,
    cookies: cookies.getAll(),
  }
}
