begin;

create table if not exists police_event (
  id uuid not null,
  url text not null,
  title text not null,
  region text not null,
  description text not null,
  publish_time timestamptz not null,
  create_time timestamptz not null,
  content_hash bytea not null,
  revision int not null,
  constraint police_event_pk
    primary key (id, revision)
);

end transaction;