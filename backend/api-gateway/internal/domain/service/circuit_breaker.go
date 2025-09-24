package service

import (
	"sync"
	"time"

	"github.com/noah-loop/backend/shared/pkg/domain"
)

// CircuitBreakerState 熔断器状态
type CircuitBreakerState int

const (
	StateClosed CircuitBreakerState = iota
	StateHalfOpen
	StateOpen
)

// CircuitBreaker 熔断器领域服务
type CircuitBreaker struct {
	serviceName     string
	maxFailures     int
	timeout         time.Duration
	halfOpenMaxReqs int
	
	state           CircuitBreakerState
	failures        int
	requests        int
	lastFailureTime time.Time
	
	mutex           sync.RWMutex
}

// CircuitBreakerConfig 熔断器配置
type CircuitBreakerConfig struct {
	ServiceName     string
	MaxFailures     int
	Timeout         time.Duration
	HalfOpenMaxReqs int
}

// NewCircuitBreaker 创建熔断器
func NewCircuitBreaker(config CircuitBreakerConfig) *CircuitBreaker {
	if config.MaxFailures <= 0 {
		config.MaxFailures = 5
	}
	if config.Timeout <= 0 {
		config.Timeout = 60 * time.Second
	}
	if config.HalfOpenMaxReqs <= 0 {
		config.HalfOpenMaxReqs = 3
	}
	
	return &CircuitBreaker{
		serviceName:     config.ServiceName,
		maxFailures:     config.MaxFailures,
		timeout:         config.Timeout,
		halfOpenMaxReqs: config.HalfOpenMaxReqs,
		state:           StateClosed,
	}
}

// CanExecute 检查是否可以执行请求
func (cb *CircuitBreaker) CanExecute() error {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()
	
	switch cb.state {
	case StateClosed:
		return nil
	case StateOpen:
		// 检查是否可以转为半开状态
		if time.Since(cb.lastFailureTime) > cb.timeout {
			cb.state = StateHalfOpen
			cb.requests = 0
			return nil
		}
		return domain.NewDomainError("CIRCUIT_BREAKER_OPEN", "Circuit breaker is open for service: "+cb.serviceName)
	case StateHalfOpen:
		// 半开状态下限制请求数量
		if cb.requests < cb.halfOpenMaxReqs {
			cb.requests++
			return nil
		}
		return domain.NewDomainError("CIRCUIT_BREAKER_HALF_OPEN_LIMIT", "Circuit breaker half-open request limit reached")
	default:
		return domain.NewDomainError("UNKNOWN_CIRCUIT_BREAKER_STATE", "Unknown circuit breaker state")
	}
}

// RecordSuccess 记录成功
func (cb *CircuitBreaker) RecordSuccess() {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()
	
	switch cb.state {
	case StateHalfOpen:
		if cb.requests >= cb.halfOpenMaxReqs {
			cb.state = StateClosed
			cb.failures = 0
			cb.requests = 0
		}
	case StateClosed:
		cb.failures = 0
	}
}

// RecordFailure 记录失败
func (cb *CircuitBreaker) RecordFailure() {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()
	
	cb.failures++
	cb.lastFailureTime = time.Now()
	
	switch cb.state {
	case StateClosed:
		if cb.failures >= cb.maxFailures {
			cb.state = StateOpen
		}
	case StateHalfOpen:
		cb.state = StateOpen
		cb.requests = 0
	}
}

// GetState 获取当前状态
func (cb *CircuitBreaker) GetState() CircuitBreakerState {
	cb.mutex.RLock()
	defer cb.mutex.RUnlock()
	
	return cb.state
}

// GetFailureCount 获取失败次数
func (cb *CircuitBreaker) GetFailureCount() int {
	cb.mutex.RLock()
	defer cb.mutex.RUnlock()
	
	return cb.failures
}

// GetServiceName 获取服务名称
func (cb *CircuitBreaker) GetServiceName() string {
	return cb.serviceName
}

// Reset 重置熔断器
func (cb *CircuitBreaker) Reset() {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()
	
	cb.state = StateClosed
	cb.failures = 0
	cb.requests = 0
	cb.lastFailureTime = time.Time{}
}

// GetStateName 获取状态名称
func (cb *CircuitBreaker) GetStateName() string {
	switch cb.GetState() {
	case StateClosed:
		return "CLOSED"
	case StateHalfOpen:
		return "HALF_OPEN"  
	case StateOpen:
		return "OPEN"
	default:
		return "UNKNOWN"
	}
}
