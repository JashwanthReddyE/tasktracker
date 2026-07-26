-- Create profiles table linked to auth.users
CREATE TABLE public.profiles (
  id UUID REFERENCES auth.users(id) ON DELETE CASCADE PRIMARY KEY,
  name TEXT NOT NULL,
  email TEXT NOT NULL,
  hue INTEGER DEFAULT 220,
  created_at TIMESTAMP WITH TIME ZONE DEFAULT timezone('utc'::text, now())
);

ALTER TABLE public.profiles ENABLE ROW LEVEL SECURITY;

-- Everyone can view profiles
CREATE POLICY "Profiles are viewable by everyone" 
  ON public.profiles FOR SELECT USING (true);

-- Users can update their own profile
CREATE POLICY "Users can update own profile" 
  ON public.profiles FOR UPDATE USING (auth.uid() = id);

-- Function to handle new user signup and create a profile automatically
CREATE OR REPLACE FUNCTION public.handle_new_user()
RETURNS trigger AS $$
BEGIN
  INSERT INTO public.profiles (id, email, name, hue)
  VALUES (
    new.id,
    new.email,
    COALESCE(new.raw_user_meta_data->>'name', 'Unknown User'),
    floor(random() * 360)
  );
  RETURN new;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- Trigger for new user signup
DROP TRIGGER IF EXISTS on_auth_user_created ON auth.users;
CREATE TRIGGER on_auth_user_created
  AFTER INSERT ON auth.users
  FOR EACH ROW EXECUTE PROCEDURE public.handle_new_user();

-- Drop old people tables since we are moving to global profiles
DROP TABLE IF EXISTS public.task_people;
DROP TABLE IF EXISTS public.people;

-- Create new task assignments table mapping tasks to profiles
CREATE TABLE public.task_assignments (
  task_id UUID REFERENCES public.tasks(id) ON DELETE CASCADE,
  user_id UUID REFERENCES public.profiles(id) ON DELETE CASCADE,
  assigned_by UUID REFERENCES auth.users(id) ON DELETE SET NULL,
  created_at TIMESTAMP WITH TIME ZONE DEFAULT timezone('utc'::text, now()),
  PRIMARY KEY (task_id, user_id)
);

ALTER TABLE public.task_assignments ENABLE ROW LEVEL SECURITY;

-- Anyone can see an assignment if they can see the task or they are the assignee
CREATE POLICY "Assignments viewable by assignee or task owner" 
  ON public.task_assignments FOR SELECT USING (
    user_id = auth.uid() OR 
    assigned_by = auth.uid() OR
    EXISTS (SELECT 1 FROM public.tasks WHERE id = task_id AND tasks.user_id = auth.uid())
  );

-- Task owners can manage assignments
CREATE POLICY "Task owners can manage assignments" 
  ON public.task_assignments FOR ALL USING (
    EXISTS (SELECT 1 FROM public.tasks WHERE id = task_id AND tasks.user_id = auth.uid())
  );

-- Update RLS for tasks to allow assignees to see them
DROP POLICY IF EXISTS "Users can view their own tasks" ON public.tasks;
CREATE POLICY "Users can view assigned and owned tasks" 
  ON public.tasks FOR SELECT USING (
    user_id = auth.uid() OR
    id IN (SELECT task_id FROM public.task_assignments WHERE user_id = auth.uid())
  );

-- Update RLS for categories to allow assignees to see the categories of their tasks
DROP POLICY IF EXISTS "Users can view their own categories" ON public.categories;
CREATE POLICY "Users can view their own categories or categories with assigned tasks" 
  ON public.categories FOR SELECT USING (
    user_id = auth.uid() OR
    id IN (
      SELECT category_id FROM public.tasks 
      WHERE id IN (SELECT task_id FROM public.task_assignments WHERE user_id = auth.uid())
    )
  );
