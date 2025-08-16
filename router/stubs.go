package router

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/traPtitech/rucQ/api"
)

// 欠けているメソッドのスタブ実装

// AdminDeleteImage 画像を削除（管理者用）
func (s *Server) AdminDeleteImage(
	_ echo.Context,
	_ api.ImageId,
	_ api.AdminDeleteImageParams,
) error {
	return echo.NewHTTPError(http.StatusNotImplemented, "AdminDeleteImage not implemented")
}

// AdminPutImage 画像を更新（管理者用）
func (s *Server) AdminPutImage(_ echo.Context, _ api.ImageId, _ api.AdminPutImageParams) error {
	return echo.NewHTTPError(http.StatusNotImplemented, "AdminPutImage not implemented")
}

// AdminPostImage 画像をアップロード（管理者用）
func (s *Server) AdminPostImage(_ echo.Context, _ api.CampId, _ api.AdminPostImageParams) error {
	return echo.NewHTTPError(http.StatusNotImplemented, "AdminPostImage not implemented")
}

// AdminPostRollCall 点呼を作成（管理者用）
func (s *Server) AdminPostRollCall(
	_ echo.Context,
	_ api.CampId,
	_ api.AdminPostRollCallParams,
) error {
	return echo.NewHTTPError(http.StatusNotImplemented, "AdminPostRollCall not implemented")
}

// AdminDeleteRoomGroup 部屋グループを削除（管理者用）
func (s *Server) AdminDeleteRoomGroup(
	_ echo.Context,
	_ api.RoomGroupId,
	_ api.AdminDeleteRoomGroupParams,
) error {
	return echo.NewHTTPError(http.StatusNotImplemented, "AdminDeleteRoomGroup not implemented")
}

// AdminDeleteRoom 部屋を削除（管理者用）
func (s *Server) AdminDeleteRoom(_ echo.Context, _ api.RoomId, _ api.AdminDeleteRoomParams) error {
	return echo.NewHTTPError(http.StatusNotImplemented, "AdminDeleteRoom not implemented")
}

// AdminGetUser ユーザー詳細を取得（管理者用）
func (s *Server) AdminGetUser(_ echo.Context, _ string, _ api.AdminGetUserParams) error {
	return echo.NewHTTPError(http.StatusNotImplemented, "AdminGetUser not implemented")
}

// GetImages 画像の一覧を取得
func (s *Server) GetImages(_ echo.Context, _ api.CampId) error {
	return echo.NewHTTPError(http.StatusNotImplemented, "GetImages not implemented")
}

// GetImage 画像を取得
func (s *Server) GetImage(_ echo.Context, _ api.ImageId) error {
	return echo.NewHTTPError(http.StatusNotImplemented, "GetImage not implemented")
}

// GetRollCalls 点呼の一覧を取得
func (s *Server) GetRollCalls(_ echo.Context, _ api.CampId) error {
	return echo.NewHTTPError(http.StatusNotImplemented, "GetRollCalls not implemented")
}

// DeleteReaction リアクションを削除
func (s *Server) DeleteReaction(
	_ echo.Context,
	_ api.RollCallId,
	_ api.DeleteReactionParams,
) error {
	return echo.NewHTTPError(http.StatusNotImplemented, "DeleteReaction not implemented")
}

// PutReaction リアクションを更新
func (s *Server) PutReaction(_ echo.Context, _ api.RollCallId, _ api.PutReactionParams) error {
	return echo.NewHTTPError(http.StatusNotImplemented, "PutReaction not implemented")
}

// GetRollCallReactions 点呼のリアクション一覧を取得
func (s *Server) GetRollCallReactions(_ echo.Context, _ api.RollCallId) error {
	return echo.NewHTTPError(http.StatusNotImplemented, "GetRollCallReactions not implemented")
}

// PostRollCallReaction 点呼にリアクション
func (s *Server) PostRollCallReaction(
	_ echo.Context,
	_ api.RollCallId,
	_ api.PostRollCallReactionParams,
) error {
	return echo.NewHTTPError(http.StatusNotImplemented, "PostRollCallReaction not implemented")
}

// StreamRollCallReactions 点呼のリアクションをストリーミング
func (s *Server) StreamRollCallReactions(_ echo.Context, _ api.RollCallId) error {
	return echo.NewHTTPError(http.StatusNotImplemented, "StreamRollCallReactions not implemented")
}
