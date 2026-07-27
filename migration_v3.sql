-- MIGRATION V3: Fix Infinite Recursion in RLS

-- 1. Create helper functions that bypass RLS to prevent infinite recursion
CREATE OR REPLACE FUNCTION public.get_user_role()
RETURNS text
LANGUAGE sql
SECURITY DEFINER
SET search_path = public
AS $$
  SELECT role FROM profiles WHERE id = auth.uid();
$$;

CREATE OR REPLACE FUNCTION public.get_user_team()
RETURNS uuid
LANGUAGE sql
SECURITY DEFINER
SET search_path = public
AS $$
  SELECT team_id FROM profiles WHERE id = auth.uid();
$$;

-- 2. Drop the recursive policies
DROP POLICY IF EXISTS "Profiles viewable by team members or admins" ON public.profiles;
DROP POLICY IF EXISTS "Profiles updatable by self or admin" ON public.profiles;

DROP POLICY IF EXISTS "Teams viewable by members or admins" ON public.teams;
DROP POLICY IF EXISTS "Teams manageable by admins" ON public.teams;

DROP POLICY IF EXISTS "Tasks viewable by team members or admins" ON public.tasks;
DROP POLICY IF EXISTS "Tasks editable by team members or admins" ON public.tasks;

DROP POLICY IF EXISTS "Categories viewable by team members or admins" ON public.categories;
DROP POLICY IF EXISTS "Categories editable by team members or admins" ON public.categories;

DROP POLICY IF EXISTS "Assignments viewable by team members" ON public.task_assignments;
DROP POLICY IF EXISTS "Assignments manageable by team members" ON public.task_assignments;

-- 3. Re-create Profiles Policies
CREATE POLICY "Profiles viewable by team members or admins" ON public.profiles FOR SELECT USING (
  team_id = public.get_user_team() OR
  public.get_user_role() = 'admin' OR
  id = auth.uid()
);

CREATE POLICY "Profiles updatable by self or admin" ON public.profiles FOR UPDATE USING (
  id = auth.uid() OR
  public.get_user_role() = 'admin'
);

-- 4. Re-create Teams Policies
CREATE POLICY "Teams viewable by members or admins" ON public.teams FOR SELECT USING (
  id = public.get_user_team() OR
  public.get_user_role() = 'admin'
);
CREATE POLICY "Teams manageable by admins" ON public.teams FOR ALL USING (
  public.get_user_role() = 'admin'
);

-- 5. Re-create Tasks Policies
CREATE POLICY "Tasks viewable by team members or admins" ON public.tasks FOR SELECT USING (
  team_id = public.get_user_team() OR
  public.get_user_role() = 'admin'
);
CREATE POLICY "Tasks editable by team members or admins" ON public.tasks FOR ALL USING (
  team_id = public.get_user_team() OR
  public.get_user_role() = 'admin'
);

-- 6. Re-create Categories Policies
CREATE POLICY "Categories viewable by team members or admins" ON public.categories FOR SELECT USING (
  team_id = public.get_user_team() OR
  public.get_user_role() = 'admin'
);
CREATE POLICY "Categories editable by team members or admins" ON public.categories FOR ALL USING (
  team_id = public.get_user_team() OR
  public.get_user_role() = 'admin'
);

-- 7. Re-create Assignments Policies
CREATE POLICY "Assignments viewable by team members" ON public.task_assignments FOR SELECT USING (
  EXISTS (
    SELECT 1 FROM public.tasks t 
    WHERE t.id = task_id AND (
      t.team_id = public.get_user_team() OR
      public.get_user_role() = 'admin'
    )
  )
);
CREATE POLICY "Assignments manageable by team members" ON public.task_assignments FOR ALL USING (
  EXISTS (
    SELECT 1 FROM public.tasks t 
    WHERE t.id = task_id AND (
      t.team_id = public.get_user_team() OR
      public.get_user_role() = 'admin'
    )
  )
);
