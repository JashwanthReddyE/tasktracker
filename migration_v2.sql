-- ==========================================
-- MIGRATION V2: Invite-Only & Multi-Team
-- ==========================================

-- 1. Create teams table
CREATE TABLE IF NOT EXISTS public.teams (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name TEXT NOT NULL,
  created_at TIMESTAMP WITH TIME ZONE DEFAULT timezone('utc'::text, now()),
  created_by UUID REFERENCES auth.users(id) ON DELETE SET NULL
);
ALTER TABLE public.teams ENABLE ROW LEVEL SECURITY;

-- 2. Insert Default General Team (if not exists)
INSERT INTO public.teams (id, name) 
VALUES ('00000000-0000-0000-0000-000000000000', 'General')
ON CONFLICT (id) DO NOTHING;

-- 3. Alter profiles table
ALTER TABLE public.profiles
ADD COLUMN IF NOT EXISTS status TEXT DEFAULT 'pending',
ADD COLUMN IF NOT EXISTS role TEXT DEFAULT 'user',
ADD COLUMN IF NOT EXISTS team_id UUID REFERENCES public.teams(id) ON DELETE SET NULL;

-- 4. Alter tasks and categories
ALTER TABLE public.tasks ADD COLUMN IF NOT EXISTS team_id UUID REFERENCES public.teams(id) ON DELETE CASCADE;
ALTER TABLE public.categories ADD COLUMN IF NOT EXISTS team_id UUID REFERENCES public.teams(id) ON DELETE CASCADE;

-- Default existing tasks/categories to General team
UPDATE public.tasks SET team_id = '00000000-0000-0000-0000-000000000000' WHERE team_id IS NULL;
UPDATE public.categories SET team_id = '00000000-0000-0000-0000-000000000000' WHERE team_id IS NULL;

-- 5. Update user creation trigger to handle admin promotion and defaults
CREATE OR REPLACE FUNCTION public.handle_new_user()
RETURNS trigger AS $$
DECLARE
  is_admin BOOLEAN;
BEGIN
  is_admin := (new.email = 'jashwanthreddyearla@gmail.com');

  INSERT INTO public.profiles (id, email, name, hue, status, role, team_id)
  VALUES (
    new.id,
    new.email,
    COALESCE(new.raw_user_meta_data->>'name', 'Unknown User'),
    floor(random() * 360),
    CASE WHEN is_admin THEN 'approved' ELSE 'pending' END,
    CASE WHEN is_admin THEN 'admin' ELSE 'user' END,
    CASE WHEN is_admin THEN '00000000-0000-0000-0000-000000000000'::uuid ELSE NULL END
  );
  RETURN new;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- Force update existing admin
UPDATE public.profiles
SET role = 'admin', status = 'approved', team_id = '00000000-0000-0000-0000-000000000000'
WHERE email = 'jashwanthreddyearla@gmail.com';

-- Force update existing users to be approved and in General team so they don't lose access
UPDATE public.profiles
SET status = 'approved', team_id = '00000000-0000-0000-0000-000000000000'
WHERE email != 'jashwanthreddyearla@gmail.com' AND status = 'pending';


-- 6. Rewrite RLS Policies for true isolation
-- Profiles
DROP POLICY IF EXISTS "Profiles are viewable by everyone" ON public.profiles;
DROP POLICY IF EXISTS "Users can update own profile" ON public.profiles;
DROP POLICY IF EXISTS "Profiles viewable by team members or admins" ON public.profiles;
DROP POLICY IF EXISTS "Profiles updatable by self or admin" ON public.profiles;

CREATE POLICY "Profiles viewable by team members or admins" ON public.profiles FOR SELECT USING (
  team_id = (SELECT p.team_id FROM public.profiles p WHERE p.id = auth.uid()) OR
  (SELECT p.role FROM public.profiles p WHERE p.id = auth.uid()) = 'admin' OR
  id = auth.uid()
);

CREATE POLICY "Profiles updatable by self or admin" ON public.profiles FOR UPDATE USING (
  id = auth.uid() OR
  (SELECT p.role FROM public.profiles p WHERE p.id = auth.uid()) = 'admin'
);

-- Teams
DROP POLICY IF EXISTS "Teams viewable by members or admins" ON public.teams;
CREATE POLICY "Teams viewable by members or admins" ON public.teams FOR SELECT USING (
  id = (SELECT p.team_id FROM public.profiles p WHERE p.id = auth.uid()) OR
  (SELECT p.role FROM public.profiles p WHERE p.id = auth.uid()) = 'admin'
);
-- Admins can insert/update teams
CREATE POLICY "Teams manageable by admins" ON public.teams FOR ALL USING (
  (SELECT p.role FROM public.profiles p WHERE p.id = auth.uid()) = 'admin'
);

-- Tasks
DROP POLICY IF EXISTS "Users can view assigned and owned tasks" ON public.tasks;
DROP POLICY IF EXISTS "Tasks viewable by team members or admins" ON public.tasks;
DROP POLICY IF EXISTS "Tasks editable by team members or admins" ON public.tasks;

CREATE POLICY "Tasks viewable by team members or admins" ON public.tasks FOR SELECT USING (
  team_id = (SELECT p.team_id FROM public.profiles p WHERE p.id = auth.uid()) OR
  (SELECT p.role FROM public.profiles p WHERE p.id = auth.uid()) = 'admin'
);

CREATE POLICY "Tasks editable by team members or admins" ON public.tasks FOR ALL USING (
  team_id = (SELECT p.team_id FROM public.profiles p WHERE p.id = auth.uid()) OR
  (SELECT p.role FROM public.profiles p WHERE p.id = auth.uid()) = 'admin'
);

-- Categories
DROP POLICY IF EXISTS "Users can view their own categories or categories with assigned tasks" ON public.categories;
DROP POLICY IF EXISTS "Categories viewable by team members or admins" ON public.categories;
DROP POLICY IF EXISTS "Categories editable by team members or admins" ON public.categories;

CREATE POLICY "Categories viewable by team members or admins" ON public.categories FOR SELECT USING (
  team_id = (SELECT p.team_id FROM public.profiles p WHERE p.id = auth.uid()) OR
  (SELECT p.role FROM public.profiles p WHERE p.id = auth.uid()) = 'admin'
);
CREATE POLICY "Categories editable by team members or admins" ON public.categories FOR ALL USING (
  team_id = (SELECT p.team_id FROM public.profiles p WHERE p.id = auth.uid()) OR
  (SELECT p.role FROM public.profiles p WHERE p.id = auth.uid()) = 'admin'
);

-- Task Assignments
DROP POLICY IF EXISTS "Assignments viewable by assignee or task owner" ON public.task_assignments;
DROP POLICY IF EXISTS "Task owners can manage assignments" ON public.task_assignments;

CREATE POLICY "Assignments viewable by team members" ON public.task_assignments FOR SELECT USING (
  EXISTS (
    SELECT 1 FROM public.tasks t 
    WHERE t.id = task_id AND (
      t.team_id = (SELECT p.team_id FROM public.profiles p WHERE p.id = auth.uid()) OR
      (SELECT p.role FROM public.profiles p WHERE p.id = auth.uid()) = 'admin'
    )
  )
);
CREATE POLICY "Assignments manageable by team members" ON public.task_assignments FOR ALL USING (
  EXISTS (
    SELECT 1 FROM public.tasks t 
    WHERE t.id = task_id AND (
      t.team_id = (SELECT p.team_id FROM public.profiles p WHERE p.id = auth.uid()) OR
      (SELECT p.role FROM public.profiles p WHERE p.id = auth.uid()) = 'admin'
    )
  )
);
