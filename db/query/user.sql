-- name: CreateUser :one
insert into users (
    username, hashed_password, full_name,email
) VALUES (
             $1,$2,$3,$4
         )
returning *;

-- name: GetUser :one
select * from users
where username= $1 limit 1;

-- name: UpdateUser :one
update users
set hashed_password = coalesce(sqlc.narg('hashed_password'),hashed_password),
    full_name = coalesce(sqlc.narg('full_name'),full_name),
    email = coalesce(sqlc.narg('email'),email)
where username = sqlc.arg('username')
returning *;