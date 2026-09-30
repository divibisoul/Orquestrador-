create table if not exists public.n07_learning_experiences (
  id uuid primary key default gen_random_uuid(),
  trace_id text not null,
  correlation_id text not null,
  source text not null,
  target text,
  capability text not null,
  event_type text not null check (event_type in ('supervised', 'feedback')),
  outcome text,
  reward double precision not null default 0,
  confidence double precision not null default 0,
  input jsonb,
  target_vector jsonb,
  provenance text not null default 'unknown',
  metadata jsonb not null default '{}'::jsonb,
  created_at timestamptz not null default now()
);

create index if not exists n07_learning_experiences_trace_idx
  on public.n07_learning_experiences(trace_id);
create index if not exists n07_learning_experiences_correlation_idx
  on public.n07_learning_experiences(correlation_id);
create index if not exists n07_learning_experiences_source_capability_idx
  on public.n07_learning_experiences(source, capability);
create index if not exists n07_learning_experiences_created_at_idx
  on public.n07_learning_experiences(created_at asc);

alter table public.n07_learning_experiences enable row level security;
revoke all on table public.n07_learning_experiences from anon, authenticated;
grant all on table public.n07_learning_experiences to service_role;
