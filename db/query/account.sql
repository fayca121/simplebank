-- name: CreateAccount :one
insert into accounts (
 owner, balance, currency
) VALUES (
  $1,$2,$3
)
returning *;

-- name: GetAccount :one
select * from accounts
where id= $1 limit 1;

-- name: GetAccountForUpdate :one
select * from accounts
where id= $1 limit 1
FOR NO KEY UPDATE;

-- name: GetAccounts :many
select * from accounts
order by id
LIMIT $1
offset $2;

-- name: UpdateAccount :one
update accounts
set balance = $2
where id = $1
returning *;

-- name: AddAccountBalance :one
update accounts
set balance = balance + sqlc.arg(amount)
where id = sqlc.arg(id)
returning *;

-- name: DeleteAccount :exec
delete from accounts
where id = $1;
