create extension if not exists vector;
create table if not exists public.memories (
  id uuid primary key default gen_random_uuid(),
  user_id uuid not null,
  session_id uuid not null,
  resumo text not null,
  embedding vector(768) not null,
  tags text[] not null default '{}',
  criado_em timestamptz not null default now()
);
create index if not exists memories_user_idx on public.memories(user_id, criado_em desc);
create index if not exists memories_embedding_ivfflat on public.memories using ivfflat (embedding vector_cosine_ops);
create or replace function public.match_memories(query_embedding vector(768), match_user_id uuid, match_threshold double precision default 0.75, match_count integer default 5)
returns table (id uuid, user_id uuid, session_id uuid, resumo text, tags text[], criado_em timestamptz, similarity double precision)
language sql stable set search_path = public as $$
  select m.id,m.user_id,m.session_id,m.resumo,m.tags,m.criado_em,1-(m.embedding <=> query_embedding) as similarity
  from public.memories m
  where m.user_id=match_user_id and 1-(m.embedding <=> query_embedding)>match_threshold
  order by m.embedding <=> query_embedding
  limit least(greatest(match_count,1),50);
$$;
alter table public.memories enable row level security;
revoke all on public.memories from anon, authenticated;
grant all on public.memories to service_role;