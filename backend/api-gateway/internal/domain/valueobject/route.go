package valueobject

import (
	"regexp"
	"strings"

	"github.com/noah-loop/backend/shared/pkg/domain"
)

// Route 路由值对象
type Route struct {
	pattern     string
	serviceName string
	method      string
	pathRewrite string
	middleware  []string
	regex       *regexp.Regexp
}

// RouteConfig 路由配置
type RouteConfig struct {
	Pattern     string
	ServiceName string
	Method      string // GET, POST, PUT, DELETE, ANY
	PathRewrite string
	Middleware  []string
}

// NewRoute 创建路由值对象
func NewRoute(config RouteConfig) (*Route, error) {
	if config.Pattern == "" {
		return nil, domain.NewDomainError("INVALID_ROUTE_PATTERN", "Route pattern cannot be empty")
	}
	
	if config.ServiceName == "" {
		return nil, domain.NewDomainError("INVALID_SERVICE_NAME", "Service name cannot be empty")
	}
	
	if config.Method == "" {
		config.Method = "ANY"
	}
	
	// 将路径模式转换为正则表达式
	regexPattern := convertToRegex(config.Pattern)
	regex, err := regexp.Compile(regexPattern)
	if err != nil {
		return nil, domain.NewDomainError("INVALID_ROUTE_REGEX", "Invalid route pattern: "+err.Error())
	}
	
	if config.Middleware == nil {
		config.Middleware = make([]string, 0)
	}
	
	return &Route{
		pattern:     config.Pattern,
		serviceName: config.ServiceName,
		method:      config.Method,
		pathRewrite: config.PathRewrite,
		middleware:  config.Middleware,
		regex:       regex,
	}, nil
}

// convertToRegex 将路径模式转换为正则表达式
func convertToRegex(pattern string) string {
	// 处理路径参数，如 /users/{id} -> /users/([^/]+)
	regex := pattern
	
	// 替换路径参数
	regex = regexp.MustCompile(`\{([^}]+)\}`).ReplaceAllString(regex, `([^/]+)`)
	
	// 处理通配符，如 /api/v1/* -> /api/v1/.*
	regex = strings.ReplaceAll(regex, "/*", "/.*")
	
	// 确保完全匹配
	if !strings.HasPrefix(regex, "^") {
		regex = "^" + regex
	}
	if !strings.HasSuffix(regex, "$") {
		regex = regex + "$"
	}
	
	return regex
}

// Matches 检查路径是否匹配
func (r *Route) Matches(path string) bool {
	return r.regex.MatchString(path)
}

// MatchesMethod 检查方法是否匹配
func (r *Route) MatchesMethod(method string) bool {
	return r.method == "ANY" || strings.EqualFold(r.method, method)
}

// GetPattern 获取路由模式
func (r *Route) GetPattern() string {
	return r.pattern
}

// GetServiceName 获取服务名称
func (r *Route) GetServiceName() string {
	return r.serviceName
}

// GetMethod 获取HTTP方法
func (r *Route) GetMethod() string {
	return r.method
}

// GetPathRewrite 获取路径重写规则
func (r *Route) GetPathRewrite() string {
	return r.pathRewrite
}

// GetMiddleware 获取中间件列表
func (r *Route) GetMiddleware() []string {
	// 返回副本以避免外部修改
	middleware := make([]string, len(r.middleware))
	copy(middleware, r.middleware)
	return middleware
}

// RewritePath 重写路径
func (r *Route) RewritePath(originalPath string) string {
	if r.pathRewrite == "" {
		return originalPath
	}
	
	// 提取路径参数
	matches := r.regex.FindStringSubmatch(originalPath)
	if len(matches) <= 1 {
		return r.pathRewrite
	}
	
	// 替换路径参数
	rewrittenPath := r.pathRewrite
	for i, match := range matches[1:] {
		placeholder := "{" + string(rune(i+1)) + "}"
		rewrittenPath = strings.ReplaceAll(rewrittenPath, placeholder, match)
	}
	
	return rewrittenPath
}

// ExtractPathParams 提取路径参数
func (r *Route) ExtractPathParams(path string) map[string]string {
	params := make(map[string]string)
	
	matches := r.regex.FindStringSubmatch(path)
	if len(matches) <= 1 {
		return params
	}
	
	// 这里简化处理，实际需要根据路由模式提取参数名
	for i, match := range matches[1:] {
		params["param"+string(rune(i+1))] = match
	}
	
	return params
}

// Equals 检查两个路由是否相等
func (r *Route) Equals(other *Route) bool {
	if other == nil {
		return false
	}
	
	return r.pattern == other.pattern &&
		r.serviceName == other.serviceName &&
		r.method == other.method &&
		r.pathRewrite == other.pathRewrite
}

// String 字符串表示
func (r *Route) String() string {
	return r.method + " " + r.pattern + " -> " + r.serviceName
}
