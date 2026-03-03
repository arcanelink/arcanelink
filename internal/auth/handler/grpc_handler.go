package handler

import (
	"context"

	"github.com/arcane/arcanelink/internal/auth/service"
	"github.com/arcane/arcanelink/pkg/models"
	pb "github.com/arcane/arcanelink/pkg/proto/auth"
)

type GRPCHandler struct {
	pb.UnimplementedAuthServiceServer
	authService *service.AuthService
}

func NewGRPCHandler(authService *service.AuthService) *GRPCHandler {
	return &GRPCHandler{
		authService: authService,
	}
}

func (h *GRPCHandler) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	resp, err := h.authService.Login(req.UserId, req.Password)
	if err != nil {
		return nil, err
	}

	return &pb.LoginResponse{
		AccessToken: resp.AccessToken,
		UserId:      resp.UserID,
		ExpiresIn:   resp.ExpiresIn,
	}, nil
}

func (h *GRPCHandler) Logout(ctx context.Context, req *pb.LogoutRequest) (*pb.LogoutResponse, error) {
	err := h.authService.Logout(req.UserId)
	if err != nil {
		return &pb.LogoutResponse{Success: false}, err
	}

	return &pb.LogoutResponse{Success: true}, nil
}

func (h *GRPCHandler) ValidateToken(ctx context.Context, req *pb.ValidateTokenRequest) (*pb.ValidateTokenResponse, error) {
	userID, err := h.authService.ValidateToken(req.Token)
	if err != nil {
		return &pb.ValidateTokenResponse{Valid: false}, nil
	}

	return &pb.ValidateTokenResponse{
		Valid:  true,
		UserId: userID,
	}, nil
}

func (h *GRPCHandler) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.CreateUserResponse, error) {
	createReq := &models.CreateUserRequest{
		UserID:   req.UserId,
		Username: req.Username,
		Password: req.Password,
	}

	user, err := h.authService.CreateUser(createReq)
	if err != nil {
		return nil, err
	}

	return &pb.CreateUserResponse{
		UserId:    user.UserID,
		CreatedAt: user.CreatedAt.Unix(),
	}, nil
}

func (h *GRPCHandler) GetUserProfile(ctx context.Context, req *pb.GetUserProfileRequest) (*pb.GetUserProfileResponse, error) {
	profile, err := h.authService.GetUserProfile(req.UserId)
	if err != nil {
		return nil, err
	}

	return &pb.GetUserProfileResponse{
		UserId:      profile.UserID,
		DisplayName: profile.DisplayName,
		AvatarUrl:   profile.AvatarURL,
		StatusMsg:   profile.StatusMsg,
	}, nil
}

func (h *GRPCHandler) UpdateUserProfile(ctx context.Context, req *pb.UpdateUserProfileRequest) (*pb.UpdateUserProfileResponse, error) {
	updateReq := &models.UpdateProfileRequest{
		DisplayName: req.DisplayName,
		AvatarURL:   req.AvatarUrl,
		StatusMsg:   req.StatusMsg,
	}

	err := h.authService.UpdateUserProfile(req.UserId, updateReq)
	if err != nil {
		return &pb.UpdateUserProfileResponse{Success: false}, err
	}

	return &pb.UpdateUserProfileResponse{Success: true}, nil
}
