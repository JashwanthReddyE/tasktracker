import { fail, redirect } from '@sveltejs/kit'
import type { Actions, PageServerLoad } from './$types'

export const load: PageServerLoad = async ({ locals: { safeGetSession } }) => {
  const { session } = await safeGetSession()
  if (!session) {
    throw redirect(303, '/login')
  }
}

export const actions: Actions = {
  updatePassword: async ({ request, locals: { supabase } }) => {
    const formData = await request.formData()
    const password = formData.get('password') as string

    if (!password || password.length < 6) {
      return fail(400, { error: 'Password must be at least 6 characters long.' })
    }

    const { error } = await supabase.auth.updateUser({
      password,
    })

    if (error) {
      return fail(400, { error: error.message })
    }

    throw redirect(303, '/')
  }
}
