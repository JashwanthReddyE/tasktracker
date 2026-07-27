export interface Category {
  id: string;
  label: string;
  position: number;
  team_id: string;
}
export interface Profile {
  id: string;
  name: string;
  email: string;
  hue: number;
  status: 'pending' | 'approved';
  role: 'admin' | 'user';
  team_id: string | null;
}
export interface Team {
  id: string;
  name: string;
}
export interface Event {
  id: string;
  type: string;
  time: string;
  text: string;
}
export interface Task {
  id: string;
  title: string;
  notes: string;
  priority: string;
  status: string;
  category_id: string;
  team_id: string;
  due_date: string;
  created_at: string;
  archived: boolean;
  position: number;
  task_assignments: { user_id: string }[];
  events: Event[];
}
