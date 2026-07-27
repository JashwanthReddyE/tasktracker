-- MIGRATION V4: Multi-Team Architecture & Deny Flow

-- 1. Create team_members table
CREATE TABLE IF NOT EXISTS public.team_members (
  team_id UUID REFERENCES public.teams(id) ON DELETE CASCADE,
  user_id UUID REFERENCES public.profiles(id) ON DELETE CASCADE,
  role TEXT DEFAULT 'member',
  joined_at TIMESTAMP WITH TIME ZONE DEFAULT timezone('utc'::text, now()),
  PRIMARY KEY (team_id, user_id)
);
ALTER TABLE public.team_members ENABLE ROW LEVEL SECURITY;

-- 2. Migrate existing users into team_members
INSERT INTO public.team_members (team_id, user_id, role)
SELECT team_id, id, role FROM public.profiles 
WHERE team_id IS NOT NULL 
ON CONFLICT DO NOTHING;

-- 3. Add requested_team_id to profiles (for the signup flow)
ALTER TABLE public.profiles ADD COLUMN IF NOT EXISTS requested_team_id UUID REFERENCES public.teams(id) ON DELETE SET NULL;

-- 4. Re-write the handle_new_user trigger to include requested_team_id
CREATE OR REPLACE FUNCTION public.handle_new_user()
RETURNS trigger AS $$
DECLARE
  is_admin BOOLEAN;
  req_team UUID;
BEGIN
  is_admin := (new.email ilike 'jashwanthreddyearla@gmail.com');
  
  BEGIN
    req_team := NULLIF(new.raw_user_meta_data->>'requested_team_id', '')::uuid;
  EXCEPTION WHEN OTHERS THEN
    req_team := NULL;
  END;

  INSERT INTO public.profiles (id, email, name, hue, status, role, team_id, requested_team_id)
  VALUES (
    new.id,
    new.email,
    COALESCE(new.raw_user_meta_data->>'name', 'Unknown User'),
    floor(random() * 360),
    CASE WHEN is_admin THEN 'approved' ELSE 'pending' END,
    CASE WHEN is_admin THEN 'admin' ELSE 'user' END,
    CASE WHEN is_admin THEN '00000000-0000-0000-0000-000000000000'::uuid ELSE NULL END,
    req_team
  );
  
  -- If admin, auto-add to team_members for General
  IF is_admin THEN
    INSERT INTO public.team_members (team_id, user_id, role)
    VALUES ('00000000-0000-0000-0000-000000000000', new.id, 'admin')
    ON CONFLICT DO NOTHING;
  END IF;
  
  RETURN new;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- 5. Helper function for RLS
CREATE OR REPLACE FUNCTION public.get_user_teams()
RETURNS SETOF uuid
LANGUAGE sql
SECURITY DEFINER
SET search_path = public
AS $$
  SELECT team_id FROM public.team_members WHERE user_id = auth.uid();
$$;

-- 6. Rewrite Policies
-- Teams (Make public for signup form)
DROP POLICY IF EXISTS "Teams viewable by members or admins" ON public.teams;
CREATE POLICY "Teams viewable by everyone for signup" ON public.teams FOR SELECT USING (true);

-- Team Members
CREATE POLICY "Team members viewable by members or admins" ON public.team_members FOR SELECT USING (
  team_id IN (SELECT public.get_user_teams()) OR
  public.get_user_role() = 'admin'
);
CREATE POLICY "Team members manageable by admins" ON public.team_members FOR ALL USING (
  public.get_user_role() = 'admin'
);

-- Profiles (allow users to see people in any of their teams)
DROP POLICY IF EXISTS "Profiles viewable by team members or admins" ON public.profiles;
CREATE POLICY "Profiles viewable by team members or admins" ON public.profiles FOR SELECT USING (
  EXISTS (
    SELECT 1 FROM public.team_members tm WHERE tm.user_id = profiles.id AND tm.team_id IN (SELECT public.get_user_teams())
  ) OR
  public.get_user_role() = 'admin' OR
  id = auth.uid()
);

-- Tasks
DROP POLICY IF EXISTS "Tasks viewable by team members or admins" ON public.tasks;
DROP POLICY IF EXISTS "Tasks editable by team members or admins" ON public.tasks;
CREATE POLICY "Tasks viewable by team members or admins" ON public.tasks FOR SELECT USING (
  team_id IN (SELECT public.get_user_teams()) OR
  public.get_user_role() = 'admin'
);
CREATE POLICY "Tasks editable by team members or admins" ON public.tasks FOR ALL USING (
  team_id IN (SELECT public.get_user_teams()) OR
  public.get_user_role() = 'admin'
);

-- Categories
DROP POLICY IF EXISTS "Categories viewable by team members or admins" ON public.categories;
DROP POLICY IF EXISTS "Categories editable by team members or admins" ON public.categories;
CREATE POLICY "Categories viewable by team members or admins" ON public.categories FOR SELECT USING (
  team_id IN (SELECT public.get_user_teams()) OR
  public.get_user_role() = 'admin'
);
CREATE POLICY "Categories editable by team members or admins" ON public.categories FOR ALL USING (
  team_id IN (SELECT public.get_user_teams()) OR
  public.get_user_role() = 'admin'
);

-- Assignments
DROP POLICY IF EXISTS "Assignments viewable by team members" ON public.task_assignments;
DROP POLICY IF EXISTS "Assignments manageable by team members" ON public.task_assignments;
CREATE POLICY "Assignments viewable by team members" ON public.task_assignments FOR SELECT USING (
  EXISTS (
    SELECT 1 FROM public.tasks t 
    WHERE t.id = task_id AND (
      t.team_id IN (SELECT public.get_user_teams()) OR
      public.get_user_role() = 'admin'
    )
  )
);
CREATE POLICY "Assignments manageable by team members" ON public.task_assignments FOR ALL USING (
  EXISTS (
    SELECT 1 FROM public.tasks t 
    WHERE t.id = task_id AND (
      t.team_id IN (SELECT public.get_user_teams()) OR
      public.get_user_role() = 'admin'
    )
  )
);
