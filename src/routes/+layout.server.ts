import { redirect } from '@sveltejs/kit'
import type { LayoutServerLoad } from './$types'

export const load: LayoutServerLoad = async ({ url, locals: { supabase, safeGetSession }, cookies }) => {
  const { session, user } = await safeGetSession()
  let profile = null

  if (user) {
    const { data } = await supabase.from('profiles').select('*').eq('id', user.id).single()
    profile = data
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
    cookies: cookies.getAll(),
  }
}
