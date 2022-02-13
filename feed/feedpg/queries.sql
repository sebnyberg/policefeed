-- name: ListEvents :many
select *
from police_event
where id = any (@ids::uuid[]);
