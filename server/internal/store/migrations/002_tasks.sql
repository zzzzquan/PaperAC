-- 任务编排表结构

create table if not exists tasks (
  id uuid primary key,
  user_id uuid not null references users(id),
  status varchar(20) not null,
  progress int not null default 0,
  x numeric(5,4) not null,
  original_filename varchar(255) not null,
  upload_path text not null,
  result_path text null,
  error_message text null,
  created_at timestamp not null default now(),
  updated_at timestamp not null default now(),
  finished_at timestamp null
);

create index if not exists idx_tasks_user_created on tasks(user_id, created_at desc);
