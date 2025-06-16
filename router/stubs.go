package router

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

// 欠けているメソッドのスタブ実装

// AdminDeleteImage 画像を削除（管理者用）
func (s *Server) AdminDeleteImage(e echo.Context, imageId ImageId, params AdminDeleteImageParams) error {
	return echo.NewHTTPError(http.StatusNotImplemented, "AdminDeleteImage not implemented")
}

// AdminPutImage 画像を更新（管理者用）
func (s *Server) AdminPutImage(e echo.Context, imageId ImageId, params AdminPutImageParams) error {
	return echo.NewHTTPError(http.StatusNotImplemented, "AdminPutImage not implemented")
}

// AdminPostImage 画像をアップロード（管理者用）
func (s *Server) AdminPostImage(e echo.Context, campId CampId, params AdminPostImageParams) error {
	return echo.NewHTTPError(http.StatusNotImplemented, "AdminPostImage not implemented")
}

// AdminGetPayments 支払い情報の一覧を取得（管理者用）
func (s *Server) AdminGetPayments(e echo.Context, campId CampId, params AdminGetPaymentsParams) error {
	return echo.NewHTTPError(http.StatusNotImplemented, "AdminGetPayments not implemented")
}

// AdminPostPayment 支払い情報を作成（管理者用）
func (s *Server) AdminPostPayment(e echo.Context, campId CampId, params AdminPostPaymentParams) error {
	return echo.NewHTTPError(http.StatusNotImplemented, "AdminPostPayment not implemented")
}

// AdminPutPayment 支払い情報を更新（管理者用）
func (s *Server) AdminPutPayment(e echo.Context, paymentId PaymentId, params AdminPutPaymentParams) error {
	return echo.NewHTTPError(http.StatusNotImplemented, "AdminPutPayment not implemented")
}

// AdminPostRollCall 点呼を作成（管理者用）
func (s *Server) AdminPostRollCall(e echo.Context, campId CampId, params AdminPostRollCallParams) error {
	return echo.NewHTTPError(http.StatusNotImplemented, "AdminPostRollCall not implemented")
}

// AdminPostRoomGroup 部屋グループを作成（管理者用）
func (s *Server) AdminPostRoomGroup(e echo.Context, campId CampId, params AdminPostRoomGroupParams) error {
	return echo.NewHTTPError(http.StatusNotImplemented, "AdminPostRoomGroup not implemented")
}

// AdminDeleteRoomGroup 部屋グループを削除（管理者用）
func (s *Server) AdminDeleteRoomGroup(e echo.Context, roomGroupId RoomGroupId, params AdminDeleteRoomGroupParams) error {
	return echo.NewHTTPError(http.StatusNotImplemented, "AdminDeleteRoomGroup not implemented")
}

// AdminPutRoomGroup 部屋グループを更新（管理者用）
func (s *Server) AdminPutRoomGroup(e echo.Context, roomGroupId RoomGroupId, params AdminPutRoomGroupParams) error {
	return echo.NewHTTPError(http.StatusNotImplemented, "AdminPutRoomGroup not implemented")
}

// AdminDeleteRoom 部屋を削除（管理者用）
func (s *Server) AdminDeleteRoom(e echo.Context, roomId RoomId, params AdminDeleteRoomParams) error {
	return echo.NewHTTPError(http.StatusNotImplemented, "AdminDeleteRoom not implemented")
}

// AdminDeleteOption 選択肢を削除（管理者用）
func (s *Server) AdminDeleteOption(e echo.Context, optionId OptionId, params AdminDeleteOptionParams) error {
	return echo.NewHTTPError(http.StatusNotImplemented, "AdminDeleteOption not implemented")
}

// AdminGetAnswers 回答の一覧を取得（管理者用）
func (s *Server) AdminGetAnswers(e echo.Context, questionId QuestionId, params AdminGetAnswersParams) error {
	return echo.NewHTTPError(http.StatusNotImplemented, "AdminGetAnswers not implemented")
}

// AdminPutAnswer 管理者が回答を更新
func (s *Server) AdminPutAnswer(e echo.Context, answerId AnswerId, params AdminPutAnswerParams) error {
	return echo.NewHTTPError(http.StatusNotImplemented, "AdminPutAnswer not implemented")
}

// AdminGetUser ユーザー詳細を取得（管理者用）
func (s *Server) AdminGetUser(e echo.Context, traqId string, params AdminGetUserParams) error {
	return echo.NewHTTPError(http.StatusNotImplemented, "AdminGetUser not implemented")
}

// GetImages 画像の一覧を取得
func (s *Server) GetImages(e echo.Context, campId CampId) error {
	return echo.NewHTTPError(http.StatusNotImplemented, "GetImages not implemented")
}

// GetImage 画像を取得
func (s *Server) GetImage(e echo.Context, imageId ImageId) error {
	return echo.NewHTTPError(http.StatusNotImplemented, "GetImage not implemented")
}

// GetDashboard 自分の参加する合宿を取得
func (s *Server) GetDashboard(e echo.Context, campId CampId, params GetDashboardParams) error {
	return echo.NewHTTPError(http.StatusNotImplemented, "GetDashboard not implemented")
}

// GetCampParticipants 合宿の参加者一覧を取得
func (s *Server) GetCampParticipants(e echo.Context, campId CampId) error {
	return echo.NewHTTPError(http.StatusNotImplemented, "GetCampParticipants not implemented")
}

// DeleteCampRegister 合宿登録を削除
func (s *Server) DeleteCampRegister(e echo.Context, campId CampId, params DeleteCampRegisterParams) error {
	return echo.NewHTTPError(http.StatusNotImplemented, "DeleteCampRegister not implemented")
}

// GetRollCalls 点呼の一覧を取得
func (s *Server) GetRollCalls(e echo.Context, campId CampId) error {
	return echo.NewHTTPError(http.StatusNotImplemented, "GetRollCalls not implemented")
}

// GetRoomGroups 部屋グループの一覧を取得
func (s *Server) GetRoomGroups(e echo.Context, campId CampId) error {
	return echo.NewHTTPError(http.StatusNotImplemented, "GetRoomGroups not implemented")
}

// GetMyAnswers 自分の回答一覧を取得
func (s *Server) GetMyAnswers(e echo.Context, questionId QuestionId, params GetMyAnswersParams) error {
	return echo.NewHTTPError(http.StatusNotImplemented, "GetMyAnswers not implemented")
}

// GetAnswers 回答の一覧を取得
func (s *Server) GetAnswers(e echo.Context, questionId QuestionId) error {
	return echo.NewHTTPError(http.StatusNotImplemented, "GetAnswers not implemented")
}

// DeleteReaction リアクションを削除
func (s *Server) DeleteReaction(e echo.Context, rollCallId RollCallId, params DeleteReactionParams) error {
	return echo.NewHTTPError(http.StatusNotImplemented, "DeleteReaction not implemented")
}

// PutReaction リアクションを更新
func (s *Server) PutReaction(e echo.Context, rollCallId RollCallId, params PutReactionParams) error {
	return echo.NewHTTPError(http.StatusNotImplemented, "PutReaction not implemented")
}

// GetRollCallReactions 点呼のリアクション一覧を取得
func (s *Server) GetRollCallReactions(e echo.Context, rollCallId RollCallId) error {
	return echo.NewHTTPError(http.StatusNotImplemented, "GetRollCallReactions not implemented")
}

// PostRollCallReaction 点呼にリアクション
func (s *Server) PostRollCallReaction(e echo.Context, rollCallId RollCallId, params PostRollCallReactionParams) error {
	return echo.NewHTTPError(http.StatusNotImplemented, "PostRollCallReaction not implemented")
}

// StreamRollCallReactions 点呼のリアクションをストリーミング
func (s *Server) StreamRollCallReactions(e echo.Context, rollCallId RollCallId) error {
	return echo.NewHTTPError(http.StatusNotImplemented, "StreamRollCallReactions not implemented")
}
