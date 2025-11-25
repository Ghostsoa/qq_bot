package user

// UserService 用户服务
type UserService struct {
	allowedQQs []int64
}

// NewUserService 创建用户服务
func NewUserService(allowedQQs []int64) *UserService {
	return &UserService{
		allowedQQs: allowedQQs,
	}
}

// CheckPermission 检查QQ号是否在白名单中
func (s *UserService) CheckPermission(qqId int64) bool {
	for _, allowedQQ := range s.allowedQQs {
		if allowedQQ == qqId {
			return true
		}
	}
	return false
}

// IsAllowed 检查是否允许使用 (别名方法)
func (s *UserService) IsAllowed(qqId int64) bool {
	return s.CheckPermission(qqId)
}

// UpdateAllowedQQs 更新白名单（热更新配置时使用）
func (s *UserService) UpdateAllowedQQs(allowedQQs []int64) {
	s.allowedQQs = allowedQQs
}

// GetAllowedQQs 获取当前白名单
func (s *UserService) GetAllowedQQs() []int64 {
	return s.allowedQQs
}
