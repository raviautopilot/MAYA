// ── API Response Envelope ──
export interface APIResponse<T = unknown> {
  data: T;
  error: string;
  status: number;
}

// ── Auth ──
export interface LoginRequest {
  email: string;
  password: string;
}

export interface LoginResponse {
  token: string;
}

export interface ChangePasswordRequest {
  old_password: string;
  new_password: string;
}

// ── Project ──
export interface Project {
  id: string;
  name: string;
  description: string;
  type: 'personal' | 'professional';
  start_date?: string;
  end_date?: string;
  estimated_hours?: number;
  created_at: string;
  updated_at: string;
  deleted_at?: string;
}

export type ProjectCreate = Pick<Project, 'name' | 'description' | 'type'> & {
  start_date?: string;
  end_date?: string;
  estimated_hours?: number;
};
export type ProjectUpdate = ProjectCreate;

// ── Board ──
export interface Board {
  id: string;
  project_id: string;
  name: string;
  swimlanes: string[];
  task_types: string[];
  is_active: boolean;
  created_at: string;
  updated_at: string;
  deleted_at?: string;
}

export type BoardCreate = Pick<Board, 'project_id' | 'name'> & {
  swimlanes?: string[];
  task_types?: string[];
  is_active?: boolean;
};
export type BoardUpdate = BoardCreate;

// ── Task ──
export interface Reminder {
  time: string;
  note: string;
}

export interface Task {
  id: string;
  board_id: string;
  swimlane: string;
  task_type: string;
  title: string;
  description: string;
  assignee_id: string;
  estimation_minutes: number;
  actual_time_minutes: number;
  cost: number;
  priority: 'Low' | 'Medium' | 'High' | 'Critical';
  reminders: Reminder[];
  scheduler_id: string;
  due_date?: string;
  created_at: string;
  updated_at: string;
  deleted_at?: string;
}

export type TaskCreate = Pick<Task, 'board_id' | 'swimlane' | 'task_type' | 'title' | 'priority'> & {
  description?: string;
  assignee_id?: string;
  estimation_minutes?: number;
  cost?: number;
  reminders?: Reminder[];
  scheduler_id?: string;
  due_date?: string;
};
export type TaskUpdate = TaskCreate & { actual_time_minutes?: number };
export type TaskPatch = Partial<Omit<Task, 'id' | 'created_at' | 'updated_at' | 'deleted_at'>>;

// ── Scheduler ──
export interface Scheduler {
  id: string;
  name: string;
  cron_expression: string;
  type: 'cron' | 'yearly' | 'monthly' | 'weekly' | 'daily';
  next_run: string;
  linked_task_template_id: string;
  created_at: string;
  updated_at: string;
  deleted_at?: string;
}

export type SchedulerCreate = Pick<Scheduler, 'name' | 'type'> & {
  cron_expression?: string;
  linked_task_template_id?: string;
};
export type SchedulerUpdate = SchedulerCreate;

// ── Resource ──
export interface Resource {
  id: string;
  name: string;
  type: 'Global' | 'Project' | 'Task';
  linked_items: string[];
  created_at: string;
  updated_at: string;
  deleted_at?: string;
}

export type ResourceCreate = Pick<Resource, 'name' | 'type'> & {
  linked_items?: string[];
};
export type ResourceUpdate = ResourceCreate;
