package gapi

import (
	"context"
	"errors"
	db "github.com/fayca121/simplebank/db/sqlc"
	"github.com/fayca121/simplebank/pb"
	"github.com/fayca121/simplebank/util"
	"github.com/fayca121/simplebank/val"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
	"time"
)

func (server *Server) UpdateUser(ctx context.Context, req *pb.UpdateUserRequest) (*pb.UpdateUserResponse, error) {
	//2- check authorization access token
	payload, err := server.authorizeUser(ctx)
	if err != nil {
		return nil, status.Errorf(codes.PermissionDenied, "unauthorized: %s", err)
	}
	if payload.Username != req.Username && payload.Role != util.BankerRole.String() {
		return nil, status.Errorf(codes.PermissionDenied, "you cannot update other user's info")
	}
	//1- validate request
	violations := validateUpdateUserRequest(req)
	if violations != nil {
		return nil, invalidArgumentError(violations)
	}

	//2- create args for db update query
	arg := db.UpdateUserParams{
		FullName: pgtype.Text{
			String: req.GetFullName(),
			Valid:  req.FullName != nil,
		},
		Email: pgtype.Text{
			String: req.GetEmail(),
			Valid:  req.Email != nil,
		},
		Username: req.GetUsername(),
	}
	//3 - check if password has been updated, if so, update arg with hashed password
	if req.Password != nil {
		hashedPassword, err := util.HashPassword(req.GetPassword())
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to hash password: %s", err)
		}
		arg.HashedPassword = pgtype.Text{
			String: hashedPassword,
			Valid:  true,
		}
		arg.PasswordChangedAt = pgtype.Timestamptz{
			Time:  time.Now(),
			Valid: true,
		}
	}
	//4- call update user query
	updatedUser, err := server.store.UpdateUser(ctx, arg)
	//5- check if error and get updated user
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, status.Errorf(codes.NotFound, "user not found : %s", err)
		}
		return nil, status.Errorf(codes.Internal, "failed to update user: %s", err)
	}

	//6- create response
	resp := &pb.UpdateUserResponse{
		User: &pb.User{
			Username:          req.Username,
			FullName:          updatedUser.FullName,
			Email:             updatedUser.Email,
			PasswordChangedAt: timestamppb.New(updatedUser.PasswordChangedAt),
			CreatedAt:         timestamppb.New(updatedUser.CreatedAt),
		},
	}

	return resp, nil
}

func validateUpdateUserRequest(req *pb.UpdateUserRequest) (violations []*errdetails.BadRequest_FieldViolation) {
	if err := val.ValidateUsername(req.GetUsername()); err != nil {
		violations = append(violations, fieldViolation("username", err))
	}
	if req.Password != nil {
		if err := val.ValidatePassword(req.GetPassword()); err != nil {
			violations = append(violations, fieldViolation("password", err))
		}
	}
	if req.FullName != nil {
		if err := val.ValidateFullName(req.GetFullName()); err != nil {
			violations = append(violations, fieldViolation("full_name", err))
		}
	}
	if req.Email != nil {
		if err := val.ValidateEmail(req.GetEmail()); err != nil {
			violations = append(violations, fieldViolation("email", err))
		}
	}
	return
}
