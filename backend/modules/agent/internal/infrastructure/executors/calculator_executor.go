package executors

import (
	"context"
	"fmt"
	"math"
	"strconv"
	
	"github.com/noah-loop/backend/modules/agent/internal/application/service"
	"github.com/noah-loop/backend/modules/agent/internal/domain"
)

// CalculatorExecutor 计算器工具执行器
type CalculatorExecutor struct{}

// NewCalculatorExecutor 创建计算器执行器
func NewCalculatorExecutor() service.ToolExecutor {
	return &CalculatorExecutor{}
}

// Execute 执行计算操作
func (e *CalculatorExecutor) Execute(ctx context.Context, request *service.ToolExecutionRequest) (*service.ToolExecutionResult, error) {
	// 解析输入参数
	operation, ok := request.Input["operation"].(string)
	if !ok {
		return nil, fmt.Errorf("operation parameter is required")
	}
	
	switch operation {
	case "add":
		return e.executeAdd(request.Input)
	case "subtract":
		return e.executeSubtract(request.Input)
	case "multiply":
		return e.executeMultiply(request.Input)
	case "divide":
		return e.executeDivide(request.Input)
	case "power":
		return e.executePower(request.Input)
	case "sqrt":
		return e.executeSqrt(request.Input)
	default:
		return nil, fmt.Errorf("unsupported operation: %s", operation)
	}
}

// GetSupportedType 获取支持的工具类型
func (e *CalculatorExecutor) GetSupportedType() domain.ToolType {
	return domain.ToolTypeCalculator
}

// executeAdd 执行加法
func (e *CalculatorExecutor) executeAdd(input map[string]interface{}) (*service.ToolExecutionResult, error) {
	a, err := e.getFloat(input, "a")
	if err != nil {
		return nil, err
	}
	
	b, err := e.getFloat(input, "b")
	if err != nil {
		return nil, err
	}
	
	result := a + b
	
	return &service.ToolExecutionResult{
		Output: map[string]interface{}{
			"result":    result,
			"operation": "add",
			"operands":  []float64{a, b},
		},
		ShouldLearn: false,
		Metadata: map[string]interface{}{
			"calculation_type": "basic_arithmetic",
		},
	}, nil
}

// executeSubtract 执行减法
func (e *CalculatorExecutor) executeSubtract(input map[string]interface{}) (*service.ToolExecutionResult, error) {
	a, err := e.getFloat(input, "a")
	if err != nil {
		return nil, err
	}
	
	b, err := e.getFloat(input, "b")
	if err != nil {
		return nil, err
	}
	
	result := a - b
	
	return &service.ToolExecutionResult{
		Output: map[string]interface{}{
			"result":    result,
			"operation": "subtract",
			"operands":  []float64{a, b},
		},
		ShouldLearn: false,
		Metadata: map[string]interface{}{
			"calculation_type": "basic_arithmetic",
		},
	}, nil
}

// executeMultiply 执行乘法
func (e *CalculatorExecutor) executeMultiply(input map[string]interface{}) (*service.ToolExecutionResult, error) {
	a, err := e.getFloat(input, "a")
	if err != nil {
		return nil, err
	}
	
	b, err := e.getFloat(input, "b")
	if err != nil {
		return nil, err
	}
	
	result := a * b
	
	return &service.ToolExecutionResult{
		Output: map[string]interface{}{
			"result":    result,
			"operation": "multiply",
			"operands":  []float64{a, b},
		},
		ShouldLearn: false,
		Metadata: map[string]interface{}{
			"calculation_type": "basic_arithmetic",
		},
	}, nil
}

// executeDivide 执行除法
func (e *CalculatorExecutor) executeDivide(input map[string]interface{}) (*service.ToolExecutionResult, error) {
	a, err := e.getFloat(input, "a")
	if err != nil {
		return nil, err
	}
	
	b, err := e.getFloat(input, "b")
	if err != nil {
		return nil, err
	}
	
	if b == 0 {
		return nil, fmt.Errorf("division by zero")
	}
	
	result := a / b
	
	return &service.ToolExecutionResult{
		Output: map[string]interface{}{
			"result":    result,
			"operation": "divide",
			"operands":  []float64{a, b},
		},
		ShouldLearn: false,
		Metadata: map[string]interface{}{
			"calculation_type": "basic_arithmetic",
		},
	}, nil
}

// executePower 执行幂运算
func (e *CalculatorExecutor) executePower(input map[string]interface{}) (*service.ToolExecutionResult, error) {
	base, err := e.getFloat(input, "base")
	if err != nil {
		return nil, err
	}
	
	exponent, err := e.getFloat(input, "exponent")
	if err != nil {
		return nil, err
	}
	
	result := math.Pow(base, exponent)
	
	return &service.ToolExecutionResult{
		Output: map[string]interface{}{
			"result":    result,
			"operation": "power",
			"base":      base,
			"exponent":  exponent,
		},
		ShouldLearn: false,
		Metadata: map[string]interface{}{
			"calculation_type": "advanced_math",
		},
	}, nil
}

// executeSqrt 执行平方根运算
func (e *CalculatorExecutor) executeSqrt(input map[string]interface{}) (*service.ToolExecutionResult, error) {
	value, err := e.getFloat(input, "value")
	if err != nil {
		return nil, err
	}
	
	if value < 0 {
		return nil, fmt.Errorf("cannot calculate square root of negative number")
	}
	
	result := math.Sqrt(value)
	
	return &service.ToolExecutionResult{
		Output: map[string]interface{}{
			"result":    result,
			"operation": "sqrt",
			"value":     value,
		},
		ShouldLearn: false,
		Metadata: map[string]interface{}{
			"calculation_type": "advanced_math",
		},
	}, nil
}

// getFloat 从输入中获取浮点数
func (e *CalculatorExecutor) getFloat(input map[string]interface{}, key string) (float64, error) {
	value, ok := input[key]
	if !ok {
		return 0, fmt.Errorf("%s parameter is required", key)
	}
	
	switch v := value.(type) {
	case float64:
		return v, nil
	case int:
		return float64(v), nil
	case string:
		parsed, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return 0, fmt.Errorf("invalid number format for %s: %s", key, v)
		}
		return parsed, nil
	default:
		return 0, fmt.Errorf("invalid type for %s, expected number", key)
	}
}
