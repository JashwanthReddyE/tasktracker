import { fail, redirect } from '@sveltejs/kit'
import type { PageServerLoad, Actions } from './$types'

export const load: PageServerLoad = async ({ locals: { supabase, safeGetSession } }) => {
  const { session, user } = await safeGetSession()

  if (!session || !user) {
    throw redirect(303, '/login')
  }

  // Load tasks with their assignments and events
  const { data: tasks, error: tasksError } = await supabase
    .from('tasks')
    .select('*, task_assignments(*), events(*)')
    .order('position', { ascending: true })

  // Load categories
  const { data: categories, error: catsError } = await supabase
    .from('categories')
    .select('*')
    .order('position', { ascending: true })

  // Load all profiles (global directory)
  const { data: profiles, error: profilesError } = await supabase
    .from('profiles')
    .select('*')
    .order('name', { ascending: true })

  // --- DEBUG CODE ---
  try {
    const fs = await import('fs');
    const { data: myProfile } = await supabase.from('profiles').select('*').eq('id', user.id).single();
    // Get foreign keys for tasks table
    const { data: fks } = await supabase.rpc('get_foreign_keys'); // if this exists, else raw query
    // Actually we can't easily query information_schema from supabase-js. 
    // Let's just dump the user and profile.
    fs.writeFileSync('debug.json', JSON.stringify({ user, myProfile }, null, 2));
  } catch (e) {
    console.error('Debug write failed', e);
  }
  // --- END DEBUG CODE ---

  if (tasksError) console.error('Error loading tasks:', tasksError)
  if (catsError) console.error('Error loading cats:', catsError)
  if (profilesError) console.error('Error loading profiles:', profilesError)

  return {
    tasks: tasks ?? [],
    categories: categories ?? [],
    people: profiles ?? [],
  }
}

export const actions: Actions = {
  createTask: async ({ request, locals: { supabase, user } }) => {
    if (!user) return fail(401, { error: 'Unauthorized' })

    const formData = await request.formData()
    const title = formData.get('title') as string
    const notes = (formData.get('notes') as string) || ''
    const priority = (formData.get('priority') as string) || 'medium'
    const status = (formData.get('status') as string) || 'todo'
    const category_id = (formData.get('category_id') as string) || ''
    const due_date = (formData.get('due_date') as string) || ''
    const position = parseInt((formData.get('position') as string) || '0', 10)
    
    // Arrays can be passed as JSON strings for simplicity in form actions
    const peopleIdsJson = formData.get('people_ids') as string
    const peopleIds = peopleIdsJson ? JSON.parse(peopleIdsJson) : []

    const { data: profile } = await supabase.from('profiles').select('team_id').eq('id', user.id).single()

    const { data: task, error } = await supabase
      .from('tasks')
      .insert({
        user_id: user.id,
        team_id: profile?.team_id,
        title,
        notes,
        priority,
        status,
        category_id,
        due_date,
        position,
        archived: false,
      })
      .select()
      .single()

    if (error) return fail(500, { error: error.message })

    // Insert task_assignments
    if (peopleIds.length > 0) {
      const taskAssignments = peopleIds.map((personId: string) => ({
        task_id: task.id,
        user_id: personId, // the assignee
        assigned_by: user.id, // the assigner
      }))
      await supabase.from('task_assignments').insert(taskAssignments)
    }

    return { success: true, task }
  },

  updateTask: async ({ request, locals: { supabase, user } }) => {
    if (!user) return fail(401, { error: 'Unauthorized' })

    const formData = await request.formData()
    const id = formData.get('id') as string
    const title = formData.get('title') as string
    const notes = formData.get('notes') as string
    const priority = formData.get('priority') as string
    const status = formData.get('status') as string
    const category_id = formData.get('category_id') as string
    const due_date = formData.get('due_date') as string
    const archived = formData.get('archived') === 'true'
    const position = parseInt(formData.get('position') as string, 10)

    const updates: any = {}
    if (title !== null) updates.title = title
    if (notes !== null) updates.notes = notes
    if (priority !== null) updates.priority = priority
    if (status !== null) updates.status = status
    if (category_id !== null) updates.category_id = category_id
    if (due_date !== null) updates.due_date = due_date
    if (formData.has('archived')) updates.archived = archived
    if (!isNaN(position)) updates.position = position

    const { error } = await supabase
      .from('tasks')
      .update(updates)
      .eq('id', id)
      .eq('user_id', user.id)

    if (error) return fail(500, { error: error.message })

    // Update assignments
    const peopleIdsJson = formData.get('people_ids') as string
    if (peopleIdsJson) {
      const peopleIds = JSON.parse(peopleIdsJson)
      // Delete old assignments
      await supabase.from('task_assignments').delete().eq('task_id', id)
      
      if (peopleIds.length > 0) {
        const taskAssignments = peopleIds.map((personId: string) => ({
          task_id: id,
          user_id: personId,
          assigned_by: user.id,
        }))
        await supabase.from('task_assignments').insert(taskAssignments)
      }
    }

    return { success: true }
  },

  deleteTask: async ({ request, locals: { supabase, user } }) => {
    if (!user) return fail(401, { error: 'Unauthorized' })

    const formData = await request.formData()
    const id = formData.get('id') as string

    const { error } = await supabase
      .from('tasks')
      .delete()
      .eq('id', id)
      .eq('user_id', user.id)

    if (error) return fail(500, { error: error.message })
    return { success: true }
  },

  replaceCategories: async ({ request, locals: { supabase, user } }) => {
    if (!user) return fail(401, { error: 'Unauthorized' })
    const formData = await request.formData()
    const categoriesJson = formData.get('categories') as string
    const categories = JSON.parse(categoriesJson || '[]')

    // Delete existing
    await supabase.from('categories').delete().eq('user_id', user.id)
    
    // Insert new
    if (categories.length > 0) {
      const { data: profile } = await supabase.from('profiles').select('team_id').eq('id', user.id).single()
      const inserts = categories.map((c: any, i: number) => ({
        user_id: user.id,
        team_id: profile?.team_id,
        id: c.id,
        label: c.label,
        position: i
      }))
      const { error } = await supabase.from('categories').insert(inserts)
      if (error) return fail(500, { error: error.message })
    }
    return { success: true }
  },

  addEvent: async ({ request, locals: { supabase, user } }) => {
    if (!user) return fail(401, { error: 'Unauthorized' })
    const formData = await request.formData()
    const task_id = formData.get('task_id') as string
    const type = formData.get('type') as string
    const text = formData.get('text') as string

    // Verify task ownership
    const { data: task } = await supabase.from('tasks').select('id').eq('id', task_id).eq('user_id', user.id).single()
    if (!task) return fail(404, { error: 'Task not found' })

    const { error } = await supabase
      .from('events')
      .insert({
        task_id,
        type,
        text,
        time: new Date().toISOString()
      })

    if (error) return fail(500, { error: error.message })
    return { success: true }
  },

  makeMeAdmin: async ({ locals: { supabase, user } }) => {
    if (!user) return fail(401, { error: 'Unauthorized' })
    const { error } = await supabase.from('profiles').update({ role: 'admin', status: 'approved' }).eq('id', user.id)
    if (error) return fail(500, { error: error.message })
    return { success: true }
  },

  switchTeam: async ({ request, locals: { supabase, user } }) => {
    if (!user) return fail(401, { error: 'Unauthorized' })
    const formData = await request.formData()
    const team_id = formData.get('team_id') as string

    if (!team_id) return fail(400, { error: 'Missing team_id' })

    const { error } = await supabase.from('profiles').update({ team_id }).eq('id', user.id)
    if (error) return fail(500, { error: error.message })
    return { success: true }
  }
}
