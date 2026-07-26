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
    const name = formData.get('name') as string
    const email = formData.get('email') as string
    const password = formData.get('password') as string

    if (!email || !password || !name) {
      return fail(400, { error: 'Please enter a name, email, and password', email, name })
    }

    const { error } = await supabase.auth.signUp({
      email,
      password,
      options: {
        data: { name }
      }
    })

    if (error) {
      return fail(400, { error: error.message, email })
    }

    return { success: 'Check your email to confirm your account (or login directly if email confirmations are disabled).' }
  },
  resetPassword: async ({ request, url, locals: { supabase } }) => {
    const formData = await request.formData()
    const email = formData.get('email') as string

    if (!email) {
      return fail(400, { error: 'Please enter your email', email })
    }

    const { error } = await supabase.auth.resetPasswordForEmail(email, {
      redirectTo: `${url.origin}/update-password`,
    })

    if (error) {
      return fail(400, { error: error.message, email })
    }

    return { success: 'Password reset link sent! Please check your email.' }
  }
}
