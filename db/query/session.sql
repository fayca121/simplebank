-- name: CreateSession :one
insert into sessions (
    id, username, refresh_token, user_agent, client_ip, expires_at
) VALUES (
             $1,$2,$3,$4,$5,$6
         )
returning *;

-- name: GetSession :one
select * from sessions
where id= $1 limit 1;