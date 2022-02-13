-- name: UpdatePoliceEvent :exec
insert into police_event (
  id,
  url,
  title,
  region,
  description,
  publish_time,
  create_time,
  revision
) values (
  @id,
  @url,
  @title,
  @region,
  @description,
  @publish_time,
  @create_time,
  @revision
);


