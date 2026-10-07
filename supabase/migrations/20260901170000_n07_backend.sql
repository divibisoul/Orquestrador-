create extension if not exists pgcrypto;

create table if not exists public.n07_runs (
  id uuid primary key default gen_random_uuid(),
  trace_id text not null unique,
  correlation_id text not null,
  source text not null,
  operation text not null,
  status text not null,
  payload jsonb,
  result jsonb,
  error text,
  metadata jsonb,
  created_at timestamptz not null default now()
);

create index if not exists n07_runs_correlation_idx on public.n07_runs(correlation_id);
create index if not exists n07_runs_operation_idx on public.n07_runs(operation);
create index if not exists n07_runs_created_at_idx on public.n07_runs(created_at desc);

alter table public.n07_runs enable row level security;
revoke all on table public.n07_runs from anon, authenticated;
grant all on table public.n07_runs to service_role;

create table if not exists public.n07_artifacts (
  id uuid primary key default gen_random_uuid(),
  cid text not null unique,
  filename text,
  size_bytes bigint,
  metadata jsonb,
  created_at timestamptz not null default now()
);

create index if not exists n07_artifacts_created_at_idx on public.n07_artifacts(created_at desc);

alter table public.n07_artifacts enable row level security;
revoke all on table public.n07_artifacts from anon, authenticated;
grant all on table public.n07_artifacts to service_role;
