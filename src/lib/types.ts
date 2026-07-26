export interface Category {
  id: string;
  label: string;
  position: number;
}
export interface Profile {
  id: string;
  name: string;
  email: string;
  hue: number;
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
  due_date: string;
  created_at: string;
  archived: boolean;
  position: number;
  task_assignments: { user_id: string }[];
  events: Event[];
}
