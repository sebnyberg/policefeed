-- name: ListEvents :many
select *
from police_event
where id = any (@ids::uuid[]);

-- name: ListRecentEvents :many
select distinct on (id) *
from police_event
where id = any (@ids::uuid[])
order by id, revision desc;
