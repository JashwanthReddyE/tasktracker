-- Fix for categories_user_id_fkey
ALTER TABLE public.categories DROP CONSTRAINT IF EXISTS categories_user_id_fkey;
ALTER TABLE public.categories ADD CONSTRAINT categories_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.profiles(id) ON DELETE CASCADE;

-- Fix for task_assignments assigned_by reference
ALTER TABLE public.task_assignments DROP CONSTRAINT IF EXISTS task_assignments_assigned_by_fkey;
ALTER TABLE public.task_assignments ADD CONSTRAINT task_assignments_assigned_by_fkey FOREIGN KEY (assigned_by) REFERENCES public.profiles(id) ON DELETE SET NULL;

-- Fix for teams created_by reference
ALTER TABLE public.teams DROP CONSTRAINT IF EXISTS teams_created_by_fkey;
ALTER TABLE public.teams ADD CONSTRAINT teams_created_by_fkey FOREIGN KEY (created_by) REFERENCES public.profiles(id) ON DELETE SET NULL;

-- Fix for events user_id reference (if it exists)
DO $$ BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.table_constraints WHERE constraint_name = 'events_user_id_fkey') THEN
        ALTER TABLE public.events DROP CONSTRAINT events_user_id_fkey;
        ALTER TABLE public.events ADD CONSTRAINT events_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.profiles(id) ON DELETE CASCADE;
    END IF;
END $$;
