package constants

const (
	MessageOK             = "ok"
	MessageUnauthorized   = "登录已失效"
	MessageForbidden      = "无权限执行此操作"
	MessageValidation     = "请求参数不合法"
	MessageNotFound       = "资源不存在"
	MessagePaymentSuccess = "支付宝沙箱支付成功"
	MessageRepairCreated  = "报修工单已提交"

	MessageRepairNotOwner     = "仅报修人可验收该工单"
	MessageRepairNotAccepting = "仅待验收工单可执行验收操作"
	MessageRepairCloseByOwner = "工单需报修人验收确认后关闭"
)
