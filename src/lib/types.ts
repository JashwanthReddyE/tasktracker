export interface Category {
  id: string;
  label: string;
  position: number;
}
export interface Person {
  id: string;
  category_id: string;
  name: string;
  hue: number;
  position: number;
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
  task_people: { person_id: string }[];
  events: Event[];
}
