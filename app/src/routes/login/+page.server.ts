import { fail, redirect } from '@sveltejs/kit'
import type { Actions } from './$types'

export const actions: Actions = {
  login: async ({ request, locals: { supabase } }) => {
    const formData = await request.formData()
    const email = formData.get('email') as string
    const password = formData.get('password') as string

    if (!email || !password) {
      return fail(400, { error: 'Please enter an email and password', email })
    }

    const { error } = await supabase.auth.signInWithPassword({
      email,
      password,
    })

    if (error) {
      return fail(400, { error: error.message, email })
    }

    throw redirect(303, '/')
  },
  signup: async ({ request, locals: { supabase } }) => {
    const formData = await request.formData()
    const email = formData.get('email') as string
    const password = formData.get('password') as string

    if (!email || !password) {
      return fail(400, { error: 'Please enter an email and password', email })
    }

    const { error } = await supabase.auth.signUp({
      email,
      password,
    })

    if (error) {
      return fail(400, { error: error.message, email })
    }

    return { success: 'Check your email to confirm your account (or login directly if email confirmations are disabled).' }
  }
}
