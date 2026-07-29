-- Fix for foreign key constraint errors when inserting via PostgREST
-- PostgREST runs as the "authenticated" role which sometimes causes issues
-- when evaluating foreign key constraints that cross schemas into auth.users.
-- Changing these to reference public.profiles (which is 1:1 with auth.users)
-- completely sidesteps this permissions issue.

ALTER TABLE public.tasks 
  DROP CONSTRAINT IF EXISTS tasks_user_id_fkey;
  
ALTER TABLE public.tasks 
  ADD CONSTRAINT tasks_user_id_fkey 
  FOREIGN KEY (user_id) REFERENCES public.profiles(id) ON DELETE CASCADE;

-- Also fix task_assignments just in case
ALTER TABLE public.task_assignments 
  DROP CONSTRAINT IF EXISTS task_assignments_assigned_by_fkey;
  
ALTER TABLE public.task_assignments 
  ADD CONSTRAINT task_assignments_assigned_by_fkey 
  FOREIGN KEY (assigned_by) REFERENCES public.profiles(id) ON DELETE SET NULL;
