-- Add Uniqueness Constraints
-- Note: If you have existing duplicate data, these ALTER TABLE statements will fail.
-- You must manually resolve existing duplicates before running this script.

-- 1. Tasks: Unique titles per team
ALTER TABLE public.tasks ADD CONSTRAINT tasks_team_title_unique UNIQUE (team_id, title);

-- 2. Categories: Unique labels per team
ALTER TABLE public.categories ADD CONSTRAINT categories_team_label_unique UNIQUE (team_id, label);

-- 3. Teams: Globally unique names
ALTER TABLE public.teams ADD CONSTRAINT teams_name_unique UNIQUE (name);

-- 4. Profiles: Globally unique user names
ALTER TABLE public.profiles ADD CONSTRAINT profiles_name_unique UNIQUE (name);


-- Update handle_new_user trigger to safely handle duplicate names during OAuth signups
-- This ensures if "John Doe" signs up but "John Doe" is already taken, they get "John Doe_a1b2" instead of failing.
CREATE OR REPLACE FUNCTION public.handle_new_user() 
RETURNS TRIGGER AS $$
DECLARE
    base_name TEXT;
    final_name TEXT;
BEGIN
    base_name := COALESCE(new.raw_user_meta_data->>'name', 'Unknown User');
    final_name := base_name;
    
    -- Loop to append a random 4-character suffix if the name is taken
    WHILE EXISTS (SELECT 1 FROM public.profiles WHERE name = final_name) LOOP
        final_name := base_name || '_' || substr(md5(random()::text), 1, 4);
    END LOOP;

    INSERT INTO public.profiles (id, email, name, role, status, team_id)
    VALUES (
      new.id, 
      new.email, 
      final_name,
      'user',
      'pending',
      '00000000-0000-0000-0000-000000000000'
    );
    RETURN new;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;
