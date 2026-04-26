package handler

import (
	"context"
	"errors"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/authorization/internal/common"
	pb "github.com/go-park-mail-ru/2026_1_Clac_Clac/pkg/contracts/auth"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"golang.org/x/oauth2"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	msgInternalError = "something went wrong"

	msgErrorParseUserLink  = "can not parse to uuid user link"
	msgDoesNotExistSession = "session does not exist"

	msgVKExchangeFailed    = "vk oauth exchange failed"
	msgVKNoEmailProvided   = "vk oauth: no email in token"
)

type AuthService interface {
	CreateSession(ctx context.Context, userLink uuid.UUID) (string, error)
	GetUserLink(ctx context.Context, sessionID string) (string, error)
	DeleteSession(ctx context.Context, sessionID string) error
	ExtendSession(ctx context.Context, sessionID string) error
}

type VkOAuth interface {
	Exchange(ctx context.Context, code string, opts ...oauth2.AuthCodeOption) (*oauth2.Token, error)
}

type Handler struct {
	srv     AuthService
	vkOAuth VkOAuth
	pb.UnimplementedAuthServiceServer
}

func NewHandler(srv AuthService, vkOAuth VkOAuth) *Handler {
	return &Handler{
		srv:     srv,
		vkOAuth: vkOAuth,
	}
}

func (h *Handler) CreateSession(ctx context.Context, req *pb.CreateSessionRequest) (*pb.CreateSessionResponse, error) {
	convertedLink, err := uuid.Parse(req.UserLink)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, msgErrorParseUserLink)
	}

	sessionID, err := h.srv.CreateSession(ctx, convertedLink)
	if err != nil {
		return nil, status.Error(codes.Internal, msgInternalError)
	}

	return &pb.CreateSessionResponse{
		SessionId: sessionID,
	}, nil
}

func (h *Handler) GetUserLink(ctx context.Context, req *pb.GetUserLinkRequest) (*pb.GetUserLinkResponse, error) {
	userLink, err := h.srv.GetUserLink(ctx, req.SessionId)
	if err != nil {
		if errors.Is(err, common.ErrorNotExistingSession) {
			return nil, status.Error(codes.NotFound, msgDoesNotExistSession)
		}

		return nil, status.Error(codes.Internal, msgInternalError)
	}

	return &pb.GetUserLinkResponse{
		UserLink: userLink,
	}, nil
}

func (h *Handler) DeleteSession(ctx context.Context, req *pb.DeleteSessionRequest) (*pb.DeleteSessionResponse, error) {
	err := h.srv.DeleteSession(ctx, req.SessionId)
	if err != nil {
		return nil, status.Error(codes.Internal, msgInternalError)
	}

	return &pb.DeleteSessionResponse{}, nil
}

func (h *Handler) ExtendSession(ctx context.Context, req *pb.ExtendSessionRequest) (*pb.ExtendSessionResponse, error) {
	err := h.srv.ExtendSession(ctx, req.SessionId)
	if err != nil {
		if errors.Is(err, common.ErrorNotExistingSession) {
			return nil, status.Error(codes.NotFound, msgDoesNotExistSession)
		}

		return nil, status.Error(codes.Internal, msgInternalError)
	}

	return &pb.ExtendSessionResponse{}, nil
}

func (h *Handler) ExchangeVKCode(ctx context.Context, req *pb.ExchangeVKCodeRequest) (*pb.ExchangeVKCodeResponse, error) {
	logger := zerolog.Ctx(ctx)

	token, err := h.vkOAuth.Exchange(ctx, req.Code)
	if err != nil {
		logger.Err(err).Msg("vk oauth exchange failed")
		return nil, status.Error(codes.Unavailable, msgVKExchangeFailed)
	}

	rawEmail := token.Extra("email")
	if rawEmail == nil {
		return nil, status.Error(codes.Unavailable, msgVKNoEmailProvided)
	}

	email, ok := rawEmail.(string)
	if !ok || email == "" {
		return nil, status.Error(codes.Unavailable, msgVKNoEmailProvided)
	}

	return &pb.ExchangeVKCodeResponse{
		AccessToken: token.AccessToken,
		Email:       email,
	}, nil
}
